package presets

type Demo struct {
	Name, Description, Music string
	Module                   bool
}

// Demos includes every repository inspected in docs/INVENTORY.md.
func Demos() []Demo {
	return []Demo{
		{"3d_doc", "Checkerboard, balls and scanline text", "3d_doc/assets/music.ym", false},
		{"bilizir-demo", "Copper, moving logo, twelve cubes and distorted text", "bilizir-demo/assets/music.ym", false},
		{"dma-3d", "Glenz vectors, stars and TCB text", "dma-3d/assets/music.ym", false},
		{"dma-is-back", "Jelly cube, tiled logos, strip scroller and CRT", "dma-is-back/assets/Mindbomb.ym", false},
		{"go-cocoisthebest", "Rotozoom, copper, cubes and MegaTwist text", "go-cocoisthebest/assets/mindbomb.ym", false},
		{"go-cuddlymenu", "Tile world, sprite sheet, camera and proportional chrome text", "go-cuddlymenu/assets/menu/menu.ym", false},
		{"go-dom-intro", "Four text scales, raster colors and star sprites", "go-dom-intro/assets/eliminator.ym", false},
		{"go-fr010", "Camera keyframes, vector text and a textured object", "go-fr010/data/Stormlord 1 - title.ym", false},
		{"go-multiscreen", "Four reusable recipes composed into viewports", "go-multiscreen/assets/music.ym", false},
		{"go-secondreality", "Sequenced Glenz, tunnel, plasma, balls and sine-field studies", "go-secondreality/internal/music/reality.fc", true},
		{"go-vectorballs", "Morphing vectorballs and reflection", "go-vectorballs/assets/Mindbomb.ym", false},
		{"grodan-kvack-kvack-demo", "Three bitmap fonts, sprite train and tiled backgrounds", "grodan-kvack-kvack-demo/assets/music.ym", false},
		{"megatwist", "Proportional scanline scroller and moving background", "megatwist/assets/music.ym", false},
		{"nonameno-demo", "Tweened text, perspective stars and small sine text", "nonameno-demo/assets/music.ym", false},
		{"phenomena-dna-scroll-intro", "Twisting front/back text strips and raster colors", "phenomena-dna-scroll-intro/assets/music.ym", false},
		{"tcb-multi-plane-3d-scroller", "Three rotating text planes above a landscape", "tcb-multi-plane-3d-scroller/assets/Thundercats.ym", false},
		{"tcb-replicants-demo", "Stars, rotating logos and deformed text", "tcb-replicants-demo/assets/Rollout.ym", false},
		{"teamg1-demo", "Plasma, textured cube, logo spiral and proportional text", "teamg1-demo/assets/music.ym", false},
		{"viva_tcb", "Repeating rotozoom, two sine scrollers and logos", "viva_tcb/assets/music.ym", false},
	}
}
func FindDemo(name string) (Demo, bool) {
	for _, r := range Demos() {
		if r.Name == name {
			return r, true
		}
	}
	return Demo{}, false
}
