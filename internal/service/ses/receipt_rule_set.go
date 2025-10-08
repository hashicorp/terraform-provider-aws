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

// @SDKResource("aws_ses_receipt_rule_set", name="Receipt Rule Set")
func resourceReceiptRuleSet() *schema.Resource {
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

func resourceReceiptRuleSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	name := d.Get("rule_set_name").(string)
	input := &ses.CreateReceiptRuleSetInput{
		RuleSetName: aws.String(name),
	}

	_, err := conn.CreateReceiptRuleSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Receipt Rule Set (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceReceiptRuleSetRead(ctx, d, meta)...)
}

func resourceReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	output, err := findReceiptRuleSetByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Receipt Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	name := aws.ToString(output.Metadata.Name)
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

func resourceReceiptRuleSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Receipt Rule Set: %s", d.Id())
	_, err := conn.DeleteReceiptRuleSet(ctx, &ses.DeleteReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findReceiptRuleSetByName(ctx context.Context, conn *ses.Client, name string) (*ses.DescribeReceiptRuleSetOutput, error) {
	input := &ses.DescribeReceiptRuleSetInput{
		RuleSetName: aws.String(name),
	}

	return findReceiptRuleSet(ctx, conn, input)
}

func findReceiptRuleSet(ctx context.Context, conn *ses.Client, input *ses.DescribeReceiptRuleSetInput) (*ses.DescribeReceiptRuleSetOutput, error) {
	output, err := conn.DescribeReceiptRuleSet(ctx, input)

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

	return output, nil
}
