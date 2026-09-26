package main

import (
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
	if r.UniqueIntervals != 3 || r.OverSlowThreshold != 1 || r.P95MS < 33 || r.IntervalCoverageSeconds < .06 {
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
