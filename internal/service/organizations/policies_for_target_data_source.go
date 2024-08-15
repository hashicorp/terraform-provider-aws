// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_policies_for_target", name="Policies For Target")
func dataSourcePoliciesForTarget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoliciesForTargetRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIDs: {
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
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	targetID := d.Get("target_id").(string)
	filter := d.Get(names.AttrFilter).(string)
	input := &organizations.ListPoliciesForTargetInput{
		Filter:   awstypes.PolicyType(filter),
		TargetId: aws.String(targetID),
	}
	policies, err := findPoliciesForTarget(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policies (%s) for target (%s): %s", filter, targetID, err)
	}

	d.SetId(targetID)
	d.Set(names.AttrIDs, tfslices.ApplyToAll(policies, func(v awstypes.PolicySummary) string {
		return aws.ToString(v.Id)
	}))

	return diags
}

func findPoliciesForTarget(ctx context.Context, conn *organizations.Client, input *organizations.ListPoliciesForTargetInput) ([]awstypes.PolicySummary, error) {
	var output []awstypes.PolicySummary

	pages := organizations.NewListPoliciesForTargetPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Policies...)
	}

	return output, nil
}
