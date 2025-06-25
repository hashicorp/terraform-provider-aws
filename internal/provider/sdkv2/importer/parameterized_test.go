// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalParameterizedSchema = map[string]*schema.Schema{
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

			d := schema.TestResourceDataRaw(t, regionalParameterizedSchema, map[string]any{
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
			d := schema.TestResourceDataWithIdentityRaw(t, regionalParameterizedSchema, identitySchema, tc.identityAttrs)

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

var globalParameterizedSchema = map[string]*schema.Schema{
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

			d := schema.TestResourceDataRaw(t, globalParameterizedSchema, map[string]any{})
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
			d := schema.TestResourceDataWithIdentityRaw(t, globalParameterizedSchema, identitySchema, tc.identityAttrs)

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
