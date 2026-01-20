// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	nameRegex := d.Get("name_regex").(string)
	var input iam.ListUsersInput
	if v, ok := d.GetOk("path_prefix"); ok {
		input.PathPrefix = aws.String(v.(string))
	}

	results, err := findUsers(ctx, conn, &input, func(v *awstypes.User) bool {
		if nameRegex != "" {
			return regexache.MustCompile(nameRegex).MatchString(aws.ToString(v.UserName))
		}

		return true
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Users: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	var arns, nms []string

	for _, r := range results {
		nms = append(nms, aws.ToString(r.UserName))
		arns = append(arns, aws.ToString(r.Arn))
	}

	d.Set(names.AttrARNs, arns)
	d.Set(names.AttrNames, nms)

	return diags
}

func findUsers(ctx context.Context, conn *iam.Client, input *iam.ListUsersInput, filter tfslices.Predicate[*awstypes.User]) ([]awstypes.User, error) {
	var output []awstypes.User

	pages := iam.NewListUsersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Users {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
