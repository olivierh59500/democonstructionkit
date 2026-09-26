// Package testutil runs pixel-readback tests while Ebitengine owns a live
// graphics context. Production packages do not import this package.
package testutil

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

var gpuSuiteActive atomic.Bool

// RequireGPU skips a visual assertion unless RunGPU started the test suite.
// Pure controller tests keep running in ordinary go test invocations.
func RequireGPU(t *testing.T) {
	t.Helper()
	if !gpuSuiteActive.Load() {
		t.Skip("pixel readback requires the dck_gpu_rendercheck tag")
	}
}

type suiteGame struct {
	tests   *testing.M
	result  chan int
	started bool
	code    int
}

func (suite *suiteGame) Layout(int, int) (int, int) { return 16, 16 }
func (suite *suiteGame) Draw(*ebiten.Image)         {}
func (suite *suiteGame) Update() error {
	if !suite.started {
		suite.started = true
		go func() { suite.result <- suite.tests.Run() }()
	}
	select {
	case suite.code = <-suite.result:
		return ebiten.Termination
	default:
		return nil
	}
}

// RunGPU must be called by a package TestMain on the process's main thread.
// The regular testing goroutine begins only after Ebitengine has started.
func RunGPU(m *testing.M) int {
	gpuSuiteActive.Store(true)
	ebiten.SetWindowSize(16, 16)
	ebiten.SetWindowTitle("DCK GPU tests")
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(false)
	suite := &suiteGame{tests: m, result: make(chan int, 1)}
	if err := ebiten.RunGame(suite); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return suite.code
}
