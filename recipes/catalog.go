// Package recipes demonstrates rebuilding the source demos from shared effects.
// These are effect studies, not pixel-identical ports of the original productions.
package recipes

import "github.com/olivierh59500/democonstructionkit/presets"

type Recipe = presets.Demo

func Catalog() []Recipe               { return presets.Demos() }
func Find(name string) (Recipe, bool) { return presets.FindDemo(name) }
