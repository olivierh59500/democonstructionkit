//go:build dck_gpu_rendercheck

package effects

import (
	"os"
	"testing"

	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func TestMain(m *testing.M) { os.Exit(testutil.RunGPU(m)) }
