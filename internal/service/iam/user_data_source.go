// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_user", name="User")
// @Tags
func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	req := &iam.GetUserInput{
		UserName: aws.String(userName),
	}

	log.Printf("[DEBUG] Reading IAM User: %v", req)
	resp, err := conn.GetUser(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting user: %s", err)
	}

	user := resp.User
	d.SetId(aws.ToString(user.UserId))
	d.Set(names.AttrARN, user.Arn)
	d.Set(names.AttrPath, user.Path)
	d.Set("permissions_boundary", "")
	if user.PermissionsBoundary != nil {
		d.Set("permissions_boundary", user.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("user_id", user.UserId)

	setTagsOut(ctx, user.Tags)

	return diags
}
