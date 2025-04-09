// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package amplify

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceApp,
			TypeName: "aws_amplify_app",
			Name:     "App",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
		{
			Factory:  resourceBackendEnvironment,
			TypeName: "aws_amplify_backend_environment",
			Name:     "Backend Environment",
		},
		{
			Factory:  resourceBranch,
			TypeName: "aws_amplify_branch",
			Name:     "Branch",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
		{
			Factory:  resourceDomainAssociation,
			TypeName: "aws_amplify_domain_association",
			Name:     "Domain Association",
		},
		{
			Factory:  resourceWebhook,
			TypeName: "aws_amplify_webhook",
			Name:     "Webhook",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Amplify
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*amplify.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*amplify.Options){
		amplify.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return amplify.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*amplify.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*amplify.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *amplify.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*amplify.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
