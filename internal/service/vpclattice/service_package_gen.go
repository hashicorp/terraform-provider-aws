// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
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
			Factory:  newResourceConfigurationResource,
			TypeName: "aws_vpclattice_resource_configuration",
			Name:     "Resource Configuration",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  newResourceGatewayResource,
			TypeName: "aws_vpclattice_resource_gateway",
			Name:     "Resource Gateway",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  newServiceNetworkResourceAssociationResource,
			TypeName: "aws_vpclattice_service_network_resource_association",
			Name:     "Service Network Resource Association",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceAuthPolicy,
			TypeName: "aws_vpclattice_auth_policy",
			Name:     "Auth Policy",
		},
		{
			Factory:  DataSourceListener,
			TypeName: "aws_vpclattice_listener",
			Name:     "Listener",
		},
		{
			Factory:  DataSourceResourcePolicy,
			TypeName: "aws_vpclattice_resource_policy",
			Name:     "Resource Policy",
		},
		{
			Factory:  dataSourceService,
			TypeName: "aws_vpclattice_service",
			Name:     "Service",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceServiceNetwork,
			TypeName: "aws_vpclattice_service_network",
			Name:     "Service Network",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceAccessLogSubscription,
			TypeName: "aws_vpclattice_access_log_subscription",
			Name:     "Access Log Subscription",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceAuthPolicy,
			TypeName: "aws_vpclattice_auth_policy",
			Name:     "Auth Policy",
		},
		{
			Factory:  resourceListener,
			TypeName: "aws_vpclattice_listener",
			Name:     "Listener",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceListenerRule,
			TypeName: "aws_vpclattice_listener_rule",
			Name:     "Listener Rule",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceResourcePolicy,
			TypeName: "aws_vpclattice_resource_policy",
			Name:     "Resource Policy",
		},
		{
			Factory:  resourceService,
			TypeName: "aws_vpclattice_service",
			Name:     "Service",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceServiceNetwork,
			TypeName: "aws_vpclattice_service_network",
			Name:     "Service Network",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceServiceNetworkServiceAssociation,
			TypeName: "aws_vpclattice_service_network_service_association",
			Name:     "Service Network Service Association",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceServiceNetworkVPCAssociation,
			TypeName: "aws_vpclattice_service_network_vpc_association",
			Name:     "Service Network VPC Association",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceTargetGroup,
			TypeName: "aws_vpclattice_target_group",
			Name:     "Target Group",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceTargetGroupAttachment,
			TypeName: "aws_vpclattice_target_group_attachment",
			Name:     "Target Group Attachment",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.VPCLattice
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*vpclattice.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*vpclattice.Options){
		vpclattice.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *vpclattice.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "vpclattice",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return vpclattice.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*vpclattice.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*vpclattice.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *vpclattice.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*vpclattice.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
