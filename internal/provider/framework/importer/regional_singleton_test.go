// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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
		"attr": schema.StringAttribute{
			Optional: true,
		},
	},
}

var regionalSingletonWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id":     framework.IDAttributeDeprecatedNoReplacement(),
		"region": resourceattribute.Region(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
	},
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

	f := importer.RegionalSingleton

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		importMethod        string // "ImportID", "Identity", or "IDWithState"
		inputRegion         string
		inputAccountID      string
		duplicateAttrs      []string
		useSchemaWithID     bool
		stateAttrs          map[string]string
		noIdentity          bool
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
			importMethod: "IDWithState",
			inputRegion:  region,
			stateAttrs: map[string]string{
				"region": region,
			},
			expectedRegion: region,
			expectError:    false,
		},
		"ImportID_Valid_NoIdentity": {
			importMethod:   "ImportID",
			inputRegion:    region,
			noIdentity:     true,
			expectedRegion: region,
			expectError:    false,
		},
		"ImportID_Invalid_WrongRegion": {
			importMethod: "IDWithState",
			inputRegion:  region,
			stateAttrs: map[string]string{
				"region": "another-region-1",
			},
			expectError:         true,
			expectedErrorPrefix: "The region passed for import,",
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
			expectedErrorPrefix: "Provider configured with Account ID",
		},

		"DuplicateAttrs_ImportID_Valid": {
			importMethod:    "ImportID",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputRegion:     region,
			expectedRegion:  region,
			expectError:     false,
		},

		"DuplicateAttrs_Identity_Valid": {
			importMethod:    "Identity",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputRegion:     region,
			expectedRegion:  region,
			expectError:     false,
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

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(identitySpec))
			}

			schema := regionalSingletonSchema
			if tc.useSchemaWithID {
				schema = regionalSingletonWithIDSchema
			}

			var response resource.ImportStateResponse
			switch tc.importMethod {
			case "ImportID":
				response = importByID(ctx, f, &client, schema, tc.inputRegion, identitySchema, identitySpec)
			case "IDWithState":
				response = importByIDWithState(ctx, f, &client, schema, tc.inputRegion, tc.stateAttrs, identitySchema, identitySpec)
			case "Identity":
				identityAttrs := make(map[string]string, 2)
				if tc.inputRegion != "" {
					identityAttrs["region"] = tc.inputRegion
				}
				if tc.inputAccountID != "" {
					identityAttrs["account_id"] = tc.inputAccountID
				}
				identity := identityFromSchema(ctx, identitySchema, identityAttrs)
				response = importByIdentity(ctx, f, &client, schema, identity, identitySpec)
			}

			if tc.expectError {
				if !response.Diagnostics.HasError() {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorPrefix != "" && !strings.HasPrefix(response.Diagnostics[0].Detail(), tc.expectedErrorPrefix) {
					t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
				}
				return
			}

			if response.Diagnostics.HasError() {
				t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
			}

			// Check attr value
			var expectedAttrValue string
			if tc.useSchemaWithID && slices.Contains(tc.duplicateAttrs, "attr") {
				expectedAttrValue = tc.expectedRegion
			}
			if e, a := expectedAttrValue, getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
				t.Errorf("expected `attr` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
					t.Errorf("expected `id` to be %q, got %q", e, a)
				}
			}

			// Check identity
			if tc.noIdentity {
				if response.Identity != nil {
					t.Error("Identity should not be set")
				}
			} else {
				if identity := response.Identity; identity == nil {
					t.Error("Identity should be set")
				} else {
					if e, a := accountID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("account_id")); e != a {
						t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
					}
					if e, a := tc.expectedRegion, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("region")); e != a {
						t.Errorf("expected Identity `region` to be %q, got %q", e, a)
					}
				}
			}
		})
	}
}
