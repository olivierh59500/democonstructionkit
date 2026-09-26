package timeline

import (
	"fmt"
	"math"
)

// ScalarCompare describes an exact threshold test after a stage advances.
type ScalarCompare uint8

const (
	ScalarAlways ScalarCompare = iota
	ScalarGreater
	ScalarGreaterEqual
	ScalarLess
	ScalarLessEqual
)

// ScalarRule applies optional value, direction and stage changes in declaration
// order. Event is an externally signaled cue; DirectionSign is -1, 0 or +1,
// with zero accepting either direction. SetValue and SetDirection distinguish
// explicit zeroes from omitted changes.
type ScalarRule struct {
	Compare       ScalarCompare
	Threshold     float64
	Event         string
	DirectionSign int
	SetValue      bool
	Value         float64
	SetDirection  bool
	Direction     float64
	ChangeStage   bool
	NextStage     int
}

type ScalarStage struct {
	Name         string
	Delta        float64
	UseDirection bool
	Rules        []ScalarRule
}

type ScalarStagesConfig struct {
	Stages                         []ScalarStage
	InitialStage                   int
	InitialValue, InitialDirection float64
}

type ScalarStageState struct {
	Stage            int
	Value, Direction float64
	Tick             int
}

// ScalarStages runs one authored stage per Step. All rules of that stage are
// checked after its delta, even if an earlier rule changes the next stage;
// the next stage's delta waits until the next Step. Signals sent after a Step
// remain pending for the following one, preserving scene/input boundaries.
type ScalarStages struct {
	config  ScalarStagesConfig
	state   ScalarStageState
	events  map[string]int
	pending []bool
}

func NewScalarStages(c ScalarStagesConfig) (*ScalarStages, error) {
	if len(c.Stages) == 0 || len(c.Stages) > 1024 || c.InitialStage < 0 || c.InitialStage >= len(c.Stages) ||
		!finiteScalar(c.InitialValue) || !finiteScalar(c.InitialDirection) {
		return nil, fmt.Errorf("timeline: invalid scalar stage configuration")
	}
	events := make(map[string]int)
	names := make(map[string]bool, len(c.Stages))
	stages := make([]ScalarStage, len(c.Stages))
	for i, stage := range c.Stages {
		if stage.Name == "" || names[stage.Name] || !finiteScalar(stage.Delta) || len(stage.Rules) > 1024 {
			return nil, fmt.Errorf("timeline: invalid scalar stage %d", i)
		}
		names[stage.Name] = true
		stages[i] = stage
		stages[i].Rules = append([]ScalarRule(nil), stage.Rules...)
		for _, rule := range stage.Rules {
			if rule.Compare > ScalarLessEqual || !finiteScalar(rule.Threshold) ||
				!finiteScalar(rule.Value) || !finiteScalar(rule.Direction) ||
				rule.DirectionSign < -1 || rule.DirectionSign > 1 ||
				rule.ChangeStage && (rule.NextStage < 0 || rule.NextStage >= len(c.Stages)) {
				return nil, fmt.Errorf("timeline: invalid scalar stage rule")
			}
			if rule.Event != "" {
				if _, ok := events[rule.Event]; !ok {
					events[rule.Event] = len(events)
				}
			}
		}
	}
	c.Stages = stages
	p := &ScalarStages{config: c, events: events, pending: make([]bool, len(events))}
	p.Reset()
	return p, nil
}

func finiteScalar(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func (p *ScalarStages) Reset() {
	p.state = ScalarStageState{Stage: p.config.InitialStage,
		Value: p.config.InitialValue, Direction: p.config.InitialDirection}
	clear(p.pending)
}

// Signal queues one known event until the next Step. An unknown event is inert.
func (p *ScalarStages) Signal(event string) bool {
	index, ok := p.events[event]
	if ok {
		p.pending[index] = true
	}
	return ok
}

func (p *ScalarStages) Step() {
	stage := p.config.Stages[p.state.Stage]
	delta := stage.Delta
	if stage.UseDirection {
		delta *= p.state.Direction
	}
	p.state.Value += delta
	for _, rule := range stage.Rules {
		if !p.matches(rule) {
			continue
		}
		if rule.SetValue {
			p.state.Value = rule.Value
		}
		if rule.SetDirection {
			p.state.Direction = rule.Direction
		}
		if rule.ChangeStage {
			p.state.Stage = rule.NextStage
		}
	}
	clear(p.pending)
	p.state.Tick++
}

func (p *ScalarStages) matches(rule ScalarRule) bool {
	if rule.Event != "" && !p.pending[p.events[rule.Event]] {
		return false
	}
	if rule.DirectionSign > 0 && p.state.Direction <= 0 ||
		rule.DirectionSign < 0 && p.state.Direction >= 0 {
		return false
	}
	switch rule.Compare {
	case ScalarGreater:
		return p.state.Value > rule.Threshold
	case ScalarGreaterEqual:
		return p.state.Value >= rule.Threshold
	case ScalarLess:
		return p.state.Value < rule.Threshold
	case ScalarLessEqual:
		return p.state.Value <= rule.Threshold
	default:
		return true
	}
}

func (p *ScalarStages) State() ScalarStageState { return p.state }
func (p *ScalarStages) StageName() string       { return p.config.Stages[p.state.Stage].Name }
