//go:build goexperiment.arenas

package btree

import (
	"arena"
)

func (s items[T]) DeepCopyWithArena(a *arena.Arena) items[T] {
	s2 := arena.MakeSlice[T](a, 0, cap(s))

	for _, item := range s {
		s2 = append(s2, item.DeepCopyWithArena(a))
	}

	return s2
}

func (n *node[T]) DeepCopyWithArena(a *arena.Arena) *node[T] {
	n2 := arena.New[node[T]](a)

	if n == nil {
		return n2
	}

	n2.items = n.items.DeepCopyWithArena(a)
	n2.children = n.children.DeepCopyWithArena(a)

	return n2
}

func (t *BTree[T]) DeepCopyWithArena(a *arena.Arena) *BTree[T] {
	t2 := arena.New[BTree[T]](a)
	t2.freelist = arena.New[FreeList[T]](a)
	t2.freelist.freelist = arena.MakeSlice[*node[T]](a, 0, cap(t.freelist.freelist))
	t2.degree = t.degree
	t2.length = t.length
	t2.root = t.root.DeepCopyWithArena(a)

	setBTreeRootRecursive(t2, t2.root)

	return t2
}
