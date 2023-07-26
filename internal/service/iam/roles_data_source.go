// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iam_roles")
func DataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRolesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	input := &iam.ListRolesInput{}

	if v, ok := d.GetOk("path_prefix"); ok {
		input.PathPrefix = aws.String(v.(string))
	}

	var results []*iam.Role

	err := conn.ListRolesPagesWithContext(ctx, input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, role := range page.Roles {
			if role == nil {
				continue
			}

			if v, ok := d.GetOk("name_regex"); ok && !regexp.MustCompile(v.(string)).MatchString(aws.StringValue(role.RoleName)) {
				continue
			}

			results = append(results, role)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM roles: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, names []string

	for _, r := range results {
		arns = append(arns, aws.StringValue(r.Arn))
		names = append(names, aws.StringValue(r.RoleName))
	}

	if err := d.Set("arns", arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	if err := d.Set("names", names); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting names: %s", err)
	}

	return diags
}
