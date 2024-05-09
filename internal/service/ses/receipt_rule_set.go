// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_receipt_rule_set")
func ResourceReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReceiptRuleSetCreate,
		ReadWithoutTimeout:   resourceReceiptRuleSetRead,
		DeleteWithoutTimeout: resourceReceiptRuleSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceReceiptRuleSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	ruleSetName := d.Get("rule_set_name").(string)

	createOpts := &ses.CreateReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.CreateReceiptRuleSetWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES rule set: %s", err)
	}

	d.SetId(ruleSetName)

	return append(diags, resourceReceiptRuleSetRead(ctx, d, meta)...)
}

func resourceReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	input := &ses.DescribeReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeReceiptRuleSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ses.ErrCodeRuleSetDoesNotExistException) {
		log.Printf("[WARN] SES Receipt Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	if resp.Metadata == nil {
		log.Print("[WARN] No Receipt Rule Set found")
		d.SetId("")
		return diags
	}

	name := aws.StringValue(resp.Metadata.Name)
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

func resourceReceiptRuleSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	log.Printf("[DEBUG] SES Delete Receipt Rule Set: %s", d.Id())
	input := &ses.DeleteReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	}
	if _, err := conn.DeleteReceiptRuleSetWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	return diags
}
