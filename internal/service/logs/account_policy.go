// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_account_policy", name="Account Policy")
// @IdentityAttribute("policy_name")
// @IdentityAttribute("policy_type")
// @ImportIDHandler("accountPolicyImportID")
// @Testing(preIdentityVersion="v6.51.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.AccountPolicy")
// @Testing(importStateIdFunc=testAccAccountPolicyImportStateIDFunc)
// @Testing(importStateIdAttribute="policy_name")
func resourceAccountPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountPolicyPut,
		ReadWithoutTimeout:   resourceAccountPolicyRead,
		UpdateWithoutTimeout: resourceAccountPolicyPut,
		DeleteWithoutTimeout: resourceAccountPolicyDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"policy_document": sdkv2.JSONDocumentSchemaRequired(),
				"policy_name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"policy_type": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.PolicyType](),
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          awstypes.ScopeAll,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
				"selection_criteria": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
			}
		},
	}
}

func resourceAccountPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("policy_name").(string)
	input := cloudwatchlogs.PutAccountPolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		PolicyType:     awstypes.PolicyType(d.Get("policy_type").(string)),
		Scope:          awstypes.Scope(d.Get(names.AttrScope).(string)),
	}

	if v, ok := d.GetOk("selection_criteria"); ok {
		input.SelectionCriteria = aws.String(v.(string))
	}

	output, err := conn.PutAccountPolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Account Policy (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.AccountPolicy.PolicyName))
	}

	return append(diags, resourceAccountPolicyRead(ctx, d, meta)...)
}

func resourceAccountPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	output, err := findAccountPolicyByTwoPartKey(ctx, conn, d.Id(), awstypes.PolicyType(d.Get("policy_type").(string)))

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourceAccountPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Logs Account Policy: %s", d.Id())
	input := cloudwatchlogs.DeleteAccountPolicyInput{
		PolicyName: aws.String(d.Id()),
		PolicyType: awstypes.PolicyType(d.Get("policy_type").(string)),
	}
	_, err := conn.DeleteAccountPolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Account Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findAccountPolicyByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, policyName string, policyType awstypes.PolicyType) (*awstypes.AccountPolicy, error) {
	input := cloudwatchlogs.DescribeAccountPoliciesInput{
		PolicyName: aws.String(policyName),
		PolicyType: policyType,
	}
	output, err := findAccountPolicy(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, err
}

func findAccountPolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) (*awstypes.AccountPolicy, error) {
	output, err := findAccountPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccountPolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) ([]awstypes.AccountPolicy, error) {
	var output []awstypes.AccountPolicy

	err := describeAccountPoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeAccountPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.AccountPolicies...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

const accountPolicyImportIDSeparator = ":"

func accountPolicyParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, accountPolicyImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <policy-name>%[2]s<policy-type>", id, accountPolicyImportIDSeparator)
}

var (
	_ inttypes.SDKv2ImportID = accountPolicyImportID{}
)

type accountPolicyImportID struct{}

func (accountPolicyImportID) Parse(id string) (string, map[string]any, error) {
	policyName, policyType, err := accountPolicyParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"policy_name": policyName,
		"policy_type": policyType,
	}

	return policyName, result, nil
}

func (accountPolicyImportID) Create(d *schema.ResourceData) string {
	return d.Get("policy_name").(string)
}
