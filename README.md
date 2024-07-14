# 🗲Channel, a channel wrapper

This package is a wrapper around the normal golang's channel, with several improvements:
- provides higher level functions, easier to use correctly
- built-in support for go context for easy termination
- provides explicit ReadOnly and WriteOnly variations to document and enforce behaviors
- carry a sticky error that can be read and acted upon by multiple readers, which normal golang channels do not support
- define a normalized way to carry an error in a channel, without having to create an explicit struct such as `type ValWithError struct { Val int, Err error }`*

`*` This is especially important because one of the core idea of golang is implicit implementation of interfaces. However, when you create such wrapper struct, you force someone creating an alternative implementation of your interface to import your package, only for that explicit struct. This can create dependency problems and is just messy.

**This package has zero dependencies.**

## Import

`go get -u github.com/MichaelMure/channel`

## Usage

### Create a channel

```go
c1 := channel.New[int]()                             // unbuffered channel
c2 := channel.NewWithSize[int](3)                    // buffered channel
c3 := channel.NewWithError[int](io.ErrUnexpectedEOF) // already closed and errored
```

### ReadOnly and WriteOnly variations

```go
c := channel.New[int]()
ro := c.ReadOnly()  // only allow reads
wo := c.WriteOnly() // only allow writes
```

### Writes

```go
c.Write(1)
err := c.WriteContext(ctx, 2)
```

### Reads

```go
// Simple reads
val, err := c.Read()
val, err = c.ReadContext(ctx)

// Iteration
err = c.Range(func(i int) error {
	// do something with i
	return nil
})
err = c.RangeContext(ctx, func(i int) error {
	// do something with i
	return nil
})

// Collect all values
all, err = c.Rest()
all, err := c.RestContext(ctx)
```

### Intercept values in a wrapper

```go
out := c.Intercept(func(i int) error {
	// do something with i
})
return out
```

### Close and error

```go
c.SetError(io.ErrUnexpectedEOF)       // explicitly set the error
c.Close()                             // simple close
c.CloseWithError(io.ErrUnexpectedEOF) // close with an error
```

## Note

Original idea and implementation by Jorropo.

## License

This project is licensed under a dual MIT/Apache-2.0 license.

MIT: https://www.opensource.org/licenses/mit
Apache-2.0: https://www.apache.org/licenses/LICENSE-2.0