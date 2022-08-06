package conns

import (
	"os"
	"testing"

	testingiface "github.com/mitchellh/go-testing-interface"
)

func TestGetWithDefault(t *testing.T) {
	envVar := "TESTENVVAR_GETWITHDEFAULT"

	t.Run("missing", func(t *testing.T) {
		want := "default"

		os.Unsetenv(envVar)

		got := GetEnvVarWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		want := "default"

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		got := GetEnvVarWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := GetEnvVarWithDefault(envVar, "default")

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})
}

func TestRequireOneOf(t *testing.T) {
	envVar1 := "TESTENVVAR_REQUIREONEOF1"
	envVar2 := "TESTENVVAR_REQUIREONEOF2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) {
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		_, _, err := RequireOneOfEnvVar(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("all empty", func(t *testing.T) {
		os.Setenv(envVar1, "")
		os.Setenv(envVar2, "")
		defer unsetEnvVars(envVars)

		_, _, err := RequireOneOfEnvVar(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("some empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, wantValue)
		defer unsetEnvVars(envVars)

		gotName, gotValue, err := RequireOneOfEnvVar(envVars, "usage")

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

	t.Run("all not empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, wantValue)
		os.Setenv(envVar2, "other")
		defer unsetEnvVars(envVars)

		gotName, gotValue, err := RequireOneOfEnvVar(envVars, "usage")

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

func TestRequire(t *testing.T) {
	envVar := "TESTENVVAR_REQUIRE"

	t.Run("missing", func(t *testing.T) {
		os.Unsetenv(envVar)

		_, err := RequireEnvVar(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty", func(t *testing.T) {
		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		_, err := RequireEnvVar(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got, err := RequireEnvVar(envVar, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func TestTestFailIfAllEmpty(t *testing.T) {
	envVar1 := "TESTENVVAR_FAILIFALLEMPTY1"
	envVar2 := "TESTENVVAR_FAILIFALLEMPTY2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) {
		defer testingifaceRecover()

		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		FailIfAllEnvVarEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("all empty", func(t *testing.T) {
		defer testingifaceRecover()

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, "")
		defer unsetEnvVars(envVars)

		FailIfAllEnvVarEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("some empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, wantValue)
		defer unsetEnvVars(envVars)

		gotName, gotValue := FailIfAllEnvVarEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar2 {
			t.Fatalf("expected name: %s, got: %s", envVar2, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})

	t.Run("all not empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, wantValue)
		os.Setenv(envVar2, "other")
		defer unsetEnvVars(envVars)

		gotName, gotValue := FailIfAllEnvVarEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar1 {
			t.Fatalf("expected name: %s, got: %s", envVar1, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})
}

func TestTestFailIfEmpty(t *testing.T) {
	envVar := "TESTENVVAR_FAILIFEMPTY"

	t.Run("missing", func(t *testing.T) {
		defer testingifaceRecover()

		os.Unsetenv(envVar)

		FailIfEnvVarEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("empty", func(t *testing.T) {
		defer testingifaceRecover()

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		FailIfEnvVarEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := FailIfEnvVarEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func TestTestSkipIfEmpty(t *testing.T) {
	envVar := "TESTENVVAR_SKIPIFEMPTY"

	t.Run("missing", func(t *testing.T) {
		mockT := &testingiface.RuntimeT{}

		os.Unsetenv(envVar)

		SkipIfEnvVarEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("empty", func(t *testing.T) {
		mockT := &testingiface.RuntimeT{}

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		SkipIfEnvVarEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := SkipIfEnvVarEmpty(&testingiface.RuntimeT{}, envVar, "usage")

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

func unsetEnvVars(envVars []string) {
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
