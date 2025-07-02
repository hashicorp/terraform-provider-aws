// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalSingleParameterizedSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"region": attribute.Region(),
}

func regionalSingleParameterizedIdentitySpec(attrName string) inttypes.Identity {
	return inttypes.Identity{
		IsGlobalResource:  true,
		IdentityAttribute: attrName,
		Attributes: []inttypes.IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     "region",
				Required: false,
			},
			{
				Name:     attrName,
				Required: true,
			},
		},
	}
}

func TestRegionalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		attrName            string
		inputID             string
		inputRegion         string
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_DefaultRegion": {
			attrName:       "name",
			inputID:        "a_name",
			inputRegion:    region,
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_RegionOverride": {
			attrName:       "name",
			inputID:        "a_name",
			inputRegion:    anotherRegion,
			expectedRegion: anotherRegion,
			expectError:    false,
		},

		"ID_DefaultRegion": {
			attrName:       "id",
			inputID:        "a_name",
			inputRegion:    region,
			expectedRegion: region,
			expectError:    false,
		},
		"ID_RegionOverride": {
			attrName:       "id",
			inputID:        "a_name",
			inputRegion:    anotherRegion,
			expectedRegion: anotherRegion,
			expectError:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			d := schema.TestResourceDataRaw(t, regionalSingleParameterizedSchema, map[string]any{
				"region": tc.inputRegion,
			})
			d.SetId(tc.inputID)

			err := importer.RegionalSingleParameterized(ctx, d, tc.attrName, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			if e, a := tc.inputID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "region"); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check name value
			var expectedNameValue string
			if tc.attrName == "name" {
				expectedNameValue = tc.inputID
			}
			if e, a := expectedNameValue, getAttributeValue(t, d, "name"); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}
		})
	}
}

func TestRegionalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		attrName            string
		identityAttrs       map[string]string
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Required": {
			attrName: "name",
			identityAttrs: map[string]string{
				"name": "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithAccountID": {
			attrName: "name",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithDefaultRegion": {
			attrName: "name",
			identityAttrs: map[string]string{
				"region": region,
				"name":   "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithRegionOverride": {
			attrName: "name",
			identityAttrs: map[string]string{
				"region": anotherRegion,
				"name":   "a_name",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"Attr_WrongAccountID": {
			attrName: "name",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
			},
			expectError: true,
		},

		"ID_Required": {
			attrName: "id",
			identityAttrs: map[string]string{
				"id": "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"ID_WithAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"ID_WithDefaultRegion": {
			attrName: "id",
			identityAttrs: map[string]string{
				"region": region,
				"id":     "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"ID_WithRegionOverride": {
			attrName: "id",
			identityAttrs: map[string]string{
				"region": anotherRegion,
				"id":     "a_name",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"ID_WrongAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := regionalSingleParameterizedIdentitySpec(tc.attrName)

			identitySchema := identity.NewIdentitySchema(identitySpec)
			d := schema.TestResourceDataWithIdentityRaw(t, regionalSingleParameterizedSchema, identitySchema, tc.identityAttrs)

			err := importer.RegionalSingleParameterized(ctx, d, tc.attrName, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			// ID must always be set for SDKv2 resources
			if e, a := tc.identityAttrs[tc.attrName], getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "region"); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check name value
			var expectedNameValue string
			if tc.attrName == "name" {
				expectedNameValue = tc.identityAttrs["name"]
			}
			if e, a := expectedNameValue, getAttributeValue(t, d, "name"); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}
		})
	}
}

var globalSingleParameterizedSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
}

func globalSingleParameterizedIdentitySpec(attrName string) inttypes.Identity {
	return inttypes.Identity{
		IsGlobalResource:  true,
		IdentityAttribute: attrName,
		Attributes: []inttypes.IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     attrName,
				Required: true,
			},
		},
	}
}

func TestGlobalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		attrName            string
		inputID             string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr": {
			attrName:    "name",
			inputID:     "a_name",
			expectError: false,
		},
		"ID": {
			attrName:    "id",
			inputID:     "a_name",
			expectError: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			d := schema.TestResourceDataRaw(t, globalSingleParameterizedSchema, map[string]any{})
			d.SetId(tc.inputID)

			err := importer.GlobalSingleParameterized(ctx, d, tc.attrName, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			if e, a := tc.inputID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check name value
			var expectedNameValue string
			if tc.attrName == "name" {
				expectedNameValue = tc.inputID
			}
			if e, a := expectedNameValue, getAttributeValue(t, d, "name"); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}
		})
	}
}

func TestGlobalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		attrName            string
		identityAttrs       map[string]string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Required": {
			attrName: "name",
			identityAttrs: map[string]string{
				"name": "a_name",
			},
			expectError: false,
		},
		"Attr_WithAccountID": {
			attrName: "name",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
			},
			expectError: false,
		},
		"Attr_WrongAccountID": {
			attrName: "name",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
			},
			expectError: true,
		},

		"ID_Required": {
			attrName: "id",
			identityAttrs: map[string]string{
				"id": "a_name",
			},
			expectError: false,
		},
		"ID_WithAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			expectError: false,
		},
		"ID_WrongAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := globalSingleParameterizedIdentitySpec(tc.attrName)

			identitySchema := identity.NewIdentitySchema(identitySpec)
			d := schema.TestResourceDataWithIdentityRaw(t, globalSingleParameterizedSchema, identitySchema, tc.identityAttrs)

			err := importer.GlobalSingleParameterized(ctx, d, tc.attrName, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			// ID must always be set for SDKv2 resources
			if e, a := tc.identityAttrs[tc.attrName], getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check name value
			var expectedNameValue string
			if tc.attrName == "name" {
				expectedNameValue = tc.identityAttrs["name"]
			}
			if e, a := expectedNameValue, getAttributeValue(t, d, "name"); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}
		})
	}
}

var regionalMultipleParameterizedSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"type": {
		Type:     schema.TypeString,
		Required: true,
	},
	"region": attribute.Region(),
}

func regionalMultipleParameterizedIdentitySpec(attrNames []string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for _, attrName := range attrNames {
		attrs = append(attrs, inttypes.StringIdentityAttribute(attrName, true))
	}
	return inttypes.RegionalParameterizedIdentity(attrs...)
}

func TestRegionalMutipleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		inputID             string
		inputRegion         string
		expectedAttrs       map[string]string
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"DefaultRegion": {
			inputID:     "a_name,a_type",
			inputRegion: region,
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"RegionOverride": {
			inputID:     "a_name,a_type",
			inputRegion: anotherRegion,
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := regionalMultipleParameterizedIdentitySpec([]string{"name", "type"})

			importSpec := inttypes.SDKv2Import{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}

			d := schema.TestResourceDataRaw(t, regionalMultipleParameterizedSchema, map[string]any{
				"region": tc.inputRegion,
			})
			d.SetId(tc.inputID)

			err := importer.RegionalMultipleParameterized(ctx, d, identitySpec.Attributes, &importSpec, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			if e, a := tc.inputID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "region"); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedAttrs {
				if e, a := expectedAttr, getAttributeValue(t, d, name); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}
		})
	}
}

func TestRegionalMutipleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		identityAttrs       map[string]string
		expectedAttrs       map[string]string
		expectedID          string
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Required": {
			identityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:     "a_name,a_type",
			expectedRegion: region,
			expectError:    false,
		},
		"WithAccountID": {
			identityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:     "a_name,a_type",
			expectedRegion: region,
			expectError:    false,
		},
		"WithDefaultRegion": {
			identityAttrs: map[string]string{
				"region": region,
				"name":   "a_name",
				"type":   "a_type",
			},
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:     "a_name,a_type",
			expectedRegion: region,
			expectError:    false,
		},
		"WithRegionOverride": {
			identityAttrs: map[string]string{
				"region": anotherRegion,
				"name":   "a_name",
				"type":   "a_type",
			},
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:     "a_name,a_type",
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"WrongAccountID": {
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
				"type":       "a_type",
			},
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := regionalMultipleParameterizedIdentitySpec([]string{"name", "type"})

			importSpec := inttypes.SDKv2Import{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}

			identitySchema := identity.NewIdentitySchema(identitySpec)
			d := schema.TestResourceDataWithIdentityRaw(t, regionalMultipleParameterizedSchema, identitySchema, tc.identityAttrs)

			err := importer.RegionalMultipleParameterized(ctx, d, identitySpec.Attributes, &importSpec, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			// ID must always be set for SDKv2 resources
			if e, a := tc.expectedID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "region"); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedAttrs {
				if e, a := expectedAttr, getAttributeValue(t, d, name); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}
		})
	}
}

var globalMultipleParameterizedSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"type": {
		Type:     schema.TypeString,
		Required: true,
	},
}

func globalMultipleParameterizedIdentitySpec(attrNames []string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for _, attrName := range attrNames {
		attrs = append(attrs, inttypes.StringIdentityAttribute(attrName, true))
	}
	return inttypes.GlobalParameterizedIdentity(attrs...)
}

func TestGlobalMutipleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		inputID             string
		expectedAttrs       map[string]string
		expectError         bool
		expectedErrorPrefix string
	}{
		"ID": {
			inputID: "a_name,a_type",
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := globalMultipleParameterizedIdentitySpec([]string{"name", "type"})

			importSpec := inttypes.SDKv2Import{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}

			d := schema.TestResourceDataRaw(t, globalMultipleParameterizedSchema, map[string]any{})
			d.SetId(tc.inputID)

			err := importer.GlobalMultipleParameterized(ctx, d, identitySpec.Attributes, &importSpec, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			if e, a := tc.inputID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedAttrs {
				if e, a := expectedAttr, getAttributeValue(t, d, name); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}
		})
	}
}

func TestGlobalMutipleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		identityAttrs       map[string]string
		expectedID          string
		expectedAttrs       map[string]string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Identity": {
			identityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID: "a_name,a_type",
			expectedAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
				region:    region,
			}

			identitySpec := globalMultipleParameterizedIdentitySpec([]string{"name", "type"})

			importSpec := inttypes.SDKv2Import{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}

			identitySchema := identity.NewIdentitySchema(identitySpec)
			d := schema.TestResourceDataWithIdentityRaw(t, globalMultipleParameterizedSchema, identitySchema, tc.identityAttrs)

			err := importer.GlobalMultipleParameterized(ctx, d, identitySpec.Attributes, &importSpec, client)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(err.Error(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}

			// Check ID value
			if e, a := tc.expectedID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedAttrs {
				if e, a := expectedAttr, getAttributeValue(t, d, name); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}
		})
	}
}

var _ inttypes.SDKv2ImportID = testImportID{}

type testImportID struct {
	t *testing.T
}

func (t testImportID) Create(d *schema.ResourceData) string {
	t.t.Helper()

	idParts := []string{
		d.Get("name").(string),
		d.Get("type").(string),
	}
	result, err := flex.FlattenResourceId(idParts, len(idParts), false)
	if err != nil {
		t.t.Fatalf("Creating test Import ID: %s", err)
	}

	return result
}

func (t testImportID) Parse(id string) (string, map[string]string, error) {
	t.t.Helper()

	parts, err := flex.ExpandResourceId(id, 2, false)
	if err != nil {
		t.t.Fatalf("Parsing test Import ID: %s", err)
	}

	return id, map[string]string{
		"name": parts[0],
		"type": parts[1],
	}, nil
}
