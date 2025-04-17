// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package auditmanager

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceControl,
			TypeName: "aws_auditmanager_control",
			Name:     "Control",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newDataSourceFramework,
			TypeName: "aws_auditmanager_framework",
			Name:     "Framework",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceAccountRegistration,
			TypeName: "aws_auditmanager_account_registration",
			Name:     "Account Registration",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceAssessment,
			TypeName: "aws_auditmanager_assessment",
			Name:     "Assessment",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceAssessmentDelegation,
			TypeName: "aws_auditmanager_assessment_delegation",
			Name:     "Assessment Delegation",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceAssessmentReport,
			TypeName: "aws_auditmanager_assessment_report",
			Name:     "Assessment Report",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceControl,
			TypeName: "aws_auditmanager_control",
			Name:     "Control",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceFramework,
			TypeName: "aws_auditmanager_framework",
			Name:     "Framework",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceFrameworkShare,
			TypeName: "aws_auditmanager_framework_share",
			Name:     "Framework Share",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  newResourceOrganizationAdminAccountRegistration,
			TypeName: "aws_auditmanager_organization_admin_account_registration",
			Name:     "Organization Admin Account Registration",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled: false,
			}),
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{}
}

func (p *servicePackage) ServicePackageName() string {
	return names.AuditManager
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*auditmanager.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*auditmanager.Options){
		auditmanager.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *auditmanager.Options) {
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

	return auditmanager.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*auditmanager.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*auditmanager.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *auditmanager.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*auditmanager.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
