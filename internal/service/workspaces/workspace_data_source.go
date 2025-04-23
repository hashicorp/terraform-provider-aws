// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_workspace", name="Workspace")
// @Tags(identifierAttribute="id")
func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"computer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				RequiredWith:  []string{names.AttrUserName},
				ConflictsWith: []string{"workspace_id"},
			},
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrUserName: {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				RequiredWith:  []string{"directory_id"},
				ConflictsWith: []string{"workspace_id"},
			},
			"user_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"volume_encryption_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"directory_id", names.AttrUserName},
			},
			"workspace_properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"root_volume_size_gib": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"running_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"running_mode_auto_stop_timeout_in_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"user_volume_size_gib": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	var workspace *types.Workspace
	var err error

	if v, ok := d.GetOk("workspace_id"); ok {
		workspace, err = findWorkspaceByID(ctx, conn, v.(string))
	}

	if v, ok := d.GetOk("directory_id"); ok {
		input := &workspaces.DescribeWorkspacesInput{
			DirectoryId: aws.String(v.(string)),
			UserName:    aws.String(d.Get(names.AttrUserName).(string)),
		}

		workspace, err = findWorkspace(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("WorkSpaces Workspace", err))
	}

	d.SetId(aws.ToString(workspace.WorkspaceId))
	d.Set("bundle_id", workspace.BundleId)
	d.Set("computer_name", workspace.ComputerName)
	d.Set("directory_id", workspace.DirectoryId)
	d.Set(names.AttrIPAddress, workspace.IpAddress)
	d.Set("root_volume_encryption_enabled", workspace.RootVolumeEncryptionEnabled)
	d.Set(names.AttrState, workspace.State)
	d.Set(names.AttrUserName, workspace.UserName)
	d.Set("user_volume_encryption_enabled", workspace.UserVolumeEncryptionEnabled)
	d.Set("volume_encryption_key", workspace.VolumeEncryptionKey)
	if err := d.Set("workspace_properties", flattenWorkspaceProperties(workspace.WorkspaceProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workspace_properties: %s", err)
	}

	return diags
}
