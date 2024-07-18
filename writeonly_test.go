package channel

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestWriteOnly_Write(t *testing.T) {
	c := New[int]()
	cwo := c.WriteOnly()

	go func() {
		cwo.Write(1)
		cwo.Write(2)
		cwo.Close()
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

func TestWriteOnly_WriteContext(t *testing.T) {
	ctx := context.Background()

	c1 := New[int]()
	cwo1 := c1.WriteOnly()

	go func() {
		err := cwo1.WriteContext(ctx, 1)
		if err != nil {
			t.Errorf("unexpected error")
		}
		err = cwo1.WriteContext(ctx, 2)
		if err != nil {
			t.Errorf("unexpected error")
		}
		cwo1.Close()
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
	cwo2 := c2.WriteOnly()

	err = cwo2.WriteContext(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error")
	}

	if c2.Len() != 0 {
		t.Errorf("channel should be empty")
	}

	cwo2.Close()
}

func TestWriteOnly_WriteChannel(t *testing.T) {
	c := NewWithSize[int](1)
	cwo := c.WriteOnly()

	cwo.WriteChannel() <- 1

	v, err := c.Read()
	if err != nil {
		t.Errorf("unexpected error")
	}
	if v != 1 {
		t.Errorf("unexpected value")
	}

	select {
	case cwo.WriteChannel() <- 2:
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

func TestWriteOnly_SetError(t *testing.T) {
	c1 := New[int]()
	cwo1 := c1.WriteOnly()

	if c1.Err() != nil {
		t.Errorf("unexpected error")
	}

	cwo1.SetError(nil)

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}

	// ---

	c2 := New[int]()
	cwo2 := c2.WriteOnly()

	cwo2.SetError(io.ErrUnexpectedEOF)

	if !errors.Is(c2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}

	// ---

	c3 := New[int]()
	cwo3 := c3.WriteOnly()

	cwo3.SetError(io.ErrUnexpectedEOF)

	shouldPanic(t, func() {
		cwo3.SetError(io.ErrClosedPipe)
	})
}

func TestWriteOnly_Close(t *testing.T) {
	c1 := New[int]()
	cwo1 := c1.WriteOnly()

	cwo1.Close()

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c2 := New[int]()
	cwo2 := c2.WriteOnly()

	cwo2.Close()

	shouldPanic(t, func() {
		cwo2.Close()
	})
}

func TestWriteOnly_CloseWithError(t *testing.T) {
	c1 := New[int]()
	cwo1 := c1.WriteOnly()

	cwo1.CloseWithError(nil)

	if c1.Err() != io.EOF {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c1.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c2 := New[int]()
	cwo2 := c2.WriteOnly()

	cwo2.CloseWithError(io.ErrUnexpectedEOF)

	if !errors.Is(c2.Err(), io.ErrUnexpectedEOF) {
		t.Errorf("incorrect error")
	}
	if _, valueFromChannel := <-c2.c; valueFromChannel {
		t.Errorf("channel should be closed")
	}

	// ---

	c3 := New[int]()
	cwo3 := c3.WriteOnly()

	cwo3.Close()

	shouldPanic(t, func() {
		cwo3.CloseWithError(io.ErrUnexpectedEOF)
	})
}
