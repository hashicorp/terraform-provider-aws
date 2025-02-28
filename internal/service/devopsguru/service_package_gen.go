// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package devopsguru

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceNotificationChannel,
			TypeName: "aws_devopsguru_notification_channel",
			Name:     "Notification Channel",
		},
		{
			Factory:  newDataSourceResourceCollection,
			TypeName: "aws_devopsguru_resource_collection",
			Name:     "Resource Collection",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceEventSourcesConfig,
			TypeName: "aws_devopsguru_event_sources_config",
			Name:     "Event Sources Config",
		},
		{
			Factory:  newResourceNotificationChannel,
			TypeName: "aws_devopsguru_notification_channel",
			Name:     "Notification Channel",
		},
		{
			Factory:  newResourceResourceCollection,
			TypeName: "aws_devopsguru_resource_collection",
			Name:     "Resource Collection",
		},
		{
			Factory:  newResourceServiceIntegration,
			TypeName: "aws_devopsguru_service_integration",
			Name:     "Service Integration",
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{}
}

func (p *servicePackage) ServicePackageName() string {
	return names.DevOpsGuru
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*devopsguru.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*devopsguru.Options){
		devopsguru.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *devopsguru.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "devopsguru",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return devopsguru.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*devopsguru.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*devopsguru.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *devopsguru.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*devopsguru.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
