// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/fwprovider/resourceattribute"
)

var regionalARNSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
		"id":     framework.IDAttributeDeprecatedNoReplacement(),
		"region": resourceattribute.Region(),
	},
}

var regionalARNIdentitySchema = identityschema.Schema{
	Attributes: map[string]identityschema.Attribute{
		"arn": identityschema.StringAttribute{
			RequiredForImport: true,
		},
	},
}

func TestRegionalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	ctx := context.Background()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		ID: "not a valid ARN",
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(regionalARNSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The import ID could not be parsed as an ARN.") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_ImportID_Invalid_WrongRegion(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State: stateFromSchema(regionalARNSchema, map[string]tftypes.Value{
			"arn":    tftypes.NewValue(tftypes.String, nil),
			"region": tftypes.NewValue(tftypes.String, "another-region-1"),
			"attr":   tftypes.NewValue(tftypes.String, nil),
			"id":     tftypes.NewValue(tftypes.String, nil),
		}),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The region passed for import,") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_ImportID_Valid_NoRegionSet(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(regionalARNSchema),
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

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_ImportID_Valid_RegionSet(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State: stateFromSchema(regionalARNSchema, map[string]tftypes.Value{
			"arn":    tftypes.NewValue(tftypes.String, nil),
			"region": tftypes.NewValue(tftypes.String, region),
			"attr":   tftypes.NewValue(tftypes.String, nil),
			"id":     tftypes.NewValue(tftypes.String, nil),
		}),
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

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	ctx := context.Background()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		Identity: identityFromSchema(regionalARNIdentitySchema, map[string]tftypes.Value{
			"arn": tftypes.NewValue(tftypes.String, "not a valid ARN"),
		}),
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(regionalARNSchema),
		Identity: emtpyIdentityFromSchema(regionalARNIdentitySchema),
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

func TestRegionalARN_Identity_Valid(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn")

	request := resource.ImportStateRequest{
		Identity: identityFromSchema(regionalARNIdentitySchema, map[string]tftypes.Value{
			"arn": tftypes.NewValue(tftypes.String, arn),
		}),
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(regionalARNSchema),
		Identity: emtpyIdentityFromSchema(regionalARNIdentitySchema),
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

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}
