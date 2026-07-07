// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsmapping_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
)

// TestCamelToSnake exercises the name-normalisation helper.
func TestCamelToSnake(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"QueueName", "queue_name"},
		{"FifoQueue", "fifo_queue"},
		{"KmsMasterKeyId", "kms_master_key_id"},
		{"SqsManagedSseEnabled", "sqs_managed_sse_enabled"},
		{"VisibilityTimeout", "visibility_timeout"},
		{"MaximumMessageSize", "maximum_message_size"},
		{"MessageRetentionPeriod", "message_retention_period"},
		{"DelaySeconds", "delay_seconds"},
		{"ContentBasedDeduplication", "content_based_deduplication"},
		{"RedriveAllowPolicy", "redrive_allow_policy"},
		{"DisplayName", "display_name"},
		{"alias", "alias"},
		{"kmsKeyArn", "kms_key_arn"},
		{"", ""},
	}

	for _, tc := range cases {
		got := awsmapping.CamelToSnake(tc.in)
		if got != tc.want {
			t.Errorf("CamelToSnake(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestLoadFile_ParsesRealMappingFile ensures the actual aws_resources.yaml
// parses without error and contains the three Phase 1 resources.
func TestLoadFile_ParsesRealMappingFile(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	wantResources := []string{
		"aws_sqs_queue",
		"aws_sns_topic",
		"aws_prometheus_workspace",
		"aws_prometheus_resource_policy",
	}
	for _, name := range wantResources {
		if _, ok := f.Resources[name]; !ok {
			t.Errorf("resource %q missing from mapping file", name)
		}
	}
}

// TestLoadFile_SQSCapabilities confirms aws_sqs_queue has expected
// lifecycle and enum extraction configuration.
func TestLoadFile_SQSCapabilities(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_sqs_queue"]
	if m == nil {
		t.Fatal("aws_sqs_queue missing")
	}
	if m.AttributeMapEnum != "QueueAttributeName" {
		t.Errorf("AttributeMapEnum = %q, want %q", m.AttributeMapEnum, "QueueAttributeName")
	}
	if m.Lifecycle.Create != "CreateQueue" {
		t.Errorf("Lifecycle.Create = %q, want %q", m.Lifecycle.Create, "CreateQueue")
	}
}

// TestLoadFile_SNSCapabilities confirms aws_sns_topic has explicit field
// extraction configured.
func TestLoadFile_SNSCapabilities(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_sns_topic"]
	if m == nil {
		t.Fatal("aws_sns_topic missing")
	}
	if m.AttributeMapEnum != "" {
		t.Errorf("AttributeMapEnum = %q, want empty", m.AttributeMapEnum)
	}
	if len(m.ExplicitFields) == 0 {
		t.Error("aws_sns_topic: ExplicitFields is empty")
	}
}

// TestLoadFile_AMPWorkspaceCapabilities confirms aws_prometheus_workspace
// is configured for lifecycle inference from smithy_resource.
func TestLoadFile_AMPWorkspaceCapabilities(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_prometheus_workspace"]
	if m == nil {
		t.Fatal("aws_prometheus_workspace missing")
	}

	if m.Lifecycle.Create != "" {
		t.Errorf("Lifecycle.Create = %q, want empty for inference", m.Lifecycle.Create)
	}
	if m.SmithyResource != "Workspace" {
		t.Errorf("SmithyResource = %q, want %q", m.SmithyResource, "Workspace")
	}
}

func TestLoadFile_AMPResourcePolicyCapabilities(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_prometheus_resource_policy"]
	if m == nil {
		t.Fatal("aws_prometheus_resource_policy missing")
	}
	if m.SmithyResource != "WorkspaceResourcePolicy" {
		t.Errorf("SmithyResource = %q, want %q", m.SmithyResource, "WorkspaceResourcePolicy")
	}
	if m.Lifecycle.Put != "" {
		t.Errorf("Lifecycle.Put = %q, want empty for inference", m.Lifecycle.Put)
	}
	if m.Lifecycle.List != "" {
		t.Errorf("Lifecycle.List = %q, want empty", m.Lifecycle.List)
	}
}

// TestMapping_TFName checks that FieldRenames overrides take priority and
// the CamelToSnake fallback is used otherwise.
func TestMapping_TFName(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_sqs_queue"]
	if m == nil {
		t.Fatal("aws_sqs_queue missing")
	}

	cases := []struct {
		awsName string
		want    string
	}{
		{"VisibilityTimeout", "visibility_timeout_seconds"}, // renamed
		{"DelaySeconds", "delay_seconds"},                   // renamed
		{"FifoQueue", "fifo_queue"},                         // renamed
		{"QueueArn", "arn"},                                 // renamed
	}
	for _, tc := range cases {
		got := m.TFName(tc.awsName)
		if got != tc.want {
			t.Errorf("TFName(%q) = %q, want %q", tc.awsName, got, tc.want)
		}
	}
}

// TestMapping_IsSuppressed checks the suppress list.
func TestMapping_IsSuppressed(t *testing.T) {
	t.Parallel()

	f, err := awsmapping.LoadFile("../../mappings/aws_resources.yaml")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := f.Resources["aws_sqs_queue"]
	if m == nil {
		t.Fatal("aws_sqs_queue missing")
	}

	if m.IsSuppressed("QueueUrl") {
		t.Error("QueueUrl should NOT be suppressed")
	}
	if m.IsSuppressed("FifoQueue") {
		t.Error("FifoQueue should NOT be suppressed")
	}
}
