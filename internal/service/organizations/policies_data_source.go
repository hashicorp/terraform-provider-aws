// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_organizations_policies")
func DataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoliciesRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	filter := d.Get("filter").(string)
	policies, err := findPolicies(ctx, conn, filter)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Policies (%s): %s", filter, err)
	}

	var policyIDs []string

	for _, v := range policies {
		policyIDs = append(policyIDs, aws.StringValue(v.Id))
	}

	d.SetId(filter)
	d.Set("ids", policyIDs)

	return diags
}

func findPolicies(ctx context.Context, conn *organizations.Organizations, filter string) ([]*organizations.PolicySummary, error) {
	input := &organizations.ListPoliciesInput{
		Filter: aws.String(filter),
	}
	var output []*organizations.PolicySummary

	err := conn.ListPoliciesPagesWithContext(ctx, input, func(page *organizations.ListPoliciesOutput, lastPage bool) bool {
		output = append(output, page.Policies...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
