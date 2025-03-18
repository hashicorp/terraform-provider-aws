// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
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
			Factory:  dataSourceDelegatedAdministrators,
			TypeName: "aws_organizations_delegated_administrators",
			Name:     "Delegated Administrators",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceDelegatedServices,
			TypeName: "aws_organizations_delegated_services",
			Name:     "Delegated Services",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganization,
			TypeName: "aws_organizations_organization",
			Name:     "Organization",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganizationalUnit,
			TypeName: "aws_organizations_organizational_unit",
			Name:     "Organizational Unit",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganizationalUnitChildAccounts,
			TypeName: "aws_organizations_organizational_unit_child_accounts",
			Name:     "Organizational Unit Child Accounts",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganizationalUnitDescendantAccounts,
			TypeName: "aws_organizations_organizational_unit_descendant_accounts",
			Name:     "Organizational Unit Descendant Accounts",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganizationalUnitDescendantOrganizationalUnits,
			TypeName: "aws_organizations_organizational_unit_descendant_organizational_units",
			Name:     "Organizational Unit Descendant Organization Units",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOrganizationalUnits,
			TypeName: "aws_organizations_organizational_units",
			Name:     "Organizational Unit",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePolicies,
			TypeName: "aws_organizations_policies",
			Name:     "Policies",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePoliciesForTarget,
			TypeName: "aws_organizations_policies_for_target",
			Name:     "Policies For Target",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePolicy,
			TypeName: "aws_organizations_policy",
			Name:     "Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceResourceTags,
			TypeName: "aws_organizations_resource_tags",
			Name:     "Resource Tags",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceAccount,
			TypeName: "aws_organizations_account",
			Name:     "Account",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceDelegatedAdministrator,
			TypeName: "aws_organizations_delegated_administrator",
			Name:     "Delegated Administrator",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceOrganization,
			TypeName: "aws_organizations_organization",
			Name:     "Organization",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceOrganizationalUnit,
			TypeName: "aws_organizations_organizational_unit",
			Name:     "Organizational Unit",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourcePolicy,
			TypeName: "aws_organizations_policy",
			Name:     "Policy",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourcePolicyAttachment,
			TypeName: "aws_organizations_policy_attachment",
			Name:     "Policy Attachment",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceResourcePolicy,
			TypeName: "aws_organizations_resource_policy",
			Name:     "Resource Policy",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Organizations
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*organizations.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*organizations.Options){
		organizations.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *organizations.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "organizations",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return organizations.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*organizations.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*organizations.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *organizations.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*organizations.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
