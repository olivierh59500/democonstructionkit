package motion

import (
	"reflect"
	"testing"
)

func TestRecycledQueueNativeBytePhasesAndStrictCrossing(t *testing.T) {
	items := [][4]byte{{0, 1, 2, 0}, {240, 3, 99, 1}, {20, 0, 130, 2}, {77, 2, 200, 3}}
	original := append([][4]byte(nil), items...)
	type pose struct{ depth, phase, x, material int }
	q, err := NewRecycledQueue(RecycledQueueConfig[[4]byte, pose]{Items: items, Count: 1, Step: 28, Spacing: 100, DepthWrap: 400, Grow: true, Recycle: func(tick int, p *[4]byte) { p[0] = 0; p[2] = byte(tick*73 + 19) }, Project: func(_ int, d int, p *[4]byte) (pose, error) {
		old := p[0]
		p[0] += byte(p[1]&3) + 5
		return pose{d, int(old), int(int8(p[2])), int(p[3])}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	items[0][0] = 33
	count, depth := 1, 0
	for tick := 365; tick < 2365; tick++ {
		depth += 28
		if depth > 100 {
			depth -= 100
			count = min(4, count+1)
			last := original[3]
			copy(original[1:], original[:3])
			original[0] = last
			original[0][0] = 0
			original[0][2] = byte(tick*73 + 19)
		}
		d := depth
		want := make([]pose, count)
		for i := 0; i < count; i++ {
			if d > 400 {
				d -= 400
			}
			old := original[i][0]
			original[i][0] += byte(original[i][1]&3) + 5
			want[i] = pose{d, int(old), int(int8(original[i][2])), int(original[i][3])}
			d += 100
		}
		if err := q.Step(tick); err != nil {
			t.Fatal(err)
		}
		if q.Count() != count || q.Depth() != depth || !reflect.DeepEqual(q.Items(), original) || !reflect.DeepEqual(q.Poses(), want) {
			t.Fatal("recycled particle timing", tick, q.Count(), q.Depth(), q.Poses(), want)
		}
	}
	if n := testing.AllocsPerRun(100, func() {
		if err := q.Step(3000); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatal("queue updates allocate", n)
	}
}

func TestRecycledQueueValidation(t *testing.T) {
	for _, c := range []RecycledQueueConfig[int, int]{{}, {Items: []int{1}, Count: 2, Spacing: 1, DepthWrap: 1, Project: func(int, int, *int) (int, error) { return 0, nil }}, {Items: []int{1}, Step: 3, Spacing: 1, DepthWrap: 1, Project: func(int, int, *int) (int, error) { return 0, nil }}} {
		if _, err := NewRecycledQueue(c); err == nil {
			t.Fatal("invalid queue", c)
		}
	}
}
