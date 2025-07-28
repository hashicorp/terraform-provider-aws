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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_active_receipt_rule_set", name="Active Receipt Rule Set")
func resourceActiveReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActiveReceiptRuleSetUpdate,
		UpdateWithoutTimeout: resourceActiveReceiptRuleSetUpdate,
		ReadWithoutTimeout:   resourceActiveReceiptRuleSetRead,
		DeleteWithoutTimeout: resourceActiveReceiptRuleSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceActiveReceiptRuleSetImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceActiveReceiptRuleSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	ruleSetName := d.Get("rule_set_name").(string)
	input := &ses.SetActiveReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.SetActiveReceiptRuleSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Active Receipt Rule Set (%s): %s", ruleSetName, err)
	}

	d.SetId(ruleSetName)

	return append(diags, resourceActiveReceiptRuleSetRead(ctx, d, meta)...)
}

func resourceActiveReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	output, err := findActiveReceiptRuleSet(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Active Receipt Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("receipt-rule-set/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("rule_set_name", output.Name)

	return diags
}

func resourceActiveReceiptRuleSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	_, err := conn.SetActiveReceiptRuleSet(ctx, &ses.SetActiveReceiptRuleSetInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Active Receipt Rule Set: %s", err)
	}

	return diags
}

func resourceActiveReceiptRuleSetImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	describeOpts := &ses.DescribeActiveReceiptRuleSetInput{}

	response, err := conn.DescribeActiveReceiptRuleSet(ctx, describeOpts)
	if err != nil {
		return nil, err
	}

	if response.Metadata == nil {
		return nil, fmt.Errorf("no active Receipt Rule Set found")
	}

	if aws.ToString(response.Metadata.Name) != d.Id() {
		return nil, fmt.Errorf("SES Receipt Rule Set (%s) belonging to SES Active Receipt Rule Set not found", d.Id())
	}

	d.Set("rule_set_name", response.Metadata.Name)

	arnValue := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("receipt-rule-set/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arnValue)

	return []*schema.ResourceData{d}, nil
}

func findActiveReceiptRuleSet(ctx context.Context, conn *ses.Client) (*awstypes.ReceiptRuleSetMetadata, error) {
	input := &ses.DescribeActiveReceiptRuleSetInput{}
	output, err := conn.DescribeActiveReceiptRuleSet(ctx, input)

	if errs.IsA[*awstypes.RuleSetDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Metadata == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Metadata, nil
}
