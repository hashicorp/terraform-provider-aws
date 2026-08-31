// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
	"maps"
	"testing"
	"unique"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

func TestCalculateTagsAll(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	testCases := map[string]struct {
		defaultTags  map[string]string
		resourceTags map[string]string
		want         map[string]string
	}{
		"system tag only": {
			resourceTags: map[string]string{
				"aws:autoscaling-group-123": "managed",
			},
			want: map[string]string{},
		},
		"system tag excluded while resource and default tags remain": {
			defaultTags: map[string]string{
				"default": "value",
			},
			resourceTags: map[string]string{
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

			got := calculateTagsAll(
				&tftags.DefaultConfig{Tags: tftags.New(ctx, testCase.defaultTags)},
				nil,
				"TestService",
				tftags.New(ctx, testCase.resourceTags),
			).Map()

			if !maps.Equal(got, testCase.want) {
				t.Errorf("tags_all = %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestTagsResourceInterceptor(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var interceptors interceptorInvocations
	sp := unique.Make(inttypes.ServicePackageResourceTags{
		IdentifierAttribute: "id",
	})
	tags := resourceTransparentTagging(sp)
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
