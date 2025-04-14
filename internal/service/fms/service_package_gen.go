// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package fms

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{
		{
			Factory:  newResourceResourceSet,
			TypeName: "aws_fms_resource_set",
			Name:     "Resource Set",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceAdminAccount,
			TypeName: "aws_fms_admin_account",
			Name:     "Admin Account",
		},
		{
			Factory:  resourcePolicy,
			TypeName: "aws_fms_policy",
			Name:     "Policy",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.FMS
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*fms.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*fms.Options){
		fms.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return fms.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*fms.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*fms.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *fms.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*fms.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
