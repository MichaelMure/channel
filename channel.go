// This is a wrapper around channels, it's useful because it adds an error field which indicate how the request terminated, this can be red by multiple consumers.
package channel

import (
	"context"
	"io"
)

type uncomparable = [0]func()

type ReadOnly[T any] struct {
	_ uncomparable
	c *C[T]
}

// Read is the same C.Read
func (c ReadOnly[T]) Read() (T, error) {
	return c.c.Read()
}

// ReadCtx is the same as C.ReadContext
func (c ReadOnly[T]) ReadContext(ctx context.Context) (T, error) {
	return c.c.ReadContext(ctx)
}

// ReadChannel is the same as C.ReadChannel
func (c ReadOnly[T]) ReadChannel() <-chan T {
	return c.c.ReadChannel()
}

// Rest is the same as C.Rest
func (c ReadOnly[T]) Rest() ([]T, error) {
	return c.c.Rest()
}

// RestContext is the same as C.RestContext
func (c ReadOnly[T]) RestContext(ctx context.Context) ([]T, error) {
	return c.c.RestContext(ctx)
}

// Err is the same as C.Err
func (c ReadOnly[T]) Err() error {
	return c.c.Err()
}

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

// Read returns io.EOF when the channel is closed.
func (c *C[T]) Read() (T, error) {
	v, ok := <-c.c
	if !ok {
		var zero T
		return zero, c.Err()
	}
	return v, nil
}

// ReadCtx returns io.EOF when the channel is closed.
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

// Rest reads all the values in the channel until it closes, and return them all at once.
func (c *C[T]) Rest() ([]T, error) {
	var res []T
	for {
		v, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

// RestContext reads all the values in the channel until it closes, and return them all at once.
func (c *C[T]) RestContext(ctx context.Context) ([]T, error) {
	var res []T
	for {
		v, err := c.ReadContext(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

// Err allows to access the error when the channel passed by ReadChannel is viewed closed.
// This is threadunsafe if called when the channel is not closed.
func (c *C[T]) Err() error {
	return c.err
}

// ReadOnly returns a ReadOnly channel view of this channel.
func (c *C[T]) ReadOnly() ReadOnly[T] {
	return ReadOnly[T]{c: c}
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

// WriteOnly returns a WriteOnly channel view of this channel.
func (c *C[T]) WriteOnly() WriteOnly[T] {
	return WriteOnly[T]{c: c}
}
