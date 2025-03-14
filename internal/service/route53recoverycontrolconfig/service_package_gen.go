// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package route53recoverycontrolconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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
	return []*itypes.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceCluster,
			TypeName: "aws_route53recoverycontrolconfig_cluster",
			Name:     "Cluster",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  resourceControlPanel,
			TypeName: "aws_route53recoverycontrolconfig_control_panel",
			Name:     "Control Panel",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  resourceRoutingControl,
			TypeName: "aws_route53recoverycontrolconfig_routing_control",
			Name:     "Routing Control",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  resourceSafetyRule,
			TypeName: "aws_route53recoverycontrolconfig_safety_rule",
			Name:     "Safety Rule",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Route53RecoveryControlConfig
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*route53recoverycontrolconfig.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*route53recoverycontrolconfig.Options){
		route53recoverycontrolconfig.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *route53recoverycontrolconfig.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "route53recoverycontrolconfig",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		func(o *route53recoverycontrolconfig.Options) {
			switch partition := config["partition"].(string); partition {
			case endpoints.AwsPartitionID:
				if region := endpoints.UsWest2RegionID; o.Region != region {
					tflog.Info(ctx, "overriding effective AWS API region", map[string]any{
						"service":         "route53recoverycontrolconfig",
						"original_region": o.Region,
						"override_region": region,
					})
					o.Region = region
				}
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return route53recoverycontrolconfig.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*route53recoverycontrolconfig.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*route53recoverycontrolconfig.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *route53recoverycontrolconfig.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*route53recoverycontrolconfig.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
