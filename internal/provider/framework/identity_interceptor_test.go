// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	ctx := t.Context()

	accountID := "123456789012"
	region := "us-west-2"
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

	stateAttrs := map[string]string{
		"name":   name,
		"region": region,
		"type":   "some_type",
	}

	identitySpec := regionalSingleParameterIdentitySpec("name")
	identitySchema := identity.NewIdentitySchema(identitySpec)

	interceptor := newIdentityInterceptor(identitySpec.Attributes)

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	identity := emtpyIdentityFromSchema(ctx, &identitySchema)

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

	diags := interceptor.create(ctx, opts)
	if len(diags) > 0 {
		t.Fatalf("unexpected diags during interception: %s", diags)
	}

	if e, a := accountID, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("account_id")); e != a {
		t.Errorf("expected Identity `account_id` to be %q, got %q", e, a)
	}
	if e, a := region, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("region")); e != a {
		t.Errorf("expected Identity `region` to be %q, got %q", e, a)
	}
	if e, a := name, getIdentityAttributeValue(ctx, t, response.Identity, path.Root("name")); e != a {
		t.Errorf("expected Identity `name` to be %q, got %q", e, a)
	}
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
	panic("not implemented")
}

func (c mockClient) IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig {
	panic("not implemented")
}

func (c mockClient) Partition(context.Context) string {
	panic("not implemented")
}

func (c mockClient) ServicePackage(_ context.Context, name string) conns.ServicePackage {
	panic("not implemented")
}

func (c mockClient) ValidateInContextRegionInPartition(ctx context.Context) error {
	panic("not implemented")
}

func (c mockClient) AwsConfig(context.Context) aws.Config {
	panic("not implemented")
}
