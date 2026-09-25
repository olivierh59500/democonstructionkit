package timeline

import "testing"

func TestEventStagesCascadeWithinOneTick(t *testing.T) {
	program, err := NewEventStages(EventStagesConfig{
		Stages: []string{"first", "second", "third", "showcase", "final"}, Initial: "first",
		Transitions: []StageTransition{
			{"first", "next", "second"}, {"second", "next", "third"},
			{"third", "next", "showcase"}, {"showcase", "next", "final"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"second", "third", "showcase", "final"} {
		if !program.Trigger("next") || !program.Active(name) {
			t.Fatalf("stage after same-tick cue = %q, want %q", program.Stage(), name)
		}
	}
	if program.Trigger("next") {
		t.Fatal("final stage unexpectedly advanced")
	}
	program.Reset()
	if !program.Active("first") {
		t.Fatal("reset did not restore initial stage")
	}
	if allocations := testing.AllocsPerRun(100, func() { program.Trigger("missing") }); allocations != 0 {
		t.Fatalf("event trigger allocates %v times", allocations)
	}
}

func TestEventStagesRejectInvalidRules(t *testing.T) {
	for _, config := range []EventStagesConfig{
		{Stages: []string{"one", "one"}, Initial: "one"},
		{Stages: []string{"one"}, Initial: "other"},
		{Stages: []string{"one", "two"}, Initial: "one", Transitions: []StageTransition{{"one", "go", "missing"}}},
		{Stages: []string{"one", "two"}, Initial: "one", Transitions: []StageTransition{{"one", "go", "two"}, {"one", "go", "one"}}},
	} {
		if _, err := NewEventStages(config); err == nil {
			t.Fatalf("accepted invalid stage program: %+v", config)
		}
	}
}
