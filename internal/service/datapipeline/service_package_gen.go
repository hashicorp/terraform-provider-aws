// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package datapipeline

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datapipeline"
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
	return []*inttypes.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourcePipeline,
			TypeName: "aws_datapipeline_pipeline",
			Name:     "Pipeline",
			Tags:     unique.Make(inttypes.ServicePackageResourceTags{}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  DataSourcePipelineDefinition,
			TypeName: "aws_datapipeline_pipeline_definition",
			Name:     "Pipeline Definition",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourcePipeline,
			TypeName: "aws_datapipeline_pipeline",
			Name:     "Pipeline",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "Pipeline",
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  ResourcePipelineDefinition,
			TypeName: "aws_datapipeline_pipeline_definition",
			Name:     "Pipeline Definition",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.DataPipeline
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*datapipeline.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*datapipeline.Options){
		datapipeline.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *datapipeline.Options) {
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

	return datapipeline.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*datapipeline.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*datapipeline.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *datapipeline.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*datapipeline.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
