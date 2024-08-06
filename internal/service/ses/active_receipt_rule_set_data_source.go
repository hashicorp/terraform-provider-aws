// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ses_active_receipt_rule_set")
func DataSourceActiveReceiptRuleSet() *schema.Resource {
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

func dataSourceActiveReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	data, err := conn.DescribeActiveReceiptRuleSetWithContext(ctx, &ses.DescribeActiveReceiptRuleSetInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", err)
	}
	if data == nil || data.Metadata == nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", tfresource.NewEmptyResultError(nil))
	}

	name := aws.StringValue(data.Metadata.Name)
	d.SetId(name)
	d.Set("rule_set_name", name)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-rule-set/%s", name),
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}
