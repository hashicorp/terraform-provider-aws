// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

type mockService struct{}

func (t *mockService) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (t *mockService) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (t *mockService) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (t *mockService) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{}
}

func (t *mockService) ServicePackageName() string {
	return "TestService"
}

func (t *mockService) ListTags(ctx context.Context, meta any, identifier string) error {
	tags := tftags.New(ctx, map[string]string{
		"tag1": "value1",
	})
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(tags)
	}

	return errors.New("test error")
}

func (t *mockService) UpdateTags(context.Context, any, string, string, any) error {
	return nil
}

func TestTagsResourceInterceptor(t *testing.T) {
	t.Parallel()

	var interceptors interceptorItems

	sp := &types.ServicePackageResourceTags{
		IdentifierAttribute: "id",
	}

	tags := tagsResourceInterceptor{
		tags:       sp,
		updateFunc: tagsUpdateFunc,
		readFunc:   tagsReadFunc,
	}

	interceptors = append(interceptors, interceptorItem{
		when:        Finally,
		why:         Update,
		interceptor: tags,
	})

	conn := &conns.AWSClient{
		ServicePackages: map[string]conns.ServicePackage{
			"Test": &mockService{},
		},
		DefaultTagsConfig: expandDefaultTags(context.Background(), map[string]interface{}{
			"tag": "",
		}),
		IgnoreTagsConfig: expandIgnoreTags(context.Background(), map[string]interface{}{
			"tag2": "tag",
		}),
	}

	bootstrapContext := func(ctx context.Context, meta any) context.Context {
		ctx = conns.NewResourceContext(ctx, "Test", "aws_test")
		if v, ok := meta.(*conns.AWSClient); ok {
			ctx = tftags.NewContext(ctx, v.DefaultTagsConfig, v.IgnoreTagsConfig)
		}

		return ctx
	}

	ctx := bootstrapContext(context.Background(), conn)
	d := &resourceData{}

	for _, v := range interceptors {
		var diags diag.Diagnostics
		_, diags = v.interceptor.run(ctx, d, conn, v.when, v.why, diags)
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

func (d *resourceData) Id() string {
	return "id"
}

func (d *resourceData) Set(string, any) error {
	return nil
}

func (d *resourceData) GetChange(key string) (interface{}, interface{}) {
	return nil, nil
}

func (d *resourceData) HasChange(key string) bool {
	return false
}
