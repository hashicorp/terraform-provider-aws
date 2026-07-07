// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfschema_test

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

const (
	fixtureSchema   = "../../testdata/schema/aws_provider.json"
	fixtureProvider = "registry.terraform.io/hashicorp/aws"
)

// TestLoadFile_ReturnsProviderSchema confirms that LoadFile successfully reads
// and parses the fixture and returns a non-nil ProviderSchema with at least
// the expected resources.
func TestLoadFile_ReturnsProviderSchema(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps == nil {
		t.Fatal("expected non-nil ProviderSchema")
	}
	if len(ps.Resources) == 0 {
		t.Fatal("expected at least one resource in ProviderSchema")
	}
}

// TestLoadFile_ResourceNames confirms that ResourceNames returns a sorted,
// stable slice — important for deterministic report output.
func TestLoadFile_ResourceNames(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := ps.ResourceNames()
	want := []string{"aws_sns_topic", "aws_sqs_queue"}
	if !slices.Equal(names, want) {
		t.Errorf("ResourceNames() = %v, want %v", names, want)
	}
}

// TestLoadFile_PrimitiveFieldsExtracted verifies that top-level primitive
// attributes (string, number→int64, bool) are present in the resource IR.
func TestLoadFile_PrimitiveFieldsExtracted(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ir := ps.Resources["aws_sns_topic"]
	if ir == nil {
		t.Fatal("aws_sns_topic not found in Resources")
	}

	wantFields := []struct {
		name     string
		ft       tfschema.FieldType
		required bool
		optional bool
		computed bool
	}{
		{"id", tfschema.FieldTypeString, false, false, true},
		{"arn", tfschema.FieldTypeString, false, false, true},
		{"name", tfschema.FieldTypeString, false, true, true},
		{"display_name", tfschema.FieldTypeString, false, true, false},
		{"fifo_topic", tfschema.FieldTypeBool, false, true, true},
		{"kms_master_key_id", tfschema.FieldTypeString, false, true, false},
	}

	for _, wf := range wantFields {
		f, ok := ir.Fields[wf.name]
		if !ok {
			t.Errorf("field %q not found in aws_sns_topic", wf.name)
			continue
		}
		if f.Type != wf.ft {
			t.Errorf("field %q: Type = %q, want %q", wf.name, f.Type, wf.ft)
		}
		if f.Required != wf.required {
			t.Errorf("field %q: Required = %v, want %v", wf.name, f.Required, wf.required)
		}
		if f.Optional != wf.optional {
			t.Errorf("field %q: Optional = %v, want %v", wf.name, f.Optional, wf.optional)
		}
		if f.Computed != wf.computed {
			t.Errorf("field %q: Computed = %v, want %v", wf.name, f.Computed, wf.computed)
		}
	}
}

// TestLoadFile_NumberMapsToInt64 confirms the cty.Number → FieldTypeInt64
// mapping used in Phase 1.
func TestLoadFile_NumberMapsToInt64(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ir := ps.Resources["aws_sqs_queue"]
	if ir == nil {
		t.Fatal("aws_sqs_queue not found in Resources")
	}

	numFields := []string{
		"visibility_timeout_seconds",
		"message_retention_seconds",
		"delay_seconds",
		"max_message_size",
	}
	for _, name := range numFields {
		f, ok := ir.Fields[name]
		if !ok {
			t.Errorf("field %q not found in aws_sqs_queue", name)
			continue
		}
		if f.Type != tfschema.FieldTypeInt64 {
			t.Errorf("field %q: Type = %q, want FieldTypeInt64", name, f.Type)
		}
	}
}

// TestLoadFile_NonPrimitiveFieldsSkipped confirms that Phase 1 drops
// collection-type attributes (maps, lists) rather than including them as
// FieldTypeUnknown entries.
func TestLoadFile_NonPrimitiveFieldsSkipped(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ir := ps.Resources["aws_sns_topic"]
	if ir == nil {
		t.Fatal("aws_sns_topic not found in Resources")
	}

	// "tags" and "tags_all" are map(string) — non-primitive, should be absent.
	for _, name := range []string{"tags", "tags_all"} {
		if _, ok := ir.Fields[name]; ok {
			t.Errorf("non-primitive field %q should not be present in Phase 1 IR", name)
		}
	}
}

// TestLoadFile_SourceIsAlwaysTerraform confirms the Source field tag on every
// ResourceIR is set to "terraform".
func TestLoadFile_SourceIsAlwaysTerraform(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name, ir := range ps.Resources {
		if ir.Source != "terraform" {
			t.Errorf("resource %q: Source = %q, want \"terraform\"", name, ir.Source)
		}
	}
}

// TestLoadFile_ShortNameFallback confirms that when a schema JSON is keyed by
// a short provider name (e.g. just "aws" instead of the full registry
// address), LoadFile still finds it.
//
// The fallback in findProvider checks the last path-segment of the requested
// source against the map keys in the JSON.  This fixture is keyed by the full
// address, so passing "aws" alone does NOT match it — the test here pins that
// expected behaviour: a clear error is returned rather than a silent mismatch.
//
// To exercise the actual short-name match path, a fixture whose top-level key
// is the bare short name (e.g. "aws") would be needed.  That will be added in
// Phase 2 when real provider JSON files are integrated.
func TestLoadFile_ShortNameFallback_ReturnsErrorForFullAddressFixture(t *testing.T) {
	t.Parallel()

	// Passing the short name "aws" against a fixture keyed by the full address
	// "registry.terraform.io/hashicorp/aws" should return a not-found error
	// because neither key matches.
	_, err := tfschema.LoadFile(fixtureSchema, "aws")
	if err == nil {
		t.Fatal("expected an error when short name does not match any fixture key, got nil")
	}
}

// TestLoadFile_ShortNameFallback_MatchesWhenKeyedShort confirms that when the
// schema JSON itself is keyed by the short name (e.g. "aws"), LoadFile
// resolves it correctly when given the full registry address.
func TestLoadFile_ShortNameFallback_MatchesWhenKeyedShort(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(
		"../../testdata/schema/aws_provider_short_name.json",
		"registry.terraform.io/hashicorp/aws",
	)
	if err != nil {
		t.Fatalf("unexpected error with short-name keyed fixture: %v", err)
	}
	if len(ps.Resources) == 0 {
		t.Fatal("expected at least one resource in short-name-keyed fixture")
	}
}

// TestLoadFile_ProviderNotFound confirms a clear error when the provider
// source is not present in the schema JSON.
func TestLoadFile_ProviderNotFound(t *testing.T) {
	t.Parallel()

	_, err := tfschema.LoadFile(fixtureSchema, "registry.terraform.io/hashicorp/nonexistent")
	if err == nil {
		t.Fatal("expected an error for nonexistent provider, got nil")
	}
}

// TestLoadFile_MissingFile confirms a clear error when the JSON file does
// not exist.
func TestLoadFile_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := tfschema.LoadFile("/nonexistent/path/schema.json", fixtureProvider)
	if err == nil {
		t.Fatal("expected an error for missing file, got nil")
	}
}

// TestResourceIR_NameAndFieldNameConsistency checks that every Field.Name
// matches its map key — a basic invariant the comparison engine relies on.
func TestResourceIR_NameAndFieldNameConsistency(t *testing.T) {
	t.Parallel()

	ps, err := tfschema.LoadFile(fixtureSchema, fixtureProvider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for resourceName, ir := range ps.Resources {
		if ir.Name != resourceName {
			t.Errorf("resource map key %q != ir.Name %q", resourceName, ir.Name)
		}
		for key, field := range ir.Fields {
			if field.Name != key {
				t.Errorf("resource %q: field map key %q != field.Name %q",
					resourceName, key, field.Name)
			}
		}
	}
}
