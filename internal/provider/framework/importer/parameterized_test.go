// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"region": resourceattribute.Region(),
	},
}

var regionalParameterizedWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": framework.IDAttributeDeprecatedNoReplacement(),
		"name": schema.StringAttribute{
			Required: true,
		},
		"region": resourceattribute.Region(),
	},
}

func regionalSingleParameterIdentitySpec(name string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentity(name)
}

func TestRegionalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	f := importer.RegionalSingleParameterized

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		attrName            string
		inputID             string
		inputRegion         string
		useSchemaWithID     bool
		noIdentity          bool
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
		"Attr_NoIdentity": {
			attrName:       "name",
			inputID:        "a_name",
			inputRegion:    region,
			noIdentity:     true,
			expectedRegion: region,
			expectError:    false,
		},

		"ID_DefaultRegion": {
			attrName:        "id",
			inputID:         "a_name",
			inputRegion:     region,
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_RegionOverride": {
			attrName:        "id",
			inputID:         "a_name",
			inputRegion:     anotherRegion,
			useSchemaWithID: true,
			expectedRegion:  anotherRegion,
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

			stateAttrs := map[string]string{
				"region": tc.inputRegion,
			}

			identitySpec := regionalSingleParameterIdentitySpec(tc.attrName)

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(identitySpec))
			}

			schema := regionalParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalParameterizedWithIDSchema
			}

			response := importByIDWithState(ctx, f, &client, schema, tc.inputID, stateAttrs, identitySchema, identitySpec)
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

			// Check name value
			var expectedNameValue string
			if !tc.useSchemaWithID {
				expectedNameValue = tc.inputID
			}
			if e, a := expectedNameValue, getAttributeValue(ctx, t, response.State, path.Root("name")); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.inputID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					if e, a := tc.inputID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.attrName)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
					}
				}
			}
		})
	}
}

func TestRegionalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.RegionalSingleParameterized

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		attrName            string
		identityAttrs       map[string]string
		useSchemaWithID     bool
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
			expectedRegion: region,
			expectError:    true,
		},

		"ID_Required": {
			attrName: "id",
			identityAttrs: map[string]string{
				"id": "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithDefaultRegion": {
			attrName: "id",
			identityAttrs: map[string]string{
				"region": region,
				"id":     "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithRegionOverride": {
			attrName: "id",
			identityAttrs: map[string]string{
				"region": anotherRegion,
				"id":     "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  anotherRegion,
			expectError:     false,
		},
		"ID_WrongAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     true,
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

			identitySpec := regionalSingleParameterIdentitySpec(tc.attrName)

			identitySchema := ptr(identity.NewIdentitySchema(identitySpec))

			schema := regionalParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalParameterizedWithIDSchema
			}
			// identityAttrs := make(map[string]string, 2)
			// if tc.inputRegion != "" {
			// 	identityAttrs["region"] = tc.inputRegion
			// }
			// if tc.inputAccountID != "" {
			// 	identityAttrs["account_id"] = tc.inputAccountID
			// }
			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrs)

			response := importByIdentity(ctx, f, &client, schema, identity, identitySpec)
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

			// Check name value
			var expectedNameValue string
			if !tc.useSchemaWithID {
				expectedNameValue = tc.identityAttrs[tc.attrName]
			}
			if e, a := expectedNameValue, getAttributeValue(ctx, t, response.State, path.Root("name")); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.identityAttrs[tc.attrName], getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
					t.Errorf("expected `id` to be %q, got %q", e, a)
				}
			}

			// Check identity
			if identity := response.Identity; identity == nil {
				t.Error("Identity should be set")
			} else {
				if e, a := accountID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("account_id")); e != a {
					t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
				}
				if e, a := tc.expectedRegion, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("region")); e != a {
					t.Errorf("expected Identity `region` to be %q, got %q", e, a)
				}
				if e, a := tc.identityAttrs[tc.attrName], getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.attrName)); e != a {
					t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
				}
			}
		})
	}
}

var globalParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
	},
}

var globalParameterizedWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": framework.IDAttributeDeprecatedNoReplacement(),
		"name": schema.StringAttribute{
			Required: true,
		},
	},
}

func globalSingleParameterIdentitySpec(name string) inttypes.Identity {
	return inttypes.GlobalSingleParameterIdentity(name)
}

func TestGlobalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	f := importer.GlobalSingleParameterized

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		attrName            string
		inputID             string
		useSchemaWithID     bool
		noIdentity          bool
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Basic": {
			attrName:    "name",
			inputID:     "a_name",
			expectError: false,
		},
		"Attr_NoIdentity": {
			attrName:    "name",
			inputID:     "a_name",
			noIdentity:  true,
			expectError: false,
		},

		"ID_Basic": {
			attrName:        "id",
			inputID:         "a_name",
			useSchemaWithID: true,
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

			stateAttrs := map[string]string{}

			identitySpec := globalSingleParameterIdentitySpec(tc.attrName)

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(identitySpec))
			}

			schema := globalParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalParameterizedWithIDSchema
			}

			response := importByIDWithState(ctx, f, &client, schema, tc.inputID, stateAttrs, identitySchema, identitySpec)
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

			// Check name value
			var expectedNameValue string
			if !tc.useSchemaWithID {
				expectedNameValue = tc.inputID
			}
			if e, a := expectedNameValue, getAttributeValue(ctx, t, response.State, path.Root("name")); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.inputID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					if e, a := tc.inputID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.attrName)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
					}
				}
			}
		})
	}
}

func TestGlobalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.GlobalSingleParameterized

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		attrName            string
		identityAttrs       map[string]string
		useSchemaWithID     bool
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
			useSchemaWithID: true,
			expectError:     false,
		},
		"ID_WithAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectError:     false,
		},
		"ID_WrongAccountID": {
			attrName: "id",
			identityAttrs: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectError:     true,
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

			identitySpec := globalSingleParameterIdentitySpec(tc.attrName)

			identitySchema := ptr(identity.NewIdentitySchema(identitySpec))

			schema := globalParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalParameterizedWithIDSchema
			}

			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrs)

			response := importByIdentity(ctx, f, &client, schema, identity, identitySpec)
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

			// Check name value
			var expectedNameValue string
			if !tc.useSchemaWithID {
				expectedNameValue = tc.identityAttrs[tc.attrName]
			}
			if e, a := expectedNameValue, getAttributeValue(ctx, t, response.State, path.Root("name")); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.identityAttrs[tc.attrName], getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
					t.Errorf("expected `id` to be %q, got %q", e, a)
				}
			}

			// Check identity
			if identity := response.Identity; identity == nil {
				t.Error("Identity should be set")
			} else {
				if e, a := accountID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("account_id")); e != a {
					t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
				}
				if e, a := tc.identityAttrs[tc.attrName], getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.attrName)); e != a {
					t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
				}
			}
		})
	}
}
