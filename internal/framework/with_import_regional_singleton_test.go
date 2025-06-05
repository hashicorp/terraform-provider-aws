// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/fwprovider/resourceattribute"
)

var regionalSingletonSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id":     framework.IDAttributeDeprecatedNoReplacement(),
		"region": resourceattribute.Region(),
	},
}

var regionalSingletonIdentitySchema = identityschema.Schema{
	Attributes: map[string]identityschema.Attribute{
		"region": identityschema.StringAttribute{
			OptionalForImport: true,
		},
	},
}

func TestRegionalSingleton_ImportID_Invalid_WrongRegion(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"

	importer := framework.WithImportRegionalSingleton{}

	response := importByIDWithState(ctx, &importer, regionalSingletonSchema, region, map[string]string{
		"region": "another-region-1",
	}, regionalSingletonIdentitySchema)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The region passed for import,") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalSingleton_ImportID_Valid_NoRegionSet(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"

	importer := framework.WithImportRegionalSingleton{}

	response := importByID(ctx, &importer, regionalSingletonSchema, region, regionalSingletonIdentitySchema)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_ImportID_Valid_RegionSet(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"

	importer := framework.WithImportRegionalSingleton{}

	response := importByIDWithState(ctx, &importer, regionalSingletonSchema, region, map[string]string{
		"region": region,
	}, regionalSingletonIdentitySchema)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Valid(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"

	importer := framework.WithImportRegionalSingleton{}

	response := importByIdentity(ctx, &importer, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{
		"region": region,
	})
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}
