// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_account_policy", name="Account Policy")
func resourceAccountPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountPolicyPut,
		ReadWithoutTimeout:   resourceAccountPolicyRead,
		UpdateWithoutTimeout: resourceAccountPolicyPut,
		DeleteWithoutTimeout: resourceAccountPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAccountPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"policy_document": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validAccountPolicyDocument,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"policy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.PolicyType](),
			},
			names.AttrScope: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ScopeAll,
				ValidateDiagFunc: enum.Validate[types.Scope](),
			},
			"selection_criteria": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAccountPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("policy_name").(string)
	input := &cloudwatchlogs.PutAccountPolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		PolicyType:     types.PolicyType(d.Get("policy_type").(string)),
		Scope:          types.Scope(d.Get(names.AttrScope).(string)),
	}

	if v, ok := d.GetOk("selection_criteria"); ok {
		input.SelectionCriteria = aws.String(v.(string))
	}

	output, err := conn.PutAccountPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Account Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AccountPolicy.PolicyName))

	return append(diags, resourceAccountPolicyRead(ctx, d, meta)...)
}

func resourceAccountPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	output, err := findAccountPolicyByTwoPartKey(ctx, conn, types.PolicyType(d.Get("policy_type").(string)), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Account Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Account Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.ToString(output.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("policy_document", policyToSet)
	d.Set("policy_name", output.PolicyName)
	d.Set("policy_type", output.PolicyType)
	d.Set(names.AttrScope, output.Scope)
	d.Set("selection_criteria", output.SelectionCriteria)

	return diags
}

func resourceAccountPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Logs Account Policy: %s", d.Id())
	_, err := conn.DeleteAccountPolicy(ctx, &cloudwatchlogs.DeleteAccountPolicyInput{
		PolicyName: aws.String(d.Id()),
		PolicyType: types.PolicyType(d.Get("policy_type").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Account Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAccountPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: '<policy-name>:<policy-type>'", d.Id())
	}

	policyName := parts[0]
	policyType := parts[1]

	d.SetId(policyName)
	d.Set("policy_type", policyType)

	return []*schema.ResourceData{d}, nil
}

func findAccountPolicyByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, policyType types.PolicyType, policyName string) (*types.AccountPolicy, error) {
	input := &cloudwatchlogs.DescribeAccountPoliciesInput{
		PolicyName: aws.String(policyName),
		PolicyType: policyType,
	}

	output, err := conn.DescribeAccountPolicies(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.AccountPolicies)
}
