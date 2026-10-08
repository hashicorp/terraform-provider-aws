// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ses_active_receipt_rule_set", name="Active Receipt Rule Set")
func dataSourceActiveReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceActiveReceiptRuleSetRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"rule_set_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceActiveReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SESClient(ctx)

	data, err := findActiveReceiptRuleSet(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", err)
	}

	name := aws.ToString(data.Name)
	d.SetId(name)
	d.Set(names.AttrARN, receiptRuleSetARN(ctx, c, name))
	d.Set("rule_set_name", name)

	return diags
}
