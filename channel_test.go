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
