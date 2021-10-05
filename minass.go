// Package minass is a MINimal ASSertion library, a small library for writing
// assertions that you can memorize in an hour. It's intended to be used with
// Go's testing package directly.
//
// Once you start an assertion with Assert(t, value), you have your choice of
// methods to assert it to. All of those methods include an optional Printf
// style variadic arguments which are printed as a message on failure, and all
// of them return a boolean indicating if the assertion passed or failed.
//
// minass doesn't panic or anything fancy on failure, so if you want to return
// when a specific assertion fails, that logic is up to you.
package minass

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// assertion contains the basic properties common to value and functional
// assertions.
type assertion struct {
	t      testingT
	invert bool
	prefix string
}

// testingT describes the behavior we require from the default testing library.
type testingT interface {
	Helper()
	Errorf(format string, args ...interface{})
}

var runtimeCaller = runtime.Caller

// newAssertion initalizes the common proprties.
func newAssertion(t testingT) *assertion {
	t.Helper()
	_, caller, line, _ := runtimeCaller(2)
	return &assertion{
		prefix: fmt.Sprintf("[%s:%d]", caller, line),
		t:      t,
	}
}

// func (a *assertion) errorf(format string, args ...interface{}) {
// 	a.t.Helper()
// 	a.t.Errorf(a.prefix+format, args...)
// }

// Invert sets the invert value to true.
func (a *assertion) Invert() *assertion {
	a.invert = true
	return a
}

// message creates a message with the given test details and parts of the
// initial assertion.
func (a *assertion) message(vals ...interface{}) *message {
	var testmsg string
	var testargs []interface{}

	if len(vals) > 0 {
		if msg, ok := vals[0].(string); ok {
			testmsg = msg
		} else {
			panic("first parameter after expected value must be a string")
		}
		testargs = vals[1:]
	}

	return &message{
		t:               a.t,
		testMessage:     testmsg,
		testArgs:        testargs,
		assertionPrefix: a.prefix,
	}
}

// ValueAssertion stores the value and the common assertion for testing.
type ValueAssertion struct {
	*assertion
	val interface{}
}

// Assert creates a new assertion about the value v.
func Assert(t testingT, v interface{}) *ValueAssertion {
	t.Helper()
	return &ValueAssertion{
		assertion: newAssertion(t),
		val:       v,
	}
}

// Not inverts the assertion so that expected result is now the opposite.
func (a *ValueAssertion) Not() *ValueAssertion {
	a.Invert()
	return a
}

func reflectIsPointer(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}

func reflectIsNil(v interface{}) bool {
	return reflect.ValueOf(v).IsNil()
}

func reflectElem(v interface{}) interface{} {
	return reflect.ValueOf(v).Elem().Interface()
}

// Nil confirms the value is nil. A string or string format and parameters may
// be provided to print a message on failure.
func (a *ValueAssertion) Nil(msgs ...interface{}) (ret bool) {
	a.t.Helper()
	msg := a.message(msgs...)
	if !reflectIsPointer(a.val) {
		msg.errorf("value provided is not a pointer but is %T", a.val)
		return false
	}

	isNil := reflectIsNil(a.val)

	if !a.invert && !isNil {
		msg.multiLineErrorf("expected nil; got:\n%+v", reflectElem(a.val))
		return false
	}

	if a.invert && isNil {
		msg.errorf("value is nil; expected not nil")
		return false
	}

	return true
}

// True prints an error if the value is not a boolean or is false. A string or
// string format and parameters may be provided to print a message on failure.
func (a *ValueAssertion) True(msgs ...interface{}) (ret bool) {
	msg := a.message(msgs...)
	b, ok := a.val.(bool)
	if !ok {
		msg.errorf("value is not boolean; is %T", a.val)
		return false
	}

	if !a.invert && !b {
		msg.errorf("value is false; expected true")
		return false
	}

	if a.invert && b {
		msg.errorf("value is true; expected false")
		return false
	}

	return true
}

// False prints an error if the value is not a boolean or is true. A string or
// string format and parameters may be provided to print a message on failure.
func (a *ValueAssertion) False(msgs ...interface{}) bool {
	return a.Not().True(msgs...)
}

// Equal is a DSL method that calls Equals.
func (a *ValueAssertion) Equal(exp interface{}, msgs ...interface{}) bool {
	return a.Equals(exp, msgs...)
}

// Equals compares the initial value to exp using reflect.DeepEqual. A string or
// string format and parameters may be provided to print a message on failure.
func (a *ValueAssertion) Equals(exp interface{}, msgs ...interface{}) (ret bool) {
	a.t.Helper()
	msg := a.message(msgs...)
	eq := reflect.DeepEqual(a.val, exp)

	if !a.invert && !eq {
		msg.multiLineErrorf("%+v\n\n\tdoes not equal\n\n%+v", a.val, exp)
		return false
	}

	if a.invert && eq {
		msg.multiLineErrorf("both values are:\n%+v", exp)
		return false
	}

	return true
}

// reflectContains performs the reflection part of checking if the value
// provided contains the expected value.
func reflectContains(container reflect.Value, exp interface{}) bool {
	switch container.Kind() {
	case reflect.Ptr:
		containerValue := container.Elem()
		if !containerValue.IsValid() {
			return false
		}
		return reflectContains(containerValue, exp)
	case reflect.Slice, reflect.Array:

		for i := 0; i < container.Len(); i += 1 {
			if exp == container.Index(i).Interface() {
				return true
			}
		}
	case reflect.Map:
		rcon := reflect.ValueOf(container)
		for _, key := range rcon.MapKeys() {
			if exp == rcon.MapIndex(key).Interface() {
				return true
			}
		}
	}

	return false
}

// Contains checks the string, slice, map, or array to ensure it contains the
// expected value, failing the test if it doesn't. A string or string format and
// parameters may be provided to print a message on failure.
func (a *ValueAssertion) Contains(exp interface{}, vals ...interface{}) (ret bool) {
	a.t.Helper()
	msg := a.message(vals...)

	if reader, ok := a.val.(io.Reader); ok {
		if data, err := io.ReadAll(reader); err == nil {
			a.val = data
		}
	}

	castString := func(v interface{}) (string, bool) {
		if sval, ok := a.val.(string); ok {
			return sval, true
		} else if dval, ok := a.val.([]byte); ok {
			return string(dval), true
		}

		return "", false
	}

	var contains bool
	var stringCompared bool

	container := a.val
	if sexp, ok := exp.(string); ok {
		if sval, ok := castString(a.val); ok {
			stringCompared = true
			container = sval
			contains = strings.Contains(sval, sexp)
		}
	}

	if !stringCompared {
		contains = reflectContains(reflect.ValueOf(a.val), exp)
	}

	if !contains && !a.invert {
		msg.multiLineErrorf("%+v\n\n\tdoes not contain\n\n%+v", container, exp)
		return false
	}

	if contains && a.invert {
		msg.multiLineErrorf("%+v\n\n\tdoes contain\n\n%+v", container, exp)
		return false
	}

	return true
}

// Contain is an alias for Contains for better reading.
func (a *ValueAssertion) Contain(exp interface{}, vals ...interface{}) (ret bool) {
	return a.Contains(exp, vals...)
}

// HasKey checks a map to ensure it has the key provided, failing the test if it
// doesn't. A string or string format and parameters may be provided to print a
// message on failure.
func (a *ValueAssertion) HasKey(key interface{}, vals ...interface{}) bool {
	a.t.Helper()
	msg := a.message(vals...)
	hasKey, err := reflectHasKey(a.val, key)
	if err != nil {
		msg.errorf("hasKey error: %s", err.Error())
		return a.Not().False()
	}

	if !hasKey && !a.invert {
		msg.errorf("%+v\n\n\tdoes not have key\n\n%+v\n", a.val, key)
		return false
	}

	if hasKey && a.invert {
		msg.errorf("%+v\n\n\tdoes have key\n\n%+v", a.val, key)
		return false
	}

	return true
}

// HaveKey is an alias for HasKey for better reading.
func (a *ValueAssertion) HaveKey(key interface{}, vals ...interface{}) bool {
	return a.HasKey(key, vals...)
}

// reflectHasKey performs the reflection part of checking if the value
// provided contains the expected key.
func reflectHasKey(value, key interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	containerType := reflect.TypeOf(value)
	keyType := reflect.TypeOf(key)

	if containerType.Kind() != reflect.Map {
		return false, fmt.Errorf("value of type %T is not a map", value)
	}

	if containerType.Key() != keyType {
		return false, fmt.Errorf("map is keyed by type %T; key provided is type %T", containerType.Key(), keyType)
	}

	containerValue := reflect.ValueOf(value)
	keyValue := reflect.ValueOf(key)
	if containerValue.MapIndex(keyValue) != (reflect.Value{}) {
		return true, nil
	}

	return false, nil
}

// valueAssertion stores the value and the common assertion for testing.
type FunctionAssertion struct {
	*assertion
	fn func() bool
}

// AssertFn starts an assetion for function.
func AssertFn(t testingT, fn func()) *FunctionAssertion {
	t.Helper()

	wrap := func() bool {
		fn()
		return true
	}

	return &FunctionAssertion{
		assertion: newAssertion(t),
		fn:        wrap,
	}
}

// Not inverts the assertion so that expected result is now the opposite.
func (a *FunctionAssertion) Not() *FunctionAssertion {
	a.Invert()
	return a
}

// deferExpectedPanic deals with a panic happening or not happening after the
// asserted function runs.
func (a *FunctionAssertion) deferExpectedPanic(retbool *bool, msg *message) {
	r := recover()
	if r == nil && !a.invert {
		msg.errorf("did not panic")
		*retbool = false
	}

	if r != nil && a.invert {
		msg.errorf("code paniced with err: %s", r)
		*retbool = false
	}

	*retbool = true
}

// Panics is an alias for Panic for better reading.
func (a *FunctionAssertion) Panics(vals ...interface{}) (ret bool) {
	return a.Panic(vals...)
}

// Panic executes the asserted function and reports an error if a panic does not
// occur.
func (a *FunctionAssertion) Panic(vals ...interface{}) (ret bool) {
	msg := a.message(vals...)
	defer a.deferExpectedPanic(&ret, msg)
	return a.fn()
}

// Promise converts the function assertion to a promise.
func (a *FunctionAssertion) Promise() *promise {
	done := make(chan bool)
	go func() { done <- a.fn() }()
	return &promise{assertion: a.assertion, done: done}
}

// A promise is an assertion that contains an chan waiting for the result of the
// execution.
type promise struct {
	*assertion
	done chan bool
}

// Not inverts the assertion so that expected result is now the opposite.
func (p *promise) Not() *promise {
	p.Invert()
	return p
}

// Wait pulls from the channel causing the asserted function to trigger. This
// will always succeed.
func (p *promise) Wait() bool {
	return <-(chan bool)(p.done)
}

// Timeout waits for either the asserted function to finish or for the timer to
// trigger, and fails the test accordingly.
func (p promise) Timeout(d time.Duration, vals ...interface{}) (ret bool) {
	msg := p.message(vals...)
	select {
	case ret = <-(chan bool)(p.done):
		if p.assertion.invert {
			msg.errorf("function didn't meet the minimum duration of %s", d)
			ret = false
		}
	case <-time.After(d):
		if !p.assertion.invert {
			msg.errorf("function reached timeout of %s", d)
		} else {
			ret = true
		}
	}
	return
}

// message contains the information to print the assertion message, including
// the test message format and args, the assertions format and args, and a
// prefix defined by the the initial assertion call.
type message struct {
	t           testingT
	testMessage string
	testArgs    []interface{}

	assertionPrefix    string
	assertionMessage   string
	assertionArgs      []interface{}
	assertionMultiline bool
}

// errorf prints the message to the testing log.
func (m *message) errorf(format string, args ...interface{}) {
	m.assertionMessage = format
	m.assertionArgs = args
	m.t.Errorf(m.String())
}

// multiLineErrorf sets the message to multiline and prints the message to the
// testing log.
func (m *message) multiLineErrorf(format string, args ...interface{}) {
	m.assertionMultiline = true
	m.errorf(format, args...)
}

/*
String prints the message in one of these formats:
[file:line] assertion message

[file:line]
test message
assertion message

[file:line]
long
	assertion
message

[file:line]
test message
long
	assertion
message
*/
func (m message) String() string {
	b := &strings.Builder{}

	fmt.Fprint(b, m.assertionPrefix)

	// If there is no message and the assertion is a single line, print short
	// form.
	if m.testMessage == "" && !m.assertionMultiline {
		fmt.Fprint(b, " ")
		fmt.Fprintf(b, m.assertionMessage, m.assertionArgs...)
		return b.String()
	}

	if m.testMessage != "" {
		fmt.Fprintln(b)
		fmt.Fprintf(b, m.testMessage, m.testArgs...)
	}

	fmt.Fprintln(b)
	fmt.Fprintf(b, m.assertionMessage, m.assertionArgs...)
	return b.String()
}
