// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newClustersDataSource,
			TypeName: "aws_ecs_clusters",
			Name:     "Clusters",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceCluster,
			TypeName: "aws_ecs_cluster",
			Name:     "Cluster",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceContainerDefinition,
			TypeName: "aws_ecs_container_definition",
			Name:     "Container Definition",
		},
		{
			Factory:  dataSourceService,
			TypeName: "aws_ecs_service",
			Name:     "Service",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceTaskDefinition,
			TypeName: "aws_ecs_task_definition",
			Name:     "Task Definition",
		},
		{
			Factory:  dataSourceTaskExecution,
			TypeName: "aws_ecs_task_execution",
			Name:     "Task Execution",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceAccountSettingDefault,
			TypeName: "aws_ecs_account_setting_default",
			Name:     "Account Setting Default",
		},
		{
			Factory:  resourceCapacityProvider,
			TypeName: "aws_ecs_capacity_provider",
			Name:     "Capacity Provider",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
		},
		{
			Factory:  resourceCluster,
			TypeName: "aws_ecs_cluster",
			Name:     "Cluster",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
		},
		{
			Factory:  resourceClusterCapacityProviders,
			TypeName: "aws_ecs_cluster_capacity_providers",
			Name:     "Cluster Capacity Providers",
		},
		{
			Factory:  resourceService,
			TypeName: "aws_ecs_service",
			Name:     "Service",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
		},
		{
			Factory:  resourceTag,
			TypeName: "aws_ecs_tag",
			Name:     "ECS Resource Tag",
		},
		{
			Factory:  resourceTaskDefinition,
			TypeName: "aws_ecs_task_definition",
			Name:     "Task Definition",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceTaskSet,
			TypeName: "aws_ecs_task_set",
			Name:     "Task Set",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.ECS
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ecs.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*ecs.Options){
		ecs.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *ecs.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "ecs",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return ecs.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*ecs.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*ecs.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *ecs.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*ecs.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
