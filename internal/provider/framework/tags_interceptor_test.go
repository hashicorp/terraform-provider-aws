// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
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

func TestTagsResourceInterceptorIgnoreTagUpdates(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resourceSchema := schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true},
		"tags":     tftags.TagsAttribute(),
		"tags_all": schema.MapAttribute{Computed: true, CustomType: tftags.MapType, ElementType: types.StringType},
	}}
	raw := func(tagsAll map[string]string) tftypes.Value {
		elems := make(map[string]tftypes.Value, len(tagsAll))
		for k, v := range tagsAll {
			elems[k] = tftypes.NewValue(tftypes.String, v)
		}
		return tftypes.NewValue(resourceSchema.Type().TerraformType(ctx), map[string]tftypes.Value{
			"id":       tftypes.NewValue(tftypes.String, "test"),
			"tags":     tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			"tags_all": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, elems),
		})
	}
	defaults := &tftags.DefaultConfig{Tags: tftags.New(ctx, map[string]string{"CreatedOn": "new", "created:by": "new", "Name": "new"})}
	ignore := &tftags.IgnoreConfig{UpdateKeys: tftags.New(ctx, []string{"CreatedOn"}), UpdateKeyPrefixes: tftags.New(ctx, []string{"created:"})}
	for _, operation := range []string{"create", "plan", "update", "inline update", "read"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			ctx := conns.NewResourceContext(t.Context(), "Test", "test", "aws_test", "")
			ctx = tftags.NewContext(ctx, defaults, ignore, nil)
			inContext, _ := tftags.FromContext(ctx)
			conn := &conns.AWSClient{}
			sp := &creationOnlyTagService{}
			conn.SetServicePackages(ctx, map[string]conns.ServicePackage{"Test": sp})
			conns.SetDefaultTagsConfig(conn, defaults)
			conns.SetIgnoreTagsConfig(conn, ignore)
			interceptor := tagsResourceInterceptor{HTags: interceptors.HTags(inttypes.ResourceTagsAttribute("id"))}
			plan := tfsdk.Plan{Raw: raw(map[string]string{"Name": "new"}), Schema: resourceSchema}
			state := tfsdk.State{Raw: raw(map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "old"}), Schema: resourceSchema}
			if operation == "inline update" {
				interceptor.HTags = interceptors.HTags(inttypes.ResourceTagsInline())
				interceptor.readTags = func(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
					inContext, _ := tftags.FromContext(ctx)
					inContext.TagsOut = option.Some(tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "old"}))
					response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("tags"), tftags.FlattenStringValueMap(ctx, map[string]string{"temporary": "read"}))...)
				}
			}
			switch operation {
			case "create":
				response := &resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: resourceSchema}}
				opts := interceptorOptions[resource.CreateRequest, resource.CreateResponse]{c: conn, request: &resource.CreateRequest{Plan: plan}, response: response, when: Before}
				interceptor.create(ctx, opts)
				if response.Diagnostics.HasError() {
					t.Fatal(response.Diagnostics)
				}
				if got := inContext.TagsIn.MustUnwrap(); !got.Equal(defaults.Tags) {
					t.Fatalf("creation tags = %v, want %v", got, defaults.Tags)
				}
				opts.when = After
				interceptor.create(ctx, opts)
				var tags tftags.Map
				response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("tags_all"), &tags)...)
				if response.Diagnostics.HasError() || !tftags.New(ctx, tags).Equal(tftags.New(ctx, map[string]string{"Name": "new"})) {
					t.Fatalf("creation state = %v, diagnostics = %v", tags, response.Diagnostics)
				}
			case "plan":
				response := &resource.ModifyPlanResponse{Plan: plan}
				interceptor.modifyPlan(ctx, interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]{c: conn, request: &resource.ModifyPlanRequest{Plan: plan, State: state}, response: response, when: Before})
				var tags tftags.Map
				response.Diagnostics.Append(response.Plan.GetAttribute(ctx, path.Root("tags_all"), &tags)...)
				if response.Diagnostics.HasError() || !tftags.New(ctx, tags).Equal(tftags.New(ctx, map[string]string{"Name": "new"})) {
					t.Fatalf("planned tags = %v, diagnostics = %v", tags, response.Diagnostics)
				}
			case "update", "inline update":
				response := &resource.UpdateResponse{}
				interceptor.update(ctx, interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]{c: conn, request: &resource.UpdateRequest{Plan: plan, State: state}, response: response, when: Before})
				if response.Diagnostics.HasError() {
					t.Fatal(response.Diagnostics)
				}
				if operation == "inline update" {
					want := tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "new"})
					if !inContext.TagsIn.MustUnwrap().Equal(want) || inContext.TagsOut.IsSome() || !state.Raw.Equal(raw(map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "old"})) {
						t.Fatal("inline update did not preserve tags in an isolated snapshot")
					}
					return
				}
				if !sp.called || !sp.old.Equal(tftags.New(ctx, map[string]string{"Name": "old"})) || !sp.new.Equal(tftags.New(ctx, map[string]string{"Name": "new"})) {
					t.Fatalf("update included creation-only tags: old=%v new=%v called=%v", sp.old, sp.new, sp.called)
				}
				want := tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "new"})
				if got := inContext.TagsIn.MustUnwrap(); !got.Equal(want) {
					t.Fatalf("resource update tags = %v, want %v", got, want)
				}
				if inContext.TagsOut.IsSome() {
					t.Fatal("preservation read leaked stale tags into update context")
				}
			case "read":
				if diags := state.SetAttribute(ctx, path.Root("tags"), tftags.FlattenStringValueMap(ctx, map[string]string{"CreatedOn": "original"})); diags.HasError() {
					t.Fatal(diags)
				}
				inContext.TagsOut = option.Some(tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "new"}))
				response := &resource.ReadResponse{State: state}
				interceptor.read(ctx, interceptorOptions[resource.ReadRequest, resource.ReadResponse]{c: conn, request: &resource.ReadRequest{State: state}, response: response, when: After})
				var tags tftags.Map
				response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("tags_all"), &tags)...)
				if response.Diagnostics.HasError() || !tftags.New(ctx, tags).Equal(tftags.New(ctx, map[string]string{"Name": "new"})) {
					t.Fatalf("read tags = %v, diagnostics = %v", tags, response.Diagnostics)
				}
				response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("tags"), &tags)...)
				if response.Diagnostics.HasError() || len(tftags.New(ctx, tags)) != 0 {
					t.Fatalf("read retained creation-only resource tags: %v", tags)
				}
			}
		})
	}
}

func TestTagsResourceInterceptorReadIgnoreTagUpdatesEmptyTags(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ignore := &tftags.IgnoreConfig{UpdateKeys: tftags.New(ctx, []string{"CreatedOn"}), UpdateKeyPrefixes: tftags.New(ctx, []string{"created:"})}
	nullTags := tftags.NewMapValueNull()
	emptyTags := tftags.FlattenStringValueMap(ctx, map[string]string{})
	protectedTags := tftags.FlattenStringValueMap(ctx, map[string]string{"CreatedOn": "original", "created:by": "original"})
	otherTags := tftags.FlattenStringValueMap(ctx, map[string]string{"Name": "test"})
	mixedTags := tftags.FlattenStringValueMap(ctx, map[string]string{"CreatedOn": "original", "Name": "test"})
	tests := []struct {
		name   string
		state  tftags.Map
		api    map[string]string
		ignore *tftags.IgnoreConfig
		want   tftags.Map
	}{
		{name: "null state without API tags", state: nullTags, ignore: ignore, want: nullTags},
		{name: "null state with protected API tags", state: nullTags, api: map[string]string{"CreatedOn": "original"}, ignore: ignore, want: nullTags},
		{name: "explicit empty state", state: emptyTags, ignore: ignore, want: emptyTags},
		{name: "protected state", state: protectedTags, api: map[string]string{"CreatedOn": "original", "created:by": "original"}, ignore: ignore, want: nullTags},
		{name: "mixed state", state: mixedTags, ignore: ignore, want: otherTags},
		{name: "unrelated state", state: otherTags, ignore: ignore, want: otherTags},
		{name: "unknown state", state: tftags.NewMapValueUnknown(), ignore: ignore, want: tftags.NewMapValueUnknown()},
		{name: "empty state without block", state: emptyTags, want: emptyTags},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := conns.NewResourceContext(t.Context(), "Test", "test", "aws_test", "")
			ctx = tftags.NewContext(ctx, nil, tt.ignore, nil)
			inContext, _ := tftags.FromContext(ctx)
			inContext.TagsOut = option.Some(tftags.New(ctx, tt.api))
			conn := &conns.AWSClient{}
			conn.SetServicePackages(ctx, map[string]conns.ServicePackage{"Test": mockServicePackage{}})
			conns.SetIgnoreTagsConfig(conn, tt.ignore)
			s := schema.Schema{Attributes: map[string]schema.Attribute{
				"tags":     tftags.TagsAttribute(),
				"tags_all": schema.MapAttribute{Computed: true, CustomType: tftags.MapType, ElementType: types.StringType},
			}}
			stateTags, err := tt.state.ToTerraformValue(ctx)
			if err != nil {
				t.Fatal(err)
			}
			state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{
				"tags":     stateTags,
				"tags_all": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			})}
			response := &resource.ReadResponse{State: state}
			interceptor := tagsResourceInterceptor{HTags: interceptors.HTags(inttypes.ResourceTagsAttribute("id"))}
			interceptor.read(ctx, interceptorOptions[resource.ReadRequest, resource.ReadResponse]{c: conn, request: &resource.ReadRequest{State: state}, response: response, when: After})
			var got tftags.Map
			response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("tags"), &got)...)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("read tags = %s, want %s", got, tt.want)
			}
		})
	}
}

type creationOnlyTagService struct {
	mockServicePackage
	called   bool
	old, new tftags.KeyValueTags
}

func (s *creationOnlyTagService) UpdateTags(ctx context.Context, _ any, _ string, old, new any) error {
	s.called = true
	s.old, s.new = tftags.New(ctx, old), tftags.New(ctx, new)
	return nil
}

func (s *creationOnlyTagService) ListTags(ctx context.Context, _ any, _ string) error {
	inContext, _ := tftags.FromContext(ctx)
	inContext.TagsOut = option.Some(tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "old"}))
	return nil
}

func TestTagsDataSourceInterceptorIgnoreTagUpdates(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	conn := &conns.AWSClient{}
	conn.SetServicePackages(ctx, map[string]conns.ServicePackage{"Test": &creationOnlyTagService{}})
	ignore := &tftags.IgnoreConfig{Keys: tftags.New(ctx, []string{"external"}), UpdateKeys: tftags.New(ctx, []string{"CreatedOn"}), UpdateKeyPrefixes: tftags.New(ctx, []string{"created:"})}
	conns.SetIgnoreTagsConfig(conn, ignore)
	ctx = conns.NewResourceContext(ctx, "Test", "test", "aws_test", "")
	ctx = tftags.NewContext(ctx, nil, ignore, nil)
	inContext, _ := tftags.FromContext(ctx)
	inContext.TagsOut = option.Some(tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "terraform", "external": "ignored"}))
	s := dataschema.Schema{Attributes: map[string]dataschema.Attribute{"tags": tftags.TagsAttributeComputedOnly()}}
	raw := tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{"tags": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil)})
	response := &datasource.ReadResponse{State: tfsdk.State{Raw: raw, Schema: s}}
	interceptor := tagsDataSourceInterceptor{HTags: interceptors.HTags(inttypes.ResourceTagsInline())}
	interceptor.read(ctx, interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]{c: conn, request: &datasource.ReadRequest{}, response: response, when: After})
	var tags tftags.Map
	response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("tags"), &tags)...)
	want := tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "terraform"})
	if response.Diagnostics.HasError() || !tftags.New(ctx, tags).Equal(want) {
		t.Fatalf("data source tags=%v want=%v diagnostics=%v", tags, want, response.Diagnostics)
	}
}
