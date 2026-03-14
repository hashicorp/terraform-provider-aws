// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iam_role_policies", name="Role Policies")
func dataSourceRolePolicies() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRolePoliciesRead,

		Schema: map[string]*schema.Schema{
			"policy_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceRolePoliciesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName := d.Get("role_name").(string)

	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	var policyNames []string

	pages := iam.NewListRolePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s) Policies: %s", roleName, err)
		}

		for _, v := range page.PolicyNames {
			if v != "" {
				policyNames = append(policyNames, v)
			}
		}
	}

	d.SetId(roleName)
	d.Set("policy_names", policyNames)

	return diags
}
