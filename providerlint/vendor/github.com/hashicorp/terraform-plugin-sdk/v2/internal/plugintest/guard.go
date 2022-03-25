package plugintest

import (
	"fmt"
)

// TestControl is an interface requiring a subset of *testing.T which is used
// by the test guards and helpers in this package. Most callers can simply
// pass their *testing.T value here, but the interface allows other
// implementations to potentially be provided instead, for example to allow
// meta-testing (testing of the test utilities themselves).
//
// This interface also describes the subset of normal test functionality the
// guards and helpers can perform: they can only create log lines, fail tests,
// and skip tests. All other test control is the responsibility of the main
// test code.
type TestControl interface {
	Helper()
	Log(args ...interface{})
	FailNow()
	SkipNow()
}

// testingT wraps a TestControl to recover some of the convenience behaviors
// that would normally come from a real *testing.T, so we can keep TestControl
// small while still having these conveniences. This is an abstraction
// inversion, but accepted because it makes the public API more convenient
// without any considerable disadvantage.
type testingT struct {
	TestControl
}

func (t testingT) Logf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
}

func (t testingT) Fatalf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
	t.FailNow()
}

func (t testingT) Skipf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
	t.SkipNow()
}
