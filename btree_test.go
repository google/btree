// Copyright 2014-2022 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.18
// +build go1.18

package btree

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
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

func TestBTreeG(t *testing.T) {
	tr := New[*testInt](*btreeDegree)
	const treeSize = 100
	for i := 0; i < 10; i++ {
		if min, ok := tr.Min(); ok || min != nil {
			t.Fatalf("empty min, got %+v", min)
		}
		if max, ok := tr.Max(); ok || max != nil {
			t.Fatalf("empty max, got %+v", max)
		}
		for _, item := range rand.Perm(treeSize) {
			i := testInt(item)
			if x, ok := tr.ReplaceOrInsert(&i); ok || x != nil {
				t.Fatal("insert found item", item)
			}
		}
		for _, item := range rand.Perm(treeSize) {
			i := testInt(item)
			if x, ok := tr.ReplaceOrInsert(&i); !ok || *x != i {
				t.Fatal("insert didn't find item", item)
			}
		}
		want := 0
		if min, ok := tr.Min(); !ok || int(*min) != want {
			t.Fatalf("min: ok %v want %+v, got %+v", ok, want, min)
		}
		want = treeSize - 1
		if max, ok := tr.Max(); !ok || int(*max) != want {
			t.Fatalf("max: ok %v want %+v, got %+v", ok, want, max)
		}
		got := testIntAll(tr)
		wantRange := intRange(treeSize, false)
		if !reflect.DeepEqual(got, wantRange) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, wantRange)
		}

		gotrev := testIntAllRev(tr)
		wantrev := intRange(treeSize, true)
		if !reflect.DeepEqual(gotrev, wantrev) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", gotrev, wantrev)
		}

		for _, item := range rand.Perm(treeSize) {
			i := testInt(item)
			if x, ok := tr.Delete(&i); !ok || int(*x) != item {
				t.Fatalf("didn't find %v", item)
			}
		}
		if got = testIntAll(tr); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
		if got = testIntAllRev(tr); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
	}
}

func ExampleBTreeG() {
	tr := New[*testInt](*btreeDegree)
	for i := 0; i < 10; i++ {
		ti := testInt(i)
		tr.ReplaceOrInsert(&ti)
	}
	fmt.Println("len:       ", tr.Len())
	v, ok := tr.Get(newTestInt(3))
	fmt.Println("get3:      ", *v, ok)
	v, ok = tr.Get(newTestInt(100))
	fmt.Println("get100:    ", v, ok)
	v, ok = tr.Delete(newTestInt(4))
	fmt.Println("del4:      ", *v, ok)
	v, ok = tr.Delete(newTestInt(100))
	fmt.Println("del100:    ", nil, ok)
	v, ok = tr.ReplaceOrInsert(newTestInt(5))
	fmt.Println("replace5:  ", *v, ok)
	v, ok = tr.ReplaceOrInsert(newTestInt(100))
	fmt.Println("replace100:", nil, ok)
	v, ok = tr.Min()
	fmt.Println("min:       ", *v, ok)
	v, ok = tr.DeleteMin()
	fmt.Println("delmin:    ", *v, ok)
	v, ok = tr.Max()
	fmt.Println("max:       ", *v, ok)
	v, ok = tr.DeleteMax()
	fmt.Println("delmax:    ", *v, ok)
	fmt.Println("len:       ", tr.Len())
	// Output:
	// len:        10
	// get3:       3 true
	// get100:     <nil> false
	// del4:       4 true
	// del100:     <nil> false
	// replace5:   5 true
	// replace100: <nil> false
	// min:        0 true
	// delmin:     0 true
	// max:        100 true
	// delmax:     100 true
	// len:        8
}

func TestDeleteMinG(t *testing.T) {
	tr := New[*testInt](3)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	for v, ok := tr.DeleteMin(); ok; v, ok = tr.DeleteMin() {
		got = append(got, v)
	}
	if want := intRange(100, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDeleteMaxG(t *testing.T) {
	tr := New[*testInt](3)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	for v, ok := tr.DeleteMax(); ok; v, ok = tr.DeleteMax() {
		got = append(got, v)
	}
	if want := intRange(100, true); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendRangeG(t *testing.T) {
	tr := New[*testInt](2)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.AscendRange(newTestInt(40), newTestInt(60), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[40:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendRange(newTestInt(40), newTestInt(60), func(a *testInt) bool {
		if *a > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendRangeG(t *testing.T) {
	tr := New[*testInt](30)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.DescendRange(newTestInt(60), newTestInt(40), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[39:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendRange(newTestInt(60), newTestInt(40), func(a *testInt) bool {
		if *a < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendLessThanG(t *testing.T) {
	tr := New[*testInt](*btreeDegree)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.AscendLessThan(newTestInt(60), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendLessThan(newTestInt(60), func(a *testInt) bool {
		if *a > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendLessOrEqualG(t *testing.T) {
	tr := New[*testInt](*btreeDegree)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.DescendLessOrEqual(newTestInt(40), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[59:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendLessOrEqual(newTestInt(60), func(a *testInt) bool {
		if *a < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendGreaterOrEqualG(t *testing.T) {
	tr := New[*testInt](*btreeDegree)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.AscendGreaterOrEqual(newTestInt(40), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[40:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendGreaterOrEqual(newTestInt(40), func(a *testInt) bool {
		if *a > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, false)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendGreaterThanG(t *testing.T) {
	tr := New[*testInt](*btreeDegree)
	for _, v := range rand.Perm(100) {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	var got []*testInt
	tr.DescendGreaterThan(newTestInt(40), func(a *testInt) bool {
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendGreaterThan(newTestInt(40), func(a *testInt) bool {
		if *a < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := intRange(100, true)[:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
}

const benchmarkTreeSize = 10000

func BenchmarkInsertG(b *testing.B) {
	b.StopTimer()
	insertP := rand.Perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New[*testInt](*btreeDegree)
		for _, item := range insertP {
			tr.ReplaceOrInsert(newTestInt(item))
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkSeekG(b *testing.B) {
	b.StopTimer()
	size := 100000
	insertP := rand.Perm(size)
	tr := New[*testInt](*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(newTestInt(item))
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		tr.AscendGreaterOrEqual(newTestInt(i%size), func(i *testInt) bool { return false })
	}
}

func BenchmarkDeleteInsertG(b *testing.B) {
	b.StopTimer()
	insertP := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(newTestInt(item))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(newTestInt(insertP[i%benchmarkTreeSize]))
		tr.ReplaceOrInsert(newTestInt(insertP[i%benchmarkTreeSize]))
	}
}

func BenchmarkDeleteG(b *testing.B) {
	b.StopTimer()
	insertP := rand.Perm(benchmarkTreeSize)
	removeP := rand.Perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New[*testInt](*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Delete(newTestInt(item))
			i++
			if i >= b.N {
				return
			}
		}
		if tr.Len() > 0 {
			panic(tr.Len())
		}
	}
}

func BenchmarkGetG(b *testing.B) {
	b.StopTimer()
	insertP := rand.Perm(benchmarkTreeSize)
	removeP := rand.Perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New[*testInt](*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Get(newTestInt(item))
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkAscendG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 0
		tr.Ascend(func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j++
			return true
		})
	}
}

func BenchmarkDescendG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 1
		tr.Descend(func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j--
			return true
		})
	}
}

func BenchmarkAscendRangeG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		tr.AscendRange(newTestInt(100), newTestInt(arr[len(arr)-100]), func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j++
			return true
		})
		if j != len(arr)-100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}

func BenchmarkDescendRangeG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		tr.DescendRange(newTestInt(arr[len(arr)-100]), newTestInt(100), func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j--
			return true
		})
		if j != 100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}

func BenchmarkAscendGreaterOrEqualG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		k := 0
		tr.AscendGreaterOrEqual(newTestInt(100), func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j++
			k++
			return true
		})
		if j != len(arr) {
			b.Fatalf("expected: %v, got %v", len(arr), j)
		}
		if k != len(arr)-100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, k)
		}
	}
}

func BenchmarkDescendLessOrEqualG(b *testing.B) {
	arr := rand.Perm(benchmarkTreeSize)
	tr := New[*testInt](*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(newTestInt(v))
	}
	sort.Ints(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		k := len(arr)
		tr.DescendLessOrEqual(newTestInt(arr[len(arr)-100]), func(item *testInt) bool {
			if int(*item) != arr[j] {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j], item)
			}
			j--
			k--
			return true
		})
		if j != -1 {
			b.Fatalf("expected: %v, got %v", -1, j)
		}
		if k != 99 {
			b.Fatalf("expected: %v, got %v", 99, k)
		}
	}
}

func BenchmarkDeleteAndRestoreG(b *testing.B) {
	items := rand.Perm(16392)
	b.ResetTimer()
	b.Run(`CopyBigFreeList`, func(b *testing.B) {
		fl := NewFreeList[*testInt](16392)
		tr := NewWithFreeList(*btreeDegree, fl)
		for _, v := range items {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dels := make([]*testInt, 0, tr.Len())
			tr.Ascend(func(b *testInt) bool {
				dels = append(dels, b)
				return true
			})
			for _, del := range dels {
				tr.Delete(del)
			}
			// tr is now empty, we make a new empty copy of it.
			tr = NewWithFreeList(*btreeDegree, fl)
			for _, v := range items {
				tr.ReplaceOrInsert(newTestInt(v))
			}
		}
	})
	b.Run(`Copy`, func(b *testing.B) {
		tr := New[*testInt](*btreeDegree)
		for _, v := range items {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dels := make([]*testInt, 0, tr.Len())
			tr.Ascend(func(b *testInt) bool {
				dels = append(dels, b)
				return true
			})
			for _, del := range dels {
				tr.Delete(del)
			}
			// tr is now empty, we make a new empty copy of it.
			tr := New[*testInt](*btreeDegree)
			for _, v := range items {
				tr.ReplaceOrInsert(newTestInt(v))
			}
		}
	})
	b.Run(`ClearBigFreelist`, func(b *testing.B) {
		fl := NewFreeList[*testInt](16392)
		tr := NewWithFreeList(*btreeDegree, fl)
		for _, v := range items {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Clear(true)
			for _, v := range items {
				tr.ReplaceOrInsert(newTestInt(v))
			}
		}
	})
	b.Run(`Clear`, func(b *testing.B) {
		tr := New[*testInt](*btreeDegree)
		for _, v := range items {
			tr.ReplaceOrInsert(newTestInt(v))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Clear(true)
			for _, v := range items {
				tr.ReplaceOrInsert(newTestInt(v))
			}
		}
	})
}
