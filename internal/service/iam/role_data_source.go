// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_role", name="Role")
// @Tags
func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRoleRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_session_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_last_used": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRegion: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_used_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	role, err := findRoleByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, role.Arn)
	d.Set("create_date", role.CreateDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set(names.AttrName, role.RoleName)
	d.Set(names.AttrPath, role.Path)
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	} else {
		d.Set("permissions_boundary", nil)
	}
	if err := d.Set("role_last_used", flattenRoleLastUsed(role.RoleLastUsed)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting role_last_used: %s", err)
	}
	d.Set("unique_id", role.RoleId)

	assumeRolePolicy, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("assume_role_policy", assumeRolePolicy)

	setTagsOut(ctx, role.Tags)

	return diags
}

func flattenRoleLastUsed(apiObject *awstypes.RoleLastUsed) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrRegion: aws.ToString(apiObject.Region),
	}

	if apiObject.LastUsedDate != nil {
		tfMap["last_used_date"] = apiObject.LastUsedDate.Format(time.RFC3339)
	}
	return []interface{}{tfMap}
}
