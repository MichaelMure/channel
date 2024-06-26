// This is a wrapper around channels, it's useful because it adds an error field which indicate how the request terminated, this can be red by multiple consumers.
package channel

import (
	"context"
	"io"
)

type uncomparable = [0]func()

// C is a read-write channel wrapper offering improvements over a normal go channel.
// It can be declined into read-only and write-only using ReadOnly and WriteOnly.
type C[T any] struct {
	c   chan T
	err error
}

// New create a new unbuffered channel.
func New[T any]() *C[T] {
	return NewWithSize[T](0)
}

// NewWithSize create a new channel of the i size.
func NewWithSize[T any](i int) *C[T] {
	return &C[T]{c: make(chan T, i)}
}

// NewWithError create an already closed and errored channel.
// It will replace err with io.EOF if err is nil.
func NewWithError[T any](err error) *C[T] {
	c := New[T]()
	c.CloseWithError(err)
	return c
}

// ReadOnly returns a ReadOnly channel view of this channel.
func (c *C[T]) ReadOnly() ReadOnly[T] {
	return ReadOnly[T]{c: c}
}

// WriteOnly returns a WriteOnly channel view of this channel.
func (c *C[T]) WriteOnly() WriteOnly[T] {
	return WriteOnly[T]{c: c}
}

// Read returns io.EOF when the channel is closed.
func (c *C[T]) Read() (T, error) {
	v, ok := <-c.c
	if !ok {
		var zero T
		return zero, c.Err()
	}
	return v, nil
}

// ReadContext returns io.EOF when the channel is closed.
func (c *C[T]) ReadContext(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case v, ok := <-c.c:
		if !ok {
			var zero T
			return zero, c.Err()
		}
		return v, nil
	}
}

// ReadChannel allows to access the underlying channel for use with select, you should use Read and ReadContext when you can.
func (c *C[T]) ReadChannel() <-chan T {
	return c.c
}

// Range iterate on the channel values in the same fashion as for a go channel.
// It will return an error if either the given function return an error, or the channel carry an error.
func (c *C[T]) Range(fn func(T) error) error {
	for {
		v, err := c.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = fn(v)
		if err != nil {
			return err
		}
	}
}

// RangeContext iterate on the channel values in the same fashion as for a go channel.
// It will return an error if the given function return an error, the channel carry an error, or if the context is cancelled.
func (c *C[T]) RangeContext(ctx context.Context, fn func(T) error) error {
	for {
		v, err := c.ReadContext(ctx)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = fn(v)
		if err != nil {
			return err
		}
	}
}

// Rest reads all the values in the channel until it closes, and return them all at once.
// If an error occurs, partial results are returned.
func (c *C[T]) Rest() ([]T, error) {
	res := append([]T(nil), make([]T, len(c.c))...) // preallocate space if known.
	for {
		v, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}
		res = append(res, v)
	}
	return res, nil
}

// RestContext reads all the values in the channel until it closes, and return them all at once.
// If an error occurs, partial results are returned.
func (c *C[T]) RestContext(ctx context.Context) ([]T, error) {
	res := append([]T(nil), make([]T, len(c.c))...) // preallocate space if known.
	for {
		v, err := c.ReadContext(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}
		res = append(res, v)
	}
	return res, nil
}

// Intercept construct another C with the given fn intercepting every value read.
// This is useful for example to write a wrapper operating on an underlying channel.
// If the intercepting function return an error, the value is not propagated into the output channel.
func (c *C[T]) Intercept(fn func(T) error) *C[T] {
	// output with zero capacity, we don't want to add buffering
	out := New[T]()

	go func() {
		defer out.Close()

		err := c.Range(func(val T) error {
			innerErr := fn(val)
			if innerErr != nil {
				return innerErr
			}
			out.Write(val)
			return nil
		})
		if err != nil {
			out.SetError(err)
		}
	}()

	return out
}

// InterceptContext construct another C with the given fn intercepting every value read.
// This is useful for example to write a wrapper operating on an underlying channel.
// If the intercepting function return an error, the value is not propagated into the output channel.
func (c *C[T]) InterceptContext(ctx context.Context, fn func(T) error) *C[T] {
	// output with zero capacity, we don't want to add buffering
	out := New[T]()

	go func() {
		defer out.Close()

		err := c.RangeContext(ctx, func(val T) error {
			innerErr := fn(val)
			if innerErr != nil {
				return innerErr
			}
			return out.WriteContext(ctx, val)
		})
		if err != nil {
			out.SetError(err)
		}
	}()

	return out
}

// Len returns the number of elements queued (unread) in the channel buffer.
// If v is nil, len(v) is zero.
func (c *C[T]) Len() int {
	return len(c.c)
}

// Err allows to access the error when the channel passed by ReadChannel is viewed closed.
// This is threadunsafe if called when the channel is not closed.
func (c *C[T]) Err() error {
	return c.err
}

// Write panic if writing to a closed channel.
func (c *C[T]) Write(v T) {
	c.c <- v
}

// WriteContext panic if writing to a closed channel.
// And error is returned when ctx.Done() was closed.
func (c *C[T]) WriteContext(ctx context.Context, v T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.c <- v:
		return nil
	}
}

// WriteChannel allows to access the underlying channel for use with select, you should use Write and WriteContext when you can.
func (c *C[T]) WriteChannel() chan<- T {
	return c.c
}

// SetError will panic if an error is already set.
// It will replace err with io.EOF if err is nil.
// It will never block.
// It is not threadsafe with any write operation.
func (c *C[T]) SetError(err error) {
	if c.err != nil {
		panic("setting error on an already errored channel")
	}
	if err == nil {
		err = io.EOF
	}
	c.err = err
}

// Close will set the error to io.EOF is it is not already set and then close the channel.
// It will never block.
// It will panic if trying to close an already closed channel.
// It is not threadsafe with any write operation.
func (c *C[T]) Close() {
	if c.err == nil {
		c.err = io.EOF
	}
	close(c.c)
}

// CloseWithError will close the channel with the provided error and then close the channel.
// It will replace err with io.EOF if err is nil.
// It will never block.
// It will panic if trying to close an already closed channel.
// It is not threadsafe with any write operation.
func (c *C[T]) CloseWithError(err error) {
	c.SetError(err)
	c.Close()
}
