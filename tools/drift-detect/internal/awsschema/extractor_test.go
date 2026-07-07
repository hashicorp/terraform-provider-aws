// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsschema_test

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsschema"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

// Paths relative to the package directory (internal/awsschema).
const (
	mappingFile       = "../../mappings/aws_resources.yaml"
	fixtureModelsRoot = "../../testdata/smithy"
)

func apiModelsBaseURL(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.FileServer(http.Dir(fixtureModelsRoot)))
	t.Cleanup(server.Close)

	return server.URL
}

// loadMapping is a test helper that loads the mapping file and returns the
// ResourceMapping for the given TF resource name.
func loadMapping(t *testing.T, tfName string) *awsmapping.ResourceMapping {
	t.Helper()
	f, err := awsmapping.LoadFile(mappingFile)
	if err != nil {
		t.Fatalf("LoadFile(%q): %v", mappingFile, err)
	}
	m, ok := f.Resources[tfName]
	if !ok {
		t.Fatalf("resource %q not found in mapping file", tfName)
	}
	return m
}

// fieldNames returns sorted field names from a ResourceIR for stable assertions.
func fieldNames(ir *tfschema.ResourceIR) []string {
	names := make([]string, 0, len(ir.Fields))
	for k := range ir.Fields {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// printIR is a test helper that prints the full IR for visual inspection.
func printIR(t *testing.T, ir *tfschema.ResourceIR) {
	t.Helper()
	t.Logf("=== %s (source: %s) — %d fields ===", ir.Name, ir.Source, len(ir.Fields))
	names := fieldNames(ir)
	for _, n := range names {
		f := ir.Fields[n]
		flags := ""
		if f.Required {
			flags += " required"
		}
		if f.Optional {
			flags += " optional"
		}
		if f.Computed {
			flags += " computed"
		}
		t.Logf("  %-45s %-8s%s", f.Name, f.Type, flags)
	}
}

// ---------------------------------------------------------------------------
// Pattern B — aws_sqs_queue (first priority per the plan)
// ---------------------------------------------------------------------------

func TestExtract_SQS_PatternB(t *testing.T) {
	t.Parallel()

	m := loadMapping(t, "aws_sqs_queue")
	ir, err := awsschema.ExtractResource("aws_sqs_queue", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	printIR(t, ir)

	// Basic invariants
	if ir.Source != "aws" {
		t.Errorf("Source = %q, want %q", ir.Source, "aws")
	}
	if len(ir.Fields) == 0 {
		t.Fatal("no fields extracted")
	}

	// Fields that must be present (from QueueAttributeName enum + renames)
	wantFields := []string{
		"visibility_timeout_seconds",
		"delay_seconds",
		"message_retention_seconds",
		"max_message_size",
		"fifo_queue",
		"kms_master_key_id",
		"arn",
		"policy",
		"redrive_policy",
		"sqs_managed_sse_enabled",
	}
	for _, want := range wantFields {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing from aws_sqs_queue IR", want)
		}
	}

	// Fields that must NOT be present (suppressed)
	suppressedFields := []string{
		"approximate_number_of_messages", // read-only metrics
		"approximate_number_of_messages_not_visible",
		"approximate_number_of_messages_delayed",
		"created_timestamp",
		"last_modified_timestamp",
		"all", // "All" enum value
	}
	for _, bad := range suppressedFields {
		if _, ok := ir.Fields[bad]; ok {
			t.Errorf("field %q should be suppressed but is present", bad)
		}
	}

	// Enum-derived fields should remain optional by default.
	for name, f := range ir.Fields {
		if f.Required {
			t.Errorf("field %q: Required = true; enum-derived fields should be optional", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Pattern C — aws_sns_topic
// ---------------------------------------------------------------------------

func TestExtract_SNS_PatternC(t *testing.T) {
	t.Parallel()

	m := loadMapping(t, "aws_sns_topic")
	ir, err := awsschema.ExtractResource("aws_sns_topic", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	printIR(t, ir)

	if ir.Source != "aws" {
		t.Errorf("Source = %q, want %q", ir.Source, "aws")
	}
	if len(ir.Fields) == 0 {
		t.Fatal("no fields extracted")
	}

	// All fields from explicit_fields in the mapping must appear
	wantFields := []string{
		"display_name",
		"kms_master_key_id",
		"fifo_topic",
		"content_based_deduplication",
		"policy",
		"delivery_policy",
		"tracing_config",
		"signature_version",
		"archive_policy",
	}
	for _, want := range wantFields {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing from aws_sns_topic IR", want)
		}
	}

	// Boolean fields should have FieldTypeBool
	for _, boolField := range []string{"fifo_topic", "content_based_deduplication"} {
		if f, ok := ir.Fields[boolField]; ok {
			if f.Type != tfschema.FieldTypeBool {
				t.Errorf("field %q: Type = %q, want bool", boolField, f.Type)
			}
		}
	}

	// String fields should have FieldTypeString
	for _, strField := range []string{"display_name", "policy", "tracing_config"} {
		if f, ok := ir.Fields[strField]; ok {
			if f.Type != tfschema.FieldTypeString {
				t.Errorf("field %q: Type = %q, want string", strField, f.Type)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Pattern A — aws_prometheus_workspace (AMP)
// ---------------------------------------------------------------------------

func TestExtract_AMP_PatternA(t *testing.T) {
	t.Parallel()

	m := loadMapping(t, "aws_prometheus_workspace")
	ir, err := awsschema.ExtractResource("aws_prometheus_workspace", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	printIR(t, ir)

	if ir.Source != "aws" {
		t.Errorf("Source = %q, want %q", ir.Source, "aws")
	}
	if len(ir.Fields) == 0 {
		t.Fatal("no fields extracted")
	}

	// alias and kms_key_arn come from CreateWorkspaceRequest
	wantFields := []string{"alias", "kms_key_arn"}
	for _, want := range wantFields {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing from aws_prometheus_workspace IR", want)
		}
	}

	// Suppressed fields must NOT appear
	suppressedFields := []string{
		"client_token", // idempotency token
		"workspace_id", // URL identifier
		"workspace",    // wrapper object in read response
	}
	for _, bad := range suppressedFields {
		if _, ok := ir.Fields[bad]; ok {
			t.Errorf("field %q should be suppressed but is present", bad)
		}
	}
}

func TestExtract_AMPWorkspace_InferLifecycleFromSmithyResource(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{
		SmithyModel:     "models/amp/service/2020-08-01/amp-2020-08-01.json",
		SmithyNamespace: "com.amazonaws.amp",
		SmithyResource:  "Workspace",
		SuppressFields:  []string{"clientToken", "workspaceId", "status", "workspace"},
		FieldRenames: map[string]string{
			"kmsKeyArn": "kms_key_arn",
			"arn":       "arn",
		},
	}

	ir, err := awsschema.ExtractResource("aws_prometheus_workspace", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	for _, want := range []string{"alias", "kms_key_arn"} {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing", want)
		}
	}

	if ir.Metadata == nil || len(ir.Metadata.Identifiers) == 0 {
		t.Fatal("identifier metadata missing")
	}
	id, ok := ir.Metadata.Identifiers["workspace_id"]
	if !ok {
		t.Fatal("workspace_id identifier metadata missing")
	}
	if id.Type != tfschema.FieldTypeString {
		t.Errorf("workspace_id identifier type = %q, want string", id.Type)
	}
}

func TestExtract_AMPResourcePolicy_InferLifecycleFromSmithyResource(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{
		SmithyModel:     "models/amp/service/2020-08-01/amp-2020-08-01.json",
		SmithyNamespace: "com.amazonaws.amp",
		SmithyResource:  "WorkspaceResourcePolicy",
		SuppressFields:  []string{"workspaceId"},
		FieldRenames: map[string]string{
			"policyDocument": "policy_document",
			"revisionId":     "revision_id",
		},
	}

	ir, err := awsschema.ExtractResource("aws_prometheus_resource_policy", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	for _, want := range []string{"policy_document", "revision_id"} {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing", want)
		}
	}

	if _, ok := ir.Fields["workspace_id"]; ok {
		t.Error("workspace_id should be suppressed in fields")
	}

	if ir.Metadata == nil || len(ir.Metadata.Identifiers) == 0 {
		t.Fatal("identifier metadata missing")
	}
	if _, ok := ir.Metadata.Identifiers["workspace_id"]; !ok {
		t.Fatal("workspace_id identifier metadata missing")
	}
}

func TestExtract_NoExtractionConfig_ReturnsError(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{}

	_, err := awsschema.ExtractResource("aws_example", m, apiModelsBaseURL(t))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no extraction configuration") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// General invariants across all resources
// ---------------------------------------------------------------------------

func TestExtract_AllResources_SourceIsAWS(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile(mappingFile)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	for tfName, m := range f.Resources {
		tfName, m := tfName, m
		t.Run(tfName, func(t *testing.T) {
			t.Parallel()
			ir, err := awsschema.ExtractResource(tfName, m, apiModelsBaseURL(t))
			if err != nil {
				t.Fatalf("ExtractResource(%q): %v", tfName, err)
			}
			if ir.Source != "aws" {
				t.Errorf("Source = %q, want aws", ir.Source)
			}
			if ir.Name != tfName {
				t.Errorf("Name = %q, want %q", ir.Name, tfName)
			}
		})
	}
}

// TestExtract_FieldNameConsistency verifies that every Field.Name matches
// its map key — an invariant the comparison engine relies on.
func TestExtract_FieldNameConsistency(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile(mappingFile)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	for tfName, m := range f.Resources {
		tfName, m := tfName, m
		t.Run(tfName, func(t *testing.T) {
			t.Parallel()
			ir, err := awsschema.ExtractResource(tfName, m, apiModelsBaseURL(t))
			if err != nil {
				t.Fatalf("ExtractResource: %v", err)
			}
			for key, field := range ir.Fields {
				if field.Name != key {
					t.Errorf("map key %q != field.Name %q", key, field.Name)
				}
			}
		})
	}
}

// TestExtract_SQS_PrintSortedFields prints all SQS fields sorted for visual
// comparison against the TF schema output. Use go test -v to see this.
func TestExtract_SQS_PrintSortedFields(t *testing.T) {
	m := loadMapping(t, "aws_sqs_queue")
	ir, err := awsschema.ExtractResource("aws_sqs_queue", m, apiModelsBaseURL(t))
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	names := fieldNames(ir)
	t.Log("\naws_sqs_queue AWS IR fields:")
	for _, n := range names {
		f := ir.Fields[n]
		t.Logf("  %-45s %s", n, f.Type)
	}
	t.Logf("Total: %d fields", len(names))
}
