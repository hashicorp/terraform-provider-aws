// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_directory")
func DataSourceDirectory() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDirectoryRead,

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"directory_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"registration_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"self_service_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"change_compute_type": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"increase_volume_size": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"rebuild_workspace": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"restart_workspace": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"switch_running_mode": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
			"workspace_access_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_type_android": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_chromeos": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_ios": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_linux": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_osx": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_web": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_windows": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type_zeroclient": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"workspace_creation_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_ou": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_internet_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_maintenance_mode": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"user_enabled_as_local_administrator": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"workspace_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	directoryID := d.Get("directory_id").(string)

	rawOutput, state, err := StatusDirectoryState(ctx, conn, directoryID)()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WorkSpaces Directory (%s): %s", directoryID, err)
	}
	if state == string(types.WorkspaceDirectoryStateDeregistered) {
		return sdkdiag.AppendErrorf(diags, "WorkSpaces directory %s was not found", directoryID)
	}

	d.SetId(directoryID)

	directory := rawOutput.(*types.WorkspaceDirectory)
	d.Set("directory_id", directory.DirectoryId)
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("registration_code", directory.RegistrationCode)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set(names.AttrAlias, directory.Alias)

	if err := d.Set(names.AttrSubnetIDs, flex.FlattenStringValueSet(directory.SubnetIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	if err := d.Set("self_service_permissions", FlattenSelfServicePermissions(directory.SelfservicePermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting self_service_permissions: %s", err)
	}

	if err := d.Set("workspace_access_properties", FlattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_access_properties: %s", err)
	}

	if err := d.Set("workspace_creation_properties", FlattenWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_creation_properties: %s", err)
	}

	if err := d.Set("ip_group_ids", flex.FlattenStringValueSet(directory.IpGroupIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_group_ids: %s", err)
	}

	if err := d.Set("dns_ip_addresses", flex.FlattenStringValueSet(directory.DnsIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dns_ip_addresses: %s", err)
	}

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags: %s", err)
	}
	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
