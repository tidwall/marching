package marching

import (
	"math/rand"
	"testing"
	"time"
)

type mockPoint struct {
	i    int
	x, y float64
}

func (p *mockPoint) Pos(interface{}) (x, y float64) {
	return p.x, p.y
}

func TestQTree(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var points []*mockPoint
	for i := 0; i < 100000; i++ {
		points = append(points, &mockPoint{i, rand.Float64(), rand.Float64()})
	}
	tr := newQTree(0, 0, 1, 1, nil)
	var start time.Time
	var dur time.Duration
	var scount int

	//
	start = time.Now()
	for i := 0; i < len(points); i++ {
		tr.insert(points[i])
	}
	dur = time.Now().Sub(start)
	//	fmt.Printf("inserted %v items in %v\n", len(points), dur)

	//
	start = time.Now()
	if tr.count() != len(points) {
		t.Fatalf("expected %v, got %v", len(points), tr.count())
	}
	dur = time.Now().Sub(start)
	//	fmt.Printf("counted %v items in %v\n", len(points), dur)

	// seach middle
	start = time.Now()
	scount = 0
	tr.search(.25, .25, .75, .75, func(item qtreeItem, ctx interface{}) bool {
		x, y := item.Pos(ctx)
		if x < .25 || x > .75 || y < .25 || y > .75 {
			t.Fatal("out of bounds item in search")
		}
		scount++
		return true
	})
	dur = time.Now().Sub(start)
	if (scount < len(points)/4-int(float64(len(points)/4)*.05)) ||
		(scount > len(points)/4+int(float64(len(points)/4)*.05)) {
		t.Fatalf("expected %v, got %v", len(points)/2, scount)
	}
	//	fmt.Printf("searched %v items in %v\n", scount, dur)

	// remove 50,000
	start = time.Now()
	randidx := rand.Perm(len(points))
	for _, i := range randidx[:len(randidx)/2] {
		if tr.remove(points[i]) != points[i] {
			t.Fatalf("item mismatch")
		}
	}
	dur = time.Now().Sub(start)
	//	fmt.Printf("removed %v items in %v\n", len(points)/2, dur)

	//
	start = time.Now()
	scount = 0
	tr.scan(func(item qtreeItem, ctx interface{}) bool {
		scount++
		return true
	})
	dur = time.Now().Sub(start)
	if scount != len(points)/2 {
		t.Fatalf("expected %v, got %v", len(points)/2, scount)
	}
	if scount != tr.count() {
		t.Fatalf("expected %v, got %v", tr.count(), scount)
	}
	//	fmt.Printf("scanned %v items in %v\n", scount, dur)

	//
	start = time.Now()
	for _, i := range randidx[len(randidx)/2:] {
		if tr.remove(points[i]) != points[i] {
			t.Fatalf("item mismatch")
		}
	}
	dur = time.Now().Sub(start)
	if tr.count() != 0 {
		t.Fatalf("expected %v, got %v", 0, tr.count())
	}
	//	fmt.Printf("removed %v items in %v\n", len(points)/2, dur)

}
