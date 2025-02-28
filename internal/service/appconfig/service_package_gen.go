// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package appconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceEnvironment,
			TypeName: "aws_appconfig_environment",
			Name:     "Environment",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceConfigurationProfile,
			TypeName: "aws_appconfig_configuration_profile",
			Name:     "Configuration Profile",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  DataSourceConfigurationProfiles,
			TypeName: "aws_appconfig_configuration_profiles",
			Name:     "Configuration Profiles",
		},
		{
			Factory:  DataSourceEnvironment,
			TypeName: "aws_appconfig_environment",
			Name:     "Environment",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  DataSourceEnvironments,
			TypeName: "aws_appconfig_environments",
			Name:     "Environments",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  ResourceApplication,
			TypeName: "aws_appconfig_application",
			Name:     "Application",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceConfigurationProfile,
			TypeName: "aws_appconfig_configuration_profile",
			Name:     "Configuration Profile",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceDeployment,
			TypeName: "aws_appconfig_deployment",
			Name:     "Deployment",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceDeploymentStrategy,
			TypeName: "aws_appconfig_deployment_strategy",
			Name:     "Deployment Strategy",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceExtension,
			TypeName: "aws_appconfig_extension",
			Name:     "Extension",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceExtensionAssociation,
			TypeName: "aws_appconfig_extension_association",
			Name:     "Extension Association",
		},
		{
			Factory:  ResourceHostedConfigurationVersion,
			TypeName: "aws_appconfig_hosted_configuration_version",
			Name:     "Hosted Configuration Version",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.AppConfig
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*appconfig.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*appconfig.Options){
		appconfig.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *appconfig.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "appconfig",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return appconfig.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*appconfig.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*appconfig.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *appconfig.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*appconfig.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
