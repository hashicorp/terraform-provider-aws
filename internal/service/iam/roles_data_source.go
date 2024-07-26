// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"reflect"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_roles", name="Roles")
func dataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRolesRead,

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

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	input := &iam.ListRolesInput{}

	if v, ok := d.GetOk("path_prefix"); ok {
		input.PathPrefix = aws.String(v.(string))
	}

	var results []awstypes.Role

	pages := iam.NewListRolesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IAM roles: %s", err)
		}

		for _, role := range page.Roles {
			if reflect.ValueOf(role).IsZero() {
				continue
			}

			if v, ok := d.GetOk("name_regex"); ok && !regexache.MustCompile(v.(string)).MatchString(aws.ToString(role.RoleName)) {
				continue
			}

			results = append(results, role)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, nms []string

	for _, r := range results {
		arns = append(arns, aws.ToString(r.Arn))
		nms = append(nms, aws.ToString(r.RoleName))
	}

	if err := d.Set(names.AttrARNs, arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	if err := d.Set(names.AttrNames, nms); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting names: %s", err)
	}

	return diags
}
