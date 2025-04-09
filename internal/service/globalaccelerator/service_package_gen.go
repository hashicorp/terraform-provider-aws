// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package globalaccelerator

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{
		{
			Factory:  newAcceleratorDataSource,
			TypeName: "aws_globalaccelerator_accelerator",
			Name:     "Accelerator",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{
		{
			Factory:  newCrossAccountAttachmentResource,
			TypeName: "aws_globalaccelerator_cross_account_attachment",
			Name:     "Cross-account Attachment",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			}),
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceCustomRoutingAccelerator,
			TypeName: "aws_globalaccelerator_custom_routing_accelerator",
			Name:     "Custom Routing Accelerator",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceAccelerator,
			TypeName: "aws_globalaccelerator_accelerator",
			Name:     "Accelerator",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			}),
		},
		{
			Factory:  resourceCustomRoutingAccelerator,
			TypeName: "aws_globalaccelerator_custom_routing_accelerator",
			Name:     "Custom Routing Accelerator",
			Tags: unique.Make(types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			}),
		},
		{
			Factory:  resourceCustomRoutingEndpointGroup,
			TypeName: "aws_globalaccelerator_custom_routing_endpoint_group",
			Name:     "Custom Routing Endpoint Group",
		},
		{
			Factory:  resourceCustomRoutingListener,
			TypeName: "aws_globalaccelerator_custom_routing_listener",
			Name:     "Custom Routing Listener",
		},
		{
			Factory:  resourceEndpointGroup,
			TypeName: "aws_globalaccelerator_endpoint_group",
			Name:     "Endpoint Group",
		},
		{
			Factory:  resourceListener,
			TypeName: "aws_globalaccelerator_listener",
			Name:     "Listener",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.GlobalAccelerator
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*globalaccelerator.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*globalaccelerator.Options){
		globalaccelerator.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *globalaccelerator.Options) {
			switch partition := config["partition"].(string); partition {
			case endpoints.AwsPartitionID:
				if region := endpoints.UsWest2RegionID; cfg.Region != region {
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": region,
					})
					o.Region = region
				}
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return globalaccelerator.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*globalaccelerator.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*globalaccelerator.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *globalaccelerator.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*globalaccelerator.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
