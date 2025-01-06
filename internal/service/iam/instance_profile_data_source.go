// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_instance_profile", name="Instance Profile")
func dataSourceInstanceProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceProfileRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceInstanceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	instanceProfile, err := findInstanceProfileByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(instanceProfile.InstanceProfileId))
	d.Set(names.AttrARN, instanceProfile.Arn)
	d.Set("create_date", fmt.Sprintf("%v", instanceProfile.CreateDate))
	d.Set(names.AttrPath, instanceProfile.Path)
	if len(instanceProfile.Roles) > 0 {
		role := instanceProfile.Roles[0]
		d.Set(names.AttrRoleARN, role.Arn)
		d.Set("role_id", role.RoleId)
		d.Set("role_name", role.RoleName)
	}

	return diags
}
