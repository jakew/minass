# Minass

### About
Minass is a minimal assertion library, to be used with the standard Go testing
library.

Goals of this project:

- Be capable of the majority of assertions needed.
- Complements the Go testing library.
- Depends **only** on the Go standard library.
- Easy to use and easy to memorize interface using a minimal DSL.

### Usage

Assertions have two constructors: 

Using `Assert(t, val)` can be used to check the value val that is provided, and
`Assert(t, func() { ...})` can be used to make assertions about runtime
properties of a function such as duration and panics. Both require the
*testing.T value passed into the function.

To check that a value is not nil, you can do this:
```go
func TestingSomeValue(t *testing.T) {
    got, err := testedFunc()
    if !Assert(t, err).Nil("expected no error, got %s", err) {
        return
    }

    Assert(t, got).Not().Nil("value returned from testedFunc was not valid")
}
```

This will first assert that the `err` value is nil. If the value is not nil, the
assertion will print an error message using `t.Errorf(...)`, failing the test.
If the assertion fails, the assertion returns `false`, so the if statement is
triggered and the test returns. If `err` _is_ nil, then the second assertion
occurs, checking that `got` is _not_ nil. If `got` is nil, an error message is
printed using `t.Errorf(...)` and the test fails.

When an assertion fails, an error message is logged. The assertion will
automatically add in the file and line where the assertion is along with the
message from the condition that failed. All of the conditions available,
including the `.Nil(...)` function we just saw, allow you to add in a message to
broadcast with the failure. The parameters for this message are always the last
ones and must be in the format of `message string, args ...interface{}`, just
like fmt.Printf.

The `.Not()` method inverts the assertion, and can be used with all conditions
as well.

#### Conditions

Conditions for value assertion `Assert(t, value)`:
- `.Nil` checks if the value is nil.
- `.Equals` and `.Equal` assert that the initial value is the same as the new
  value.
- `.True` and `.False` assert that the initial value is both binary and the
  respective truthyness. `.False` is just an alias for `.Not().True()`.
- `.Equals` and `.Equal` assert that the value is correct. This uses
  reflect.DeepEqual so values are compared, not just pointers.
- `.Contains` and `Contain` use reflect to loop through the contents of a
  string, map, slice, or array to asserts that the value provided in the
  condition exists as one of the values.
- `.HasKey` is specific to maps and asserts that the map has the key provided.

Conditions for function assertion `AssertFn(t, func() {...})`:
- `.Panics` and `.Panic` assert that the function executed panics.

## Contributing
If you find a bug or want a new assertion type or condition, feel free to make a
ticket. If you want to help out, check out the tickets that are looking for
help.