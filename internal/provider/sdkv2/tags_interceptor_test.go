// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

type mockService struct{}

var (
	_ tftags.ServiceTagLister  = &mockService{}
	_ tftags.ServiceTagUpdater = &mockService{}
)

func (t *mockService) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{}
}

func (t *mockService) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{}
}

func (t *mockService) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{}
}

func (t *mockService) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{}
}

func (t *mockService) ServicePackageName() string {
	return "TestService"
}

func (t *mockService) ListTags(ctx context.Context, meta any, identifier string) error {
	tags := tftags.New(ctx, map[string]string{
		"tag1": "value1",
	})
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return errors.New("test error")
}

func (t *mockService) UpdateTags(context.Context, any, string, any, any) error {
	return nil
}

func TestTagsResourceInterceptor(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var interceptors interceptorInvocations
	sp := inttypes.ResourceTagsAttribute("id")
	tags := resourceTransparentTagging(sp, nil)
	interceptors = append(interceptors, interceptorInvocation{
		when:        Finally,
		why:         Update,
		interceptor: tags,
	})

	conn := &conns.AWSClient{}
	conn.SetServicePackages(ctx, map[string]conns.ServicePackage{
		"Test": &mockService{},
	})
	conns.SetDefaultTagsConfig(conn, expandDefaultTags(ctx, map[string]any{
		"tag": "",
	}))
	conns.SetIgnoreTagsConfig(conn, expandIgnoreTags(ctx, map[string]any{
		"tag2": "tag",
	}))

	bootstrapContext := func(ctx context.Context, meta any) context.Context {
		ctx = conns.NewResourceContext(ctx, "Test", "test", "aws_test", "")
		if v, ok := meta.(*conns.AWSClient); ok {
			ctx = tftags.NewContext(ctx, v.DefaultTagsConfig(ctx), v.IgnoreTagsConfig(ctx), v.TagPolicyConfig(ctx))
		}

		return ctx
	}

	ctx = bootstrapContext(ctx, conn)
	d := &resourceData{}

	for _, v := range interceptors {
		opts := crudInterceptorOptions{
			c:    conn,
			d:    d,
			when: v.when,
			why:  v.why,
		}
		diags := v.interceptor.(crudInterceptor).run(ctx, opts)
		if got, want := len(diags), 1; got != want {
			t.Errorf("length of diags = %v, want %v", got, want)
		}
	}
}

type resourceData struct{}

func (d *resourceData) GetRawConfig() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"tags": cty.MapVal(map[string]cty.Value{
			"tag1": cty.StringVal("value1"),
		}),
	})
}

func (d *resourceData) GetRawPlan() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		// `tags` is set from the user's configuration, while `tags_all` is
		// computed (unknown) in the plan when, for example, an empty string
		// tag value forces tags_all to be re-computed.
		"tags": cty.MapVal(map[string]cty.Value{
			"tag1": cty.StringVal("value1"),
		}),
		"tags_all": cty.MapVal(map[string]cty.Value{
			"tag1": cty.UnknownVal(cty.String),
		}),
	})
}

func (d *resourceData) GetRawState() cty.Value {
	return cty.Value{}
}

func (d *resourceData) Get(key string) any {
	return nil
}

func (d *resourceData) GetOk(key string) (any, bool) {
	return nil, false
}

func (d *resourceData) Id() string {
	return "id"
}

func (d *resourceData) Set(string, any) error {
	return nil
}

func (d *resourceData) GetChange(key string) (any, any) {
	return nil, nil
}

func (d *resourceData) HasChange(key string) bool {
	return false
}

func (d *resourceData) HasChanges(keys ...string) bool {
	return false
}

func (d *resourceData) Identity() (*schema.IdentityData, error) {
	return nil, nil
}

func TestTagsResourceInterceptorIgnoreTagUpdates(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		when    when
		why     why
		unknown bool
	}{
		{"create", Before, Create, false},
		{"update", Before, Update, false},
		{"computed update", Finally, Update, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			sp := &creationOnlyTagService{}
			conn := &conns.AWSClient{}
			conn.SetServicePackages(ctx, map[string]conns.ServicePackage{"Test": sp})
			defaults := &tftags.DefaultConfig{Tags: tftags.New(ctx, map[string]string{"CreatedOn": "new", "created:by": "new", "Name": "new"})}
			ignore := &tftags.IgnoreConfig{UpdateKeys: tftags.New(ctx, []string{"CreatedOn"}), UpdateKeyPrefixes: tftags.New(ctx, []string{"created:"})}
			conns.SetDefaultTagsConfig(conn, defaults)
			conns.SetIgnoreTagsConfig(conn, ignore)
			ctx = conns.NewResourceContext(ctx, "Test", "test", "aws_test", "")
			ctx = tftags.NewContext(ctx, defaults, ignore, nil)
			d := &creationOnlyTagData{unknown: tc.unknown}
			diags := resourceTransparentTagging(inttypes.ResourceTagsAttribute("id"), nil).run(ctx, crudInterceptorOptions{c: conn, d: d, when: tc.when, why: tc.why})
			if diags.HasError() {
				t.Fatal(diags)
			}
			if tc.why == Create {
				inContext, _ := tftags.FromContext(ctx)
				if got := inContext.TagsIn.MustUnwrap(); !got.Equal(defaults.Tags) {
					t.Fatalf("creation tags = %v, want %v", got, defaults.Tags)
				}
				if sp.called {
					t.Fatal("creation invoked update interceptor")
				}
				return
			}
			if !sp.called {
				t.Fatal("tag update was not called")
			}
			if sp.ignore == nil || !sp.ignore.Keys.Equal(ignore.UpdateKeys) || !sp.ignore.KeyPrefixes.Equal(ignore.UpdateKeyPrefixes) || len(ignore.Keys) != 0 || len(ignore.KeyPrefixes) != 0 {
				t.Fatal("tag updater did not isolate creation-only filters for propagation waiters")
			}
			if tc.when == Before {
				inContext, _ := tftags.FromContext(ctx)
				want := tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "new"})
				if !inContext.TagsIn.MustUnwrap().Equal(want) {
					t.Fatalf("replacement update tags = %v, want %v", inContext.TagsIn.MustUnwrap(), want)
				}
				if inContext.TagsOut.IsSome() {
					t.Fatal("preservation read leaked stale tags into update context")
				}
			}
			if !sp.old.Equal(tftags.New(ctx, map[string]string{"Name": "old"})) || !sp.new.Equal(tftags.New(ctx, map[string]string{"Name": "new"})) {
				t.Fatalf("update included creation-only tags: old=%v new=%v", sp.old, sp.new)
			}
		})
	}
}

type creationOnlyTagService struct {
	mockService
	called   bool
	ignore   *tftags.IgnoreConfig
	old, new tftags.KeyValueTags
}

func (s *creationOnlyTagService) UpdateTags(ctx context.Context, _ any, _ string, old, new any) error {
	s.called = true
	if inContext, ok := tftags.FromContext(ctx); ok {
		s.ignore = inContext.IgnoreConfig
	}
	s.old, s.new = tftags.New(ctx, old), tftags.New(ctx, new)
	return nil
}

func (s *creationOnlyTagService) ListTags(ctx context.Context, _ any, _ string) error {
	inContext, _ := tftags.FromContext(ctx)
	inContext.TagsOut = option.Some(tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "original", "Name": "new"}))
	return nil
}

type creationOnlyTagData struct {
	resourceData
	unknown bool
}

func (d *creationOnlyTagData) Get(key string) any {
	if key == "tags" {
		return map[string]any{}
	}
	return nil
}

func (d *creationOnlyTagData) GetRawState() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"id":       cty.StringVal("id"),
		"tags":     cty.MapValEmpty(cty.String),
		"tags_all": cty.MapVal(map[string]cty.Value{"CreatedOn": cty.StringVal("original"), "created:by": cty.StringVal("original"), "Name": cty.StringVal("old")}),
	})
}

func (d *creationOnlyTagData) GetRawConfig() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{"tags": cty.MapValEmpty(cty.String)})
}

func (d *creationOnlyTagData) GetRawPlan() cty.Value {
	tagsAll := cty.MapVal(map[string]cty.Value{"Name": cty.StringVal("new")})
	if d.unknown {
		tagsAll = cty.UnknownVal(cty.Map(cty.String))
	}
	return cty.ObjectVal(map[string]cty.Value{"tags": cty.MapValEmpty(cty.String), "tags_all": tagsAll})
}

func (d *creationOnlyTagData) HasChange(key string) bool {
	return key == "tags_all"
}

func (d *creationOnlyTagData) GetChange(key string) (any, any) {
	return map[string]any{"CreatedOn": "original", "created:by": "original", "Name": "old"}, map[string]any{"CreatedOn": "new", "created:by": "new", "Name": "new"}
}

func TestSetTagsAllIgnoreTagUpdates(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	conn := &conns.AWSClient{}
	conns.SetDefaultTagsConfig(conn, &tftags.DefaultConfig{Tags: tftags.New(ctx, map[string]string{"CreatedOn": "changed", "Name": "example"})})
	conns.SetIgnoreTagsConfig(conn, &tftags.IgnoreConfig{UpdateKeys: tftags.New(ctx, []string{"CreatedOn"})})
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{"tags": tftags.TagsSchema(), "tags_all": tftags.TagsSchemaComputed()},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, _ any) error {
			return setTagsAll().run(ctx, customizeDiffInterceptorOptions{c: conn, d: d, when: Before, why: CustomizeDiff})
		},
	}
	state := &terraform.InstanceState{
		ID:         "test",
		Attributes: map[string]string{"tags.%": "0", "tags_all.%": "1", "tags_all.Name": "example"},
		RawPlan: cty.ObjectVal(map[string]cty.Value{
			"tags":     cty.NullVal(cty.Map(cty.String)),
			"tags_all": cty.UnknownVal(cty.Map(cty.String)),
		}),
	}
	diff, err := r.Diff(ctx, state, terraform.NewResourceConfigRaw(map[string]any{}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff != nil && !diff.Empty() {
		t.Fatalf("creation-only default tag produced a diff: %v", diff)
	}
}

func TestTagsResourceInterceptorIgnoreTagUpdatesInline(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                    string
		defaults, current, want map[string]string
	}{
		{"replacement", map[string]string{"CreatedOn": "new", "Name": "new"}, map[string]string{"CreatedOn": "original", "Name": "old"}, map[string]string{"CreatedOn": "original", "Name": "new"}},
		{"only creation tag", map[string]string{"CreatedOn": "new"}, map[string]string{"CreatedOn": "original"}, map[string]string{"CreatedOn": "original"}},
		{"missing creation tag", map[string]string{"CreatedOn": "new", "Name": "new"}, map[string]string{}, map[string]string{"Name": "new"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			conn := &conns.AWSClient{}
			conn.SetServicePackages(ctx, map[string]conns.ServicePackage{"Test": &creationOnlyTagService{}})
			defaults := &tftags.DefaultConfig{Tags: tftags.New(ctx, tc.defaults)}
			ignore := &tftags.IgnoreConfig{UpdateKeys: tftags.New(ctx, []string{"CreatedOn"})}
			conns.SetDefaultTagsConfig(conn, defaults)
			conns.SetIgnoreTagsConfig(conn, ignore)
			ctx = conns.NewResourceContext(ctx, "Test", "test", "aws_test", "")
			ctx = tftags.NewContext(ctx, defaults, ignore, nil)
			var reads int
			r := &schema.Resource{
				Schema: map[string]*schema.Schema{"tags": tftags.TagsSchema(), "tags_all": tftags.TagsSchemaComputed()},
				ReadWithoutTimeout: func(ctx context.Context, data *schema.ResourceData, _ any) diag.Diagnostics {
					reads++
					inContext, _ := tftags.FromContext(ctx)
					inContext.TagsOut = option.Some(tftags.New(ctx, tc.current))
					if err := data.Set("tags_all", tc.current); err != nil {
						return sdkdiag.AppendFromErr(nil, err)
					}
					return nil
				},
			}
			d := &creationOnlyTagData{}
			diags := resourceTransparentTagging(inttypes.ResourceTagsInline(), r).run(ctx, crudInterceptorOptions{c: conn, d: d, when: Before, why: Update})
			if diags.HasError() {
				t.Fatal(diags)
			}
			inContext, _ := tftags.FromContext(ctx)
			if reads != 1 || !inContext.TagsIn.MustUnwrap().Equal(tftags.New(ctx, tc.want)) || inContext.TagsOut.IsSome() {
				t.Fatalf("reads=%d replacement tags=%v want=%v stale TagsOut=%v", reads, inContext.TagsIn.MustUnwrap(), tc.want, inContext.TagsOut)
			}
		})
	}
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
	d := &tagDataSourceData{}
	diags := dataSourceTransparentTagging(inttypes.ResourceTagsInline()).run(ctx, crudInterceptorOptions{c: conn, d: d, when: After, why: Read})
	want := tftags.New(ctx, map[string]string{"CreatedOn": "original", "created:by": "terraform"})
	if diags.HasError() || !tftags.New(ctx, d.tags).Equal(want) {
		t.Fatalf("data source tags=%v want=%v diagnostics=%v", d.tags, want, diags)
	}
}

type tagDataSourceData struct {
	resourceData
	tags any
}

func (d *tagDataSourceData) Set(key string, value any) error {
	if key == "tags" {
		d.tags = value
	}
	return nil
}
