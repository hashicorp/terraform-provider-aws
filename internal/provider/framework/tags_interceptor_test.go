// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type mockRequiredTagsClient struct {
	mockClient
}

func (c mockRequiredTagsClient) DefaultTagsConfig(ctx context.Context) *tftags.DefaultConfig {
	return nil
}

func (c mockRequiredTagsClient) IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig {
	return nil
}

func (c mockRequiredTagsClient) ServicePackage(_ context.Context, name string) conns.ServicePackage {
	return mockServicePackage{}
}

func (c mockRequiredTagsClient) TagPolicyConfig(ctx context.Context) *tftags.TagPolicyConfig {
	return &tftags.TagPolicyConfig{
		Severity: "error",
		RequiredTags: map[string]tftags.KeyValueTags{
			"aws_test": {
				"foo": nil,
				"bar": nil,
			},
		},
	}
}

type mockServicePackage struct{}

func (sp mockServicePackage) FrameworkDataSources(context.Context) []*types.ServicePackageFrameworkDataSource {
	return nil
}

func (sp mockServicePackage) FrameworkResources(context.Context) []*types.ServicePackageFrameworkResource {
	return nil
}

func (sp mockServicePackage) SDKDataSources(context.Context) []*types.ServicePackageSDKDataSource {
	return nil
}

func (sp mockServicePackage) SDKResources(context.Context) []*types.ServicePackageSDKResource {
	return nil
}

func (sp mockServicePackage) ServicePackageName() string {
	return "Test"
}

func Test_resourceValidateRequiredTagsInterceptor(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	bootstrapContext := func(ctx context.Context, meta any) context.Context {
		ctx = conns.NewResourceContext(ctx, "Test", "test", "aws_test", "")
		if v, ok := meta.(awsClient); ok {
			ctx = tftags.NewContext(ctx, v.DefaultTagsConfig(ctx), v.IgnoreTagsConfig(ctx), v.TagPolicyConfig(ctx))
		}

		return ctx
	}

	resourceSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"tags": tftags.TagsAttribute(),
		},
	}

	attrs := map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
		"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
	}
	rawVal := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), attrs)

	tests := []struct {
		name      string
		opts      interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]
		wantDiags diag.Diagnostics
	}{
		{
			name: "create, missing tags",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw state is null on creation
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
			wantDiags: diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				path.Root(names.AttrTags),
				"Missing Required Tags",
				"An organizational tag policy requires the following tags for aws_test: [bar foo]",
			),
			},
		},
		{
			name: "update, no tags change",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
		},
		{
			name: "destroy",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    rawVal,
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw plan is null on destroy
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw plan is null on destroy
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := resourceValidateRequiredTags()
			ctx = bootstrapContext(ctx, tt.opts.c)
			r.modifyPlan(ctx, tt.opts)

			if !tt.opts.response.Diagnostics.Equal(tt.wantDiags) {
				t.Errorf("response diagnostics not equal. got: %s want: %s", tt.opts.response.Diagnostics, tt.wantDiags)
			}
		})
	}
}
