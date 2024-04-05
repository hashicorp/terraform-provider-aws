// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package route53recoveryreadiness

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceCell,
			TypeName: "aws_route53recoveryreadiness_cell",
			Name:     "Cell",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
		{
			Factory:  ResourceReadinessCheck,
			TypeName: "aws_route53recoveryreadiness_readiness_check",
			Name:     "Readiness Check",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
		{
			Factory:  ResourceRecoveryGroup,
			TypeName: "aws_route53recoveryreadiness_recovery_group",
			Name:     "Recovery Group",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
		{
			Factory:  ResourceResourceSet,
			TypeName: "aws_route53recoveryreadiness_resource_set",
			Name:     "Resource Set",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "arn",
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Route53RecoveryReadiness
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
