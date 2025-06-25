// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalSingletonSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"region": resourceattribute.Region(),
	},
}

func regionalSingletonImporter(identitySpec inttypes.Identity) (importer framework.WithImportRegionalSingleton) {
	importer.SetIdentitySpec(identitySpec)
	return
}

func TestRegionalSingleton_ImportID_Invalid_WrongRegion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	identitySpec := inttypes.RegionalSingletonIdentity()

	resImporter := regionalSingletonImporter(identitySpec)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	response := importByIDWithState(ctx, &resImporter, regionalSingletonSchema, region, map[string]string{
		"region": "another-region-1",
	}, identitySchema)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The region passed for import,") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalSingleton_ImportID_Valid_DefaultRegion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	identitySpec := inttypes.RegionalSingletonIdentity()

	resImporter := regionalSingletonImporter(identitySpec)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	response := importByID(ctx, &resImporter, regionalSingletonSchema, region, identitySchema)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_ImportID_Valid_RegionOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	identitySpec := inttypes.RegionalSingletonIdentity()

	resImporter := regionalSingletonImporter(identitySpec)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	response := importByIDWithState(ctx, &resImporter, regionalSingletonSchema, region, map[string]string{
		"region": region,
	}, identitySchema)
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	identitySpec := inttypes.RegionalSingletonIdentity()

	resImporter := regionalSingletonImporter(identitySpec)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	response := importByIdentity(ctx, &resImporter, regionalSingletonSchema, identitySchema, map[string]string{
		"region": region,
	})
	if response.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
	}

	if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}
