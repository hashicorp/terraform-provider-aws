// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package backup

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	backup_sdkv2 "github.com/aws/aws-sdk-go-v2/service/backup"
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
			Factory: newLogicallyAirGappedVaultResource,
			Name:    "Logically Air Gapped Vault",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceFramework,
			TypeName: "aws_backup_framework",
			Name:     "Framework",
		},
		{
			Factory:  dataSourcePlan,
			TypeName: "aws_backup_plan",
			Name:     "Plan",
		},
		{
			Factory:  DataSourceReportPlan,
			TypeName: "aws_backup_report_plan",
		},
		{
			Factory:  DataSourceSelection,
			TypeName: "aws_backup_selection",
		},
		{
			Factory:  DataSourceVault,
			TypeName: "aws_backup_vault",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceFramework,
			TypeName: "aws_backup_framework",
			Name:     "Framework",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceGlobalSettings,
			TypeName: "aws_backup_global_settings",
			Name:     "Global Settings",
		},
		{
			Factory:  resourcePlan,
			TypeName: "aws_backup_plan",
			Name:     "Plan",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRegionSettings,
			TypeName: "aws_backup_region_settings",
			Name:     "Region Settings",
		},
		{
			Factory:  resourceReportPlan,
			TypeName: "aws_backup_report_plan",
			Name:     "Report Plan",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceSelection,
			TypeName: "aws_backup_selection",
		},
		{
			Factory:  ResourceVault,
			TypeName: "aws_backup_vault",
			Name:     "Vault",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceVaultLockConfiguration,
			TypeName: "aws_backup_vault_lock_configuration",
		},
		{
			Factory:  ResourceVaultNotifications,
			TypeName: "aws_backup_vault_notifications",
		},
		{
			Factory:  ResourceVaultPolicy,
			TypeName: "aws_backup_vault_policy",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Backup
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*backup_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return backup_sdkv2.NewFromConfig(cfg,
		backup_sdkv2.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
	), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
