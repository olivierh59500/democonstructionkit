package timeline

import "fmt"

// StageTransition changes the active stage as soon as Event is emitted from
// From. Later checks in the same simulation tick observe To immediately.
type StageTransition struct{ From, Event, To string }

// EventStagesConfig keeps stage names and transition rules as editor-friendly
// data. The production supplies named control, music or text-cursor events.
type EventStagesConfig struct {
	Stages      []string
	Initial     string
	Transitions []StageTransition
}

type stageEvent struct{ stage, event string }

// EventStages owns a stage cursor. Trigger does not advance a hidden clock;
// callers choose exact event and draw/update boundaries.
type EventStages struct {
	initial, current string
	next             map[stageEvent]string
}

func NewEventStages(config EventStagesConfig) (*EventStages, error) {
	if len(config.Stages) == 0 || len(config.Stages) > 1024 || len(config.Transitions) > 1<<16 {
		return nil, fmt.Errorf("timeline: invalid event stage dimensions")
	}
	known := make(map[string]bool, len(config.Stages))
	for _, name := range config.Stages {
		if name == "" || known[name] {
			return nil, fmt.Errorf("timeline: empty or duplicate stage name")
		}
		known[name] = true
	}
	if !known[config.Initial] {
		return nil, fmt.Errorf("timeline: unknown initial stage")
	}
	program := &EventStages{initial: config.Initial, current: config.Initial, next: make(map[stageEvent]string, len(config.Transitions))}
	for _, transition := range config.Transitions {
		key := stageEvent{transition.From, transition.Event}
		if transition.Event == "" || !known[transition.From] || !known[transition.To] || program.next[key] != "" {
			return nil, fmt.Errorf("timeline: invalid or duplicate stage transition")
		}
		program.next[key] = transition.To
	}
	return program, nil
}

func (program *EventStages) Stage() string           { return program.current }
func (program *EventStages) Active(name string) bool { return program.current == name }

// Trigger returns true when the current stage accepts event. It changes the
// active stage immediately, allowing authored same-tick cascades.
func (program *EventStages) Trigger(event string) bool {
	next := program.next[stageEvent{program.current, event}]
	if next == "" {
		return false
	}
	program.current = next
	return true
}

func (program *EventStages) Reset() { program.current = program.initial }
