// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package redshift

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceDataShares,
			TypeName: "aws_redshift_data_shares",
			Name:     "Data Shares",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newDataSourceProducerDataShares,
			TypeName: "aws_redshift_producer_data_shares",
			Name:     "Producer Data Shares",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceDataShareAuthorization,
			TypeName: "aws_redshift_data_share_authorization",
			Name:     "Data Share Authorization",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourceDataShareConsumerAssociation,
			TypeName: "aws_redshift_data_share_consumer_association",
			Name:     "Data Share Consumer Association",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourceLogging,
			TypeName: "aws_redshift_logging",
			Name:     "Logging",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourceSnapshotCopy,
			TypeName: "aws_redshift_snapshot_copy",
			Name:     "Snapshot Copy",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceCluster,
			TypeName: "aws_redshift_cluster",
			Name:     "Cluster",
			Tags:     unique.Make(inttypes.ServicePackageResourceTags{}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceClusterCredentials,
			TypeName: "aws_redshift_cluster_credentials",
			Name:     "Cluster Credentials",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceOrderableCluster,
			TypeName: "aws_redshift_orderable_cluster",
			Name:     "Orderable Cluster",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  dataSourceSubnetGroup,
			TypeName: "aws_redshift_subnet_group",
			Name:     "Subnet Group",
			Tags:     unique.Make(inttypes.ServicePackageResourceTags{}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceAuthenticationProfile,
			TypeName: "aws_redshift_authentication_profile",
			Name:     "Authentication Profile",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceCluster,
			TypeName: "aws_redshift_cluster",
			Name:     "Cluster",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceClusterIAMRoles,
			TypeName: "aws_redshift_cluster_iam_roles",
			Name:     "Cluster IAM Roles",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceClusterSnapshot,
			TypeName: "aws_redshift_cluster_snapshot",
			Name:     "Cluster Snapshot",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceEndpointAccess,
			TypeName: "aws_redshift_endpoint_access",
			Name:     "Endpoint Access",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceEndpointAuthorization,
			TypeName: "aws_redshift_endpoint_authorization",
			Name:     "Endpoint Authorization",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceEventSubscription,
			TypeName: "aws_redshift_event_subscription",
			Name:     "Event Subscription",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceHSMClientCertificate,
			TypeName: "aws_redshift_hsm_client_certificate",
			Name:     "HSM Client Certificate",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceHSMConfiguration,
			TypeName: "aws_redshift_hsm_configuration",
			Name:     "HSM Configuration",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceParameterGroup,
			TypeName: "aws_redshift_parameter_group",
			Name:     "Parameter Group",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourcePartner,
			TypeName: "aws_redshift_partner",
			Name:     "Partner",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceResourcePolicy,
			TypeName: "aws_redshift_resource_policy",
			Name:     "Resource Policy",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceScheduledAction,
			TypeName: "aws_redshift_scheduled_action",
			Name:     "Scheduled Action",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSnapshotCopyGrant,
			TypeName: "aws_redshift_snapshot_copy_grant",
			Name:     "Snapshot Copy Grant",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSnapshotSchedule,
			TypeName: "aws_redshift_snapshot_schedule",
			Name:     "Snapshot Schedule",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSnapshotScheduleAssociation,
			TypeName: "aws_redshift_snapshot_schedule_association",
			Name:     "Snapshot Schedule Association",
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceSubnetGroup,
			TypeName: "aws_redshift_subnet_group",
			Name:     "Subnet Group",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
		{
			Factory:  resourceUsageLimit,
			TypeName: "aws_redshift_usage_limit",
			Name:     "Usage Limit",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: &inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Redshift
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*redshift.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*redshift.Options){
		redshift.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *redshift.Options) {
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

	return redshift.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*redshift.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*redshift.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *redshift.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*redshift.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
