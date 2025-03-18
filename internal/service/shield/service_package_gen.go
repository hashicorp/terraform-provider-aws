// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package shield

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceProtection,
			TypeName: "aws_shield_protection",
			Name:     "Protection",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newApplicationLayerAutomaticResponseResource,
			TypeName: "aws_shield_application_layer_automatic_response",
			Name:     "Application Layer Automatic Response",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newDRTAccessLogBucketAssociationResource,
			TypeName: "aws_shield_drt_access_log_bucket_association",
			Name:     "DRT Log Bucket Association",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newDRTAccessRoleARNAssociationResource,
			TypeName: "aws_shield_drt_access_role_arn_association",
			Name:     "DRT Role ARN Association",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newProactiveEngagementResource,
			TypeName: "aws_shield_proactive_engagement",
			Name:     "Proactive Engagement",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceSubscription,
			TypeName: "aws_shield_subscription",
			Name:     "Subscription",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceProtection,
			TypeName: "aws_shield_protection",
			Name:     "Protection",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  ResourceProtectionGroup,
			TypeName: "aws_shield_protection_group",
			Name:     "Protection Group",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: "protection_group_arn",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  ResourceProtectionHealthCheckAssociation,
			TypeName: "aws_shield_protection_health_check_association",
			Name:     "Protection Health Check Association",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Shield
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*shield.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*shield.Options){
		shield.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *shield.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "shield",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		func(o *shield.Options) {
			switch partition := config["partition"].(string); partition {
			case endpoints.AwsPartitionID:
				if region := endpoints.UsEast1RegionID; o.Region != region {
					tflog.Info(ctx, "overriding effective AWS API region", map[string]any{
						"service":         "shield",
						"original_region": o.Region,
						"override_region": region,
					})
					o.Region = region
				}
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return shield.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*shield.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*shield.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *shield.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*shield.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
