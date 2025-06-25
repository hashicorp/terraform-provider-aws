// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
)

var regionalARNSchema = map[string]*schema.Schema{
	"arn": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"region": attribute.Region(),
}

// lintignore:S013 // Identity Schemas cannot specify Computed, Optional, or Required
var regionalARNIdentitySchema = map[string]*schema.Schema{
	"arn": {
		Type:              schema.TypeString,
		RequiredForImport: true,
	},
}

func TestRegionalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalARNSchema, map[string]any{})
	rd.SetId("not a valid ARN")

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if !strings.HasPrefix(err.Error(), "could not parse import ID") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_ImportID_Invalid_WrongRegion(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalARNSchema, map[string]any{
		"region": "another-region-1",
	})
	rd.SetId(arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "a-region-1",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String())

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if !strings.HasPrefix(err.Error(), "the region passed for import") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_ImportID_Valid_DefaultRegion(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalARNSchema, map[string]any{})
	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd.SetId(arn)

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_ImportID_Valid_RegionOverride(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	rd := schema.TestResourceDataRaw(t, regionalARNSchema, map[string]any{
		"region": region,
	})
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd.SetId(arn)

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_Identity_Invalid_AttributeNotSet(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataWithIdentityRaw(t, regionalARNSchema, regionalARNIdentitySchema, map[string]string{})

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if err.Error() != fmt.Sprintf("identity attribute %q is required", "arn") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataWithIdentityRaw(t, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
		"arn": "not a valid ARN",
	})

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if !strings.HasPrefix(err.Error(), fmt.Sprintf("identity attribute %q: could not parse", "arn")) {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalARN_Identity_Valid(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalARNSchema, map[string]any{})
	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd.SetId(arn)

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id", "attr"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalARNSchema, regionalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	err := importer.RegionalARN(context.Background(), rd, "arn", []string{"id", "attr"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

var globalARNSchema = map[string]*schema.Schema{
	"arn": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
}

// lintignore:S013 // Identity Schemas cannot specify Computed, Optional, or Required
var globalARNIdentitySchema = map[string]*schema.Schema{
	"arn": {
		Type:              schema.TypeString,
		RequiredForImport: true,
	},
}

func TestGlobalARN_ImportID_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, globalARNSchema, map[string]any{})
	rd.SetId("not a valid ARN")

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if !strings.HasPrefix(err.Error(), "could not parse import ID") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestGlobalARN_ImportID_Valid(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, globalARNSchema, map[string]any{})
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd.SetId(arn)

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestGlobalARN_Identity_Invalid_AttributeNotSet(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataWithIdentityRaw(t, globalARNSchema, globalARNIdentitySchema, map[string]string{})

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if err.Error() != fmt.Sprintf("identity attribute %q is required", "arn") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestGlobalARN_Identity_Invalid_NotAnARN(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataWithIdentityRaw(t, globalARNSchema, globalARNIdentitySchema, map[string]string{
		"arn": "not a valid ARN",
	})

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		if !strings.HasPrefix(err.Error(), fmt.Sprintf("identity attribute %q: could not parse", "arn")) {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestGlobalARN_Identity_Valid(t *testing.T) {
	t.Parallel()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd := schema.TestResourceDataWithIdentityRaw(t, globalARNSchema, globalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestGlobalARN_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, globalARNSchema, map[string]any{})
	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd.SetId(arn)

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id", "attr"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestGlobalARN_DuplicateAttrs_Identity_Valid(t *testing.T) {
	t.Parallel()

	arn := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: "123456789012",
		Resource:  "res-abc123",
	}.String()
	rd := schema.TestResourceDataWithIdentityRaw(t, globalARNSchema, globalARNIdentitySchema, map[string]string{
		"arn": arn,
	})

	err := importer.GlobalARN(context.Background(), rd, "arn", []string{"id", "attr"})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := arn, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("arn"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := arn, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}
