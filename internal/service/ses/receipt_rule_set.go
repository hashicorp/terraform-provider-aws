// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	ruleSetName := d.Get("rule_set_name").(string)

	createOpts := &ses.CreateReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.CreateReceiptRuleSet(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES rule set: %s", err)
	}

	d.SetId(ruleSetName)

	return append(diags, resourceReceiptRuleSetRead(ctx, d, meta)...)
}

func resourceReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	input := &ses.DescribeReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeReceiptRuleSet(ctx, input)

	if errs.IsA[*awstypes.RuleSetDoesNotExistException](err) {
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

	name := aws.ToString(resp.Metadata.Name)
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
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] SES Delete Receipt Rule Set: %s", d.Id())
	input := &ses.DeleteReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	}
	if _, err := conn.DeleteReceiptRuleSet(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	return diags
}
