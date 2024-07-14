package channel

import "context"

// ReadOnly is a read-only channel wrapper offering improvements over a normal go channel.
type ReadOnly[T any] struct {
	_ uncomparable
	c *C[T]
}

// Read is the same C.Read
func (c ReadOnly[T]) Read() (T, error) {
	return c.c.Read()
}

// ReadContext is the same as C.ReadContext
func (c ReadOnly[T]) ReadContext(ctx context.Context) (T, error) {
	return c.c.ReadContext(ctx)
}

// ReadChannel is the same as C.ReadChannel
func (c ReadOnly[T]) ReadChannel() <-chan T {
	return c.c.ReadChannel()
}

// Range is the same as C.Range
func (c ReadOnly[T]) Range(fn func(T) error) error {
	return c.c.Range(fn)
}

// RangeContext is the same as C.RangeContext
func (c ReadOnly[T]) RangeContext(ctx context.Context, fn func(T) error) error {
	return c.c.RangeContext(ctx, fn)
}

// Rest is the same as C.Rest
func (c ReadOnly[T]) Rest() ([]T, error) {
	return c.c.Rest()
}

// RestContext is the same as C.RestContext
func (c ReadOnly[T]) RestContext(ctx context.Context) ([]T, error) {
	return c.c.RestContext(ctx)
}

// Intercept is the same as C.Intercept, but return a ReadOnly.
func (c ReadOnly[T]) Intercept(fn func(T) error) ReadOnly[T] {
	return c.c.Intercept(fn).ReadOnly()
}

// InterceptContext is the same as C.InterceptContext, but return a ReadOnly.
func (c ReadOnly[T]) InterceptContext(ctx context.Context, fn func(T) error) ReadOnly[T] {
	return c.c.InterceptContext(ctx, fn).ReadOnly()
}

// Drain is the same as C.Drain
func (c ReadOnly[T]) Drain() {
	c.c.Drain()
}

// DrainContext is the same as C.DrainContext
func (c ReadOnly[T]) DrainContext(ctx context.Context) {
	c.c.DrainContext(ctx)
}

// Len is the same as C.Len
func (c ReadOnly[T]) Len() int {
	return c.c.Len()
}

// Cap is the same as C.Cap
func (c ReadOnly[T]) Cap() int {
	return c.c.Cap()
}

// Err is the same as C.Err
func (c ReadOnly[T]) Err() error {
	return c.c.Err()
}
