//go:build goexperiment.arenas

package btree

import "arena"

// Item is the BTree objects interface.
type Item[T any] interface {
	Less(T) bool
	DeepCopy() T
	DeepCopyWithArena(*arena.Arena) T
}
