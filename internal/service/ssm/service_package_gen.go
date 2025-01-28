// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*types.ServicePackageEphemeralResource {
	return []*types.ServicePackageEphemeralResource{
		{
			Factory:  newEphemeralParameter,
			TypeName: "aws_ssm_parameter",
			Name:     "Parameter",
		},
	}
}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourcePatchBaselines,
			TypeName: "aws_ssm_patch_baselines",
			Name:     "Patch Baselines",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceDocument,
			TypeName: "aws_ssm_document",
			Name:     "Document",
		},
		{
			Factory:  dataSourceInstances,
			TypeName: "aws_ssm_instances",
			Name:     "Instances",
		},
		{
			Factory:  dataSourceMaintenanceWindows,
			TypeName: "aws_ssm_maintenance_windows",
			Name:     "Maintenance Windows",
		},
		{
			Factory:  dataSourceParameter,
			TypeName: "aws_ssm_parameter",
			Name:     "Parameter",
		},
		{
			Factory:  dataSourceParametersByPath,
			TypeName: "aws_ssm_parameters_by_path",
			Name:     "Parameters By Path",
		},
		{
			Factory:  dataSourcePatchBaseline,
			TypeName: "aws_ssm_patch_baseline",
			Name:     "Patch Baseline",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceActivation,
			TypeName: "aws_ssm_activation",
			Name:     "Activation",
			Tags:     &types.ServicePackageResourceTags{},
		},
		{
			Factory:  resourceAssociation,
			TypeName: "aws_ssm_association",
			Name:     "Association",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "Association",
			},
		},
		{
			Factory:  resourceDefaultPatchBaseline,
			TypeName: "aws_ssm_default_patch_baseline",
			Name:     "Default Patch Baseline",
		},
		{
			Factory:  resourceDocument,
			TypeName: "aws_ssm_document",
			Name:     "Document",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "Document",
			},
		},
		{
			Factory:  resourceMaintenanceWindow,
			TypeName: "aws_ssm_maintenance_window",
			Name:     "Maintenance Window",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "MaintenanceWindow",
			},
		},
		{
			Factory:  resourceMaintenanceWindowTarget,
			TypeName: "aws_ssm_maintenance_window_target",
			Name:     "Maintenance Window Target",
		},
		{
			Factory:  resourceMaintenanceWindowTask,
			TypeName: "aws_ssm_maintenance_window_task",
			Name:     "Maintenance Window Task",
		},
		{
			Factory:  resourceParameter,
			TypeName: "aws_ssm_parameter",
			Name:     "Parameter",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "Parameter",
			},
		},
		{
			Factory:  resourcePatchBaseline,
			TypeName: "aws_ssm_patch_baseline",
			Name:     "Patch Baseline",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "PatchBaseline",
			},
		},
		{
			Factory:  resourcePatchGroup,
			TypeName: "aws_ssm_patch_group",
			Name:     "Patch Group",
		},
		{
			Factory:  resourceResourceDataSync,
			TypeName: "aws_ssm_resource_data_sync",
			Name:     "Resource Data Sync",
		},
		{
			Factory:  resourceServiceSetting,
			TypeName: "aws_ssm_service_setting",
			Name:     "Service Setting",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.SSM
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ssm.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return ssm.NewFromConfig(cfg,
		ssm.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
	), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
