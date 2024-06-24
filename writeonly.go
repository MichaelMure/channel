package channel

import "context"

// WriteOnly is a write-only channel wrapper offering improvements over a normal go channel.
type WriteOnly[T any] struct {
	_ uncomparable
	c *C[T]
}

// Write is the same as C.Write
func (c WriteOnly[T]) Write(v T) {
	c.c.Write(v)
}

// WriteContext is the same as C.WriteContext
func (c WriteOnly[T]) WriteContext(ctx context.Context, v T) error {
	return c.c.WriteContext(ctx, v)
}

// WriteChannel is the same as C.WriteChannel
func (c WriteOnly[T]) WriteChannel() chan<- T {
	return c.c.WriteChannel()
}

// SetError is the same as C.SetError
func (c WriteOnly[T]) SetError(err error) {
	c.c.SetError(err)
}

// Close is the same as C.Close
func (c WriteOnly[T]) Close() {
	c.c.Close()
}

// CloseWithError is the same as C.CloseWithError
func (c WriteOnly[T]) CloseWithError(err error) {
	c.c.CloseWithError(err)
}
