// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package secretsmanager

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*types.ServicePackageEphemeralResource {
	return []*types.ServicePackageEphemeralResource{
		{
			Factory:  newEphemeralRandomPassword,
			TypeName: "aws_secretsmanager_random_password",
			Name:     "Random Password",
		},
		{
			Factory:  newEphemeralSecretVersion,
			TypeName: "aws_secretsmanager_secret_version",
			Name:     "Secret Version",
		},
	}
}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceSecretVersions,
			TypeName: "aws_secretsmanager_secret_versions",
			Name:     "Secret Versions",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceRandomPassword,
			TypeName: "aws_secretsmanager_random_password",
			Name:     "Random Password",
		},
		{
			Factory:  dataSourceSecret,
			TypeName: "aws_secretsmanager_secret",
			Name:     "Secret",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
		{
			Factory:  dataSourceSecretRotation,
			TypeName: "aws_secretsmanager_secret_rotation",
			Name:     "Secret Rotation",
		},
		{
			Factory:  dataSourceSecretVersion,
			TypeName: "aws_secretsmanager_secret_version",
			Name:     "Secret Version",
		},
		{
			Factory:  dataSourceSecrets,
			TypeName: "aws_secretsmanager_secrets",
			Name:     "Secrets",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceSecret,
			TypeName: "aws_secretsmanager_secret",
			Name:     "Secret",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
		},
		{
			Factory:  resourceSecretPolicy,
			TypeName: "aws_secretsmanager_secret_policy",
			Name:     "Secret Policy",
		},
		{
			Factory:  resourceSecretRotation,
			TypeName: "aws_secretsmanager_secret_rotation",
			Name:     "Secret Rotation",
		},
		{
			Factory:  resourceSecretVersion,
			TypeName: "aws_secretsmanager_secret_version",
			Name:     "Secret Version",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.SecretsManager
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*secretsmanager.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*secretsmanager.Options){
		secretsmanager.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return secretsmanager.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*secretsmanager.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*secretsmanager.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *secretsmanager.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*secretsmanager.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
