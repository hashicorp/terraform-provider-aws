// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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

func globalARNImporter() (importer framework.WithImportByARN) {
	importer.SetIdentitySpec(
		inttypes.GlobalARNIdentity(),
	)
	return
}

func globalARNImporterWithDuplicateAttrs(attrs ...string) (importer framework.WithImportByARN) {
	importer.SetIdentitySpec(
		inttypes.GlobalARNIdentity(
			inttypes.WithIdentityDuplicateAttrs(attrs...),
		),
	)
	return
}

type mockClient struct {
	accountID string
	region    string
}

func (c mockClient) AccountID(_ context.Context) string {
	return c.accountID
}

func (c mockClient) Region(_ context.Context) string {
	return c.region
}

func TestImportByARN_GlobalARN_ImportID_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporter()

	response := importByID(ctx, &resImporter, globalARNSchema, arn, globalARNIdentitySchema)
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

func TestImportByARN_GlobalARN_ImportID_Valid_NoIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporter()

	response := importByIDNoIdentity(ctx, &resImporter, globalARNSchema, arn)
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

func TestImportByARN_GlobalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporter()

	response := importByID(ctx, &resImporter, globalARNSchema, "not a valid ARN", globalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		if response.Diagnostics[0].Summary() != importer.InvalidResourceImportIDValue {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_GlobalARN_Identity_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporter()

	response := importByIdentity(ctx, &resImporter, globalARNSchema, globalARNIdentitySchema, map[string]string{
		"arn": arn,
	})
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

func TestImportByARN_GlobalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporter()

	response := importByIdentity(ctx, &resImporter, globalARNSchema, globalARNIdentitySchema, map[string]string{
		"arn": "not a valid ARN",
	})
	if response.Diagnostics.HasError() {
		if response.Diagnostics[0].Summary() != "Invalid Identity Attribute Value" {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_GlobalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByID(ctx, &resImporter, globalARNWithIDSchema, arn, globalARNIdentitySchema)
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

func TestImportByARN_GlobalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := globalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByIdentity(ctx, &resImporter, globalARNWithIDSchema, globalARNIdentitySchema, map[string]string{
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

func regionalARNImporter() (importer framework.WithImportByARN) {
	importer.SetIdentitySpec(
		inttypes.RegionalARNIdentity(),
	)
	return
}

func regionalARNImporterWithDuplicateAttrs(attrs ...string) (importer framework.WithImportByARN) {
	importer.SetIdentitySpec(
		inttypes.RegionalARNIdentity(
			inttypes.WithIdentityDuplicateAttrs(attrs...),
		),
	)
	return
}

func TestImportByARN_RegionalARN_ImportID_Valid_DefaultRegion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByID(ctx, &resImporter, regionalARNSchema, arn, regionalARNIdentitySchema)
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

func TestImportByARN_RegionalARN_ImportID_Valid_RegionOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByIDWithState(ctx, &resImporter, regionalARNSchema, arn, map[string]string{
		"region": region,
	}, regionalARNIdentitySchema)
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

func TestImportByARN_RegionalARN_ImportID_Valid_NoIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByIDNoIdentity(ctx, &resImporter, regionalARNSchema, arn)
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

	if response.Identity != nil {
		t.Error("Identity should not be set")
	}
}

func TestImportByARN_RegionalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByID(ctx, &resImporter, regionalARNSchema, "not a valid ARN", regionalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The import ID could not be parsed as an ARN.") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_RegionalARN_ImportID_Invalid_WrongRegion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByIDWithState(ctx, &resImporter, regionalARNSchema, arn, map[string]string{
		"region": "another-region-1",
	}, regionalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The region passed for import,") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_RegionalARN_Identity_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporter()

	response := importByIdentity(ctx, &resImporter, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
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

func TestImportByARN_RegionalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByID(ctx, &resImporter, regionalARNWithIDSchema, arn, regionalARNIdentitySchema)
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

func TestImportByARN_RegionalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	client := mockClient{
		accountID: "123456789012",
	}
	ctx = importer.Context(ctx, &client)

	resImporter := regionalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByIdentity(ctx, &resImporter, regionalARNWithIDSchema, regionalARNIdentitySchema, map[string]string{
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
