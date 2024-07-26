// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_users", name="Users")
func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			names.AttrNames: {
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

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	nameRegex := d.Get("name_regex").(string)
	pathPrefix := d.Get("path_prefix").(string)

	results, err := FindUsers(ctx, conn, nameRegex, pathPrefix)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM users: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, nms []string

	for _, r := range results {
		nms = append(nms, aws.ToString(r.UserName))
		arns = append(arns, aws.ToString(r.Arn))
	}

	if err := d.Set(names.AttrNames, nms); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting names: %s", err)
	}

	if err := d.Set(names.AttrARNs, arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	return diags
}
