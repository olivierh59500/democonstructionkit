package scrolling

import (
	"math"
	"testing"
)

func TestFeedbackDNAValidatesDimensionsAndOwnsProfile(t *testing.T) {
	profile := []int{0, 1, 4, 3}
	c := FeedbackDNAConfig{Width: 19, Height: 8, HorizontalSpeed: 3, VerticalSpeed: 1, ColumnWidth: 2, Direction: -1, InsertY: 5, Profile: profile}
	f, err := NewFeedbackDNA(c)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	profile[0] = 7
	if f.config.Profile[0] != 0 {
		t.Fatal("profile aliases caller memory")
	}
	for _, change := range []func(*FeedbackDNAConfig){func(c *FeedbackDNAConfig) { c.Direction = 0 }, func(c *FeedbackDNAConfig) { c.HorizontalSpeed = 20 }, func(c *FeedbackDNAConfig) { c.InsertY = math.NaN() }, func(c *FeedbackDNAConfig) { c.ColumnWidth = 0 }} {
		bad := c
		change(&bad)
		if effect, err := NewFeedbackDNA(bad); err == nil {
			effect.Close()
			t.Fatal("invalid configuration accepted")
		}
	}
}
