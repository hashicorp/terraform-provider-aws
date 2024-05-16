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

// @SDKDataSource("aws_organizations_organizational_unit_child_accounts")
func DataSourceOrganizationalUnitChildAccounts() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitChildAccountsRead,

		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceOrganizationalUnitChildAccountsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	parentID := d.Get("parent_id").(string)
	accounts, err := findAccountsForParent(ctx, conn, parentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Accounts for parent (%s): %s", parentID, err)
	}

	d.SetId(parentID)

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}

	return diags
}

func findAccountsForParent(ctx context.Context, conn *organizations.Organizations, id string) ([]*organizations.Account, error) {
	input := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(id),
	}
	var output []*organizations.Account

	err := conn.ListAccountsForParentPagesWithContext(ctx, input, func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
		output = append(output, page.Accounts...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
