package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
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
	return []*types.ServicePackageSDKResource{
		{
			Factory:  nil,
			TypeName: "aws_test",
			Name:     "Test",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
	}
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

func (t *mockService) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return nil
}

func TestTagsInterceptor(t *testing.T) {
	t.Parallel()

	var interceptors interceptorItems

	tags := tagsInterceptor{tags: &types.ServicePackageResourceTags{
		IdentifierAttribute: "arn",
	}}

	interceptors = append(interceptors, interceptorItem{
		when:        Finally,
		why:         Update,
		interceptor: tags,
	})

	var update schema.UpdateContextFunc = func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics
		return diags
	}
	bootstrapContext := func(ctx context.Context, meta any) context.Context {
		return ctx
	}

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

	r := &schema.Resource{}
	r.Schema = map[string]*schema.Schema{
		names.AttrTags:    tftags.TagsSchema(),
		names.AttrTagsAll: tftags.TagsSchemaComputed(),
	}

	// diff := &schema.ResourceDiff{}

	// _ = diff.SetNew("tags_all", map[string]string{"one": "tag"})
	//r.CustomizeDiff = func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	//	t.Log(diff)
	//	return diff.SetNewComputed(names.AttrTagsAll)
	//}
	//err := r.CustomizeDiff(context.Background(), diff, conn)
	//if err != nil {
	//	t.Fatal(err)
	//}
	diags := interceptedHandler(bootstrapContext, interceptors, update, Update)(context.Background(), r.TestResourceData(), conn)
	if got, want := len(diags), 0; got != want {
		t.Errorf("length of diags = %v, want %v", got, want)
	}
	outputTags := r.TestResourceData().Get("tags_all").(map[string]interface{})
	t.Log(outputTags)
}

type differ struct{}

func (d *differ) GetRawConfig() cty.Value {
	return cty.MapVal(map[string]cty.Value{
		"tags_all": cty.StringVal(""),
	})
}

func (d *differ) GetRawPlan() cty.Value {
	return cty.MapVal(map[string]cty.Value{
		"tags_all": cty.MapVal(map[string]cty.Value{
			"tag1": cty.UnknownVal(cty.String),
		}),
	})
}

func (d *differ) GetRawState() cty.Value {
	return cty.Value{}
}

func (d *differ) Get(key string) any {
	return nil
}

func (d *differ) Id() string {
	return "id"
}

func (d *differ) Set(string, any) error {
	return nil
}
