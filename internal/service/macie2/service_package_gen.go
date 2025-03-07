// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package macie2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
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
			Factory:  newOrganizationConfigurationResource,
			TypeName: "aws_macie2_organization_configuration",
			Name:     "Organization Configuration",
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceAccount,
			TypeName: "aws_macie2_account",
			Name:     "Account",
		},
		{
			Factory:  resourceClassificationExportConfiguration,
			TypeName: "aws_macie2_classification_export_configuration",
			Name:     "Classification Export Configuration",
		},
		{
			Factory:  resourceClassificationJob,
			TypeName: "aws_macie2_classification_job",
			Name:     "Classification Job",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "job_arn",
			},
		},
		{
			Factory:  resourceCustomDataIdentifier,
			TypeName: "aws_macie2_custom_data_identifier",
			Name:     "Custom Data Identifier",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceFindingsFilter,
			TypeName: "aws_macie2_findings_filter",
			Name:     "Findings Filter",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceInvitationAccepter,
			TypeName: "aws_macie2_invitation_accepter",
			Name:     "Invitation Accepter",
		},
		{
			Factory:  resourceMember,
			TypeName: "aws_macie2_member",
			Name:     "Member",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceOrganizationAdminAccount,
			TypeName: "aws_macie2_organization_admin_account",
			Name:     "Organization Admin Account",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Macie2
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*macie2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*macie2.Options){
		macie2.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return macie2.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*macie2.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*macie2.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *macie2.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*macie2.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
