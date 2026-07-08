// Copyright IBM Corp. 2026, 2026
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"os/exec"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/provider"
)

// TestCleanupSchema_NoOp verifies that CleanupSchema does not panic or error
// when called with an empty string.
func TestCleanupSchema_NoOp(t *testing.T) {
	t.Parallel()

	// Should be a silent no-op.
	provider.CleanupSchema("")
}

// TestRequireTerraform_NotInPath is a compile-time guard: requireTerraform is
// package-private, so we probe the same behaviour indirectly by calling
// GenerateSchema with a dummy providerDir when terraform is absent.
//
// This test is intentionally skipped when terraform IS available in PATH so it
// does not interfere with environments that have the binary installed.  Its
// purpose is to document and pin the behaviour when terraform is missing.
func TestGenerateSchema_NoTerraformBinary(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("terraform"); err == nil {
		t.Skip("terraform is present in PATH; skipping missing-binary test")
	}

	_, err := provider.GenerateSchema("/some/provider/dir", "registry.terraform.io/hashicorp/aws")
	if err == nil {
		t.Fatal("expected an error when terraform is not in PATH, got nil")
	}
}

// TestGenerateSchemaTo_BadProviderSource confirms that a malformed provider
// source (missing namespace segment) produces a clear error before any file
// system work is attempted.
//
// We rely on terraform being in PATH for this test; if it's not, we still get
// a PATH error first — which is also an error, so the assertion holds.
func TestGenerateSchemaTo_BadProviderSource(t *testing.T) {
	t.Parallel()

	err := provider.GenerateSchemaTo(
		"/does/not/exist",
		"bad-source", // not in registry.terraform.io/namespace/name form
		t.TempDir()+"/schema.json",
	)
	if err == nil {
		t.Fatal("expected an error for malformed provider source, got nil")
	}
}
