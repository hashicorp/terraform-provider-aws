// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
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
			inputRegion:    region,
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
			expectedErrorPrefix: "identity attribute \"account_id\": Provider configured with Account ID",
		},

		"DuplicateAttrs_ImportID_Valid": {
			importMethod:   "ImportID",
			duplicateAttrs: []string{"attr"},
			inputRegion:    region,
			expectedRegion: region,
			expectError:    false,
		},

		"DuplicateAttrs_Identity_Valid": {
			importMethod:   "Identity",
			duplicateAttrs: []string{"attr"},
			inputRegion:    region,
			expectedRegion: region,
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

			identitySpec := regionalSingletonIdentitySpec(tc.duplicateAttrs...)

			var d *schema.ResourceData
			switch tc.importMethod {
			case "ImportID":
				d = schema.TestResourceDataRaw(t, regionalSingletonSchema, tc.stateAttrs)
				d.SetId(tc.inputRegion)

			case "Identity":
				identitySchema := identity.NewIdentitySchema(identitySpec)
				identityAttrs := make(map[string]string, 2)
				if tc.inputRegion != "" {
					identityAttrs["region"] = tc.inputRegion
				}
				if tc.inputAccountID != "" {
					identityAttrs["account_id"] = tc.inputAccountID
				}
				d = schema.TestResourceDataWithIdentityRaw(t, regionalSingletonSchema, identitySchema, identityAttrs)
			}

			err := importer.RegionalSingleton(ctx, d, identitySpec, client)
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

func globalSingletonIdentitySpec(attrs ...string) inttypes.Identity {
	var opts []inttypes.IdentityOptsFunc
	if len(attrs) > 0 {
		opts = append(opts, inttypes.WithIdentityDuplicateAttrs(attrs...))
	}
	return inttypes.GlobalSingletonIdentity(opts...)
}

func TestGlobalSingleton(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		importMethod        string // "ImportID" or "Identity"
		inputAccountID      string
		duplicateAttrs      []string
		expectError         bool
		expectedErrorPrefix string
	}{
		"ImportID_Valid_AccountID": {
			importMethod:   "ImportID",
			inputAccountID: accountID,
			expectError:    false,
		},
		"ImportID_Valid_AcceptsAnything": {
			importMethod:   "ImportID",
			inputAccountID: "some value",
			expectError:    false,
		},

		"Identity_Valid_ExplicitAccountID": {
			importMethod:   "Identity",
			inputAccountID: accountID,
			expectError:    false,
		},
		"Identity_Valid_NoExplicitAttributes": {
			importMethod:   "Identity",
			inputAccountID: "",
			expectError:    false,
		},
		"Identity_Invalid_WrongAccountID": {
			importMethod:        "Identity",
			inputAccountID:      "987654321098",
			expectError:         true,
			expectedErrorPrefix: "identity attribute \"account_id\": Provider configured with Account ID",
		},

		"DuplicateAttrs_ImportID_Valid": {
			importMethod:   "ImportID",
			duplicateAttrs: []string{"attr"},
			inputAccountID: accountID,
			expectError:    false,
		},

		"DuplicateAttrs_Identity_Valid": {
			importMethod:   "Identity",
			duplicateAttrs: []string{"attr"},
			inputAccountID: accountID,
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

			identitySpec := globalSingletonIdentitySpec(tc.duplicateAttrs...)

			var d *schema.ResourceData
			switch tc.importMethod {
			case "ImportID":
				d = schema.TestResourceDataRaw(t, globalSingletonSchema, map[string]any{})
				d.SetId(tc.inputAccountID)

			case "Identity":
				identitySchema := identity.NewIdentitySchema(identitySpec)
				identityAttrs := make(map[string]string, 2)
				if tc.inputAccountID != "" {
					identityAttrs["account_id"] = tc.inputAccountID
				}
				d = schema.TestResourceDataWithIdentityRaw(t, globalSingletonSchema, identitySchema, identityAttrs)
			}

			err := importer.GlobalSingleton(ctx, d, identitySpec, client)
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
				expectedAttrValue = accountID
			}
			if e, a := expectedAttrValue, getAttributeValue(t, d, "attr"); e != a {
				t.Errorf("expected `attr` to be %q, got %q", e, a)
			}

			// Check ID value
			if e, a := accountID, getAttributeValue(t, d, "id"); e != a {
				t.Errorf("expected `id` to be %q, got %q", e, a)
			}
		})
	}
}

func getAttributeValue(t *testing.T, d *schema.ResourceData, name string) string {
	t.Helper()

	if name == "id" {
		return d.Id()
	}
	return d.Get(name).(string)
}
