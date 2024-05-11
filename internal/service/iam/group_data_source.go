// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_group", name="Group")
func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrUserName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	groupName := d.Get(names.AttrGroupName).(string)

	req := &iam.GetGroupInput{
		GroupName: aws.String(groupName),
	}

	var users []awstypes.User
	var group *awstypes.Group

	pages := iam.NewGetGroupPaginator(conn, req)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting group: %s", err)
		}
		if group == nil {
			group = page.Group
		}

		users = append(users, page.Users...)
	}

	if group == nil {
		return sdkdiag.AppendErrorf(diags, "no IAM group found")
	}

	d.SetId(aws.ToString(group.GroupId))
	d.Set(names.AttrARN, group.Arn)
	d.Set(names.AttrPath, group.Path)
	d.Set("group_id", group.GroupId)
	if err := d.Set("users", dataSourceGroupUsersRead(users)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting users: %s", err)
	}

	return diags
}

func dataSourceGroupUsersRead(iamUsers []awstypes.User) []map[string]interface{} {
	users := make([]map[string]interface{}, 0, len(iamUsers))
	for _, i := range iamUsers {
		u := make(map[string]interface{})
		u[names.AttrARN] = aws.ToString(i.Arn)
		u["user_id"] = aws.ToString(i.UserId)
		u[names.AttrUserName] = aws.ToString(i.UserName)
		u[names.AttrPath] = aws.ToString(i.Path)
		users = append(users, u)
	}
	return users
}
