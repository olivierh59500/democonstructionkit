package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCollectionProbeBoundaries(t *testing.T) {
	for _, tc := range []struct{ demo, screen, scenario string }{
		{"go-cuddlymenu", "../menu", "idle"},
		{"go-uniondemo", "loader:../../menu", "idle"},
		{"go-uniondemo", "loader:intro", "idle"},
		{"bilizir-demo", "intro", "idle"},
		{"go-uniondemo", "menu", "controls"},
		{"go-cuddlymenu", "intro", "controls"},
		{"go-uniondemo", "delta", "controls"},
		{"go-uniondemo", "", "idle"},
		{"go-cuddlymenu", "", "controls"},
		{"go-cuddlymenu", "dna", "random"},
	} {
		if _, _, err := selectProbe(tc.demo, tc.screen, tc.scenario); err == nil {
			t.Errorf("accepted invalid fixture %+v", tc)
		}
	}
}

func TestCollectionFixturesParse(t *testing.T) {
	for _, tc := range []struct{ demo, screen, scenario, target string }{
		{"go-cuddlymenu", "intro", "idle", "dck/screens"},
		{"go-cuddlymenu", "spreadpoint", "idle", "dck/screens"},
		{"go-cuddlymenu", "megaball", "controls", "dck/screens"},
		{"go-cuddlymenu", "loader:DNA_DEMO", "idle", "dck/loader"},
		{"go-uniondemo", "intro", "idle", "internal/screens"},
		{"go-uniondemo", "delta", "idle", "internal/screens"},
		{"go-uniondemo", "hidden", "controls", "internal/screens"},
		{"go-uniondemo", "tnt3", "controls", "internal/screens"},
		{"go-uniondemo", "starballs", "controls", "internal/screens"},
		{"go-uniondemo", "tnt2", "controls", "internal/screens"},
		{"go-uniondemo", "replicants", "controls", "internal/screens"},
		{"go-uniondemo", "diskcopier", "controls", "internal/screens"},
		{"go-uniondemo", "menu", "idle", "internal/menu"},
		{"go-uniondemo", "loader:replicants", "idle", "internal/loader"},
	} {
		t.Run(tc.demo+"/"+tc.screen+"/"+tc.scenario, func(t *testing.T) {
			p, options, err := selectProbe(tc.demo, tc.screen, tc.scenario)
			if err != nil {
				t.Fatal(err)
			}
			if options.Target != tc.target || p.Width <= 0 || p.Height <= 0 {
				t.Fatalf("invalid target or surface: %+v %+v", p, options)
			}
			code := fmt.Sprintf("package fixture\nimport(\"github.com/hajimehoshi/ebiten/v2\";%s)\nfunc makeGame()(ebiten.Game,error){%s}\n%s", p.Imports, p.Factory, options.Code)
			if _, err := parser.ParseFile(token.NewFileSet(), "fixture.go", code, parser.AllErrors); err != nil {
				t.Fatal(err)
			}
			if tc.screen == "delta" && (!strings.Contains(code, "sound.Open") || !strings.Contains(code, "io.ReadFull") || !strings.Contains(code, "YMRegisters")) {
				t.Fatal("music-driven screen would be captured with frozen voice levels")
			}
		})
	}
}
