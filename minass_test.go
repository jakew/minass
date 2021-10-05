package minass

import (
	"fmt"
	"time"
)

func init() {
	// Replace the default runtime.Caller with a deterministic version.
	runtimeCaller = func(skip int) (pc uintptr, file string, line int, ok bool) {
		file = "/full/path/to/test/file_test.go"
		line = 123
		return
	}
}

type exampleTestingT struct{}

func (t *exampleTestingT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (exampleTestingT) Helper() {}

var t = &exampleTestingT{}

var wantToBeNotNil *struct{} = nil

func ExampleValueAssertion_Nil() {
	val := "value"
	// Example with no message.
	Assert(t, "value").Nil()

	// Example with no message.
	Assert(t, wantToBeNotNil).Not().Nil()

	// Example with a message.
	Assert(t, &val).Nil("expecting this to not have a value")

	// Example with a message.
	Assert(t, wantToBeNotNil).Not().Nil("expected this to have a value")

	// Output:
	// [/full/path/to/test/file_test.go:123] value provided is not a pointer but is string
	// [/full/path/to/test/file_test.go:123] value is nil; expected not nil
	// [/full/path/to/test/file_test.go:123]
	// expecting this to not have a value
	// expected nil; got:
	// value
	// [/full/path/to/test/file_test.go:123]
	// expected this to have a value
	// value is nil; expected not nil
}

var (
	wantToBeTrue  = false
	wantToBeFalse = true
)

func ExampleValueAssertion_True_NoMessage() {
	// Example with no message.
	Assert(t, wantToBeTrue).True()

	// Example with no message.
	Assert(t, wantToBeFalse).False()

	// Same as .False() but less terse.
	Assert(t, wantToBeFalse).Not().True()

	// Example simple message.
	Assert(t, wantToBeFalse).False("we wanted a false here")

	// Example failing test.
	Assert(t, wantToBeTrue).True("the first value should be true; it was %t", wantToBeTrue)

	// Example failing test.
	Assert(t, wantToBeFalse).True("the second value should be false; it was %t", wantToBeFalse)

	// Example failing with early return.
	if !Assert(t, wantToBeTrue).True("the third value should be true; it was %t", wantToBeTrue) {
		return
	}

	// This never executes.
	Assert(t, wantToBeTrue).True()

	// Output:
	// [/full/path/to/test/file_test.go:123] value is false; expected true
	// [/full/path/to/test/file_test.go:123] value is true; expected false
	// [/full/path/to/test/file_test.go:123] value is true; expected false
	// [/full/path/to/test/file_test.go:123]
	// we wanted a false here
	// value is true; expected false
	// [/full/path/to/test/file_test.go:123]
	// the first value should be true; it was false
	// value is false; expected true
	// [/full/path/to/test/file_test.go:123]
	// the third value should be true; it was false
	// value is false; expected true
}

func ExampleValueAssertion_True_BadExamples() {
	// Example wrong value type.
	Assert(t, "string value").True()

	// Output:
	// [/full/path/to/test/file_test.go:123] value is not boolean; is string
}

var wantedText, gotText = "wanted", "got"

func ExampleValueAssertion_Equals() {
	Assert(t, gotText).Equals(wantedText, "we want \"%s\", not \"%s\"", wantedText, gotText)

	Assert(t, false).Equals(wantedText)

	Assert(t, []string{gotText}).Equals(wantedText)

	Assert(t, wantedText).Not().Equal(wantedText)

	// Output:
	// [/full/path/to/test/file_test.go:123]
	// we want "wanted", not "got"
	// got
	//
	// 	does not equal
	//
	// wanted
	// [/full/path/to/test/file_test.go:123]
	// false
	//
	// 	does not equal
	//
	// wanted
	// [/full/path/to/test/file_test.go:123]
	// [got]
	//
	// 	does not equal
	//
	// wanted
	// [/full/path/to/test/file_test.go:123]
	// both values are:
	// wanted
}

var gotSlice, wantedSlice = []string{gotText}, []string{wantedText}

func ExampleValueAssertion_Contains() {
	Assert(t, gotSlice).Contains(wantedText, "This should have the wanted text.")

	Assert(t, wantedSlice).Contain(wantedText)

	Assert(t, wantedSlice).Not().Contain(wantedText)

	// Output:
	// [/full/path/to/test/file_test.go:123]
	// This should have the wanted text.
	// [got]
	//
	// 	does not contain
	//
	// wanted
	// [/full/path/to/test/file_test.go:123]
	// [wanted]
	//
	// 	does contain
	//
	// wanted
}

var wantedKey = "wantedKey"
var gotMap = map[string]string{"gotKey": gotText}
var wantedMap = map[string]string{"wantedKey": wantedText}

func ExampleValueAssertion_HasKey() {
	if Assert(t, wantedMap).HasKey(wantedKey) {
		Assert(t, wantedMap[wantedKey]).Contain(wantedText)
	}

	Assert(t, wantedMap).Not().HaveKey(wantedKey, "We don't want this key here for reasons.")

	if !Assert(t, gotMap).HasKey(wantedKey, "The map doesn't have the key.") {
		return
	}

	Assert(t, gotMap[wantedKey]).Contain(wantedText)

	// Output:
	// [/full/path/to/test/file_test.go:123]
	// We don't want this key here for reasons.
	// map[wantedKey:wanted]
	//
	// 	does have key
	//
	// wantedKey
	// [/full/path/to/test/file_test.go:123]
	// The map doesn't have the key.
	// map[gotKey:got]
	//
	// 	does not have key
	//
	// wantedKey
}

func ExampleFunctionAssertion_Panics() {
	AssertFn(t, func() {
		panic("at the disco")
	}).Not().Panic()

	AssertFn(t, func() {
		// Keep Calm
	}).Panics()

	// Output:
	// [/full/path/to/test/file_test.go:123] code paniced with err: at the disco
	// [/full/path/to/test/file_test.go:123] did not panic
}

func ExampleFunctionAssertion_Promise() {
	p := AssertFn(t, func() {
		fmt.Println("executing")
	}).Promise()

	fmt.Println("...")

	p.Wait()
	fmt.Println("waited")

	AssertFn(t, func() {
		<-time.After(time.Millisecond / 2)
	}).Promise().Timeout(time.Millisecond / 3)

	AssertFn(t, func() {
		<-time.After(time.Millisecond / 2)
	}).Promise().Timeout(time.Millisecond, "waiting for the func to finish.")

	// Output:
	// ...
	// executing
	// waited
	// [/full/path/to/test/file_test.go:123] function reached timeout of 333.333µs
}
