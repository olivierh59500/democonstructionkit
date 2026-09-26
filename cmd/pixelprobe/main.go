// Command pixelprobe measures presented-frame intervals of an Android demo.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type interval [2]int64

type report struct {
	Package                 string  `json:"package"`
	Serial                  string  `json:"serial,omitempty"`
	Layer                   string  `json:"layer"`
	RefreshHz               float64 `json:"refresh_hz"`
	Samples                 int     `json:"samples"`
	UniqueIntervals         int     `json:"unique_intervals"`
	IntervalCoverageSeconds float64 `json:"interval_coverage_seconds"`
	ElapsedSeconds          float64 `json:"elapsed_seconds"`
	MeanMS                  float64 `json:"mean_ms"`
	P50MS                   float64 `json:"p50_ms"`
	P95MS                   float64 `json:"p95_ms"`
	P99MS                   float64 `json:"p99_ms"`
	MaxMS                   float64 `json:"max_ms"`
	SlowThresholdMS         float64 `json:"slow_threshold_ms"`
	OverSlowThreshold       int     `json:"over_slow_threshold"`
}

func selectLayer(output, pkg string) (string, error) {
	if pkg == "" {
		return "", errors.New("package is required")
	}
	var layer string
	for _, line := range strings.Split(output, "\n") {
		start := strings.Index(line, "RequestedLayerState{")
		if start < 0 {
			continue
		}
		name := line[start+len("RequestedLayerState{"):]
		if end := strings.Index(name, " parentId="); end >= 0 {
			name = name[:end]
		}
		if strings.Contains(name, "SurfaceView["+pkg+"/") && strings.Contains(name, "(BLAST)") {
			layer = name
		}
	}
	if layer == "" {
		return "", fmt.Errorf("no active BLAST SurfaceView for %s", pkg)
	}
	return layer, nil
}

// parseLatency reads SurfaceFlinger's refresh period followed by desired,
// actual-present and frame-ready times in nanoseconds. Only the second column
// measures when a frame became visible to the user.
func parseLatency(output string) (int64, []interval, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return 0, nil, errors.New("missing SurfaceFlinger latency rows")
	}
	period, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
	if err != nil || period <= 0 {
		return 0, nil, errors.New("invalid SurfaceFlinger refresh period")
	}
	var previous int64
	intervals := make([]interval, 0, len(lines)-2)
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		actual, parseErr := strconv.ParseInt(fields[1], 10, 64)
		if parseErr != nil || actual <= 0 || actual == math.MaxInt64 {
			previous = 0
			continue
		}
		if difference := actual - previous; previous > 0 && difference > 0 && difference < 10*int64(time.Second) {
			intervals = append(intervals, interval{previous, actual})
		}
		previous = actual
	}
	if len(intervals) == 0 {
		return 0, nil, errors.New("no valid presented-frame intervals")
	}
	return period, intervals, nil
}

func summarize(pkg, serial, layer string, period int64, unique map[interval]struct{}, samples int, elapsed time.Duration, slowMS float64) (report, error) {
	if period <= 0 || len(unique) == 0 || slowMS <= 0 {
		return report{}, errors.New("no valid frame samples")
	}
	gaps := make([]float64, 0, len(unique))
	var total float64
	for pair := range unique {
		gap := float64(pair[1]-pair[0]) / 1e6
		gaps = append(gaps, gap)
		total += gap
	}
	sort.Float64s(gaps)
	percentile := func(q float64) float64 {
		return gaps[max(0, int(math.Ceil(q*float64(len(gaps))))-1)]
	}
	r := report{
		Package: pkg, Serial: serial, Layer: layer,
		RefreshHz: float64(time.Second) / float64(period),
		Samples:   samples, UniqueIntervals: len(gaps),
		IntervalCoverageSeconds: total / 1000, ElapsedSeconds: elapsed.Seconds(),
		MeanMS: total / float64(len(gaps)), P50MS: percentile(.5),
		P95MS: percentile(.95), P99MS: percentile(.99), MaxMS: gaps[len(gaps)-1],
		SlowThresholdMS: slowMS,
	}
	for _, gap := range gaps {
		if gap > slowMS {
			r.OverSlowThreshold++
		}
	}
	return r, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func adbCommand(ctx context.Context, executable, serial string, args ...string) (string, error) {
	if serial != "" {
		args = append([]string{"-s", serial}, args...)
	}
	request, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(request, executable, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("adb %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func measure(ctx context.Context, adb, serial, pkg, layer string, samples int, intervalTime time.Duration, slowMS float64) (report, error) {
	if pkg == "" || samples < 1 || samples > 10000 || intervalTime < 0 || slowMS <= 0 {
		return report{}, errors.New("invalid package, sample count, interval or slow threshold")
	}
	if layer == "" {
		listing, err := adbCommand(ctx, adb, serial, "shell", "dumpsys SurfaceFlinger --list")
		if err != nil {
			return report{}, err
		}
		layer, err = selectLayer(listing, pkg)
		if err != nil {
			return report{}, err
		}
	}
	started := time.Now()
	unique := make(map[interval]struct{}, samples*64)
	var period int64
	for sample := 0; sample < samples; sample++ {
		output, err := adbCommand(ctx, adb, serial, "shell", "dumpsys SurfaceFlinger --latency "+shellQuote(layer))
		if err != nil {
			return report{}, err
		}
		currentPeriod, intervals, err := parseLatency(output)
		if err != nil {
			return report{}, fmt.Errorf("sample %d: %w", sample+1, err)
		}
		if period != 0 && currentPeriod != period {
			return report{}, fmt.Errorf("display refresh period changed during measurement")
		}
		period = currentPeriod
		for _, pair := range intervals {
			unique[pair] = struct{}{}
		}
		if sample+1 < samples {
			select {
			case <-ctx.Done():
				return report{}, ctx.Err()
			case <-time.After(intervalTime):
			}
		}
	}
	return summarize(pkg, serial, layer, period, unique, samples, time.Since(started), slowMS)
}

func main() {
	pkg := flag.String("package", "", "Android application ID to measure")
	adb := flag.String("adb", "adb", "path to Android Debug Bridge")
	serial := flag.String("serial", "", "Android device serial; optional with one device")
	layer := flag.String("layer", "", "SurfaceFlinger layer; auto-detected from package when empty")
	samples := flag.Int("samples", 12, "number of recent-frame samples")
	intervalTime := flag.Duration("interval", 2*time.Second, "time between samples")
	slowMS := flag.Float64("slow-ms", 20, "presented-frame interval reported as slow")
	outputPath := flag.String("out", "", "optional JSON report path")
	flag.Parse()
	r, err := measure(context.Background(), *adb, *serial, *pkg, *layer, *samples, *intervalTime, *slowMS)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err == nil && *outputPath != "" {
		err = os.WriteFile(*outputPath, append(data, '\n'), 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
