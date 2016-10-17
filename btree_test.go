// Copyright 2014 Google Inc.
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

package btree

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

func init() {
	seed := time.Now().Unix()
	fmt.Println(seed)
	rand.Seed(seed)
}

func sorted(orig []Item) []Item {
	result := make([]Item, len(orig))
	copy(result, orig)
	sort.Sort(byInts(result))
	return result
}

func difference(orig, subtract []Item) (result []Item) {
	var idx int
	subLen := len(subtract)
	for _, item := range orig {
		for ; idx < subLen && subtract[idx].Less(item); idx++ {
		}
		if idx >= subLen || item != subtract[idx] {
			result = append(result, item)
		}
	}
	return
}

// perm returns a random permutation of n Int items in the range [0, n).
func perm(n int) (out []Item) {
	for _, v := range rand.Perm(n) {
		out = append(out, Int(v))
	}
	return
}

// permf applies f to each element in a random permutation of range [0, n).
func permf(n int, f func(i int) int) (out []Item) {
	for _, v := range rand.Perm(n) {
		out = append(out, Int(f(v)))
	}
	return
}

// rang returns an ordered list of Int items in the range [0, n).
func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, Int(i))
	}
	return
}

type ascender interface {
	Ascend(ItemIterator)
}

// all extracts all items from a tree in order as a slice.
func all(t ascender) (out []Item) {
	t.Ascend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

// rangerev returns a reversed ordered list of Int items in the range [0, n).
func rangrev(n int) (out []Item) {
	for i := n - 1; i >= 0; i-- {
		out = append(out, Int(i))
	}
	return
}

// allrev extracts all items from a tree in reverse order as a slice.
func allrev(t *BTree) (out []Item) {
	t.Descend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

func TestImmutableBTree(t *testing.T) {
	builder := CopyOf(NewImmutable(4))
	const treeSize = 1024
	const sizeIncr = 32
	for i := 0; i < 10; i++ {
		if min := builder.Min(); min != nil {
			t.Fatalf("empty min, got %+v", min)
		}
		if max := builder.Max(); max != nil {
			t.Fatalf("empty max, got %+v", max)
		}
		trees := make([]*ImmutableBTree, treeSize/sizeIncr)
		aPerm := perm(treeSize)
		for i, item := range aPerm {
			if i%sizeIncr == 0 {
				trees[i/sizeIncr] = builder.Copy()
			}
			if x := builder.ReplaceOrInsert(item); x != nil {
				t.Fatal("insert found item", item)
			}
		}
		for _, item := range perm(treeSize) {
			if x := builder.ReplaceOrInsert(item); x == nil {
				t.Fatal("insert didn't find item", item)
			}
		}
		fullTree := builder.Copy()
		if min, want := fullTree.Min(), Item(Int(0)); min != want {
			t.Fatalf("min: want %+v, got %+v", want, min)
		}
		if max, want := fullTree.Max(), Item(Int(treeSize-1)); max != want {
			t.Fatalf("max: want %+v, got %+v", want, max)
		}
		got := all(fullTree)
		want := rang(treeSize)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}
		// Now check partial trees
		for i, partialTree := range trees {
			got := all(partialTree)
			if i == 0 {
				if len(got) > 0 {
					t.Fatalf("Expected empty, got %v", got)
				}
			} else {
				want := sorted(aPerm[:(i * sizeIncr)])
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
				}
			}
		}
		builder.Set(fullTree)
		aPerm = perm(treeSize)
		for i, item := range aPerm {
			if i%sizeIncr == 0 {
				trees[i/sizeIncr] = builder.Copy()
			}
			if x := builder.Delete(item); x == nil {
				t.Fatalf("didn't find %v", item)
			}
		}
		// Now check partial trees
		allNumbers := rang(treeSize)
		for i, partialTree := range trees {
			got := all(partialTree)
			want := difference(allNumbers, sorted(aPerm[:(i*sizeIncr)]))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
			}
		}
		if got = all(builder); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
	}
}

func TestBTree(t *testing.T) {
	tr := New(*btreeDegree)
	const treeSize = 10000
	for i := 0; i < 10; i++ {
		if min := tr.Min(); min != nil {
			t.Fatalf("empty min, got %+v", min)
		}
		if max := tr.Max(); max != nil {
			t.Fatalf("empty max, got %+v", max)
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x != nil {
				t.Fatal("insert found item", item)
			}
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x == nil {
				t.Fatal("insert didn't find item", item)
			}
		}
		if min, want := tr.Min(), Item(Int(0)); min != want {
			t.Fatalf("min: want %+v, got %+v", want, min)
		}
		if max, want := tr.Max(), Item(Int(treeSize-1)); max != want {
			t.Fatalf("max: want %+v, got %+v", want, max)
		}
		got := all(tr)
		want := rang(treeSize)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		gotrev := allrev(tr)
		wantrev := rangrev(treeSize)
		if !reflect.DeepEqual(gotrev, wantrev) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		for _, item := range perm(treeSize) {
			if x := tr.Delete(item); x == nil {
				t.Fatalf("didn't find %v", item)
			}
		}
		if got = all(tr); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
	}
}

func TestImmutableBTreeBuilderReuse(t *testing.T) {
	builder := CopyOf(NewImmutable(3))
	for i := 0; i < 1000; i += 2 {
		builder.ReplaceOrInsert(Int(i))
	}
	twos := builder.Copy()
	for i := 0; i < 1000; i += 4 {
		builder.Delete(Int(i))
	}
	for i := 5; i < 1000; i += 10 {
		builder.ReplaceOrInsert(Int(i))
	}
	minus4sPlusOdd5s := builder.Copy()
	builder.Set(twos)
	for i := 0; i < 1000; i += 6 {
		builder.Delete(Int(i))
	}
	for i := 7; i < 1000; i += 14 {
		builder.ReplaceOrInsert(Int(i))
	}
	minus6sPlusOdd7s := builder.Copy()

	var want []Item

	// Verify twos
	for i := 0; i < 1000; i += 2 {
		want = append(want, Int(i))
	}
	if got := twos.Len(); got != len(want) {
		t.Fatalf("Expected %v, got %v", len(want), got)
	}
	got := all(twos)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
	}

	// Verify minus4sPlusOdd5s
	want = nil
	for i := 2; i < 1000; i += 4 {
		want = append(want, Int(i))
	}
	for i := 5; i < 1000; i += 10 {
		want = append(want, Int(i))
	}
	want = sorted(want)
	if got := minus4sPlusOdd5s.Len(); got != len(want) {
		t.Fatalf("Expected %v, got %v", len(want), got)
	}
	got = all(minus4sPlusOdd5s)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
	}

	// Verify minus6sPlusOdd7s
	want = nil
	for i := 2; i < 1000; i += 2 {
		if i%6 != 0 {
			want = append(want, Int(i))
		}
	}
	for i := 7; i < 1000; i += 14 {
		want = append(want, Int(i))
	}
	want = sorted(want)
	if got := minus6sPlusOdd7s.Len(); got != len(want) {
		t.Fatalf("Expected %v, got %v", len(want), got)
	}
	got = all(minus6sPlusOdd7s)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
	}
}

func TestInsertExistingImmutableBTree(t *testing.T) {
	const initialSize = 10000
	const batchSize = 100
	// 0,2,4,6,...,19998
	insertP := permf(initialSize, func(i int) int { return 2 * i })
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	var trees [10]*ImmutableBTree
	var batches [10][]Item
	for i := range trees {
		// 100 random numbers taken from 1,3,5,7,...,19999
		batches[i] = permf(
			initialSize,
			func(i int) int { return 2*i + 1 })[:batchSize]
		builder := CopyOf(tr)
		for _, item := range batches[i] {
			builder.ReplaceOrInsert(item)
		}
		trees[i] = builder.Copy()
	}
	// Test each of the trees
	for i := range trees {
		got := all(trees[i])
		var want []Item
		want = append(want, insertP...)
		want = append(want, batches[i]...)
		want = sorted(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}
		if got, want := trees[i].Len(), initialSize+batchSize; got != want {
			t.Fatalf("got size: %v\nwant: %v", got, want)
		}
	}
	// Test tr
	got := all(tr)
	want := sorted(insertP)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
	}
	if got, want := tr.Len(), initialSize; got != want {
		t.Fatalf("got size: %v\nwant: %v", got, want)
	}
}

func TestDeleteExistingImmutableBTree(t *testing.T) {
	const initialSize = 10000
	const batchSize = 100
	// 0,1,2,3,...,9999
	insertP := perm(initialSize)
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	var trees [10]*ImmutableBTree
	var batches [10][]Item
	for i := range trees {
		// 100 random numbers taken from 0,1,2,...,9999
		batches[i] = perm(initialSize)[:batchSize]
		builder := CopyOf(tr)
		for _, item := range batches[i] {
			builder.Delete(item)
		}
		trees[i] = builder.Copy()
	}
	// Test each of the trees
	for i := range trees {
		got := all(trees[i])
		want := difference(
			rang(initialSize), sorted(batches[i]))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}
		if got, want := trees[i].Len(), initialSize-batchSize; got != want {
			t.Fatalf("got size: %v\nwant: %v", got, want)
		}
	}
	// Test tr
	got := all(tr)
	want := sorted(insertP)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
	}
	if got, want := tr.Len(), initialSize; got != want {
		t.Fatalf("got size: %v\nwant: %v", got, want)
	}
}

func ExampleImmutableBTree() {
	builder := CopyOf(NewImmutable(*btreeDegree))
	for i := Int(0); i < 10; i++ {
		builder.ReplaceOrInsert(i)
	}
	zeroTo9 := builder.Copy()
	builder.DeleteMax()
	builder.ReplaceOrInsert(Int(100))
	builder.ReplaceOrInsert(Int(50))
	no9But50And100 := builder.Copy()
	builder.Set(zeroTo9)
	builder.DeleteMin()
	builder.Delete(Int(7))
	no0no7 := builder.Copy()
	fmt.Println("len:       ", zeroTo9.Len())
	fmt.Println("get3:      ", zeroTo9.Get(Int(3)))
	fmt.Println("get100:    ", zeroTo9.Get(Int(100)))
	fmt.Println("min:       ", zeroTo9.Min())
	fmt.Println("max:       ", zeroTo9.Max())
	fmt.Println()
	fmt.Println("len:       ", no9But50And100.Len())
	fmt.Println("get9:      ", no9But50And100.Get(Int(9)))
	fmt.Println("get50:     ", no9But50And100.Get(Int(50)))
	fmt.Println("min:       ", no9But50And100.Min())
	fmt.Println("max:       ", no9But50And100.Max())
	fmt.Println()
	fmt.Println("len:       ", no0no7.Len())
	fmt.Println("get7:      ", no0no7.Get(Int(7)))
	fmt.Println("get4:      ", no0no7.Get(Int(4)))
	fmt.Println("min:       ", no0no7.Min())
	fmt.Println("max:       ", no0no7.Max())

	// Output:
	// len:        10
	// get3:       3
	// get100:     <nil>
	// min:        0
	// max:        9
	//
	// len:        11
	// get9:       <nil>
	// get50:      50
	// min:        0
	// max:        100
	//
	// len:        8
	// get7:       <nil>
	// get4:       4
	// min:        1
	// max:        9
}

func ExampleBTree() {
	tr := New(*btreeDegree)
	for i := Int(0); i < 10; i++ {
		tr.ReplaceOrInsert(i)
	}
	fmt.Println("len:       ", tr.Len())
	fmt.Println("get3:      ", tr.Get(Int(3)))
	fmt.Println("get100:    ", tr.Get(Int(100)))
	fmt.Println("del4:      ", tr.Delete(Int(4)))
	fmt.Println("del100:    ", tr.Delete(Int(100)))
	fmt.Println("replace5:  ", tr.ReplaceOrInsert(Int(5)))
	fmt.Println("replace100:", tr.ReplaceOrInsert(Int(100)))
	fmt.Println("min:       ", tr.Min())
	fmt.Println("delmin:    ", tr.DeleteMin())
	fmt.Println("max:       ", tr.Max())
	fmt.Println("delmax:    ", tr.DeleteMax())
	fmt.Println("len:       ", tr.Len())
	// Output:
	// len:        10
	// get3:       3
	// get100:     <nil>
	// del4:       4
	// del100:     <nil>
	// replace5:   5
	// replace100: <nil>
	// min:        0
	// delmin:     0
	// max:        100
	// delmax:     100
	// len:        8
}

func TestDeleteMin(t *testing.T) {
	tr := New(3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMin(); v != nil; v = tr.DeleteMin() {
		got = append(got, v)
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDeleteMax(t *testing.T) {
	tr := New(3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMax(); v != nil; v = tr.DeleteMax() {
		got = append(got, v)
	}
	// Reverse our list.
	for i := 0; i < len(got)/2; i++ {
		got[i], got[len(got)-i-1] = got[len(got)-i-1], got[i]
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestImmutableDeleteMin(t *testing.T) {
	builder := CopyOf(NewImmutable(3))
	for _, v := range perm(100) {
		builder.ReplaceOrInsert(v)
	}
	zeroTo99 := builder.Copy()
	var got []Item
	for v := builder.DeleteMin(); v != nil; v = builder.DeleteMin() {
		got = append(got, v)
	}
	empty := builder.Copy()
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = all(zeroTo99)
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange2:\n got: %v\nwant: %v", got, want)
	}
	got = all(empty)
	if want := rang(0); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange3:\n got: %v\nwant: %v", got, want)
	}
}

func TestImmutableDeleteMax(t *testing.T) {
	builder := CopyOf(NewImmutable(3))
	for _, v := range perm(100) {
		builder.ReplaceOrInsert(v)
	}
	zeroTo99 := builder.Copy()
	var got []Item
	for v := builder.DeleteMax(); v != nil; v = builder.DeleteMax() {
		got = append(got, v)
	}
	empty := builder.Copy()
	// Reverse our list.
	for i := 0; i < len(got)/2; i++ {
		got[i], got[len(got)-i-1] = got[len(got)-i-1], got[i]
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = all(zeroTo99)
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange2:\n got: %v\nwant: %v", got, want)
	}
	got = all(empty)
	if want := rang(0); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange3:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendRange(t *testing.T) {
	tr := New(2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendRange(t *testing.T) {
	tr := New(2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendLessThan(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendLessThan(Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendLessThan(Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendLessOrEqual(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendLessOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[59:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendLessOrEqual(Int(60), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendGreaterOrEqual(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendGreaterThan(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
}

const benchmarkTreeSize = 10000

func BenchmarkCopy(b *testing.B) {
	b.StopTimer()
	insert := perm(benchmarkTreeSize)
	builder := New(*btreeDegree)
	for _, item := range insert {
		builder.ReplaceOrInsert(item)
	}
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := builder.Copy()
		i++
		if tr.Len() != benchmarkTreeSize {
			panic(tr.Len())
		}
	}
}

func BenchmarkInsert(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New(*btreeDegree)
		for _, item := range insertP {
			tr.ReplaceOrInsert(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkImmutableInsert1(b *testing.B) {
	b.StopTimer()
	// 0,2,4,6,...,19998
	insertP := permf(benchmarkTreeSize, func(i int) int { return 2 * i })
	// Of form 2*n + 1 n is in [0, 9999)
	itemToInsert := Int(2*rand.Intn(benchmarkTreeSize) + 1)
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	expectedNewTrSize := benchmarkTreeSize + 1
	b.StartTimer()
	i := 0
	for i < b.N {
		builder := CopyOf(tr)
		builder.ReplaceOrInsert(itemToInsert)
		i++
		newTr := builder.Copy()
		if newTr.Len() != expectedNewTrSize {
			panic(newTr.Len())
		}
	}
}

func BenchmarkImmutableInsert100(b *testing.B) {
	const batchSize = 100
	b.StopTimer()
	// 0,2,4,6,...,19998
	insertP := permf(benchmarkTreeSize, func(i int) int { return 2 * i })
	// 100 random numbers taken from 1,3,5,7,...,19999
	insertB := permf(
		benchmarkTreeSize,
		func(i int) int { return 2*i + 1 })[:batchSize]
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	expectedNewTrSize := benchmarkTreeSize + batchSize
	b.StartTimer()
	i := 0
	for i < b.N {
		builder := CopyOf(tr)
		for _, item := range insertB {
			builder.ReplaceOrInsert(item)
			i++
			if i >= b.N {
				return
			}
		}
		newTr := builder.Copy()
		if newTr.Len() != expectedNewTrSize {
			panic(newTr.Len())
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	removeP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New(*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Delete(item)
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

func BenchmarkImmutableDelete1(b *testing.B) {
	b.StopTimer()
	// 0,1,2,3,...,9999
	insertP := perm(benchmarkTreeSize)
	itemToDelete := Int(rand.Intn(benchmarkTreeSize))
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	expectedNewTrSize := benchmarkTreeSize - 1
	b.StartTimer()
	i := 0
	for i < b.N {
		builder := CopyOf(tr)
		builder.Delete(itemToDelete)
		i++
		newTr := builder.Copy()
		if newTr.Len() != expectedNewTrSize {
			panic(newTr.Len())
		}
	}
}

func BenchmarkImmutableDelete100(b *testing.B) {
	const batchSize = 100
	b.StopTimer()
	// 0,1,2,3,...,9999
	insertP := perm(benchmarkTreeSize)
	// 100 random numbers taken from 0,1,2,3,...,9999
	deleteB := perm(benchmarkTreeSize)[:batchSize]
	builder := CopyOf(NewImmutable(*btreeDegree))
	for _, item := range insertP {
		builder.ReplaceOrInsert(item)
	}
	tr := builder.Copy()
	expectedNewTrSize := benchmarkTreeSize - batchSize
	b.StartTimer()
	i := 0
	for i < b.N {
		builder := CopyOf(tr)
		for _, item := range deleteB {
			builder.Delete(item)
			i++
			if i >= b.N {
				return
			}
		}
		newTr := builder.Copy()
		if newTr.Len() != expectedNewTrSize {
			panic(newTr.Len())
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	removeP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New(*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Get(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

type byInts []Item

func (a byInts) Len() int {
	return len(a)
}

func (a byInts) Less(i, j int) bool {
	return a[i].(Int) < a[j].(Int)
}

func (a byInts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func BenchmarkAscend(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 0
		tr.Ascend(func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j++
			return true
		})
	}
}

func BenchmarkDescend(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 1
		tr.Descend(func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j--
			return true
		})
	}
}
func BenchmarkAscendRange(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		tr.AscendRange(Int(100), arr[len(arr)-100], func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j++
			return true
		})
		if j != len(arr)-100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}

func BenchmarkDescendRange(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		tr.DescendRange(arr[len(arr)-100], Int(100), func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j--
			return true
		})
		if j != 100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}
func BenchmarkAscendGreaterOrEqual(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		k := 0
		tr.AscendGreaterOrEqual(Int(100), func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
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
func BenchmarkDescendLessOrEqual(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		k := len(arr)
		tr.DescendLessOrEqual(arr[len(arr)-100], func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
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
