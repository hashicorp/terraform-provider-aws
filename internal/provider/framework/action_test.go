// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// testAction implements action.Action for testing
type testAction struct{}

func (t *testAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "aws_test_action"
}

func (t *testAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.UnlinkedSchema{
		Description: "Test action for framework integration",
		Attributes: map[string]schema.Attribute{
			"test_param": schema.StringAttribute{
				Required:    true,
				Description: "Test parameter",
			},
		},
	}
}

func (t *testAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	// Test implementation - just validate we can access the config
	var config testActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Test action executed successfully",
	})
}

// testActionModel represents the configuration model for the test action
type testActionModel struct {
	TestParam types.String `tfsdk:"test_param"`
}

// Implement ActionValidateModel interface
func (t *testAction) ValidateModel(ctx context.Context, schema *schema.UnlinkedSchema) diag.Diagnostics {
	var diags diag.Diagnostics
	// Basic validation - ensure required attributes exist
	if _, ok := schema.Attributes["test_param"]; !ok {
		diags.AddError("Missing required attribute", "test_param attribute is required")
	}
	return diags
}

// Ensure testAction implements required interfaces
var (
	_ action.Action                 = (*testAction)(nil)
	_ framework.ActionValidateModel = (*testAction)(nil)
)

func TestWrappedAction_Basic(t *testing.T) {
	ctx := context.Background()

	// Create test action
	inner := &testAction{}

	// Create wrapped action with minimal options
	opts := wrappedActionOptions{
		bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
			return ctx, nil
		},
		interceptors: interceptorInvocations{},
		typeName:     "aws_test_action",
	}

	wrapped := newWrappedAction(inner, opts)

	// Test Metadata
	metaReq := action.MetadataRequest{
		ProviderTypeName: "aws",
	}
	var metaResp action.MetadataResponse
	wrapped.Metadata(ctx, metaReq, &metaResp)

	if metaResp.TypeName != "aws_test_action" {
		t.Errorf("Expected TypeName 'aws_test_action', got '%s'", metaResp.TypeName)
	}

	// Test Schema
	schemaReq := action.SchemaRequest{}
	var schemaResp action.SchemaResponse
	wrapped.Schema(ctx, schemaReq, &schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Errorf("Schema method returned errors: %v", schemaResp.Diagnostics)
	}

	if unlinkedSchema, ok := schemaResp.Schema.(schema.UnlinkedSchema); ok {
		if _, exists := unlinkedSchema.Attributes["test_param"]; !exists {
			t.Error("Expected 'test_param' attribute in schema")
		}
	} else {
		t.Error("Expected UnlinkedSchema type")
	}
}

func TestActionInterceptors_RegionInjection(t *testing.T) {
	ctx := context.Background()

	// Create test action
	inner := &testAction{}

	// Create wrapped action with region interceptor
	opts := wrappedActionOptions{
		bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
			return ctx, nil
		},
		interceptors: interceptorInvocations{
			actionInjectRegionAttribute(),
		},
		typeName: "aws_test_action",
	}

	wrapped := newWrappedAction(inner, opts)

	// Test Schema with region injection
	schemaReq := action.SchemaRequest{}
	var schemaResp action.SchemaResponse
	wrapped.Schema(ctx, schemaReq, &schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Errorf("Schema method returned errors: %v", schemaResp.Diagnostics)
	}

	if unlinkedSchema, ok := schemaResp.Schema.(schema.UnlinkedSchema); ok {
		if _, exists := unlinkedSchema.Attributes["region"]; !exists {
			t.Error("Expected 'region' attribute to be injected into schema")
		}
		if _, exists := unlinkedSchema.Attributes["test_param"]; !exists {
			t.Error("Expected original 'test_param' attribute to remain in schema")
		}
	} else {
		t.Error("Expected UnlinkedSchema type")
	}
}
