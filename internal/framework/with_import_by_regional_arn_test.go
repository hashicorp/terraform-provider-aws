// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
		"region": resourceattribute.Region(),
	},
}

var regionalARNWithIDSchema = schema.Schema{
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
	importer.SetARNAttributeName("arn", nil)

	response := importByID(ctx, &importer, regionalARNSchema, "not a valid ARN", regionalARNIdentitySchema)
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
	importer.SetARNAttributeName("arn", nil)

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State: stateFromSchema(regionalARNSchema, map[string]string{
			"region": "another-region-1",
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
	importer.SetARNAttributeName("arn", nil)

	response := importByID(ctx, &importer, regionalARNSchema, arn, regionalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
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
	importer.SetARNAttributeName("arn", nil)

	request := resource.ImportStateRequest{
		ID: arn,
	}
	response := resource.ImportStateResponse{
		State: stateFromSchema(regionalARNSchema, map[string]string{
			"region": region,
		}),
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
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	ctx := context.Background()

	importer := framework.WithImportByRegionalARN{}
	importer.SetARNAttributeName("arn", nil)

	response := importByIdentity(ctx, &importer, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
		"arn": "not a valid ARN",
	})
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
	importer.SetARNAttributeName("arn", nil)

	response := importByIdentity(ctx, &importer, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
		"arn": arn,
	})
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := arn, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
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
	importer.SetARNAttributeName("arn", []string{"id", "attr"})

	response := importByID(ctx, &importer, regionalARNWithIDSchema, arn, regionalARNIdentitySchema)
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
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
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
	importer.SetARNAttributeName("arn", []string{"id", "attr"})

	response := importByIdentity(ctx, &importer, regionalARNWithIDSchema, regionalARNIdentitySchema, map[string]string{
		"arn": arn,
	})
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
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}
