package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSelectLayerUsesActiveBLASTSurface(t *testing.T) {
	listing := strings.Join([]string{
		"RequestedLayerState{Background for aa SurfaceView[demo.pkg/demo.pkg.MainActivity]#5 parentId=2}",
		"RequestedLayerState{aa SurfaceView[demo.pkg/demo.pkg.MainActivity](BLAST)#4 parentId=3}",
		"RequestedLayerState{bb SurfaceView[other.pkg/other.pkg.MainActivity](BLAST)#8 parentId=7}",
	}, "\n")
	layer, err := selectLayer(listing, "demo.pkg")
	if err != nil || layer != "aa SurfaceView[demo.pkg/demo.pkg.MainActivity](BLAST)#4" {
		t.Fatalf("selected layer = %q, error = %v", layer, err)
	}
	if _, err := selectLayer(listing, "missing.pkg"); err == nil {
		t.Fatal("accepted a missing layer")
	}
	if err := changedLayer(listing, "demo.pkg", layer); err != nil {
		t.Fatalf("stable layer was reported as changed: %v", err)
	}
	if err := changedLayer(strings.ReplaceAll(listing, "(BLAST)#4", "(BLAST)#9"), "demo.pkg", layer); err == nil || !strings.Contains(err.Error(), "recreated") {
		t.Fatalf("surface recreation was not identified: %v", err)
	}
}

func TestFiniteSurfaceCanReturnExplicitPartialReport(t *testing.T) {
	directory := t.TempDir()
	state := filepath.Join(directory, "calls")
	t.Setenv("PIXELPROBE_TEST_STATE", state)
	adb := filepath.Join(directory, "adb")
	script := `#!/bin/sh
count=$(cat "$PIXELPROBE_TEST_STATE" 2>/dev/null || echo 0)
case "$2" in
  *--list*)
    if [ "$count" -lt 2 ]; then
      echo 'RequestedLayerState{aa SurfaceView[demo.pkg/demo.pkg.MainActivity](BLAST)#4 parentId=3}'
    fi ;;
  *--latency*)
    count=$((count + 1))
    echo "$count" > "$PIXELPROBE_TEST_STATE"
    if [ "$count" -eq 1 ]; then
      printf '16666667\n100 100000000 90\n110 133333334 95\n'
    else
      echo invalid
    fi ;;
esac
`
	if err := os.WriteFile(adb, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	r, err := measure(context.Background(), adb, "", "demo.pkg", "", 3, 0, 20, true)
	if err != nil || !r.EndedEarly || r.Samples != 1 || r.RequestedSamples != 3 ||
		r.StopReason == "" || r.OverSlowThreshold != 1 || len(r.SlowIntervals) != 1 ||
		r.SlowIntervals[0].AtUTC == "" {
		t.Fatalf("finite surface report = %+v, error = %v", r, err)
	}
	if err := os.Remove(state); err != nil {
		t.Fatal(err)
	}
	if _, err := measure(context.Background(), adb, "", "demo.pkg", "", 3, 0, 20, false); err == nil || !strings.Contains(err.Error(), "disappeared") {
		t.Fatalf("ordinary probe accepted a missing surface: %v", err)
	}
}

func TestParseLatencyAndSummarizeUniqueIntervals(t *testing.T) {
	first := "16666667\n0\t0\t0\n100\t100000000\t90\n110\t116666667\t95\n120\t133333334\t100\n"
	second := "16666667\n110\t116666667\t95\n120\t133333334\t100\n130\t166666668\t110\n"
	unique := map[interval]struct{}{}
	for _, input := range []string{first, second} {
		period, pairs, err := parseLatency(input)
		if err != nil || period != 16666667 {
			t.Fatalf("latency parse period = %d, error = %v", period, err)
		}
		for _, pair := range pairs {
			unique[pair] = struct{}{}
		}
	}
	r, err := summarize("demo.pkg", "serial", "layer", 16666667, unique, 2, 3*time.Second, 20)
	if err != nil {
		t.Fatal(err)
	}
	if r.UniqueIntervals != 3 || r.OverSlowThreshold != 1 || r.P95MS < 33 || r.IntervalCoverageSeconds < .06 ||
		len(r.SlowIntervals) != 1 || r.SlowIntervals[0].PresentNS != 166666668 || r.SlowIntervals[0].DurationMS < 33 {
		t.Fatalf("unexpected frame summary: %+v", r)
	}
	if _, _, err := parseLatency("invalid\n1 2 3\n"); err == nil {
		t.Fatal("accepted an invalid refresh period")
	}
}

func TestShellQuoteProtectsLayerName(t *testing.T) {
	if got, want := shellQuote("view's layer"), "'view'\\''s layer'"; got != want {
		t.Fatalf("shell quote = %q, want %q", got, want)
	}
}
