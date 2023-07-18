// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_iam_role")
func DataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRoleRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_session_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
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
						"region": {
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
			"tags": tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	output, err := conn.GetRoleWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", name, err)
	}

	d.Set("arn", output.Role.Arn)
	if err := d.Set("create_date", output.Role.CreateDate.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting create_date: %s", err)
	}

	if err := d.Set("role_last_used", flattenRoleLastUsed(output.Role.RoleLastUsed)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting role_last_used: %s", err)
	}

	d.Set("description", output.Role.Description)
	d.Set("max_session_duration", output.Role.MaxSessionDuration)
	d.Set("name", output.Role.RoleName)
	d.Set("path", output.Role.Path)
	d.Set("permissions_boundary", "")
	if output.Role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", output.Role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", output.Role.RoleId)

	assumRolePolicy, err := url.QueryUnescape(aws.StringValue(output.Role.AssumeRolePolicyDocument))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing assume role policy document: %s", err)
	}
	if err := d.Set("assume_role_policy", assumRolePolicy); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting assume_role_policy: %s", err)
	}

	tags := KeyValueTags(ctx, output.Role.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(name)

	return diags
}

func flattenRoleLastUsed(apiObject *iam.RoleLastUsed) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"region": aws.StringValue(apiObject.Region),
	}

	if apiObject.LastUsedDate != nil {
		tfMap["last_used_date"] = apiObject.LastUsedDate.Format(time.RFC3339)
	}
	return []interface{}{tfMap}
}
