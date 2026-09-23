package presets

import "testing"

func TestNonamenoStarsRecipeAcceptsEditableParameters(t *testing.T) {
	c := DefaultNonamenoStarsConfig()
	c.Count = 3
	c.Speed = 4
	c.Focal = 150
	c.TrailWidth = 2
	field, err := NonamenoProjectedStars(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(field.Field.Points) != 3 || field.Velocity.Z != -4 || field.View.Camera.Focal != 150 {
		t.Fatalf("editable radial field was not compiled: %+v", field)
	}
	if reset := field.Field.Spawn(1, true); reset.Z != c.DepthMax {
		t.Fatalf("reset depth = %g, want %g", reset.Z, c.DepthMax)
	}
	c.Count = -1
	if _, err := NonamenoProjectedStars(c); err == nil {
		t.Fatal("accepted a negative star count")
	}
}
