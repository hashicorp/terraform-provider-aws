// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

type importStater interface {
	ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse)
}

func importByID(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string, identitySchema identityschema.Schema) resource.ImportStateResponse {
	identity := emtpyIdentityFromSchema(ctx, identitySchema)

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIDWithState(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string, stateAttrs map[string]string, identitySchema identityschema.Schema) resource.ImportStateResponse {
	identity := emtpyIdentityFromSchema(ctx, identitySchema)

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIDNoIdentity(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string) resource.ImportStateResponse {
	request := resource.ImportStateRequest{
		ID:       id,
		Identity: nil,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIdentity(ctx context.Context, importer importStater, resourceSchema schema.Schema, identitySchema identityschema.Schema, identityAttrs map[string]string) resource.ImportStateResponse {
	identity := identityFromSchema(ctx, identitySchema, identityAttrs)

	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func getAttributeValue(ctx context.Context, t *testing.T, state tfsdk.State, path path.Path) string {
	t.Helper()

	var attrVal types.String
	if diags := state.GetAttribute(ctx, path, &attrVal); diags.HasError() {
		t.Fatalf("Unexpected error getting attribute %q: %s", path, fwdiag.DiagnosticsError(diags))
	}
	return attrVal.ValueString()
}

func emtpyStateFromSchema(ctx context.Context, schema schema.Schema) tfsdk.State {
	return tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}
}

func stateFromSchema(ctx context.Context, schema schema.Schema, values map[string]string) tfsdk.State {
	val := make(map[string]tftypes.Value)
	for name := range maps.Keys(schema.Attributes) {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), val),
		Schema: schema,
	}
}

func emtpyIdentityFromSchema(ctx context.Context, schema identityschema.Schema) *tfsdk.ResourceIdentity {
	return &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}
}

func identityFromSchema(ctx context.Context, schema identityschema.Schema, values map[string]string) *tfsdk.ResourceIdentity {
	val := make(map[string]tftypes.Value)
	for name := range maps.Keys(schema.Attributes) {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), val),
		Schema: schema,
	}
}
