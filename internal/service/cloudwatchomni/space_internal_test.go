// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSpaceResource_Schema(t *testing.T) {
	ctx := context.Background()

	r, err := newSpaceResource(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema errors: %v", resp.Diagnostics.Errors())
	}

	// Verify required attributes exist
	requiredAttrs := []string{"name", "domain_id", "data_access_role_arn"}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("required attribute %s not found in schema", attr)
		}
	}

	// Verify optional attributes exist
	optionalAttrs := []string{"agent_core_evaluation_role_arn"}
	for _, attr := range optionalAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("optional attribute %s not found in schema", attr)
		}
	}

	// Verify computed attributes exist
	computedAttrs := []string{"space_id", "space_arn", "domain_arn", "status"}
	for _, attr := range computedAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("computed attribute %s not found in schema", attr)
		}
	}

	// Verify encryption configuration block exists
	if _, exists := resp.Schema.Blocks["encryption_configuration"]; !exists {
		t.Error("encryption_configuration block not found in schema")
	}

	// Tags are deliberately unsupported: CreateSpace accepts them but the Space
	// response type omits them and the service has no tag read/write APIs, so a
	// tags attribute could never refresh, update, or survive import.
	for _, attr := range []string{"tags", "tags_all"} {
		if _, exists := resp.Schema.Attributes[attr]; exists {
			t.Errorf("attribute %s should not be in schema: CloudWatch Omni has no tagging API", attr)
		}
	}

	// The resource is identified by space_id and has no separate id attribute;
	// ImportState passes through to space_id.
	if _, exists := resp.Schema.Attributes["id"]; exists {
		t.Error("unexpected id attribute: the resource is identified by space_id")
	}
}

func TestSpaceResource_Metadata(t *testing.T) {
	ctx := context.Background()

	r, err := newSpaceResource(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	req := resource.MetadataRequest{
		ProviderTypeName: "aws",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	expectedTypeName := "aws_cloudwatchomni_space"
	if resp.TypeName != expectedTypeName {
		t.Errorf("expected TypeName %s, got %s", expectedTypeName, resp.TypeName)
	}
}
