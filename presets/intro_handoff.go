package presets

import "github.com/olivierh59500/democonstructionkit/timeline"

// FadedIntroHandoff returns a data-only intro/main transition. The soundtrack
// cue becomes ready strictly above threshold; the caller chooses its player.
func FadedIntroHandoff(step, threshold float64) timeline.IntroHandoffConfig {
	return timeline.IntroHandoffConfig{
		FadeStart: 0, FadeStep: step, FadeMax: 1,
		Cue: timeline.IntroCueAboveFade, CueThreshold: threshold,
	}
}

// ImmediateIntroHandoff enters an opaque main scene and releases its one-shot
// soundtrack cue on the same update as the intro completion.
func ImmediateIntroHandoff() timeline.IntroHandoffConfig {
	return timeline.IntroHandoffConfig{FadeStart: 1, FadeMax: 1, Cue: timeline.IntroCueOnEntry}
}
