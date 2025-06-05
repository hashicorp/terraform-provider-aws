// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

var globalARNSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
	},
}

var globalARNWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
		"id": framework.IDAttributeDeprecatedNoReplacement(),
	},
}

var globalARNIdentitySchema = identityschema.Schema{
	Attributes: map[string]identityschema.Attribute{
		"arn": identityschema.StringAttribute{
			RequiredForImport: true,
		},
	},
}

func TestGlobalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	ctx := context.Background()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", nil)

	request := resource.ImportStateRequest{
		ID:       "not a valid ARN",
		Identity: nil,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		if response.Diagnostics[0].Summary() != "Invalid Resource Import ID Value" {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestGlobalARN_ImportID_Valid_NoIdentity(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", nil)

	request := resource.ImportStateRequest{
		ID:       arn,
		Identity: nil,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}

	if response.Identity != nil {
		t.Error("Identity should not be set")
	}
}

func TestGlobalARN_ImportID_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", nil)

	identity := emtpyIdentityFromSchema(globalARNIdentitySchema)

	request := resource.ImportStateRequest{
		ID:       arn,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}

	if identity := response.Identity; identity == nil {
		t.Error("Identity should be set")
	} else {
		var arnVal string
		identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
		if e, a := arn, arnVal; e != a {
			t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
		}
	}
}

func TestGlobalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	ctx := context.Background()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", nil)

	request := resource.ImportStateRequest{
		Identity: identityFromSchema(globalARNIdentitySchema, map[string]string{
			"arn": "not a valid ARN",
		}),
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNSchema),
		Identity: emtpyIdentityFromSchema(globalARNIdentitySchema),
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		if response.Diagnostics[0].Summary() != "Invalid Import Attribute Value" {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestGlobalARN_Identity_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", nil)

	identity := identityFromSchema(globalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}

	if identity := response.Identity; identity == nil {
		t.Error("Identity should be set")
	} else {
		var arnVal string
		identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
		if e, a := arn, arnVal; e != a {
			t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
		}
	}
}

func TestGlobalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", []string{"id", "attr"})

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNWithIDSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestGlobalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByGlobalARN{}
	importer.SetARNAttributeName("arn", []string{"id", "attr"})

	identity := identityFromSchema(globalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(globalARNWithIDSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}

	if identity := response.Identity; identity == nil {
		t.Error("Identity should be set")
	} else {
		var arnVal string
		identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
		if e, a := arn, arnVal; e != a {
			t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
		}
	}
}

func emtpyStateFromSchema(schema schema.Schema) tfsdk.State {
	return tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil),
		Schema: schema,
	}
}

func stateFromSchema(schema schema.Schema, values map[string]string) tfsdk.State {
	val := make(map[string]tftypes.Value)
	for name := range maps.Keys(schema.Attributes) {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(context.Background()), val),
		Schema: schema,
	}
}

func emtpyIdentityFromSchema(schema identityschema.Schema) *tfsdk.ResourceIdentity {
	return &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil),
		Schema: schema,
	}
}

func identityFromSchema(schema identityschema.Schema, values map[string]string) *tfsdk.ResourceIdentity {
	val := make(map[string]tftypes.Value)
	for name := range maps.Keys(schema.Attributes) {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(context.Background()), val),
		Schema: schema,
	}
}

func getAttributeValue(ctx context.Context, t *testing.T, state tfsdk.State, path path.Path) string {
	t.Helper()

	var attrVal types.String
	if diags := state.GetAttribute(ctx, path, &attrVal); diags.HasError() {
		t.Fatalf("Unexpected error getting attribute %q: %s", path, fwdiag.DiagnosticsError(diags))
	}
	return attrVal.ValueString()
}
