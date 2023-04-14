// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package workspaces

import (
	"context"

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
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceBundle,
			TypeName: "aws_workspaces_bundle",
		},
		{
			Factory:  DataSourceDirectory,
			TypeName: "aws_workspaces_directory",
		},
		{
			Factory:  DataSourceImage,
			TypeName: "aws_workspaces_image",
		},
		{
			Factory:  DataSourceWorkspace,
			TypeName: "aws_workspaces_workspace",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceDirectory,
			TypeName: "aws_workspaces_directory",
			Name:     "Directory",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "id",
			},
		},
		{
			Factory:  ResourceIPGroup,
			TypeName: "aws_workspaces_ip_group",
			Name:     "IP Group",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "id",
			},
		},
		{
			Factory:  ResourceWorkspace,
			TypeName: "aws_workspaces_workspace",
			Name:     "Workspace",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: "id",
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.WorkSpaces
}

var ServicePackage = &servicePackage{}
