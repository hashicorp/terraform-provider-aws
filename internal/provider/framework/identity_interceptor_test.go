// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func TestIdentityInterceptor(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"region": resourceattribute.Region(),
			"type": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	stateAttrs := map[string]string{
		"name":   name,
		"region": region,
		"type":   "some_type",
	}

	testOperations := map[string]struct {
		operation func(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics)
	}{
		"create": {
			operation: create,
		},
		"read": {
			operation: read,
		},
	}

	for tname, tc := range testOperations {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			operation := tc.operation

			testCases := map[string]struct {
				attrName     string
				identitySpec inttypes.Identity
			}{
				"same names": {
					attrName:     "name",
					identitySpec: regionalSingleParameterIdentitySpec("name"),
				},
				"name mapped": {
					attrName:     "resource_name",
					identitySpec: regionalSingleParameterIdentitySpecNameMapped("resource_name", "name"),
				},
			}

			for tname, tc := range testCases {
				t.Run(tname, func(t *testing.T) {
					t.Parallel()
					ctx := t.Context()

					identitySchema := identity.NewIdentitySchema(tc.identitySpec)

					interceptor := newIdentityInterceptor(tc.identitySpec.Attributes)

					identity := emtpyIdentityFromSchema(ctx, &identitySchema)

					responseIdentity, diags := operation(ctx, interceptor, resourceSchema, stateAttrs, identity, client)
					if len(diags) > 0 {
						t.Fatalf("unexpected diags during interception: %s", diags)
					}

					if e, a := accountID, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root("account_id")); e != a {
						t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
					}
					if e, a := region, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root("region")); e != a {
						t.Errorf("expected Identity `region` to be %q, got %q", e, a)
					}
					if e, a := name, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root(tc.attrName)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
					}
				})
			}
		})
	}
}

func create(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics) {
	request := resource.CreateRequest{
		Config:   configFromSchema(ctx, resourceSchema, stateAttrs),
		Plan:     planFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	response := resource.CreateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	opts := interceptorOptions[resource.CreateRequest, resource.CreateResponse]{
		c:        client,
		request:  &request,
		response: &response,
		when:     After,
	}

	interceptor.create(ctx, opts)

	return response.Identity, response.Diagnostics
}

func read(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics) {
	request := resource.ReadRequest{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	response := resource.ReadResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	opts := interceptorOptions[resource.ReadRequest, resource.ReadResponse]{
		c:        client,
		request:  &request,
		response: &response,
		when:     After,
	}

	interceptor.read(ctx, opts)

	return response.Identity, response.Diagnostics
}

func TestIdentityInterceptor_OnError(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"region": resourceattribute.Region(),
			"type": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	testOperations := map[string]struct {
		operation  func(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics)
		stateAttrs map[string]string
	}{
		"create": {
			operation: createOnError,
			stateAttrs: map[string]string{
				"name":   name,
				"region": region,
				"type":   "some_type",
			},
		},
		"update": {
			operation: updateOnError,
			stateAttrs: map[string]string{
				"name":   name,
				"region": region,
				"type":   "some_type",
			},
		},
	}

	for tname, tc := range testOperations {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()

			operation := tc.operation
			stateAttrs := tc.stateAttrs

			testCases := map[string]struct {
				attrName     string
				identitySpec inttypes.Identity
			}{
				"same names": {
					attrName:     "name",
					identitySpec: regionalSingleParameterIdentitySpec("name"),
				},
				"name mapped": {
					attrName:     "resource_name",
					identitySpec: regionalSingleParameterIdentitySpecNameMapped("resource_name", "name"),
				},
			}

			for tname, tc := range testCases {
				t.Run(tname, func(t *testing.T) {
					t.Parallel()
					ctx := t.Context()

					identitySchema := identity.NewIdentitySchema(tc.identitySpec)

					interceptor := newIdentityInterceptor(tc.identitySpec.Attributes)

					identity := emtpyIdentityFromSchema(ctx, &identitySchema)

					responseIdentity, _ := operation(ctx, interceptor, resourceSchema, stateAttrs, identity, client)

					if e, a := accountID, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root("account_id")); e != a {
						t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
					}
					if e, a := region, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root("region")); e != a {
						t.Errorf("expected Identity `region` to be %q, got %q", e, a)
					}
					if e, a := name, getIdentityAttributeValue(ctx, t, responseIdentity, path.Root(tc.attrName)); e != a {
						t.Errorf("expected Identity `%s` to be %q, got %q", tc.attrName, e, a)
					}
				})
			}
		})
	}
}

func createOnError(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics) {
	request := resource.CreateRequest{
		Config:   configFromSchema(ctx, resourceSchema, stateAttrs),
		Plan:     planFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	response := resource.CreateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
		Diagnostics: diag.Diagnostics{
			diag.NewErrorDiagnostic("summary", "detail"),
		},
	}
	opts := interceptorOptions[resource.CreateRequest, resource.CreateResponse]{
		c:        client,
		request:  &request,
		response: &response,
		when:     OnError,
	}

	interceptor.create(ctx, opts)

	return response.Identity, response.Diagnostics
}

func updateOnError(ctx context.Context, interceptor identityInterceptor, resourceSchema schema.Schema, stateAttrs map[string]string, identity *tfsdk.ResourceIdentity, client awsClient) (*tfsdk.ResourceIdentity, diag.Diagnostics) {
	request := resource.UpdateRequest{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	response := resource.UpdateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
		Diagnostics: diag.Diagnostics{
			diag.NewErrorDiagnostic("summary", "detail"),
		},
	}
	opts := interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]{
		c:        client,
		request:  &request,
		response: &response,
		when:     OnError,
	}

	interceptor.update(ctx, opts)

	return response.Identity, response.Diagnostics
}

func getIdentityAttributeValue(ctx context.Context, t *testing.T, identity *tfsdk.ResourceIdentity, path path.Path) string {
	t.Helper()

	var attrVal types.String
	if diags := identity.GetAttribute(ctx, path, &attrVal); diags.HasError() {
		t.Fatalf("Unexpected error getting Identity attribute %q: %s", path, fwdiag.DiagnosticsError(diags))
	}
	return attrVal.ValueString()
}

func regionalSingleParameterIdentitySpec(name string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentity(name)
}

func regionalSingleParameterIdentitySpecNameMapped(identityAttrName, resourceAttrName string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentityWithMappedName(identityAttrName, resourceAttrName)
}

func stateFromSchema(ctx context.Context, schema schema.Schema, values map[string]string) tfsdk.State {
	val := make(map[string]tftypes.Value)
	for name := range schema.Attributes {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), val),
		Schema: schema,
	}
}

func configFromSchema(ctx context.Context, schema schema.Schema, values map[string]string) tfsdk.Config {
	val := make(map[string]tftypes.Value)
	for name := range schema.Attributes {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return tfsdk.Config{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), val),
		Schema: schema,
	}
}

func planFromSchema(ctx context.Context, schema schema.Schema, values map[string]string) tfsdk.Plan {
	val := make(map[string]tftypes.Value)
	for name := range schema.Attributes {
		if v, ok := values[name]; ok {
			val[name] = tftypes.NewValue(tftypes.String, v)
		} else {
			val[name] = tftypes.NewValue(tftypes.String, nil)
		}
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), val),
		Schema: schema,
	}
}

func emtpyIdentityFromSchema(ctx context.Context, schema *identityschema.Schema) *tfsdk.ResourceIdentity {
	return &tfsdk.ResourceIdentity{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}
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

func (c mockClient) DefaultTagsConfig(ctx context.Context) *tftags.DefaultConfig {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) Partition(context.Context) string {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) ServicePackage(_ context.Context, name string) conns.ServicePackage {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) TagPolicyConfig(ctx context.Context) *tftags.TagPolicyConfig {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) ValidateInContextRegionInPartition(ctx context.Context) error {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) AwsConfig(context.Context) aws.Config { // nosemgrep:ci.aws-in-func-name
	panic("not implemented") //lintignore:R009
}

func TestIdentityIsFullyNull(t *testing.T) {
	t.Parallel()

	attributes := []inttypes.IdentityAttribute{
		inttypes.StringIdentityAttribute("account_id", false),
		inttypes.StringIdentityAttribute("region", false),
		inttypes.StringIdentityAttribute("bucket", true),
	}

	// Create identity schema once for all test cases
	identitySchema := &identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"account_id": identityschema.StringAttribute{},
			"region":     identityschema.StringAttribute{},
			"bucket":     identityschema.StringAttribute{},
		},
	}

	ctx := context.Background()

	// Helper function to create identity with values
	createIdentityWithValues := func(values map[string]string) *tfsdk.ResourceIdentity {
		if values == nil {
			return nil
		}
		identity := emtpyIdentityFromSchema(ctx, identitySchema)
		for attrName, value := range values {
			if value != "" {
				diags := identity.SetAttribute(ctx, path.Root(attrName), value)
				if diags.HasError() {
					t.Fatalf("unexpected error setting %s in identity: %s", attrName, fwdiag.DiagnosticsError(diags))
				}
			}
		}
		return identity
	}

	testCases := map[string]struct {
		identity    *tfsdk.ResourceIdentity
		expectNull  bool
		description string
	}{
		"all_null": {
			identity:    createIdentityWithValues(map[string]string{}),
			expectNull:  true,
			description: "All attributes null should return true",
		},
		"some_null": {
			identity: createIdentityWithValues(map[string]string{
				"account_id": "123456789012",
				// region and bucket remain null
			}),
			expectNull:  false,
			description: "Some attributes set should return false",
		},
		"all_set": {
			identity: createIdentityWithValues(map[string]string{
				"account_id": "123456789012",
				"region":     "us-west-2", // lintignore:AWSAT003
				"bucket":     "test-bucket",
			}),
			expectNull:  false,
			description: "All attributes set should return false",
		},
		"empty_string_values": {
			identity: createIdentityWithValues(map[string]string{
				"account_id": "",
				"region":     "",
				"bucket":     "",
			}),
			expectNull:  true,
			description: "Empty string values should be treated as null",
		},
		"nil_identity": {
			identity:    createIdentityWithValues(nil),
			expectNull:  true,
			description: "Nil identity should return true",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			result := identityIsFullyNull(ctx, tc.identity, attributes)
			if result != tc.expectNull {
				t.Errorf("%s: expected identityIsFullyNull to return %v, got %v",
					tc.description, tc.expectNull, result)
			}
		})
	}
}
