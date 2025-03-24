// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceActiveReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	data, err := findActiveReceiptRuleSet(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", err)
	}

	name := aws.ToString(data.Name)
	d.SetId(name)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("receipt-rule-set/%s", name),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("rule_set_name", name)

	return diags
}
