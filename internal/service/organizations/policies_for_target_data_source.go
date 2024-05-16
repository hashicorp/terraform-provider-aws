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

// @SDKDataSource("aws_organizations_policies_for_target")
func DataSourcePoliciesForTarget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoliciesForTargetRead,

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
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcePoliciesForTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	targetID := d.Get("target_id").(string)
	filter := d.Get("filter").(string)
	policies, err := findPoliciesForTarget(ctx, conn, targetID, filter)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Policies (%s) for target (%s): %s", filter, targetID, err)
	}

	var policyIDs []string

	for _, v := range policies {
		policyIDs = append(policyIDs, aws.StringValue(v.Id))
	}

	d.SetId(targetID)

	d.Set("ids", policyIDs)

	return diags
}

func findPoliciesForTarget(ctx context.Context, conn *organizations.Organizations, targetID string, filter string) ([]*organizations.PolicySummary, error) {
	input := &organizations.ListPoliciesForTargetInput{
		Filter:   aws.String(filter),
		TargetId: aws.String(targetID),
	}
	var output []*organizations.PolicySummary

	err := conn.ListPoliciesForTargetPagesWithContext(ctx, input, func(page *organizations.ListPoliciesForTargetOutput, lastPage bool) bool {
		output = append(output, page.Policies...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
