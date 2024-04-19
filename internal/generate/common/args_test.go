// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"
)

func TestArgsEmptyString(t *testing.T) {
	t.Parallel()

	input := ``
	args := ParseArgs(input)

	if got, want := len(args.Positional), 0; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 0; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
}

func TestArgsSinglePositional(t *testing.T) {
	t.Parallel()

	input := `aws_instance`
	args := ParseArgs(input)

	if got, want := len(args.Positional), 1; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := args.Positional[0], "aws_instance"; got != want {
		t.Errorf("Positional[0] = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 0; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
}

func TestArgsSingleQuotedPositional(t *testing.T) {
	t.Parallel()

	input := `"aws_instance"`
	args := ParseArgs(input)

	if got, want := len(args.Positional), 1; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := args.Positional[0], "aws_instance"; got != want {
		t.Errorf("Positional[0] = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 0; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
}

func TestArgsSingleKeyword(t *testing.T) {
	t.Parallel()

	input := `vv=42`
	args := ParseArgs(input)

	if got, want := len(args.Positional), 0; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 1; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
	if got, want := args.Keyword["vv"], "42"; got != want {
		t.Errorf("Keyword[vv] = %v, want %v", got, want)
	}
}

func TestArgsMultipleKeywords(t *testing.T) {
	t.Parallel()

	input := `vv=42,type=aws_instance`
	args := ParseArgs(input)

	if got, want := len(args.Positional), 0; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 2; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
	if got, want := args.Keyword["vv"], "42"; got != want {
		t.Errorf("Keyword[vv] = %v, want %v", got, want)
	}
	if got, want := args.Keyword["type"], "aws_instance"; got != want {
		t.Errorf("Keyword[type] = %v, want %v", got, want)
	}
}

func TestArgsPostionalAndKeywords(t *testing.T) {
	t.Parallel()

	input := `first, vv=42 ,type=aws_instance,2`
	args := ParseArgs(input)

	if got, want := len(args.Positional), 2; got != want {
		t.Errorf("length of Positional = %v, want %v", got, want)
	}
	if got, want := args.Positional[0], "first"; got != want {
		t.Errorf("Positional[0] = %v, want %v", got, want)
	}
	if got, want := args.Positional[1], "2"; got != want {
		t.Errorf("Positional[1] = %v, want %v", got, want)
	}
	if got, want := len(args.Keyword), 2; got != want {
		t.Errorf("length of Keyword = %v, want %v", got, want)
	}
	if got, want := args.Keyword["vv"], "42"; got != want {
		t.Errorf("Keyword[vv] = %v, want %v", got, want)
	}
	if got, want := args.Keyword["type"], "aws_instance"; got != want {
		t.Errorf("Keyword[type] = %v, want %v", got, want)
	}
}
