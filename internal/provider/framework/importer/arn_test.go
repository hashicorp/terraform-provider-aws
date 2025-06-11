// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var globalARNSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
	},
}

var globalARNWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
		"id": framework.IDAttributeDeprecatedNoReplacement(),
	},
}

func globalARNIdentitySpec(attrs ...string) inttypes.Identity {
	var opts []inttypes.IdentityOptsFunc
	if len(attrs) > 0 {
		opts = append(opts, inttypes.WithIdentityDuplicateAttrs(attrs...))
	}
	return inttypes.GlobalARNIdentity(opts...)
}

type mockClient struct {
	accountID string
}

func (c *mockClient) AccountID(_ context.Context) string {
	return c.accountID
}

func TestGlobalARN(t *testing.T) {
	t.Parallel()

	f := importer.GlobalARN

	accountID := "123456789012"
	validARN := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: accountID,
		Resource:  "res-abc123",
	}.String()

	testCases := map[string]struct {
		importMethod         string // "ImportID" or "Identity"
		inputARN             string
		duplicateAttrs       []string
		useSchemaWithID      bool
		noIdentity           bool
		expectError          bool
		expectedErrorSummary string
	}{
		"ImportID_Valid": {
			importMethod: "ImportID",
			inputARN:     validARN,
			expectError:  false,
		},
		"ImportID_Valid_NoIdentity": {
			importMethod: "ImportID",
			inputARN:     validARN,
			noIdentity:   true,
			expectError:  false,
		},
		"ImportID_Invalid_NotAnARN": {
			importMethod:         "ImportID",
			inputARN:             "not a valid ARN",
			expectError:          true,
			expectedErrorSummary: importer.InvalidResourceImportIDValue,
		},
		"Identity_Valid": {
			importMethod: "Identity",
			inputARN:     validARN,
			expectError:  false,
		},
		"Identity_Invalid_NotAnARN": {
			importMethod:         "Identity",
			inputARN:             "not a valid ARN",
			expectError:          true,
			expectedErrorSummary: "Invalid Identity Attribute Value",
		},
		"DuplicateAttrs_ImportID_Valid": {
			importMethod:    "ImportID",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			expectError:     false,
		},
		"DuplicateAttrs_Identity_Valid": {
			importMethod:    "Identity",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			expectError:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
			}

			identitySpec := globalARNIdentitySpec(tc.duplicateAttrs...)

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				x := identity.NewIdentitySchema(identitySpec)
				identitySchema = &x
			}

			var response resource.ImportStateResponse
			schema := globalARNSchema
			if tc.useSchemaWithID {
				schema = globalARNWithIDSchema
			}

			switch tc.importMethod {
			case "ImportID":
				response = importARNByID(ctx, f, &client, schema, tc.inputARN, identitySchema, identitySpec)
			case "Identity":
				identity := identityFromSchema(ctx, identitySchema, map[string]string{
					"arn": tc.inputARN,
				})
				response = importARNByIdentity(ctx, f, &client, schema, identity, identitySpec)
			}

			if tc.expectError {
				if !response.Diagnostics.HasError() {
					t.Fatal("Expected error, got none")
				}
				if tc.expectedErrorSummary != "" && response.Diagnostics[0].Summary() != tc.expectedErrorSummary {
					t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
				}
				return
			}

			if response.Diagnostics.HasError() {
				t.Fatalf("Unexpected error: %s", fwdiag.DiagnosticsError(response.Diagnostics))
			}

			// Check ARN value
			if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
				t.Errorf("expected `arn` to be %q, got %q", e, a)
			}

			// Check attr value
			expectedAttrValue := ""
			if len(tc.duplicateAttrs) > 0 && tc.useSchemaWithID {
				for _, attr := range tc.duplicateAttrs {
					if attr == "attr" {
						expectedAttrValue = tc.inputARN
						break
					}
				}
			}
			if e, a := expectedAttrValue, getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
				t.Errorf("expected `attr` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					var arnVal string
					identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
					if e, a := tc.inputARN, arnVal; e != a {
						t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
					}
				}
			}
		})
	}
}

var regionalARNSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
		"region": resourceattribute.Region(),
	},
}

var regionalARNWithIDSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"arn": framework.ARNAttributeComputedOnly(),
		"attr": schema.StringAttribute{
			Optional: true,
		},
		"id":     framework.IDAttributeDeprecatedNoReplacement(),
		"region": resourceattribute.Region(),
	},
}

func regionalARNIdentitySpec(attrs ...string) inttypes.Identity {
	var opts []inttypes.IdentityOptsFunc
	if len(attrs) > 0 {
		opts = append(opts, inttypes.WithIdentityDuplicateAttrs(attrs...))
	}
	return inttypes.RegionalARNIdentity(opts...)
}

func TestRegionalARN(t *testing.T) {
	t.Parallel()

	f := importer.RegionalARN

	accountID := "123456789012"
	region := "a-region-1"
	validARN := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    region,
		AccountID: accountID,
		Resource:  "res-abc123",
	}.String()

	testCases := map[string]struct {
		importMethod        string // "ImportID", "Identity", or "IDWithState"
		inputARN            string
		duplicateAttrs      []string
		useSchemaWithID     bool
		stateAttrs          map[string]string
		noIdentity          bool
		expectError         bool
		expectedErrorPrefix string
	}{
		"ImportID_Valid_DefaultRegion": {
			importMethod: "ImportID",
			inputARN:     validARN,
			expectError:  false,
		},
		"ImportID_Valid_RegionOverride": {
			importMethod: "IDWithState",
			inputARN:     validARN,
			stateAttrs: map[string]string{
				"region": region,
			},
			expectError: false,
		},
		"ImportID_Valid_NoIdentity": {
			importMethod: "ImportID",
			inputARN:     validARN,
			noIdentity:   true,
			expectError:  false,
		},
		"ImportID_Invalid_NotAnARN": {
			importMethod:        "ImportID",
			inputARN:            "not a valid ARN",
			expectError:         true,
			expectedErrorPrefix: "The import ID could not be parsed as an ARN.",
		},
		"ImportID_Invalid_WrongRegion": {
			importMethod: "IDWithState",
			inputARN:     validARN,
			stateAttrs: map[string]string{
				"region": "another-region-1",
			},
			expectError:         true,
			expectedErrorPrefix: "The region passed for import,",
		},
		"Identity_Valid": {
			importMethod: "Identity",
			inputARN:     validARN,
			expectError:  false,
		},
		"Identity_Invalid_NotAnARN": {
			importMethod:        "Identity",
			inputARN:            "not a valid ARN",
			expectError:         true,
			expectedErrorPrefix: "Identity attribute \"arn\" could not be parsed as an ARN.",
		},
		"DuplicateAttrs_ImportID_Valid": {
			importMethod:    "ImportID",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			expectError:     false,
		},
		"DuplicateAttrs_Identity_Valid": {
			importMethod:    "Identity",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			expectError:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
			}

			identitySpec := regionalARNIdentitySpec(tc.duplicateAttrs...)

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				x := identity.NewIdentitySchema(identitySpec)
				identitySchema = &x
			}

			var response resource.ImportStateResponse
			schema := regionalARNSchema
			if tc.useSchemaWithID {
				schema = regionalARNWithIDSchema
			}

			switch tc.importMethod {
			case "ImportID":
				response = importARNByID(ctx, f, &client, schema, tc.inputARN, identitySchema, identitySpec)
			case "IDWithState":
				response = importARNByIDWithState(ctx, f, &client, schema, tc.inputARN, tc.stateAttrs, identitySchema, identitySpec)
			case "Identity":
				identity := identityFromSchema(ctx, identitySchema, map[string]string{
					"arn": tc.inputARN,
				})
				response = importARNByIdentity(ctx, f, &client, schema, identity, identitySpec)
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

			// Check ARN value
			if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
				t.Errorf("expected `arn` to be %q, got %q", e, a)
			}

			// Check attr value
			expectedAttrValue := ""
			if len(tc.duplicateAttrs) > 0 && tc.useSchemaWithID {
				for _, attr := range tc.duplicateAttrs {
					if attr == "attr" {
						expectedAttrValue = tc.inputARN
						break
					}
				}
			}
			if e, a := expectedAttrValue, getAttributeValue(ctx, t, response.State, path.Root("attr")); e != a {
				t.Errorf("expected `attr` to be %q, got %q", e, a)
			}

			// Check region value
			if e, a := region, getAttributeValue(ctx, t, response.State, path.Root("region")); e != a {
				t.Errorf("expected `region` to be %q, got %q", e, a)
			}

			// Check ID value if using schema with ID
			if tc.useSchemaWithID {
				if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					var arnVal string
					identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
					if e, a := tc.inputARN, arnVal; e != a {
						t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
					}
				}
			}
		})
	}
}

var regionalResourceWithGlobalARNFormatSchema = regionalARNSchema

var regionalResourceWithGlobalARNFormatWithIDSchema = regionalARNWithIDSchema

func regionalResourceWithGlobalARNFormatIdentitySpec(attrs ...string) inttypes.Identity {
	var opts []inttypes.IdentityOptsFunc
	if len(attrs) > 0 {
		opts = append(opts, inttypes.WithIdentityDuplicateAttrs(attrs...))
	}
	return inttypes.RegionalResourceWithGlobalARNFormat(opts...)
}

func TestRegionalARNWithGlobalFormat(t *testing.T) {
	t.Parallel()

	f := importer.RegionalARNWithGlobalFormat

	accountID := "123456789012"
	defaultRegion := "a-region-1"
	anotherRegion := "another-region-1"
	validARN := arn.ARN{
		Partition: "aws",
		Service:   "a-service",
		Region:    "",
		AccountID: accountID,
		Resource:  "res-abc123",
	}.String()

	testCases := map[string]struct {
		importMethod        string // "ImportID" or "Identity"
		inputARN            string
		duplicateAttrs      []string
		useSchemaWithID     bool
		inputRegion         string
		noIdentity          bool
		expectedRegion      string
		expectError         bool
		expectedErrorPrefix string
	}{
		"ImportID_Valid_DefaultRegion": {
			importMethod:   "ImportID",
			inputARN:       validARN,
			inputRegion:    defaultRegion,
			expectedRegion: defaultRegion,
			expectError:    false,
		},
		"ImportID_Valid_RegionOverride": {
			importMethod:   "ImportID",
			inputARN:       validARN,
			inputRegion:    anotherRegion,
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"ImportID_Valid_NoIdentity": {
			importMethod:   "ImportID",
			inputARN:       validARN,
			inputRegion:    defaultRegion,
			noIdentity:     true,
			expectedRegion: defaultRegion,
			expectError:    false,
		},
		"ImportID_Invalid_NotAnARN": {
			importMethod:        "ImportID",
			inputARN:            "not a valid ARN",
			inputRegion:         defaultRegion,
			expectError:         true,
			expectedErrorPrefix: "The import ID could not be parsed as an ARN.",
		},

		"Identity_Valid_DefaultRegion": {
			importMethod:   "Identity",
			inputARN:       validARN,
			inputRegion:    defaultRegion,
			expectedRegion: defaultRegion,
			expectError:    false,
		},
		"Identity_Valid_RegionOverride": {
			importMethod:   "Identity",
			inputARN:       validARN,
			inputRegion:    anotherRegion,
			expectedRegion: anotherRegion,
			expectError:    false,
		},
		"Identity_Invalid_NotAnARN": {
			importMethod:        "Identity",
			inputARN:            "not a valid ARN",
			inputRegion:         defaultRegion,
			expectError:         true,
			expectedErrorPrefix: "Identity attribute \"arn\" could not be parsed as an ARN.",
		},

		"DuplicateAttrs_ImportID_Valid": {
			importMethod:    "ImportID",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			inputRegion:     defaultRegion,
			expectedRegion:  defaultRegion,
			expectError:     false,
		},

		"DuplicateAttrs_Identity_Valid": {
			importMethod:    "Identity",
			duplicateAttrs:  []string{"id", "attr"},
			useSchemaWithID: true,
			inputARN:        validARN,
			inputRegion:     defaultRegion,
			expectedRegion:  defaultRegion,
			expectError:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			client := mockClient{
				accountID: accountID,
			}

			identitySpec := regionalResourceWithGlobalARNFormatIdentitySpec(tc.duplicateAttrs...)

			var identitySchema *identityschema.Schema
			if !tc.noIdentity {
				x := identity.NewIdentitySchema(identitySpec)
				identitySchema = &x
			}

			var response resource.ImportStateResponse
			schema := regionalResourceWithGlobalARNFormatSchema
			if tc.useSchemaWithID {
				schema = regionalResourceWithGlobalARNFormatWithIDSchema
			}

			switch tc.importMethod {
			case "ImportID":
				stateAttrs := map[string]string{
					"region": tc.inputRegion,
				}
				response = importARNByIDWithState(ctx, f, &client, schema, tc.inputARN, stateAttrs, identitySchema, identitySpec)
			case "Identity":
				identity := identityFromSchema(ctx, identitySchema, map[string]string{
					"region": tc.inputRegion,
					"arn":    tc.inputARN,
				})
				response = importARNByIdentity(ctx, f, &client, schema, identity, identitySpec)
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

			// Check ARN value
			if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("arn")); e != a {
				t.Errorf("expected `arn` to be %q, got %q", e, a)
			}

			// Check attr value
			expectedAttrValue := ""
			if len(tc.duplicateAttrs) > 0 && tc.useSchemaWithID {
				for _, attr := range tc.duplicateAttrs {
					if attr == "attr" {
						expectedAttrValue = tc.inputARN
						break
					}
				}
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
				if e, a := tc.inputARN, getAttributeValue(ctx, t, response.State, path.Root("id")); e != a {
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
					if e, a := tc.expectedRegion, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("region")); e != a {
						t.Errorf("expected Identity `region` to be %q, got %q", e, a)
					}

					var arnVal string
					identity.GetAttribute(ctx, path.Root("arn"), &arnVal)
					if e, a := tc.inputARN, arnVal; e != a {
						t.Errorf("expected Identity `arn` to be %q, got %q", e, a)
					}
				}
			}
		})
	}
}

type importARNFunc func(ctx context.Context, client importer.AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse)

func importARNByID(ctx context.Context, f importARNFunc, client importer.AWSClient, resourceSchema schema.Schema, id string, identitySchema *identityschema.Schema, identitySpec inttypes.Identity) resource.ImportStateResponse {
	var identity *tfsdk.ResourceIdentity
	if identitySchema != nil {
		identity = emtpyIdentityFromSchema(ctx, identitySchema)
	}

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func importARNByIDWithState(ctx context.Context, f importARNFunc, client importer.AWSClient, resourceSchema schema.Schema, id string, stateAttrs map[string]string, identitySchema *identityschema.Schema, identitySpec inttypes.Identity) resource.ImportStateResponse {
	var identity *tfsdk.ResourceIdentity
	if identitySchema != nil {
		identity = emtpyIdentityFromSchema(ctx, identitySchema)
	}

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func importARNByIdentity(ctx context.Context, f importARNFunc, client importer.AWSClient, resourceSchema schema.Schema, identity *tfsdk.ResourceIdentity, identitySpec inttypes.Identity) resource.ImportStateResponse {
	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func getAttributeValue(ctx context.Context, t *testing.T, state tfsdk.State, path path.Path) string {
	t.Helper()

	var attrVal types.String
	if diags := state.GetAttribute(ctx, path, &attrVal); diags.HasError() {
		t.Fatalf("Unexpected error getting attribute %q: %s", path, fwdiag.DiagnosticsError(diags))
	}
	return attrVal.ValueString()
}

func getIdentityAttributeValue(ctx context.Context, t *testing.T, identity *tfsdk.ResourceIdentity, path path.Path) string {
	t.Helper()

	var attrVal types.String
	if diags := identity.GetAttribute(ctx, path, &attrVal); diags.HasError() {
		t.Fatalf("Unexpected error getting Identity attribute %q: %s", path, fwdiag.DiagnosticsError(diags))
	}
	return attrVal.ValueString()
}
