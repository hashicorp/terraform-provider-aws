package envvar_test

import (
	"os"
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

func TestGetWithDefault(t *testing.T) {
	envVar := "TESTENVVAR_GETWITHDEFAULT"

	t.Run("missing", func(t *testing.T) {
		want := "default"

		os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		want := "default"

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, "default")

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})
}
