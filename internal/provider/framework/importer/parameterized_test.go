// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalSingleParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"region": resourceattribute.Region(),
	},
}

var regionalSingleParameterizedWithIDSchema = schema.Schema{
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

func regionalSingleParameterIdentitySpecNameMapped(identityAttrName, resourceAttrName string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentityWithMappedName(identityAttrName, resourceAttrName)
}

func regionalSingleParameterIdentitySpecWithDuplicates(name string, duplicateAttrs []string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentity(name,
		inttypes.WithIdentityDuplicateAttrs(duplicateAttrs...),
	)
}

func TestRegionalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	f := importer.SingleParameterized

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		identityAttrName    string
		resourceAttrName    string
		duplicateAttrs      []string
		inputID             string
		inputRegion         string
		useSchemaWithID     bool
		noIdentity          bool
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_DefaultRegion": {
			identityAttrName: "name",
			inputID:          "a_name",
			inputRegion:      region,
			expectedRegion:   region,
			expectError:      false,
		},
		"Attr_RegionOverride": {
			identityAttrName: "name",
			inputID:          "a_name",
			inputRegion:      anotherRegion,
			expectedRegion:   anotherRegion,
			expectError:      false,
		},
		"Attr_NoIdentity": {
			identityAttrName: "name",
			inputID:          "a_name",
			inputRegion:      region,
			noIdentity:       true,
			expectedRegion:   region,
			expectError:      false,
		},
		"Attr_NameMapped": {
			identityAttrName: "id_name",
			resourceAttrName: "name",
			inputID:          "a_name",
			inputRegion:      region,
			expectedRegion:   region,
			expectError:      false,
		},

		"ID_DefaultRegion": {
			identityAttrName: "id",
			inputID:          "a_name",
			inputRegion:      region,
			useSchemaWithID:  true,
			expectedRegion:   region,
			expectError:      false,
		},
		"ID_RegionOverride": {
			identityAttrName: "id",
			inputID:          "a_name",
			inputRegion:      anotherRegion,
			useSchemaWithID:  true,
			expectedRegion:   anotherRegion,
			expectError:      false,
		},

		"Attr_DuplicateID": {
			identityAttrName: "name",
			duplicateAttrs:   []string{"id"},
			inputID:          "a_name",
			inputRegion:      region,
			useSchemaWithID:  true,
			expectedRegion:   region,
			expectError:      false,
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

			var identitySpec inttypes.Identity
			if len(tc.duplicateAttrs) > 0 {
				identitySpec = regionalSingleParameterIdentitySpecWithDuplicates(tc.identityAttrName, tc.duplicateAttrs)
			} else if tc.resourceAttrName == "" || tc.resourceAttrName == tc.identityAttrName {
				identitySpec = regionalSingleParameterIdentitySpec(tc.identityAttrName)
			} else {
				identitySpec = regionalSingleParameterIdentitySpecNameMapped(tc.identityAttrName, tc.resourceAttrName)
			}

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(identitySpec))
			}

			schema := regionalSingleParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalSingleParameterizedWithIDSchema
			}

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
			}

			response := importByIDWithState(ctx, f, &client, schema, tc.inputID, stateAttrs, identitySchema, identitySpec, &importSpec)
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
			if tc.identityAttrName != "id" {
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
					expectedIdentityAttrs := map[string]string{
						"account_id":        accountID,
						"region":            tc.expectedRegion,
						tc.identityAttrName: tc.inputID,
					}

					var obj types.Object
					if diags := identity.Get(ctx, &obj); diags.HasError() {
						t.Fatalf("Unexpected error getting identity attributes: %s", fwdiag.DiagnosticsError(diags))
					}

					actualIdentityAttrs := make(map[string]string)
					for attrName, attrValue := range obj.Attributes() {
						if v, ok := attrValue.(types.String); !ok {
							t.Fatalf("expected string attribute, had %T", attrValue)
						} else {
							actualIdentityAttrs[attrName] = v.ValueString()
						}
					}

					if diff := cmp.Diff(actualIdentityAttrs, expectedIdentityAttrs); diff != "" {
						t.Fatalf("Unexpected identity attributes (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestRegionalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.SingleParameterized

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		identityAttrName    string
		identityAttrValues  map[string]string
		resourceAttrName    string
		duplicateAttrs      []string
		useSchemaWithID     bool
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Required": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"name": "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithAccountID": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithDefaultRegion": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"region": region,
				"name":   "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Attr_WithRegionOverride": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"region": anotherRegion,
				"name":   "a_name",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"Attr_WrongAccountID": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
			},
			expectedRegion: region,
			expectError:    true,
		},

		"ID_Required": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"id": "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithAccountID": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithDefaultRegion": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"region": region,
				"id":     "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     false,
		},
		"ID_WithRegionOverride": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"region": anotherRegion,
				"id":     "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  anotherRegion,
			expectError:     false,
		},
		"ID_WrongAccountID": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectedRegion:  region,
			expectError:     true,
		},

		"name mapped": {
			identityAttrName: "id_name",
			resourceAttrName: "name",
			identityAttrValues: map[string]string{
				"id_name": "a_name",
			},
			expectedRegion: region,
			expectError:    false,
		},

		"Attr_DuplicateID": {
			identityAttrName: "name",
			duplicateAttrs:   []string{"id"},
			identityAttrValues: map[string]string{
				"name": "a_name",
			},
			useSchemaWithID: true,
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

			var identitySpec inttypes.Identity
			if len(tc.duplicateAttrs) > 0 {
				identitySpec = regionalSingleParameterIdentitySpecWithDuplicates(tc.identityAttrName, tc.duplicateAttrs)
			} else if tc.resourceAttrName == "" || tc.resourceAttrName == tc.identityAttrName {
				identitySpec = regionalSingleParameterIdentitySpec(tc.identityAttrName)
			} else {
				identitySpec = regionalSingleParameterIdentitySpecNameMapped(tc.identityAttrName, tc.resourceAttrName)
			}

			identitySchema := ptr(identity.NewIdentitySchema(identitySpec))

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
			}

			schema := regionalSingleParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalSingleParameterizedWithIDSchema
			}
			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrValues)

			response := importByIdentity(ctx, f, &client, schema, identity, identitySpec, &importSpec)
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
			if tc.identityAttrName != "id" {
				expectedNameValue = tc.identityAttrValues[tc.identityAttrName]
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
				if e, a := tc.identityAttrValues[tc.identityAttrName], getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
				if e, a := tc.identityAttrValues[tc.identityAttrName], getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.identityAttrName)); e != a {
					t.Errorf("expected Identity `%s` to be %q, got %q", tc.identityAttrName, e, a)
				}
			}
		})
	}
}

var globalSingleParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
	},
}

var globalSingleParameterizedWithIDSchema = schema.Schema{
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

func globalSingleParameterIdentitySpecNameMapped(identityAttrName, resourceAttrName string) inttypes.Identity {
	return inttypes.GlobalSingleParameterIdentityWithMappedName(identityAttrName, resourceAttrName)
}

func globalSingleParameterIdentitySpecWithDuplicates(name string, duplicateAttrs []string) inttypes.Identity {
	return inttypes.GlobalSingleParameterIdentity(name,
		inttypes.WithIdentityDuplicateAttrs(duplicateAttrs...),
	)
}

func TestGlobalSingleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	f := importer.SingleParameterized

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		identityAttrName    string
		resourceAttrName    string
		duplicateAttrs      []string
		inputID             string
		useSchemaWithID     bool
		noIdentity          bool
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Basic": {
			identityAttrName: "name",
			inputID:          "a_name",
			expectError:      false,
		},
		"Attr_NoIdentity": {
			identityAttrName: "name",
			inputID:          "a_name",
			noIdentity:       true,
			expectError:      false,
		},
		"Attr_NameMapped": {
			identityAttrName: "id_name",
			resourceAttrName: "name",
			inputID:          "a_name",
			expectError:      false,
		},

		"ID_Basic": {
			identityAttrName: "id",
			inputID:          "a_name",
			useSchemaWithID:  true,
			expectError:      false,
		},

		"Attr_DuplicateID": {
			identityAttrName: "name",
			duplicateAttrs:   []string{"id"},
			inputID:          "a_name",
			useSchemaWithID:  true,
			expectError:      false,
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

			var identitySpec inttypes.Identity
			if len(tc.duplicateAttrs) > 0 {
				identitySpec = globalSingleParameterIdentitySpecWithDuplicates(tc.identityAttrName, tc.duplicateAttrs)
			} else if tc.resourceAttrName == "" || tc.resourceAttrName == tc.identityAttrName {
				identitySpec = globalSingleParameterIdentitySpec(tc.identityAttrName)
			} else {
				identitySpec = globalSingleParameterIdentitySpecNameMapped(tc.identityAttrName, tc.resourceAttrName)
			}

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(identitySpec))
			}

			schema := globalSingleParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalSingleParameterizedWithIDSchema
			}

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
			}

			response := importByIDWithState(ctx, f, &client, schema, tc.inputID, stateAttrs, identitySchema, identitySpec, &importSpec)
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
			if tc.identityAttrName != "id" {
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
					expectedIdentityAttrs := map[string]string{
						"account_id":        accountID,
						tc.identityAttrName: tc.inputID,
					}

					var obj types.Object
					if diags := identity.Get(ctx, &obj); diags.HasError() {
						t.Fatalf("Unexpected error getting identity attributes: %s", fwdiag.DiagnosticsError(diags))
					}

					actualIdentityAttrs := make(map[string]string)
					for attrName, attrValue := range obj.Attributes() {
						if v, ok := attrValue.(types.String); !ok {
							t.Fatalf("expected string attribute, had %T", attrValue)
						} else {
							actualIdentityAttrs[attrName] = v.ValueString()
						}
					}

					if diff := cmp.Diff(actualIdentityAttrs, expectedIdentityAttrs); diff != "" {
						t.Fatalf("Unexpected identity attributes (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestGlobalSingleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.SingleParameterized

	accountID := "123456789012"
	region := "a-region-1"

	testCases := map[string]struct {
		identityAttrName    string
		identityAttrValues  map[string]string
		resourceAttrName    string
		duplicateAttrs      []string
		useSchemaWithID     bool
		expectError         bool
		expectedErrorPrefix string
	}{
		"Attr_Required": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"name": "a_name",
			},
			expectError: false,
		},
		"Attr_WithAccountID": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
			},
			expectError: false,
		},
		"Attr_WrongAccountID": {
			identityAttrName: "name",
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
			},
			expectError: true,
		},

		"ID_Required": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"id": "a_name",
			},
			useSchemaWithID: true,
			expectError:     false,
		},
		"ID_WithAccountID": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectError:     false,
		},
		"ID_WrongAccountID": {
			identityAttrName: "id",
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"id":         "a_name",
			},
			useSchemaWithID: true,
			expectError:     true,
		},

		"name mapped": {
			identityAttrName: "id_name",
			resourceAttrName: "name",
			identityAttrValues: map[string]string{
				"id_name": "a_name",
			},
			expectError: false,
		},

		"Attr_DuplicateID": {
			identityAttrName: "name",
			duplicateAttrs:   []string{"id"},
			identityAttrValues: map[string]string{
				"name": "a_name",
			},
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

			var identitySpec inttypes.Identity
			if len(tc.duplicateAttrs) > 0 {
				identitySpec = globalSingleParameterIdentitySpecWithDuplicates(tc.identityAttrName, tc.duplicateAttrs)
			} else if tc.resourceAttrName == "" || tc.resourceAttrName == tc.identityAttrName {
				identitySpec = globalSingleParameterIdentitySpec(tc.identityAttrName)
			} else {
				identitySpec = globalSingleParameterIdentitySpecNameMapped(tc.identityAttrName, tc.resourceAttrName)
			}

			identitySchema := ptr(identity.NewIdentitySchema(identitySpec))

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
			}

			schema := globalSingleParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalSingleParameterizedWithIDSchema
			}

			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrValues)

			response := importByIdentity(ctx, f, &client, schema, identity, identitySpec, &importSpec)
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
			if tc.identityAttrName != "id" {
				expectedNameValue = tc.identityAttrValues[tc.identityAttrName]
			}
			if e, a := expectedNameValue, getAttributeValue(ctx, t, response.State, path.Root("name")); e != a {
				t.Errorf("expected `name` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.identityAttrValues[tc.identityAttrName], getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
				if e, a := tc.identityAttrValues[tc.identityAttrName], getIdentityAttributeValue(ctx, t, response.Identity, path.Root(tc.identityAttrName)); e != a {
					t.Errorf("expected Identity `%s` to be %q, got %q", tc.identityAttrName, e, a)
				}
			}
		})
	}
}

var regionalMultipleParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
		"region": resourceattribute.Region(),
	},
}

var regionalMultipleParameterizedWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": framework.IDAttributeDeprecatedNoReplacement(),
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
		"region": resourceattribute.Region(),
	},
}

func regionalMultipleParameterizedIdentitySpec(attrNames []string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for _, attrName := range attrNames {
		attrs = append(attrs, inttypes.StringIdentityAttribute(attrName, true))
	}
	return inttypes.RegionalParameterizedIdentity(attrs)
}

func regionalMultipleParameterizedIdentitySpecWithMappedName(attrNames map[string]string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for identityAttrName, resourceAttrName := range attrNames {
		if identityAttrName == resourceAttrName {
			attrs = append(attrs, inttypes.StringIdentityAttribute(identityAttrName, true))
		} else {
			attrs = append(attrs, inttypes.StringIdentityAttributeWithMappedName(identityAttrName, true, resourceAttrName))
		}
	}
	return inttypes.RegionalParameterizedIdentity(attrs)
}

func TestRegionalMutipleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		inputID               string
		inputRegion           string
		identitySpec          inttypes.Identity
		useSchemaWithID       bool
		noIdentity            bool
		expectedResourceAttrs map[string]string
		expectedIdentityAttrs map[string]string
		expectedRegion        string
		expectedID            string
		expectError           bool
	}{
		"DefaultRegion": {
			inputID:      "a_name,a_type",
			inputRegion:  region,
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"RegionOverride": {
			inputID:      "a_name,a_type",
			inputRegion:  anotherRegion,
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"name mapped": {
			inputID:     "a_name,a_type",
			inputRegion: anotherRegion,
			identitySpec: regionalMultipleParameterizedIdentitySpecWithMappedName(map[string]string{
				"id_name": "name",
				"type":    "type",
			}),
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"id_name": "a_name",
				"type":    "a_type",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"NoIdentity": {
			inputID:      "a_name,a_type",
			inputRegion:  region,
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			noIdentity:   true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"Invalid": {
			inputID:      "invalid",
			inputRegion:  region,
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectError:  true,
		},

		"WithIDAttr_DefaultRegion": {
			inputID:         "a_name,a_type",
			inputRegion:     region,
			identitySpec:    regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID: true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectedID:     "a_name,a_type",
			expectError:    false,
		},
		"WithIDAttr_TrimmedID": {
			inputID:         "trim:a_name,a_type",
			inputRegion:     region,
			identitySpec:    regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID: true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectedID:     "a_name,a_type",
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

			stateAttrs := map[string]string{
				"region": tc.inputRegion,
			}

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(tc.identitySpec))
			}

			schema := regionalMultipleParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalMultipleParameterizedWithIDSchema
			}

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}
			if tc.useSchemaWithID {
				importSpec.SetIDAttr = true
			}

			response := importByIDWithState(ctx, importer.MultipleParameterized, &client, schema, tc.inputID, stateAttrs, identitySchema, tc.identitySpec, &importSpec)
			if tc.expectError {
				if !response.Diagnostics.HasError() {
					t.Fatal("Expected error, got none")
				}
				return
			}

			if response.Diagnostics.HasError() {
				t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
			}

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedResourceAttrs {
				if e, a := expectedAttr, getAttributeValue(ctx, t, response.State, path.Root(name)); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.expectedID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					expectedIdentityAttrs := tc.expectedIdentityAttrs
					expectedIdentityAttrs["account_id"] = accountID
					expectedIdentityAttrs["region"] = tc.expectedRegion

					var obj types.Object
					if diags := identity.Get(ctx, &obj); diags.HasError() {
						t.Fatalf("Unexpected error getting identity attributes: %s", fwdiag.DiagnosticsError(diags))
					}

					actualIdentityAttrs := make(map[string]string)
					for attrName, attrValue := range obj.Attributes() {
						if v, ok := attrValue.(types.String); !ok {
							t.Fatalf("expected string attribute, had %T", attrValue)
						} else {
							actualIdentityAttrs[attrName] = v.ValueString()
						}
					}

					if diff := cmp.Diff(actualIdentityAttrs, expectedIdentityAttrs); diff != "" {
						t.Fatalf("Unexpected identity attributes (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestRegionalMutipleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.MultipleParameterized

	accountID := "123456789012"
	region := "a-region-1"
	anotherRegion := "another-region-1"

	testCases := map[string]struct {
		identityAttrValues    map[string]string
		identitySpec          inttypes.Identity
		useSchemaWithID       bool
		useImportIDCreator    bool
		expectedIdentityAttrs map[string]string
		expectedResourceAttrs map[string]string
		expectedRegion        string
		expectedID            string
		expectError           bool
		expectedErrorPrefix   string
	}{
		"Required": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     region,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"WithAccountID": {
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     region,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"WithDefaultRegion": {
			identityAttrValues: map[string]string{
				"region": region,
				"name":   "a_name",
				"type":   "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     region,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectError:    false,
		},
		"WithRegionOverride": {
			identityAttrValues: map[string]string{
				"region": anotherRegion,
				"name":   "a_name",
				"type":   "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     anotherRegion,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"WrongAccountID": {
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
				"type":       "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectError:  true,
		},

		"WithIDAttr_DefaultRegion": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec:       regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID:    true,
			useImportIDCreator: true,
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     region,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedRegion: region,
			expectedID:     "a_name,a_type",
			expectError:    false,
		},
		"WithIDAttr_NoImportIDCreate": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec:       regionalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID:    true,
			useImportIDCreator: false,
			expectError:        true,
		},

		"name mapped": {
			identityAttrValues: map[string]string{
				"id_name": "a_name",
				"type":    "a_type",
			},
			identitySpec: regionalMultipleParameterizedIdentitySpecWithMappedName(map[string]string{
				"id_name": "name",
				"type":    "type",
			}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"region":     region,
				"id_name":    "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:     "a_name,a_type",
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

			identitySchema := ptr(identity.NewIdentitySchema(tc.identitySpec))

			schema := regionalMultipleParameterizedSchema
			if tc.useSchemaWithID {
				schema = regionalMultipleParameterizedWithIDSchema
			}

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}
			if tc.useSchemaWithID {
				importSpec.SetIDAttr = true
			}
			if tc.useImportIDCreator {
				importSpec.ImportID = testImportIDCreator{
					testImportID: testImportID{t: t},
				}
			}

			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrValues)

			response := importByIdentity(ctx, f, &client, schema, identity, tc.identitySpec, &importSpec)
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

			// Check region value
			if e, a := tc.expectedRegion, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check attr values
			for name, expectedAttr := range tc.expectedResourceAttrs {
				if e, a := expectedAttr, getAttributeValue(ctx, t, response.State, path.Root(name)); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.expectedID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
					t.Errorf("expected `id` to be %q, got %q", e, a)
				}
			}

			// Check identity
			if identity := response.Identity; identity == nil {
				t.Error("Identity should be set")
			} else {
				for name, expectedAttr := range tc.expectedIdentityAttrs {
					if e, a := expectedAttr, getIdentityAttributeValue(ctx, t, response.Identity, path.Root(name)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", name, e, a)
					}
				}
			}
		})
	}
}

var globalMultipleParameterizedSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
	},
}

var globalMultipleParameterizedWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": framework.IDAttributeDeprecatedNoReplacement(),
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
	},
}

func globalMultipleParameterizedIdentitySpec(attrNames []string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for _, attrName := range attrNames {
		attrs = append(attrs, inttypes.StringIdentityAttribute(attrName, true))
	}
	return inttypes.GlobalParameterizedIdentity(attrs)
}

func globalMultipleParameterizedIdentitySpecWithMappedName(attrNames map[string]string) inttypes.Identity {
	var attrs []inttypes.IdentityAttribute
	for identityAttrName, resourceAttrName := range attrNames {
		if identityAttrName == resourceAttrName {
			attrs = append(attrs, inttypes.StringIdentityAttribute(identityAttrName, true))
		} else {
			attrs = append(attrs, inttypes.StringIdentityAttributeWithMappedName(identityAttrName, true, resourceAttrName))
		}
	}
	return inttypes.GlobalParameterizedIdentity(attrs)
}

func TestGlobalMutipleParameterized_ByImportID(t *testing.T) {
	t.Parallel()

	f := importer.MultipleParameterized

	accountID := "123456789012"

	testCases := map[string]struct {
		inputID               string
		identitySpec          inttypes.Identity
		useSchemaWithID       bool
		noIdentity            bool
		expectedResourceAttrs map[string]string
		expectedIdentityAttrs map[string]string
		expectedID            string
		expectError           bool
		expectedErrorPrefix   string
	}{
		"Basic": {
			inputID:      "a_name,a_type",
			identitySpec: globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
		"name mapped": {
			inputID: "a_name,a_type",
			identitySpec: globalMultipleParameterizedIdentitySpecWithMappedName(map[string]string{
				"id_name": "name",
				"type":    "type",
			}),
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"id_name": "a_name",
				"type":    "a_type",
			},
			expectError: false,
		},
		"NoIdentity": {
			inputID:      "a_name,a_type",
			identitySpec: globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			noIdentity:   true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:  "a_name,a_type",
			expectError: false,
		},

		"WithIDAttr": {
			inputID:         "a_name,a_type",
			identitySpec:    globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID: true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:  "a_name,a_type",
			expectError: false,
		},
		"WithIDAttr_TrimmedID": {
			inputID:         "trim:a_name,a_type",
			identitySpec:    globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID: true,
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedIdentityAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:  "a_name,a_type",
			expectError: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
			}

			stateAttrs := map[string]string{}

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				identitySchema = ptr(identity.NewIdentitySchema(tc.identitySpec))
			}

			schema := globalMultipleParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalMultipleParameterizedWithIDSchema
			}

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}
			if tc.useSchemaWithID {
				importSpec.SetIDAttr = true
			}

			response := importByIDWithState(ctx, f, &client, schema, tc.inputID, stateAttrs, identitySchema, tc.identitySpec, &importSpec)
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

			// Check attr values
			for name, expectedAttr := range tc.expectedResourceAttrs {
				if e, a := expectedAttr, getAttributeValue(ctx, t, response.State, path.Root(name)); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.expectedID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					for name, expectedAttr := range tc.expectedIdentityAttrs {
						if e, a := expectedAttr, getIdentityAttributeValue(ctx, t, response.Identity, path.Root(name)); e != a {
							t.Errorf("expected Identity `%s` to be %q, got %q", name, e, a)
						}
					}

					expectedIdentityAttrs := tc.expectedIdentityAttrs
					expectedIdentityAttrs["account_id"] = accountID

					var obj types.Object
					if diags := identity.Get(ctx, &obj); diags.HasError() {
						t.Fatalf("Unexpected error getting identity attributes: %s", fwdiag.DiagnosticsError(diags))
					}

					actualIdentityAttrs := make(map[string]string)
					for attrName, attrValue := range obj.Attributes() {
						if v, ok := attrValue.(types.String); !ok {
							t.Fatalf("expected string attribute, had %T", attrValue)
						} else {
							actualIdentityAttrs[attrName] = v.ValueString()
						}
					}

					if diff := cmp.Diff(actualIdentityAttrs, expectedIdentityAttrs); diff != "" {
						t.Fatalf("Unexpected identity attributes (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestGlobalMutipleParameterized_ByIdentity(t *testing.T) {
	t.Parallel()

	f := importer.MultipleParameterized

	accountID := "123456789012"

	testCases := map[string]struct {
		identityAttrValues    map[string]string
		identitySpec          inttypes.Identity
		useSchemaWithID       bool
		useImportIDCreator    bool
		expectedID            string
		expectedIdentityAttrs map[string]string
		expectedResourceAttrs map[string]string
		expectError           bool
		expectedErrorPrefix   string
	}{
		"Required": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec: globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedID:   "a_name,a_type",
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
		"WithAccountID": {
			identityAttrValues: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			identitySpec: globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectedID:   "a_name,a_type",
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
		"WrongAccountID": {
			identityAttrValues: map[string]string{
				"account_id": "987654321098",
				"name":       "a_name",
				"type":       "a_type",
			},
			identitySpec: globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			expectError:  true,
		},

		"WithIDAttr_Required": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec:       globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID:    true,
			useImportIDCreator: true,
			expectedID:         "a_name,a_type",
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"name":       "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectError: false,
		},
		"WithIDAttr_NoImportIDCreate": {
			identityAttrValues: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			identitySpec:       globalMultipleParameterizedIdentitySpec([]string{"name", "type"}),
			useSchemaWithID:    true,
			useImportIDCreator: false,
			expectError:        true,
		},

		"name mapped": {
			identityAttrValues: map[string]string{
				"id_name": "a_name",
				"type":    "a_type",
			},
			identitySpec: globalMultipleParameterizedIdentitySpecWithMappedName(map[string]string{
				"id_name": "name",
				"type":    "type",
			}),
			expectedIdentityAttrs: map[string]string{
				"account_id": accountID,
				"id_name":    "a_name",
				"type":       "a_type",
			},
			expectedResourceAttrs: map[string]string{
				"name": "a_name",
				"type": "a_type",
			},
			expectedID:  "a_name,a_type",
			expectError: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
			}

			identitySchema := ptr(identity.NewIdentitySchema(tc.identitySpec))

			importSpec := inttypes.FrameworkImport{
				WrappedImport: true,
				ImportID:      testImportID{t: t},
			}
			if tc.useSchemaWithID {
				importSpec.SetIDAttr = true
			}
			if tc.useImportIDCreator {
				importSpec.ImportID = testImportIDCreator{
					testImportID: testImportID{t: t},
				}
			}

			schema := globalMultipleParameterizedSchema
			if tc.useSchemaWithID {
				schema = globalMultipleParameterizedWithIDSchema
			}

			identity := identityFromSchema(ctx, identitySchema, tc.identityAttrValues)

			response := importByIdentity(ctx, f, &client, schema, identity, tc.identitySpec, &importSpec)
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

			// Check attr values
			for name, expectedAttr := range tc.expectedResourceAttrs {
				if e, a := expectedAttr, getAttributeValue(ctx, t, response.State, path.Root(name)); e != a {
					t.Errorf("expected `%s` to be %q, got %q", name, e, a)
				}
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.expectedID, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
					t.Errorf("expected `id` to be %q, got %q", e, a)
				}
			}

			// Check identity
			if identity := response.Identity; identity == nil {
				t.Error("Identity should be set")
			} else {
				for name, expectedAttr := range tc.expectedIdentityAttrs {
					if e, a := expectedAttr, getIdentityAttributeValue(ctx, t, response.Identity, path.Root(name)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", name, e, a)
					}
				}
			}
		})
	}
}

var _ inttypes.ImportIDParser = testImportID{}

type testImportID struct {
	t *testing.T
}

func (t testImportID) Parse(id string) (string, map[string]any, error) {
	t.t.Helper()

	if id == "invalid" {
		return "", nil, errors.New("invalid ID")
	}

	id = strings.TrimPrefix(id, "trim:")

	parts, err := flex.ExpandResourceId(id, 2, false)
	if err != nil {
		t.t.Fatalf("Parsing test Import ID: %s", err)
	}

	return id, map[string]any{
		"name": parts[0],
		"type": parts[1],
	}, nil
}

var _ inttypes.FrameworkImportIDCreator = testImportIDCreator{}

type testImportIDCreator struct {
	testImportID
}

func (t testImportIDCreator) Create(ctx context.Context, state tfsdk.State) string {
	t.t.Helper()

	parts := []string{
		getAttributeValue(ctx, t.t, state, path.Root("name")),
		getAttributeValue(ctx, t.t, state, path.Root("type")),
	}

	id, err := flex.FlattenResourceId(parts, 2, false)
	if err != nil {
		t.t.Fatalf("Unexpected error creating composite ID: %s", err)
	}

	return id
}
