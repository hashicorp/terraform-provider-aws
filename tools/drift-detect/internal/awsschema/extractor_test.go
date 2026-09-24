// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package awsschema_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

// fixtureServiceRenames maps TF service name segments to their fixture model
// directory names when the two differ (e.g. "prometheus" → "amp").
var fixtureServiceRenames = map[string]string{
	"prometheus": "amp",
}

func resolveFixtureModelAndNamespace(tfName string) (string, string, error) {
	parts := strings.SplitN(tfName, "_", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid terraform resource name %q", tfName)
	}

	tfService := parts[1]
	awsService := tfService
	if renamed, ok := fixtureServiceRenames[tfService]; ok {
		awsService = renamed
	}

	globPattern := filepath.Join(
		fixtureModelsRoot,
		"models",
		awsService,
		"service",
		"*",
		awsService+"-*.json",
	)
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return "", "", fmt.Errorf("glob fixture model path: %w", err)
	}
	if len(matches) == 0 {
		return "", "", fmt.Errorf("no fixture model found for service %q", awsService)
	}
	sort.Strings(matches)
	modelFile := matches[len(matches)-1]

	modelPath, err := filepath.Rel(fixtureModelsRoot, modelFile)
	if err != nil {
		return "", "", fmt.Errorf("computing fixture model path: %w", err)
	}

	namespace := "com.amazonaws." + strings.ReplaceAll(awsService, "-", "")
	return filepath.ToSlash(modelPath), namespace, nil
}

func extractResourceWithDerivedModel(t *testing.T, tfName string, m *awsmapping.ResourceMapping) (*tfschema.ResourceIR, error) {
	t.Helper()

	modelPath, namespace, err := resolveFixtureModelAndNamespace(tfName)
	if err != nil {
		return nil, err
	}

	model, err := awsschema.LoadModel(apiModelsBaseURL(t), modelPath)
	if err != nil {
		return nil, err
	}

	return awsschema.ExtractResource(tfName, m, model, namespace)
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
// SQS — enum-backed attribute map
// ---------------------------------------------------------------------------

func TestExtract_SQS_PatternB(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{
		Lifecycle: awsmapping.Lifecycle{
			Create: "CreateQueue",
			Read:   "GetQueueAttributes",
			Update: "SetQueueAttributes",
			Delete: "DeleteQueue",
			List:   "ListQueues",
		},
		AttributeMapEnum: "QueueAttributeName",
		SuppressFields: []string{
			"ApproximateNumberOfMessages",
			"ApproximateNumberOfMessagesNotVisible",
			"ApproximateNumberOfMessagesDelayed",
			"CreatedTimestamp",
			"LastModifiedTimestamp",
			"NextToken",
			"All",
		},
		FieldRenames: map[string]string{
			"VisibilityTimeout":             "visibility_timeout_seconds",
			"MaximumMessageSize":            "max_message_size",
			"MessageRetentionPeriod":        "message_retention_seconds",
			"ReceiveMessageWaitTimeSeconds": "receive_wait_time_seconds",
			"QueueArn":                      "arn",
			"QueueUrl":                      "url",
			"QueueName":                     "name",
			"QueueNamePrefix":               "name_prefix",
		},
	}
	ir, err := extractResourceWithDerivedModel(t, "aws_sqs_queue", m)
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
// SNS — explicit fields for an untyped attribute map
// ---------------------------------------------------------------------------

func TestExtract_SNS_PatternC(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{
		Lifecycle: awsmapping.Lifecycle{
			Create: "CreateTopic",
			Read:   "GetTopicAttributes",
		},
		FieldRenames: map[string]string{"topic_arn": "arn"},
		ExplicitFields: []awsmapping.ExplicitField{
			{Name: "DisplayName", Type: "string"},
			{Name: "KmsMasterKeyId", Type: "string"},
			{Name: "FifoTopic", Type: "bool"},
			{Name: "ContentBasedDeduplication", Type: "bool"},
			{Name: "Policy", Type: "string"},
			{Name: "DeliveryPolicy", Type: "string"},
			{Name: "TracingConfig", Type: "string"},
			{Name: "SignatureVersion", Type: "string"},
			{Name: "ArchivePolicy", Type: "string"},
		},
	}
	ir, err := extractResourceWithDerivedModel(t, "aws_sns_topic", m)
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
// AMP — Smithy resource lifecycle extraction
// ---------------------------------------------------------------------------

func TestExtract_AMP_PatternA(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{
		SmithyResource: "Workspace",
		SuppressFields: []string{"clientToken", "workspaceId", "status", "workspace"},
		FieldRenames: map[string]string{
			"workspaceId": "id",
			"kmsKeyArn":   "kms_key_arn",
		},
	}
	ir, err := extractResourceWithDerivedModel(t, "aws_prometheus_workspace", m)
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
		SmithyResource: "Workspace",
		SuppressFields: []string{"clientToken", "workspaceId", "status", "workspace"},
		FieldRenames: map[string]string{
			"kmsKeyArn": "kms_key_arn",
			"arn":       "arn",
		},
	}

	ir, err := extractResourceWithDerivedModel(t, "aws_prometheus_workspace", m)
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
		SmithyResource: "WorkspaceResourcePolicy",
		SuppressFields: []string{"workspaceId"},
		FieldRenames: map[string]string{
			"policyDocument": "policy_document",
			"revisionId":     "revision_id",
		},
	}

	ir, err := extractResourceWithDerivedModel(t, "aws_prometheus_resource_policy", m)
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}

	for _, want := range []string{"policy_document", "revision_id"} {
		if _, ok := ir.Fields[want]; !ok {
			t.Errorf("field %q missing", want)
		}
	}

	// These fields come from the resource-level `operations` target
	// (CreateWorkspace) and are not present in put/read/delete.
	if _, ok := ir.Fields["alias"]; !ok {
		t.Error("field \"alias\" missing from inferred resource operations")
	}
	if _, ok := ir.Fields["kms_key_arn"]; !ok {
		t.Error("field \"kms_key_arn\" missing from inferred resource operations")
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

	_, err := awsschema.ExtractResource("aws_example", &awsmapping.ResourceMapping{}, nil, "")
	if err == nil {
		t.Fatal("expected an error for a mapping without extraction configuration")
	}
	if !strings.Contains(err.Error(), "no extraction configuration") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtract_ExplicitFieldsWithoutModel(t *testing.T) {
	t.Parallel()

	m := &awsmapping.ResourceMapping{ExplicitFields: []awsmapping.ExplicitField{
		{Name: "Enabled", Type: "bool", Required: true},
		{Name: "Count", Type: "int64"},
		{Name: "Ratio", Type: "float64"},
		{Name: "Description", Type: "string"},
	}}
	ir, err := awsschema.ExtractResource("aws_example", m, nil, "")
	if err != nil {
		t.Fatalf("ExtractResource: %v", err)
	}
	if len(ir.Fields) != 4 {
		t.Fatalf("field count = %d, want 4", len(ir.Fields))
	}
	if got := ir.Fields["enabled"]; got.Type != tfschema.FieldTypeBool || !got.Required || got.Optional {
		t.Errorf("enabled field = %#v", got)
	}
	if got := ir.Fields["count"]; got.Type != tfschema.FieldTypeInt64 || !got.Optional || got.Required {
		t.Errorf("count field = %#v", got)
	}
	if got := ir.Fields["ratio"]; got.Type != tfschema.FieldTypeFloat64 {
		t.Errorf("ratio type = %q, want float64", got.Type)
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
			ir, err := extractResourceWithDerivedModel(t, tfName, m)
			if err != nil {
				if strings.Contains(err.Error(), "no extraction configuration") {
					t.Skip("mapping relies on resource-service discovery before extraction")
				}
				if strings.Contains(err.Error(), "no fixture model found") {
					t.Skip(err.Error())
				}
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
			ir, err := extractResourceWithDerivedModel(t, tfName, m)
			if err != nil {
				if strings.Contains(err.Error(), "no extraction configuration") {
					t.Skip("mapping relies on resource-service discovery before extraction")
				}
				if strings.Contains(err.Error(), "no fixture model found") {
					t.Skip(err.Error())
				}
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
