// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalSingletonSchema = map[string]*schema.Schema{
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"region": attribute.Region(),
}

// lintignore:S013 // Identity Schemas cannot specify Computed, Optional, or Required
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

func regionalSingletonIdentitySpec(attrs ...string) inttypes.Identity {
	var opts []inttypes.IdentityOptsFunc
	if len(attrs) > 0 {
		opts = append(opts, inttypes.WithIdentityDuplicateAttrs(attrs...))
	}
	return inttypes.RegionalSingletonIdentity(opts...)
}

func TestRegionalSingleton(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		importMethod        string // "ImportID" or "Identity"
		inputRegion         string
		inputAccountID      string
		duplicateAttrs      []string
		stateAttrs          map[string]any
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"ImportID_Valid_DefaultRegion": {
			importMethod:   "ImportID",
			inputRegion:    "",
			expectedRegion: region,
			expectError:    false,
		},
		"ImportID_Valid_RegionOverride": {
			importMethod: "ImportID",
			inputRegion:  region,
			stateAttrs: map[string]any{
				"region": region,
			},
			expectedRegion: region,
			expectError:    false,
		},
		"ImportID_Invalid_WrongRegion": {
			importMethod: "ImportID",
			inputRegion:  region,
			stateAttrs: map[string]any{
				"region": "another-region-1",
			},
			expectError:         true,
			expectedErrorPrefix: "the region passed for import",
		},

		"Identity_Valid_ExplicitRegion": {
			importMethod:   "Identity",
			inputRegion:    region,
			inputAccountID: "",
			expectedRegion: region,
			expectError:    false,
		},
		"Identity_Valid_ExplicitAccountID": {
			importMethod:   "Identity",
			inputRegion:    "",
			inputAccountID: accountID,
			expectedRegion: region,
			expectError:    false,
		},
		"Identity_Valid_ExplicitRegionAndAccountID": {
			importMethod:   "Identity",
			inputRegion:    region,
			inputAccountID: accountID,
			expectedRegion: region,
			expectError:    false,
		},
		"Identity_Valid_NoExplicitAttributes": {
			importMethod:   "Identity",
			inputRegion:    "",
			inputAccountID: "",
			expectedRegion: region,
			expectError:    false,
		},
		"Identity_Invalid_WrongAccountID": {
			importMethod:        "Identity",
			inputAccountID:      "987654321098",
			expectedRegion:      region,
			expectError:         true,
			expectedErrorPrefix: fmt.Sprintf("identity attribute \"account_id\": Provider configured with Account ID"),
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

			var (
				d   *schema.ResourceData
				err error
			)
			switch tc.importMethod {
			case "ImportID":
				d = schema.TestResourceDataRaw(t, regionalSingletonSchema, tc.stateAttrs)
				d.SetId(region)

				err = importer.RegionalSingleton(ctx, d, client)

			case "Identity":
				identitySpec := regionalSingletonIdentitySpec(tc.duplicateAttrs...)
				identitySchema := identity.NewIdentitySchema(identitySpec)
				identityAttrs := make(map[string]string, 2)
				if tc.inputRegion != "" {
					identityAttrs["region"] = tc.inputRegion
				}
				if tc.inputAccountID != "" {
					identityAttrs["account_id"] = tc.inputAccountID
				}
				d = schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, identitySchema, identityAttrs)

				err = importer.RegionalSingleton(ctx, d, client)
			}
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

			// Check attr value
			var expectedAttrValue string
			if slices.Contains(tc.duplicateAttrs, "attr") {
				expectedAttrValue = tc.expectedRegion
			}
			if e, a := expectedAttrValue, getAttributeValue(t, d, "attr"); e != a {
				t.Errorf("expected `attr` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "region"); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check ID value
			if e, a := tc.expectedRegion, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}
		})
	}
}

var globalSingletonSchema = map[string]*schema.Schema{
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
}

// lintignore:S013 // Identity Schemas cannot specify Computed, Optional, or Required
var globalSingletonIdentitySchema = map[string]*schema.Schema{
	"account_id": {
		Type:              schema.TypeString,
		OptionalForImport: true,
	},
}

func TestGlobalSingleton_ImportID_Valid_AcceptsAnything(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	client := mockClient{
		accountID: accountID,
		region:    "a-region-1",
	}
	rd := schema.TestResourceDataRaw(t, globalSingletonSchema, map[string]any{})
	rd.SetId("a value")

	err := importer.GlobalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := accountID, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
}

func TestGlobalSingleton_ImportID_Valid_AccountID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	client := mockClient{
		accountID: accountID,
		region:    "a-region-1",
	}
	rd := schema.TestResourceDataRaw(t, globalSingletonSchema, map[string]any{})
	rd.SetId(accountID)

	err := importer.GlobalSingleton(context.Background(), rd, client)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := accountID, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
}

func TestGlobalSingleton_Identity_Valid_AttributeNotSet(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func getAttributeValue(t *testing.T, d *schema.ResourceData, name string) string {
	t.Helper()

	if name == "id" {
		return d.Id()
	}
	return d.Get(name).(string)
}
