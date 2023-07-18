// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package envvar

import (
	"os"
	"testing"

	testingiface "github.com/mitchellh/go-testing-interface"
)

func TestGetWithDefault(t *testing.T) { //nolint:paralleltest
	envVar := "TESTENVVAR_GETWITHDEFAULT"

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		want := "default"

		os.Unsetenv(envVar)

		got := GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("empty", func(t *testing.T) { //nolint:paralleltest
		want := "default"

		t.Setenv(envVar, "")

		got := GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("not empty", func(t *testing.T) { //nolint:paralleltest
		want := "notempty"

		t.Setenv(envVar, want)

		got := GetWithDefault(envVar, "default")

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})
}

func TestRequireOneOf(t *testing.T) { //nolint:paralleltest
	envVar1 := "TESTENVVAR_REQUIREONEOF1"
	envVar2 := "TESTENVVAR_REQUIREONEOF2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		_, _, err := RequireOneOf(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("all empty", func(t *testing.T) { //nolint:paralleltest
		t.Setenv(envVar1, "")
		t.Setenv(envVar2, "")

		_, _, err := RequireOneOf(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("some empty", func(t *testing.T) { //nolint:paralleltest
		wantValue := "pickme"

		t.Setenv(envVar1, "")
		t.Setenv(envVar2, wantValue)

		gotName, gotValue, err := RequireOneOf(envVars, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if gotName != envVar2 {
			t.Fatalf("expected name: %s, got: %s", envVar2, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})

	t.Run("all not empty", func(t *testing.T) { //nolint:paralleltest
		wantValue := "pickme"

		t.Setenv(envVar1, wantValue)
		t.Setenv(envVar2, "other")

		gotName, gotValue, err := RequireOneOf(envVars, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if gotName != envVar1 {
			t.Fatalf("expected name: %s, got: %s", envVar1, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})
}

func TestRequire(t *testing.T) { //nolint:paralleltest
	envVar := "TESTENVVAR_REQUIRE"

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		os.Unsetenv(envVar)

		_, err := Require(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty", func(t *testing.T) { //nolint:paralleltest
		t.Setenv(envVar, "")

		_, err := Require(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("not empty", func(t *testing.T) { //nolint:paralleltest
		want := "notempty"

		t.Setenv(envVar, want)

		got, err := Require(envVar, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func TestTestFailIfAllEmpty(t *testing.T) { //nolint:paralleltest
	envVar1 := "TESTENVVAR_FAILIFALLEMPTY1"
	envVar2 := "TESTENVVAR_FAILIFALLEMPTY2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		defer testingifaceRecover()

		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		FailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("all empty", func(t *testing.T) { //nolint:paralleltest
		defer testingifaceRecover()

		t.Setenv(envVar1, "")
		t.Setenv(envVar2, "")

		FailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("some empty", func(t *testing.T) { //nolint:paralleltest
		wantValue := "pickme"

		t.Setenv(envVar1, "")
		t.Setenv(envVar2, wantValue)

		gotName, gotValue := FailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar2 {
			t.Fatalf("expected name: %s, got: %s", envVar2, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})

	t.Run("all not empty", func(t *testing.T) { //nolint:paralleltest
		wantValue := "pickme"

		t.Setenv(envVar1, wantValue)
		t.Setenv(envVar2, "other")

		gotName, gotValue := FailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar1 {
			t.Fatalf("expected name: %s, got: %s", envVar1, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})
}

func TestTestFailIfEmpty(t *testing.T) { //nolint:paralleltest
	envVar := "TESTENVVAR_FAILIFEMPTY"

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		defer testingifaceRecover()

		os.Unsetenv(envVar)

		FailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("empty", func(t *testing.T) { //nolint:paralleltest
		defer testingifaceRecover()

		t.Setenv(envVar, "")

		FailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("not empty", func(t *testing.T) { //nolint:paralleltest
		want := "notempty"

		t.Setenv(envVar, want)

		got := FailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func TestTestSkipIfEmpty(t *testing.T) { //nolint:paralleltest
	envVar := "TESTENVVAR_SKIPIFEMPTY"

	t.Run("missing", func(t *testing.T) { //nolint:paralleltest
		mockT := &testingiface.RuntimeT{}

		os.Unsetenv(envVar)

		SkipIfEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("empty", func(t *testing.T) { //nolint:paralleltest
		mockT := &testingiface.RuntimeT{}

		t.Setenv(envVar, "")

		SkipIfEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("not empty", func(t *testing.T) { //nolint:paralleltest
		want := "notempty"

		t.Setenv(envVar, want)

		got := SkipIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func testingifaceRecover() {
	r := recover()

	// this string is hardcoded in github.com/mitchellh/go-testing-interface
	if s, ok := r.(string); !ok || s != "testing.T failed, see logs for output (if any)" {
		panic(r)
	}
}
