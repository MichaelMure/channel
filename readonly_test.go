package channel

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestReadOnly_Read(t *testing.T) {
	c := makeProducerNoClose(2)
	cro := c.ReadOnly()

	v, err := cro.Read()
	if err != nil {
		t.Error(err)
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = cro.Read()
	if err != nil {
		t.Error(err)
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}

	c.Close()

	v, err = cro.Read()
	if err != io.EOF {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_ReadContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducerNoClose(2)
	cro1 := c1.ReadOnly()

	v, err := cro1.ReadContext(ctx)
	if err != nil {
		t.Error(err)
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	v, err = cro1.ReadContext(ctx)
	if err != nil {
		t.Error(err)
	}
	if v != 2 {
		t.Errorf("unexpected value")
	}

	c1.Close()

	v, err = cro1.ReadContext(ctx)
	if !errors.Is(err, io.EOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c2 := makeProducer(1)
	cro2 := c2.ReadOnly()

	cancel()

	_, err = cro2.ReadContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_ReadChannel(t *testing.T) {
	c := makeProducerNoClose(2)
	cro := c.ReadOnly()

	v := <-cro.ReadChannel()
	if v != 1 {
		t.Errorf("unexpected value")
	}

	select {
	case v = <-cro.ReadChannel():
	default:
	}

	if v != 2 {
		t.Errorf("unexpected value")
	}
}

func TestReadOnly_Range(t *testing.T) {
	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	var all []int

	err := cro1.Range(func(val int) error {
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
	cro2 := c2.ReadOnly()

	err = cro2.Range(func(i int) error {
		cro2.Drain()
		return io.ErrUnexpectedEOF
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	err = cro3.Range(func(i int) error {
		t.Errorf("should be unreachable")
		return nil
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_RangeContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	var all []int

	err := cro1.RangeContext(ctx, func(val int) error {
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
	cro2 := c2.ReadOnly()

	err = cro2.RangeContext(ctx, func(i int) error {
		cro2.Drain()
		return io.ErrUnexpectedEOF
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	err = cro3.RangeContext(ctx, func(i int) error {
		t.Errorf("should be unreachable")
		return nil
	})

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c4 := makeProducer(1)
	cro4 := c4.ReadOnly()

	cancel()

	err = cro4.RangeContext(ctx, func(i int) error {
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_Rest(t *testing.T) {
	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	all, err := cro1.Rest()
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
	cro2 := c2.ReadOnly()

	_, err = cro2.Rest()

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_RestContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	all, err := cro1.RestContext(ctx)
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
	cro2 := c2.ReadOnly()

	_, err = cro2.RestContext(ctx)

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c3 := makeProducer(1)
	cro3 := c3.ReadOnly()

	cancel()

	_, err = cro3.RestContext(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("incorrect error")
	}
}

func TestReadOnly_Intercept(t *testing.T) {
	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	var intercepted []int

	out := cro1.Intercept(func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c2 := makeProducer(2)
	cro2 := c2.ReadOnly()

	intercepted = nil

	out = cro2.Intercept(func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	intercepted = nil

	out = cro3.Intercept(func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}
}

func TestReadOnly_InterceptContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(2)
	cro1 := c1.ReadOnly()

	var intercepted []int

	out := cro1.InterceptContext(ctx, func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c2 := makeProducer(2)
	cro2 := c2.ReadOnly()

	intercepted = nil

	out = cro2.InterceptContext(ctx, func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	intercepted = nil

	out = cro3.InterceptContext(ctx, func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c4 := makeProducer(1)
	cro4 := c4.ReadOnly()

	intercepted = nil

	cancel()

	out = cro4.InterceptContext(ctx, func(i int) error {
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
	if _, valueFromChannel := <-out.ReadChannel(); valueFromChannel {
		t.Errorf("out channel should be closed")
	}
}

func TestReadOnly_Drain(t *testing.T) {
	c1 := makeProducer(10)
	cro1 := c1.ReadOnly()

	cro1.Drain()

	all, err := cro1.Rest()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-cro1.ReadChannel(); valueFromChannel {
		t.Errorf("channel should be closed")
	}
}

func TestReadOnly_DrainContext(t *testing.T) {
	ctx := context.Background()

	c1 := makeProducer(10)
	cro1 := c1.ReadOnly()

	cro1.DrainContext(ctx)

	all, err := cro1.RestContext(ctx)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 0 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-cro1.ReadChannel(); valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	ctx, cancel := context.WithCancel(context.Background())

	c2 := makeProducer(1)
	cro2 := c2.ReadOnly()

	cancel()

	cro2.DrainContext(ctx)

	// there should still be values in the channel, as the DrainContext execution didn't
	// proceed, due to the cancelled context
	all, err = cro2.Rest() // no context on purpose
	if err != nil {
		t.Errorf("unexpected error")
	}
	if len(all) != 1 {
		t.Errorf("incorrect len")
	}
	if _, valueFromChannel := <-cro2.ReadChannel(); valueFromChannel {
		t.Errorf("channel should be closed")
	}
}

func TestReadOnly_Len(t *testing.T) {
	c1 := New[int]()
	cro1 := c1.ReadOnly()

	if cro1.Len() != 0 {
		t.Errorf("incorrect len")
	}

	// ---

	c2 := NewWithSize[int](5)
	cro2 := c2.ReadOnly()

	if cro2.Len() != 0 {
		t.Errorf("incorrect len")
	}

	c2.Write(1)
	c2.Write(2)

	if cro2.Len() != 2 {
		t.Errorf("incorrect len")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	if cro3.Len() != 0 {
		t.Errorf("incorrect len")
	}
}

func TestReadOnly_Cap(t *testing.T) {
	c1 := New[int]()
	cro1 := c1.ReadOnly()

	if cro1.Cap() != 0 {
		t.Errorf("incorrect cap")
	}

	// ---

	c2 := NewWithSize[int](5)
	cro2 := c2.ReadOnly()

	if cro2.Cap() != 5 {
		t.Errorf("incorrect cap")
	}

	c2.Write(1)
	c2.Write(2)

	if cro2.Cap() != 5 {
		t.Errorf("incorrect cap")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	if cro3.Cap() != 0 {
		t.Errorf("incorrect cap")
	}
}

func TestReadOnly_Err(t *testing.T) {
	c1 := New[int]()
	cro1 := c1.ReadOnly()

	if cro1.Err() != nil {
		t.Errorf("unexpected error")
	}

	c1.Close()

	if !errors.Is(cro1.Err(), io.EOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c2 := New[int]()
	cro2 := c2.ReadOnly()

	c2.SetError(io.ErrUnexpectedEOF)

	if !errors.Is(cro2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := NewWithError[int](io.ErrUnexpectedEOF)
	cro3 := c3.ReadOnly()

	if !errors.Is(cro3.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
}
