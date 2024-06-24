package channel

import "fmt"

func ExampleC() {
	// Create a buffered channel
	c := NewWithSize[int](3)

	// Get the read-only and write-only variant. This allows to enforce the restriction in your typed interfaces.
	wo := c.WriteOnly()
	ro := c.ReadOnly()

	// Now write some values
	wo.Write(1)
	wo.Write(2)
	wo.Write(3)

	// ... and read them from the other side. We'll also return an error for later on the last value.
	err := ro.Range(func(i int) error {
		fmt.Println(i)
		if i == 3 {
			return fmt.Errorf("I don't like 3")
		}
		return nil
	})

	// We got an error while reading values:
	fmt.Println(err)

	// Major improvement compared to raw channel, that error is sticky, so all consumers of the channel can
	// be aware of it by checking the sticky error:
	err = ro.Err()

	// Output: 1
	// 2
	// 3
	// I don't like 3
}

func ExampleC_Intercept() {
	c := NewWithSize[int](3)

	wrapped := c.Intercept(func(i int) error {
		fmt.Println(i)
		return nil
	})

	c.Write(1)
	c.Write(2)
	c.Write(3)

	_, _ = wrapped.Read()
	_, _ = wrapped.Read()
	_, _ = wrapped.Read()

	// Output: 1
	// 2
	// 3
}
