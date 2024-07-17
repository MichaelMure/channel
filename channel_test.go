package channel

import (
	"context"
	"errors"
	"io"
	"testing"
)

func makeProducer(count int) *C[int] {
	c := New[int]()

	go func() {
		defer c.Close()
		for i := 1; i < count+1; i++ {
			c.Write(i)
		}
	}()

	return c
}

func makeProducerNoClose(count int) *C[int] {
	c := New[int]()

	go func() {
		for i := 1; i < count+1; i++ {
			c.Write(i)
		}
	}()

	return c
}

func shouldPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func TestC_New(t *testing.T) {
	c := New[int]()

	if len(c.c) != 0 {
		t.Errorf("incorrect len")
	}
	if cap(c.c) != 0 {
		t.Errorf("incorrect cap")
	}
}

func TestC_NewWithSize(t *testing.T) {
	c := NewWithSize[int](2)

	c.Write(123)

	if len(c.c) != 1 {
		t.Errorf("incorrect len")
	}
	if cap(c.c) != 2 {
		t.Errorf("incorrect cap")
	}
}

func TestC_NewWithError(t *testing.T) {
	c := NewWithError[int](io.ErrUnexpectedEOF)

	if len(c.c) != 0 {
		t.Errorf("incorrect len")
	}
	if cap(c.c) != 0 {
		t.Errorf("incorrect cap")
	}
	_, err := c.Read()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestC_Read(t *testing.T) {
	c := makeProducerNoClose(2)

	v, err := c.Read()
	if err != nil {
		t.Error(err)
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = c.Read()
	if err != nil {
		t.Error(err)
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}

	c.Close()

	v, err = c.Read()
	if err != io.EOF {
		t.Errorf("incorrect error")
	}
}

func TestC_ReadContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducerNoClose(2)

	v, err := c1.ReadContext(ctx)
	if err != nil {
		t.Error(err)
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = c1.ReadContext(ctx)
	if err != nil {
		t.Error(err)
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}

	c1.Close()

	v, err = c1.ReadContext(ctx)
	if !errors.Is(err, io.EOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c2 := makeProducer(1)

	cancel()

	_, err = c2.ReadContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestC_ReadChannel(t *testing.T) {
	c := makeProducerNoClose(2)

	v := <-c.ReadChannel()
	if v != 1 {
		t.Errorf("unexpected value")
	}

	select {
	case v = <-c.ReadChannel():
	default:
	}

	if v != 2 {
		t.Errorf("unexpected value")
	}
}

func TestC_Range(t *testing.T) {
	c1 := makeProducer(2)

	var all []int

	err := c1.Range(func(val int) error {
		all = append(all, val)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}

	// ---

	c2 := makeProducer(2)

	err = c2.Range(func(i int) error {
		c2.Drain()
		return io.ErrUnexpectedEOF
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	err = c3.Range(func(i int) error {
		t.Errorf("should be unreachable")
		return nil
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestC_RangeContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)

	var all []int

	err := c1.RangeContext(ctx, func(val int) error {
		all = append(all, val)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}

	// ---

	c2 := makeProducer(2)

	err = c2.RangeContext(ctx, func(i int) error {
		c2.Drain()
		return io.ErrUnexpectedEOF
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	err = c3.RangeContext(ctx, func(i int) error {
		t.Errorf("should be unreachable")
		return nil
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c4 := makeProducer(1)

	cancel()

	err = c4.RangeContext(ctx, func(i int) error {
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestC_Rest(t *testing.T) {
	c1 := makeProducer(2)

	all, err := c1.Rest()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}

	// ---

	c2 := NewWithError[int](io.ErrUnexpectedEOF)

	_, err = c2.Rest()

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestC_RestContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)

	all, err := c1.RestContext(ctx)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}

	// ---

	c2 := NewWithError[int](io.ErrUnexpectedEOF)

	_, err = c2.RestContext(ctx)

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c3 := makeProducer(1)

	cancel()

	_, err = c3.RestContext(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestC_Intercept(t *testing.T) {
	c1 := makeProducer(2)

	var intercepted []int

	out := c1.Intercept(func(i int) error {
		intercepted = append(intercepted, i)
		return nil
	})

	all, err := out.Rest()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}
	if len(intercepted) != 2 {
		t.Errorf("incorrect len")
	}
	if intercepted[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if intercepted[1] != 2 {
		t.Errorf("incorrect second value")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c2 := makeProducer(2)

	intercepted = nil

	out = c2.Intercept(func(i int) error {
		if i == 2 {
			return io.ErrUnexpectedEOF
		}
		intercepted = append(intercepted, i)
		return nil
	})

	all, err = out.Rest()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if len(all) != 1 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if len(intercepted) != 1 {
		t.Errorf("incorrect len")
	}
	if intercepted[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if !errors.Is(out.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	intercepted = nil

	out = c3.Intercept(func(i int) error {
		intercepted = append(intercepted, i)
		return nil
	})

	all, err = out.Rest()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if len(intercepted) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}
}

func TestC_InterceptContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)

	var intercepted []int

	out := c1.InterceptContext(ctx, func(i int) error {
		intercepted = append(intercepted, i)
		return nil
	})

	all, err := out.RestContext(ctx)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 2 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if all[1] != 2 {
		t.Errorf("incorrect second value")
	}
	if len(intercepted) != 2 {
		t.Errorf("incorrect len")
	}
	if intercepted[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if intercepted[1] != 2 {
		t.Errorf("incorrect second value")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c2 := makeProducer(2)

	intercepted = nil

	out = c2.InterceptContext(ctx, func(i int) error {
		if i == 2 {
			return io.ErrUnexpectedEOF
		}
		intercepted = append(intercepted, i)
		return nil
	})

	all, err = out.RestContext(ctx)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if len(all) != 1 {
		t.Errorf("incorrect len")
	}
	if all[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if len(intercepted) != 1 {
		t.Errorf("incorrect len")
	}
	if intercepted[0] != 1 {
		t.Errorf("incorrect first value")
	}
	if !errors.Is(out.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	intercepted = nil

	out = c3.InterceptContext(ctx, func(i int) error {
		intercepted = append(intercepted, i)
		return nil
	})

	all, err = out.RestContext(ctx)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if len(intercepted) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c4 := makeProducer(1)

	intercepted = nil

	cancel()

	out = c4.InterceptContext(ctx, func(i int) error {
		intercepted = append(intercepted, i)
		return nil
	})

	all, err = out.RestContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if len(intercepted) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-out.c; valueFromChannel {
		t.Errorf("out channel should be closed")
	}
}

func TestC_Drain(t *testing.T) {
	c1 := makeProducer(10)

	c1.Drain()

	all, err := c1.Rest()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}
}

func TestC_DrainContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(10)

	c1.DrainContext(ctx)

	all, err := c1.RestContext(ctx)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c2 := makeProducer(1)

	cancel()

	c2.DrainContext(ctx)

	// there should still be values in the channel, as the DrainContext execution didn't
	// proceed, due to the cancelled context
	all, err = c2.Rest() // no context on purpose
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 1 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-c2.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}
}

func TestC_Len(t *testing.T) {
	c1 := New[int]()

	if c1.Len() != 0 {
		t.Errorf("incorrect len")
	}

	// ---

	c2 := NewWithSize[int](5)

	if c2.Len() != 0 {
		t.Errorf("incorrect len")
	}

	c2.Write(1)
	c2.Write(2)

	if c2.Len() != 2 {
		t.Errorf("incorrect len")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	if c3.Len() != 0 {
		t.Errorf("incorrect len")
	}
}

func TestC_Cap(t *testing.T) {
	c1 := New[int]()

	if c1.Cap() != 0 {
		t.Errorf("incorrect cap")
	}

	// ---

	c2 := NewWithSize[int](5)

	if c2.Cap() != 5 {
		t.Errorf("incorrect cap")
	}

	c2.Write(1)
	c2.Write(2)

	if c2.Cap() != 5 {
		t.Errorf("incorrect cap")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	if c3.Cap() != 0 {
		t.Errorf("incorrect cap")
	}
}

func TestC_Err(t *testing.T) {
	c1 := New[int]()

	if c1.Err() != nil {
		t.Errorf("unexpected error")
	}

	c1.Close()

	if !errors.Is(c1.Err(), io.EOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c2 := New[int]()

	c2.SetError(io.ErrUnexpectedEOF)

	if !errors.Is(c2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)

	if !errors.Is(c3.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestWrite(t *testing.T) {
	c := New[int]()

	go func() {
		c.Write(1)
		c.Write(2)
		c.Close()
	}()

	v, err := c.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = c.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}
}

func TestC_WriteContext(t *testing.T) {
	ctx := context.Background()

	c1 := New[int]()

	go func() {
		err := c1.WriteContext(ctx, 1)
		if err != nil {
			t.Errorf("unexpected error")
		}
		err = c1.WriteContext(ctx, 2)
		if err != nil {
			t.Errorf("unexpected error")
		}
		c1.Close()
	}()

	v, err := c1.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = c1.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c2 := New[int]()

	err = c2.WriteContext(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error")
	}

	if c2.Len() != 0 {
		t.Errorf("channel should be empty")
	}

	c2.Close()
}

func TestC_WriteChannel(t *testing.T) {
	c := NewWithSize[int](1)

	c.WriteChannel() <- 1

	v, err := c.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	select {
	case c.WriteChannel() <- 2:
	default:
	}

	v, err = c.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}
}

func TestC_SetError(t *testing.T) {
	c1 := New[int]()

	if c1.Err() != nil {
		t.Errorf("unexpected error")
	}

	c1.SetError(nil)

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}

	// ---

	c2 := New[int]()

	c2.SetError(io.ErrUnexpectedEOF)

	if !errors.Is(c2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := New[int]()
	c3.SetError(io.ErrUnexpectedEOF)

	shouldPanic(t, func() {
		c3.SetError(io.ErrClosedPipe)
	})
}

func TestC_Close(t *testing.T) {
	c1 := New[int]()

	c1.Close()

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c2 := New[int]()
	c2.Close()

	shouldPanic(t, func() {
		c2.Close()
	})
}

func TestC_CloseWithError(t *testing.T) {
	c1 := New[int]()

	c1.CloseWithError(nil)

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c2 := New[int]()

	c2.CloseWithError(io.ErrUnexpectedEOF)

	if !errors.Is(c2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c2.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c3 := New[int]()
	c3.Close()

	shouldPanic(t, func() {
		c3.CloseWithError(io.ErrUnexpectedEOF)
	})
}
