// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_policy")
func DataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policyID := d.Get("policy_id").(string)
	policy, err := findPolicyByID(ctx, conn, policyID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy (%s): %s", policyID, err)
	}

	policySummary := policy.PolicySummary
	d.SetId(aws.StringValue(policySummary.Id))
	d.Set(names.AttrARN, policySummary.Arn)
	d.Set("aws_managed", policySummary.AwsManaged)
	d.Set("content", policy.Content)
	d.Set(names.AttrDescription, policySummary.Description)
	d.Set(names.AttrName, policySummary.Name)
	d.Set(names.AttrType, policySummary.Type)

	return diags
}
