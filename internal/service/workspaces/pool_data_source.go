// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_pool", name="Pool")
// @Tags(identifierAttribute="id")
func dataSourcePool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoolRead,
		Schema: map[string]*schema.Schema{
			"application_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrS3BucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"settings_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_user_sessions": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					names.AttrID,
					names.AttrName,
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ExactlyOneOf: []string{
					names.AttrID,
					names.AttrName,
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"timeout_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disconnect_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"idle_disconnect_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_user_duration_in_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const (
	DSNamePool = "Pool Data Source"
)

func dataSourcePoolRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	var out *types.WorkspacesPool
	var err error

	if v, ok := d.GetOk(names.AttrID); ok {
		poolID := v.(string)
		out, err = findPoolByID(ctx, conn, poolID)
		d.SetId(poolID)
	} else if v, ok := d.GetOk(names.AttrName); ok {
		poolName := v.(string)
		out, err = findPoolByName(ctx, conn, poolName)

		if out != nil {
			d.SetId(aws.ToString(out.PoolId))
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionReading, ResNamePool, d.Id(), err)
	}

	if err := d.Set("application_settings", flattenApplicationSettings(out.ApplicationSettings)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}
	d.Set(names.AttrARN, out.PoolArn)
	d.Set("bundle_id", out.BundleId)
	if err := d.Set("capacity", flattenCapacity(out.CapacityStatus)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}
	d.Set(names.AttrDescription, out.Description)
	d.Set("directory_id", out.DirectoryId)
	d.Set(names.AttrID, out.PoolId)
	d.Set(names.AttrName, out.PoolName)
	d.Set(names.AttrState, out.State)
	if err := d.Set("timeout_settings", flattenTimeoutSettings(out.TimeoutSettings)); err != nil {
		return create.AppendDiagError(diags, names.WorkSpaces, create.ErrActionSetting, ResNamePool, d.Id(), err)
	}

	return diags
}
