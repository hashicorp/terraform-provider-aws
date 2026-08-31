// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"maps"
	"testing"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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

func (sp mockServicePackage) FrameworkDataSources(context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return nil
}

func (sp mockServicePackage) FrameworkResources(context.Context) []*inttypes.ServicePackageFrameworkResource {
	return nil
}

func (sp mockServicePackage) SDKDataSources(context.Context) []*inttypes.ServicePackageSDKDataSource {
	return nil
}

func (sp mockServicePackage) SDKResources(context.Context) []*inttypes.ServicePackageSDKResource {
	return nil
}

func (sp mockServicePackage) ServicePackageName() string {
	return "Test"
}

type mockTagsClient struct {
	mockRequiredTagsClient
	defaultTags *tftags.DefaultConfig
	ignoreTags  *tftags.IgnoreConfig
}

func (c mockTagsClient) DefaultTagsConfig(context.Context) *tftags.DefaultConfig {
	return c.defaultTags
}

func (c mockTagsClient) IgnoreTagsConfig(context.Context) *tftags.IgnoreConfig {
	return c.ignoreTags
}

func TestTagsResourceInterceptorModifyPlan(t *testing.T) {
	t.Parallel()

	resourceSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrTags: tftags.TagsAttribute(),
			names.AttrTagsAll: schema.MapAttribute{
				CustomType:  tftags.MapType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
	tagsType := tftypes.Map{ElementType: tftypes.String}

	testCases := map[string]struct {
		defaultTags map[string]string
		planTags    map[string]string
		want        map[string]string
	}{
		"system tag only": {
			planTags: map[string]string{
				"aws:autoscaling-group-123": "managed",
			},
			want: map[string]string{},
		},
		"system tag excluded while resource and default tags remain": {
			defaultTags: map[string]string{
				"default": "value",
			},
			planTags: map[string]string{
				"aws:autoscaling-group-123": "managed",
				"resource":                  "value",
			},
			want: map[string]string{
				"default":  "value",
				"resource": "value",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := conns.NewResourceContext(t.Context(), "Test", "test", "aws_test", "")
			client := mockTagsClient{
				defaultTags: &tftags.DefaultConfig{Tags: tftags.New(ctx, testCase.defaultTags)},
			}
			ctx = tftags.NewContext(ctx, client.DefaultTagsConfig(ctx), client.IgnoreTagsConfig(ctx), nil)

			planTags := make(map[string]tftypes.Value, len(testCase.planTags))
			for k, v := range testCase.planTags {
				planTags[k] = tftypes.NewValue(tftypes.String, v)
			}
			rawPlan := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), map[string]tftypes.Value{
				names.AttrTags:    tftypes.NewValue(tagsType, planTags),
				names.AttrTagsAll: tftypes.NewValue(tagsType, tftypes.UnknownValue),
			})

			response := &resource.ModifyPlanResponse{
				Plan: tfsdk.Plan{
					Raw:    rawPlan,
					Schema: resourceSchema,
				},
			}
			r := resourceTransparentTagging(unique.Make(inttypes.ServicePackageResourceTags{})).(resourceModifyPlanInterceptor)
			r.modifyPlan(ctx, interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: client,
				request: &resource.ModifyPlanRequest{
					Plan: tfsdk.Plan{
						Raw:    rawPlan,
						Schema: resourceSchema,
					},
				},
				response: response,
				when:     Before,
			})
			if response.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %s", response.Diagnostics)
			}

			var gotTagsAll tftags.Map
			response.Diagnostics.Append(response.Plan.GetAttribute(ctx, path.Root(names.AttrTagsAll), &gotTagsAll)...)
			if response.Diagnostics.HasError() {
				t.Fatalf("reading planned tags_all: %s", response.Diagnostics)
			}

			if got := tftags.New(ctx, gotTagsAll).Map(); !maps.Equal(got, testCase.want) {
				t.Errorf("planned tags_all = %#v, want %#v", got, testCase.want)
			}
		})
	}
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

	// Null tags
	attrs := map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
		"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
	}
	rawVal := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), attrs)

	// Partial required tags
	attrsPartial := map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
		"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"bar": tftypes.NewValue(tftypes.String, nil),
		}),
	}
	rawValPartial := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), attrsPartial)

	// All required tags
	attrsRequired := map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
		"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"foo": tftypes.NewValue(tftypes.String, nil),
			"bar": tftypes.NewValue(tftypes.String, nil),
		}),
	}
	rawValRequired := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), attrsRequired)

	// Unknown tag values
	attrsUnknown := map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
		"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"foo": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"bar": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		}),
	}
	rawValUnknown := tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), attrsUnknown)

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
			name: "create, partial tags",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw state is null on creation
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
			wantDiags: diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				path.Root(names.AttrTags),
				"Missing Required Tags",
				"An organizational tag policy requires the following tags for aws_test: [foo]",
			),
			},
		},
		{
			name: "create, required tags",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw state is null on creation
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
		},
		{
			name: "create, unknown tag values",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValUnknown,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), nil), // Raw state is null on creation
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValUnknown,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValUnknown,
						Schema: resourceSchema,
					},
				},
				when: Before,
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
			name: "update, add required",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
		},
		{
			name: "update, remove required",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValRequired,
						Schema: resourceSchema,
					},
				},
				when: Before,
			},
			wantDiags: diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				path.Root(names.AttrTags),
				"Missing Required Tags",
				"An organizational tag policy requires the following tags for aws_test: [foo]",
			),
			},
		},
		{
			name: "update, unknown tag values",
			opts: interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{
				c: mockRequiredTagsClient{},
				request: &resource.ModifyPlanRequest{
					Config: tfsdk.Config{
						Raw:    rawValUnknown,
						Schema: resourceSchema,
					},
					State: tfsdk.State{
						Raw:    rawValPartial,
						Schema: resourceSchema,
					},
					Plan: tfsdk.Plan{
						Raw:    rawValUnknown,
						Schema: resourceSchema,
					},
				},
				response: &resource.ModifyPlanResponse{
					Plan: tfsdk.Plan{
						Raw:    rawValUnknown,
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
