// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func TestImportByARN_GlobalARN_ImportID_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := globalARNImporter()

	response := importByID(ctx, &importer, globalARNSchema, arn, globalARNIdentitySchema)
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
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := globalARNImporter()

	response := importByIDNoIdentity(ctx, &importer, globalARNSchema, arn)
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
	ctx := context.Background()

	importer := globalARNImporter()

	response := importByID(ctx, &importer, globalARNSchema, "not a valid ARN", globalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		if response.Diagnostics[0].Summary() != "Invalid Resource Import ID Value" {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_GlobalARN_Identity_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := globalARNImporter()

	response := importByIdentity(ctx, &importer, globalARNSchema, globalARNIdentitySchema, map[string]string{
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
	ctx := context.Background()

	importer := globalARNImporter()

	response := importByIdentity(ctx, &importer, globalARNSchema, globalARNIdentitySchema, map[string]string{
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

func TestImportByARN_GlobalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := globalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByID(ctx, &importer, globalARNWithIDSchema, arn, globalARNIdentitySchema)
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
	ctx := context.Background()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := globalARNImporterWithDuplicateAttrs("id", "attr")

	response := importByIdentity(ctx, &importer, globalARNWithIDSchema, globalARNIdentitySchema, map[string]string{
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

func TestImportByARN_RegionalARN_ImportID_Valid_DefaultRegion(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporter()

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
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporter()

	response := importByIDWithState(ctx, &importer, regionalARNSchema, arn, map[string]string{
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
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporter()

	response := importByIDNoIdentity(ctx, &importer, regionalARNSchema, arn)
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
	ctx := context.Background()

	importer := regionalARNImporter()

	response := importByID(ctx, &importer, regionalARNSchema, "not a valid ARN", regionalARNIdentitySchema)
	if response.Diagnostics.HasError() {
		if !strings.HasPrefix(response.Diagnostics[0].Detail(), "The import ID could not be parsed as an ARN.") {
			t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestImportByARN_RegionalARN_ImportID_Invalid_WrongRegion(t *testing.T) {
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporter()

	response := importByIDWithState(ctx, &importer, regionalARNSchema, arn, map[string]string{
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
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporter()

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
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporterWithDuplicateAttrs("id", "attr")

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
	ctx := context.Background()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()

	importer := regionalARNImporterWithDuplicateAttrs("id", "attr")

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
