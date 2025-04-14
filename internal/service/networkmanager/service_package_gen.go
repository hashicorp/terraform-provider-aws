// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package networkmanager

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newDirectConnectGatewayAttachmentResource,
			TypeName: "aws_networkmanager_dx_gateway_attachment",
			Name:     "Direct Connect Gateway Attachment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceConnection,
			TypeName: "aws_networkmanager_connection",
			Name:     "Connection",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceConnections,
			TypeName: "aws_networkmanager_connections",
			Name:     "Connections",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceCoreNetworkPolicyDocument,
			TypeName: "aws_networkmanager_core_network_policy_document",
			Name:     "Core Network Policy Document",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceDevice,
			TypeName: "aws_networkmanager_device",
			Name:     "Device",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceDevices,
			TypeName: "aws_networkmanager_devices",
			Name:     "Devices",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceGlobalNetwork,
			TypeName: "aws_networkmanager_global_network",
			Name:     "Global Network",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceGlobalNetworks,
			TypeName: "aws_networkmanager_global_networks",
			Name:     "Global Networks",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceLink,
			TypeName: "aws_networkmanager_link",
			Name:     "Link",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceLinks,
			TypeName: "aws_networkmanager_links",
			Name:     "Links",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceSite,
			TypeName: "aws_networkmanager_site",
			Name:     "Site",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceSites,
			TypeName: "aws_networkmanager_sites",
			Name:     "Sites",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceAttachmentAccepter,
			TypeName: "aws_networkmanager_attachment_accepter",
			Name:     "Attachment Accepter",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceConnectAttachment,
			TypeName: "aws_networkmanager_connect_attachment",
			Name:     "Connect Attachment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceConnectPeer,
			TypeName: "aws_networkmanager_connect_peer",
			Name:     "Connect Peer",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceConnection,
			TypeName: "aws_networkmanager_connection",
			Name:     "Connection",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceCoreNetwork,
			TypeName: "aws_networkmanager_core_network",
			Name:     "Core Network",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceCoreNetworkPolicyAttachment,
			TypeName: "aws_networkmanager_core_network_policy_attachment",
			Name:     "Core Network Policy Attachment",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceCustomerGatewayAssociation,
			TypeName: "aws_networkmanager_customer_gateway_association",
			Name:     "Customer Gateway Association",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceDevice,
			TypeName: "aws_networkmanager_device",
			Name:     "Device",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceGlobalNetwork,
			TypeName: "aws_networkmanager_global_network",
			Name:     "Global Network",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceLink,
			TypeName: "aws_networkmanager_link",
			Name:     "Link",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceLinkAssociation,
			TypeName: "aws_networkmanager_link_association",
			Name:     "Link Association",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSite,
			TypeName: "aws_networkmanager_site",
			Name:     "Site",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSiteToSiteVPNAttachment,
			TypeName: "aws_networkmanager_site_to_site_vpn_attachment",
			Name:     "Site To Site VPN Attachment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceTransitGatewayConnectPeerAssociation,
			TypeName: "aws_networkmanager_transit_gateway_connect_peer_association",
			Name:     "Transit Gateway Connect Peer Association",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceTransitGatewayPeering,
			TypeName: "aws_networkmanager_transit_gateway_peering",
			Name:     "Transit Gateway Peering",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceTransitGatewayRegistration,
			TypeName: "aws_networkmanager_transit_gateway_registration",
			Name:     "Transit Gateway Registration",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceTransitGatewayRouteTableAttachment,
			TypeName: "aws_networkmanager_transit_gateway_route_table_attachment",
			Name:     "Transit Gateway Route Table Attachment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceVPCAttachment,
			TypeName: "aws_networkmanager_vpc_attachment",
			Name:     "VPC Attachment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      true,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.NetworkManager
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*networkmanager.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*networkmanager.Options){
		networkmanager.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *networkmanager.Options) {
			if region := config[names.AttrRegion].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         p.ServicePackageName(),
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return networkmanager.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*networkmanager.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*networkmanager.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *networkmanager.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*networkmanager.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
