package presets

import "github.com/olivierh59500/democonstructionkit/catalog"

// Demo is renderer-independent production metadata.
type Demo = catalog.Demo

// Demos returns production metadata; headless tools can import catalog directly.
func Demos() []Demo                     { return catalog.Demos() }
func FindDemo(name string) (Demo, bool) { return catalog.FindDemo(name) }
