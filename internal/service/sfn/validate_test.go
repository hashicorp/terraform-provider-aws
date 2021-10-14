package sfn

import (
	"strings"
	"testing"
)

func TestValidStateMachineName(t *testing.T) {
	validTypes := []string{
		"foo",
		"BAR",
		"FooBar123",
		"FooBar123Baz-_",
	}

	invalidTypes := []string{
		"foo bar",
		"foo<bar>",
		"foo{bar}",
		"foo[bar]",
		"foo*bar",
		"foo?bar",
		"foo#bar",
		"foo%bar",
		"foo\bar",
		"foo^bar",
		"foo|bar",
		"foo~bar",
		"foo$bar",
		"foo&bar",
		"foo,bar",
		"foo:bar",
		"foo;bar",
		"foo/bar",
		strings.Repeat("W", 81), // length > 80
	}

	for _, v := range validTypes {
		_, errors := validStateMachineName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Step Function State Machine name: %v", v, errors)
		}
	}

	for _, v := range invalidTypes {
		_, errors := validStateMachineName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Step Function State Machine name", v)
		}
	}
}
