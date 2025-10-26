// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEmptyBucketAction_Schema(t *testing.T) {
	ctx := context.Background()
	
	// Create the action
	actionInstance, err := newEmptyBucketAction(ctx)
	if err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}

	// Test schema generation
	req := action.SchemaRequest{}
	resp := &action.SchemaResponse{}
	
	actionInstance.Schema(ctx, req, resp)
	
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema generation failed: %v", resp.Diagnostics)
	}
	
	// Verify required attributes exist
	schema := resp.Schema
	if _, ok := schema.Attributes["bucket_name"]; !ok {
		t.Error("bucket_name attribute missing from schema")
	}
	if _, ok := schema.Attributes["prefix"]; !ok {
		t.Error("prefix attribute missing from schema")
	}
	if _, ok := schema.Attributes["batch_size"]; !ok {
		t.Error("batch_size attribute missing from schema")
	}
	if _, ok := schema.Attributes["timeout"]; !ok {
		t.Error("timeout attribute missing from schema")
	}
}

func TestEmptyBucketAction_ModelValidation(t *testing.T) {
	// Test that the model struct is properly defined
	model := emptyBucketModel{
		BucketName: types.StringValue("test-bucket"),
		Prefix:     types.StringValue("test/"),
		BatchSize:  types.Int64Value(500),
		Timeout:    types.Int64Value(1800),
	}
	
	if model.BucketName.ValueString() != "test-bucket" {
		t.Error("BucketName not properly set")
	}
	if model.Prefix.ValueString() != "test/" {
		t.Error("Prefix not properly set")
	}
	if model.BatchSize.ValueInt64() != 500 {
		t.Error("BatchSize not properly set")
	}
	if model.Timeout.ValueInt64() != 1800 {
		t.Error("Timeout not properly set")
	}
}

func TestEmptyBucketAction_HelperFunctions(t *testing.T) {
	ctx := context.Background()
	
	// Create the action to test helper methods
	actionInstance := &emptyBucketAction{}
	
	// Test deleteObjects with empty slice
	count, err := actionInstance.deleteObjects(ctx, nil, "test-bucket", nil)
	if err != nil {
		t.Errorf("deleteObjects with empty slice should not error: %v", err)
	}
	if count != 0 {
		t.Errorf("deleteObjects with empty slice should return 0, got %d", count)
	}
}
