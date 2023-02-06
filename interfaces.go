//go:build !goexperiment.arenas

package btree

// Item is the BTree objects interface.
type Item[T any] interface {
	Less(T) bool
	DeepCopy() T
}
