//go:build goexperiment.arenas

package btree

import (
	"arena"
	"flag"
	"math/rand"
	"runtime"
	"testing"
)

var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

type testInt int

func newTestInt(i int) *testInt {
	ti := testInt(i)
	return &ti
}

func (n *testInt) Less(n2 *testInt) bool {
	return *n < *n2
}

func (n *testInt) DeepCopy() *testInt {
	var np testInt
	np = testInt(*n)
	return &np
}

func (n *testInt) DeepCopyWithArena(a *arena.Arena) *testInt {
	np := arena.New[testInt](a)
	*np = *n
	return np
}

func intRange(s int, reverse bool) []*testInt {
	out := make([]*testInt, s)
	for i := 0; i < s; i++ {
		v := i
		if reverse {
			v = s - i - 1
		}
		ti := testInt(v)
		out[i] = &ti
	}
	return out
}

func testIntAll(t *BTree[*testInt]) (out []*testInt) {
	t.Ascend(func(a *testInt) bool {
		out = append(out, a)
		return true
	})
	return
}

func testIntAllRev(t *BTree[*testInt]) (out []*testInt) {
	t.Descend(func(a *testInt) bool {
		out = append(out, a)
		return true
	})
	return
}

func BenchmarkBothDeepCopy(b *testing.B) {
	items := rand.Perm(16392)

	tr := New[*testInt](*btreeDegree)

	for _, v := range items {
		tr.ReplaceOrInsert(newTestInt(v))
	}

	b.ResetTimer()

	b.Run(`DeepCoopy`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tr2 := tr.DeepCopy()
			tr2.Len()
			tr2 = nil
		}
		runtime.GC()
	})

	b.Run(`DeepCopyWithArena`, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			a := arena.NewArena()
			tr2 := tr.DeepCopyWithArena(a)
			tr2.Len()
			a.Free()
		}
		runtime.GC()
	})
}

func BenchmarkDeepCopy(b *testing.B) {
	items := rand.Perm(16392)

	tr := New[*testInt](*btreeDegree)

	for _, v := range items {
		tr.ReplaceOrInsert(newTestInt(v))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tr2 := tr.DeepCopy()
		tr2.Len()
		tr2 = nil
	}

	runtime.GC()
}

func BenchmarkDeepCopyWithArena(b *testing.B) {
	items := rand.Perm(16392)

	tr := New[*testInt](*btreeDegree)

	for _, v := range items {
		tr.ReplaceOrInsert(newTestInt(v))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a := arena.NewArena()
		tr2 := tr.DeepCopyWithArena(a)
		tr2.Len()
		a.Free()
	}

	runtime.GC()
}
