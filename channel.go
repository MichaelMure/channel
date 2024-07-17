// Package channel contains a wrapper around channels, simplifying and expending usage.
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

// Read perform a blocking read on the channel.
// If the channel is simply closed, the io.EOF error is returned.
// If the channel has been closed with an error, that error is returned.
func (c *C[T]) Read() (T, error) {
	v, ok := <-c.c
	if !ok {
		var zero T
		return zero, c.Err()
	}
	return v, nil
}

// ReadContext performs a blocking read on the channel, but returns if the context expires.
// If the channel is simply closed, the io.EOF error is returned.
// If the channel has been closed with an error, that error is returned.
// If the context expires, the context's error is returned. This does not close the channel.
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

// ReadChannel allows accessing the underlying channel for use with golang's select.
// You should use Read or ReadContext when you can.
func (c *C[T]) ReadChannel() <-chan T {
	return c.c
}

// Range calls the given function on the channel values until the channel get closed.
// Important: if the given function returns an error, the channel can be left with
// values to be drained if that was the last possible reader, which can create a
// goroutine and/or memory leak. You can use Drain or DrainContext to make sure it
// doesn't happen.
// If the channel is simply closed, no error is returned.
// If the channel had been closed with an error, this error is returned.
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

// RangeContext calls the given function on the channel values until the channel get closed or the context expires.
// Important: if the given function returns an error, the channel can be left with
// values to be drained if that was the last possible reader, which can create a
// goroutine and/or memory leak. You can use Drain or DrainContext to make sure it
// doesn't happen.
// If the channel is simply closed, no error is returned.
// If the channel had been closed with an error, this error is returned.
// If the context expires, the iteration stops and the context's error is returned. This does not close the channel.
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
// If the channel is simply closed, no error is returned.
// If the channel had been closed with an error, this error is returned.
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
// If the channel is simply closed, no error is returned.
// If the channel had been closed with an error, this error is returned.
// If the context expires, the iteration stops and the context's error is returned along with possibly partial results.
// This does not close the channel.
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
// This is useful, for example, to write a wrapper operating on an underlying channel.
// If the intercepting function returns an error, the output channel is closed with that error.
// Any error set on the underlying channel is replicated onto the output channel.
func (c *C[T]) Intercept(fn func(T) error) *C[T] {
	// output with zero capacity, we don't want to add another buffering
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
// This is useful, for example, to write a wrapper operating on an underlying channel.
// If the intercepting function returns an error, the value is not propagated into the output channel.
// Any error set on the underlying channel is replicated onto the output channel.
// If the context expires, the iteration stops and the context's error is returned. This does not close the underlying
// channel, but does close the output channel.
func (c *C[T]) InterceptContext(ctx context.Context, fn func(T) error) *C[T] {
	// output with zero capacity, we don't want to add another buffering
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

// Drain read and discard values from the channel until it's closed.
func (c *C[T]) Drain() {
	for {
		_, err := c.Read()
		if err != nil {
			return
		}
	}
}

// DrainContext read and discard values from the channel until it's closed or the context expires.
func (c *C[T]) DrainContext(ctx context.Context) {
	for {
		_, err := c.ReadContext(ctx)
		if err != nil {
			return
		}
	}
}

// Len returns the number of elements queued (unread) in the channel buffer.
// If the channel is not buffered, Len returns zero.
func (c *C[T]) Len() int {
	return len(c.c)
}

// Cap returns the capacity of the channel buffer.
// If the channel is not buffered, Cap returns zero.
func (c *C[T]) Cap() int {
	return cap(c.c)
}

// Err allows accessing the error when the channel passed by ReadChannel is viewed closed.
// This is not thread-safe if called when the channel is not closed.
func (c *C[T]) Err() error {
	return c.err
}

// Write performs a blocking write of a value to the channel.
// It panics if writing to a closed channel.
func (c *C[T]) Write(v T) {
	c.c <- v
}

// WriteContext performs a blocking write of a value to the channel, but returns if the context expires.
// It panics if writing to a closed channel.
// If the context expires, the context's error is returned. This does not close the channel.
func (c *C[T]) WriteContext(ctx context.Context, v T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.c <- v:
		return nil
	}
}

// WriteChannel allows accessing the underlying channel for use with golang's select.
// You should use Write or WriteContext when you can.
func (c *C[T]) WriteChannel() chan<- T {
	return c.c
}

// SetError records the given error as the sticky error of this channel.
// It will panic if an error has already been set.
// It will replace err with io.EOF if err is nil.
// It will never block.
// It is not thread-safe with any write operation.
func (c *C[T]) SetError(err error) {
	if c.err != nil {
		panic("setting error on an already errored channel")
	}
	if err == nil {
		err = io.EOF
	}
	c.err = err
}

// Close will set the sticky error to io.EOF if it is not already set and then close the channel.
// It will never block.
// It will panic if trying to close an already closed channel.
// It is not thread-safe with any write operation.
func (c *C[T]) Close() {
	if c.err == nil {
		c.err = io.EOF
	}
	close(c.c)
}

// CloseWithError will close the channel with the given error and then close the channel.
// It will replace err with io.EOF if err is nil.
// It will never block.
// It will panic if trying to close an already closed channel.
// It is not thread-safe with any write operation.
func (c *C[T]) CloseWithError(err error) {
	c.SetError(err)
	c.Close()
}
