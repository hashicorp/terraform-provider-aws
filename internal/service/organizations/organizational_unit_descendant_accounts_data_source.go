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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_unit_descendant_accounts", name="Organizational Unit Descendant Accounts")
func dataSourceOrganizationalUnitDescendantAccounts() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitDescendantAccountsRead,

		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEmail: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
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

func dataSourceOrganizationalUnitDescendantAccountsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	parentID := d.Get("parent_id").(string)
	accounts, err := findAllAccountsForParentAndBelow(ctx, conn, parentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Accounts for parent (%s) and descendants: %s", parentID, err)
	}

	d.SetId(parentID)

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}

	return diags
}

// findAllAccountsForParent recurses down an OU tree, returning all accounts at the specified parent and below.
func findAllAccountsForParentAndBelow(ctx context.Context, conn *organizations.Client, id string) ([]awstypes.Account, error) {
	var output []awstypes.Account

	accounts, err := findAccountsForParentByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	output = append(output, accounts...)

	ous, err := findOrganizationalUnitsForParentByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	for _, ou := range ous {
		accounts, err = findAllAccountsForParentAndBelow(ctx, conn, aws.ToString(ou.Id))

		if err != nil {
			return nil, err
		}

		output = append(output, accounts...)
	}

	return output, nil
}
