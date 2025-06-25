// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
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
		ctx = conns.NewResourceContext(ctx, "Test", "aws_test", "")
		if v, ok := meta.(*conns.AWSClient); ok {
			ctx = tftags.NewContext(ctx, v.DefaultTagsConfig(ctx), v.IgnoreTagsConfig(ctx))
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
		"tags_all": cty.MapVal(map[string]cty.Value{
			"tag1": cty.UnknownVal(cty.String),
		}),
	})
}

func (d *resourceData) GetRawState() cty.Value { // nosemgrep:ci.aws-in-func-name
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
