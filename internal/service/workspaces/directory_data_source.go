// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_directory", name="Directory")
// @Tags(identifierAttribute="id")
func dataSourceDirectory() *schema.Resource {
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
			"saml_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"relay_state_parameter_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_access_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
			names.AttrTags: tftags.TagsSchemaComputed(),
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

func dataSourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	directoryID := d.Get("directory_id").(string)
	directory, err := findDirectoryByID(ctx, conn, directoryID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Directory (%s): %s", directoryID, err)
	}

	d.SetId(directoryID)
	d.Set(names.AttrAlias, directory.Alias)
	d.Set("directory_id", directory.DirectoryId)
	d.Set("directory_name", directory.DirectoryName)
	d.Set("directory_type", directory.DirectoryType)
	d.Set("dns_ip_addresses", directory.DnsIpAddresses)
	d.Set("iam_role_id", directory.IamRoleId)
	d.Set("ip_group_ids", directory.IpGroupIds)
	d.Set("registration_code", directory.RegistrationCode)
	if err := d.Set("self_service_permissions", flattenSelfservicePermissions(directory.SelfservicePermissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting self_service_permissions: %s", err)
	}
	if err := d.Set("saml_properties", flattenSAMLProperties(directory.SamlProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_properties: %s", err)
	}
	d.Set(names.AttrSubnetIDs, directory.SubnetIds)
	if err := d.Set("workspace_access_properties", flattenWorkspaceAccessProperties(directory.WorkspaceAccessProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_access_properties: %s", err)
	}
	if err := d.Set("workspace_creation_properties", flattenDefaultWorkspaceCreationProperties(directory.WorkspaceCreationProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_creation_properties: %s", err)
	}
	d.Set("workspace_security_group_id", directory.WorkspaceSecurityGroupId)

	return diags
}
