// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSpaceResource_Schema(t *testing.T) {
	t.Parallel()

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
	requiredAttrs := []string{names.AttrName, "domain_id", "data_access_role_arn"}
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
	computedAttrs := []string{"space_id", "space_arn", "domain_arn", names.AttrStatus}
	for _, attr := range computedAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("computed attribute %s not found in schema", attr)
		}
	}

	// Verify encryption configuration block exists
	if _, exists := resp.Schema.Blocks[names.AttrEncryptionConfiguration]; !exists {
		t.Error("encryption_configuration block not found in schema")
	}

	// Tags are deliberately unsupported: CreateSpace accepts them but the Space
	// response type omits them and the service has no tag read/write APIs, so a
	// tags attribute could never refresh, update, or survive import.
	for _, attr := range []string{names.AttrTags, names.AttrTagsAll} {
		if _, exists := resp.Schema.Attributes[attr]; exists {
			t.Errorf("attribute %s should not be in schema: CloudWatch Omni has no tagging API", attr)
		}
	}

	// The resource is identified by space_id and has no separate id attribute;
	// ImportState passes through to space_id.
	if _, exists := resp.Schema.Attributes[names.AttrID]; exists {
		t.Error("unexpected id attribute: the resource is identified by space_id")
	}
}
