// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_account", name="Account")
// @Tags(identifierAttribute="account_id")
// @IdentityAttribute("account_id")
// @Testing(tagsTest=false, identityTest=false)
func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccountRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)

	account, err := findAccountByID(ctx, conn, accountID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s): %s", d.Id(), err)
	}
	d.SetId(aws.ToString(account.Id))

	parentAccountID, err := findParentAccountID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AWS Organizations Account (%s) parent: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, account.Arn)
	d.Set(names.AttrEmail, account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", aws.ToTime(account.JoinedTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, account.Name)
	d.Set("parent_id", parentAccountID)
	d.Set(names.AttrState, account.State)

	return diags
}
