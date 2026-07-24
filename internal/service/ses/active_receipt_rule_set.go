// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"rule_set_name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 64),
				},
			}
		},
	}
}

func resourceActiveReceiptRuleSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	ruleSetName := d.Get("rule_set_name").(string)
	input := ses.SetActiveReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.SetActiveReceiptRuleSet(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES Active Receipt Rule Set (%s): %s", ruleSetName, err)
	}

	d.SetId(ruleSetName)

	return append(diags, resourceActiveReceiptRuleSetRead(ctx, d, meta)...)
}

func resourceActiveReceiptRuleSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.SESClient(ctx)

	output, err := findActiveReceiptRuleSet(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SES Active Receipt Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Active Receipt Rule Set: %s", err)
	}

	d.Set(names.AttrARN, receiptRuleSetARN(ctx, c, d.Id()))
	d.Set("rule_set_name", output.Name)

	return diags
}

func resourceActiveReceiptRuleSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Active Receipt Rule Set: %s", d.Id())
	var input ses.SetActiveReceiptRuleSetInput
	_, err := conn.SetActiveReceiptRuleSet(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Active Receipt Rule Set: %s", err)
	}

	return diags
}

func resourceActiveReceiptRuleSetImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	// Region may be overridden in the import block.
	var optFns []func(*ses.Options)
	if v, ok := d.GetOk(names.AttrRegion); ok {
		optFns = append(optFns, func(o *ses.Options) {
			o.Region = v.(string)
		})
	}

	output, err := findActiveReceiptRuleSet(ctx, conn, optFns...)

	if err != nil {
		return nil, err
	}

	if got, want := aws.ToString(output.Name), d.Id(); got != want {
		return nil, fmt.Errorf("SES Receipt Rule Set (%s) is not the SES Active Receipt Rule Set (%s)", want, got)
	}

	return []*schema.ResourceData{d}, nil
}

func findActiveReceiptRuleSet(ctx context.Context, conn *ses.Client, optFns ...func(*ses.Options)) (*awstypes.ReceiptRuleSetMetadata, error) {
	var input ses.DescribeActiveReceiptRuleSetInput
	output, err := conn.DescribeActiveReceiptRuleSet(ctx, &input, optFns...)

	if errs.IsA[*awstypes.RuleSetDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Metadata == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Metadata, nil
}
