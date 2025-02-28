// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package location

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
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
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceGeofenceCollection,
			TypeName: "aws_location_geofence_collection",
			Name:     "Geofence Collection",
		},
		{
			Factory:  DataSourceMap,
			TypeName: "aws_location_map",
			Name:     "Map",
		},
		{
			Factory:  DataSourcePlaceIndex,
			TypeName: "aws_location_place_index",
			Name:     "Place Index",
		},
		{
			Factory:  DataSourceRouteCalculator,
			TypeName: "aws_location_route_calculator",
			Name:     "Route Calculator",
		},
		{
			Factory:  DataSourceTracker,
			TypeName: "aws_location_tracker",
			Name:     "Tracker",
		},
		{
			Factory:  DataSourceTrackerAssociation,
			TypeName: "aws_location_tracker_association",
			Name:     "Tracker Association",
		},
		{
			Factory:  DataSourceTrackerAssociations,
			TypeName: "aws_location_tracker_associations",
			Name:     "Tracker Associations",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  ResourceGeofenceCollection,
			TypeName: "aws_location_geofence_collection",
			Name:     "Geofence Collection",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "collection_arn",
			},
		},
		{
			Factory:  ResourceMap,
			TypeName: "aws_location_map",
			Name:     "Map",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "map_arn",
			},
		},
		{
			Factory:  ResourcePlaceIndex,
			TypeName: "aws_location_place_index",
			Name:     "Map",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "index_arn",
			},
		},
		{
			Factory:  ResourceRouteCalculator,
			TypeName: "aws_location_route_calculator",
			Name:     "Route Calculator",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "calculator_arn",
			},
		},
		{
			Factory:  ResourceTracker,
			TypeName: "aws_location_tracker",
			Name:     "Route Calculator",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "tracker_arn",
			},
		},
		{
			Factory:  ResourceTrackerAssociation,
			TypeName: "aws_location_tracker_association",
			Name:     "Tracker Association",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Location
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*location.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*location.Options){
		location.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *location.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "location",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return location.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*location.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*location.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *location.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*location.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
