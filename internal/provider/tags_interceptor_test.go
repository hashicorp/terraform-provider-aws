package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-cty/cty"
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

	sp := &types.ServicePackageResourceTags{
		IdentifierAttribute: "arn",
	}

	tags := tagsInterceptor{tags: sp}

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

	d := &differ{}

	_, _ = finalTagsUpdate(context.Background(), d, &mockService{}, nil, nil, sp, "TestService", "Test", conn)
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
