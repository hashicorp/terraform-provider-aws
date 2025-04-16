// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{
		{
			Factory:  newResourcePolicyResource,
			TypeName: "aws_dynamodb_resource_policy",
			Name:     "Resource Policy",
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceTable,
			TypeName: "aws_dynamodb_table",
			Name:     "Table",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  dataSourceTableItem,
			TypeName: "aws_dynamodb_table_item",
			Name:     "Table Item",
		},
		{
			Factory:  DataSourceTableQuery,
			TypeName: "aws_dynamodb_table_query",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceContributorInsights,
			TypeName: "aws_dynamodb_contributor_insights",
			Name:     "Contributor Insights",
		},
		{
			Factory:  resourceGlobalTable,
			TypeName: "aws_dynamodb_global_table",
			Name:     "Global Table",
		},
		{
			Factory:  resourceKinesisStreamingDestination,
			TypeName: "aws_dynamodb_kinesis_streaming_destination",
			Name:     "Kinesis Streaming Destination",
		},
		{
			Factory:  resourceTable,
			TypeName: "aws_dynamodb_table",
			Name:     "Table",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceTableExport,
			TypeName: "aws_dynamodb_table_export",
			Name:     "Table Export",
		},
		{
			Factory:  resourceTableItem,
			TypeName: "aws_dynamodb_table_item",
			Name:     "Table Item",
		},
		{
			Factory:  resourceTableReplica,
			TypeName: "aws_dynamodb_table_replica",
			Name:     "Table Replica",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceTag,
			TypeName: "aws_dynamodb_tag",
			Name:     "DynamoDB Resource Tag",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.DynamoDB
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*dynamodb.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*dynamodb.Options){
		dynamodb.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return dynamodb.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*dynamodb.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*dynamodb.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *dynamodb.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*dynamodb.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
