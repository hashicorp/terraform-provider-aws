// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/internal/attribute"
)

var regionalSingletonSchema = map[string]*schema.Schema{
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"region": attribute.Region(),
}

var regionalSingletonIdentitySchema = map[string]*schema.Schema{
	"region": {
		Type:              schema.TypeString,
		OptionalForImport: true,
	},
}

type mockClient struct {
	accountID string
	region    string
}

func (c mockClient) AccountID(ctx context.Context) string {
	return c.accountID
}

func (c mockClient) Region(ctx context.Context) string {
	return c.region
}

func TestRegionalSingleton_ImportID_Invalid_WrongRegion(t *testing.T) {
	region := "a-region-1"

	rd := schema.TestResourceDataRaw(t, regionalSingletonSchema, map[string]any{
		"region": "another-region-1",
	})
	rd.SetId(region)

	err := importer.RegionalSingleton(context.Background(), rd, nil)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "the region passed for import") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalSingleton_ImportID_Valid_NoRegionSet(t *testing.T) {
	region := "a-region-1"

	rd := schema.TestResourceDataRaw(t, regionalSingletonSchema, map[string]any{})
	rd.SetId(region)

	err := importer.RegionalSingleton(context.Background(), rd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_ImportID_Valid_RegionSet(t *testing.T) {
	region := "a-region-1"

	rd := schema.TestResourceDataRaw(t, regionalSingletonSchema, map[string]any{
		"region": region,
	})
	rd.SetId(region)

	err := importer.RegionalSingleton(context.Background(), rd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Invalid_WrongAccountID(t *testing.T) {
	region := "a-region-1"
	client := mockClient{
		accountID: "123456789012",
		region:    region,
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{
		"account_id": "123450054321",
	})

	err := importer.RegionalSingleton(context.Background(), rd, client)
	if err != nil {
		if !strings.Contains(err.Error(), "Provider configured with Account ID") {
			t.Fatalf("Unexpected error: %s", err)
		}
	}

}

func TestRegionalSingleton_Identity_Valid_AttributesNotSet(t *testing.T) {
	region := "a-region-1"
	client := mockClient{
		accountID: "123456789012",
		region:    region,
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{})

	err := importer.RegionalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Valid_AccountIDSet(t *testing.T) {
	accountID := "123456789012"
	region := "a-region-1"
	client := mockClient{
		accountID: accountID,
		region:    region,
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{
		"account_id": accountID,
	})

	err := importer.RegionalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Valid_RegionSet(t *testing.T) {
	accountID := "123456789012"
	region := "a-region-1"
	client := mockClient{
		accountID: accountID,
		region:    "another-region-1",
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{
		"region": region,
	})

	err := importer.RegionalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

func TestRegionalSingleton_Identity_Valid_AccountIDAndRegionSet(t *testing.T) {
	accountID := "123456789012"
	region := "a-region-1"
	client := mockClient{
		accountID: accountID,
		region:    "another-region-1",
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, regionalSingletonIdentitySchema, map[string]string{
		"account_id": accountID,
		"region":     region,
	})

	err := importer.RegionalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := region, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
}

var globalSingletonSchema = map[string]*schema.Schema{
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
}

var globalSingletonIdentitySchema = map[string]*schema.Schema{
	"account_id": {
		Type:              schema.TypeString,
		OptionalForImport: true,
	},
}

func TestGlobalSingleton_ImportID_Valid_AcceptsAnything(t *testing.T) {
	rd := schema.TestResourceDataRaw(t, globalSingletonSchema, map[string]any{})
	rd.SetId("a value")

	err := importer.GlobalSingleton(context.Background(), rd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if rd.Id() == "" {
		t.Error("expected `id`")
	}
}

func TestGlobalSingleton_ImportID_Valid_AccountID(t *testing.T) {
	accountID := "123456789012"

	rd := schema.TestResourceDataRaw(t, globalSingletonSchema, map[string]any{})
	rd.SetId(accountID)

	err := importer.GlobalSingleton(context.Background(), rd, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := accountID, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
}

func TestGlobalSingleton_Identity_Valid_AttributeNotSet(t *testing.T) {
	accountID := "123456789012"
	client := mockClient{
		accountID: accountID,
		region:    "a-region-1",
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, globalSingletonSchema, globalSingletonIdentitySchema, map[string]string{})

	err := importer.GlobalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := accountID, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
}

func TestGlobalSingleton_Identity_Valid_AccountIDSet(t *testing.T) {
	accountID := "123456789012"
	region := "a-region-1"
	client := mockClient{
		accountID: accountID,
		region:    region,
	}
	rd := schema.TestResourceDataWithIdentityRaw(t, globalSingletonSchema, globalSingletonIdentitySchema, map[string]string{
		"account_id": accountID,
	})

	err := importer.GlobalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := accountID, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
}
