// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_data_protection_policy", name="Data Protection Policy")
func resourceDataProtectionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataProtectionPolicyPut,
		ReadWithoutTimeout:   resourceDataProtectionPolicyRead,
		UpdateWithoutTimeout: resourceDataProtectionPolicyPut,
		DeleteWithoutTimeout: resourceDataProtectionPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrLogGroupName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogGroupName,
			},
			"policy_document": sdkv2.JSONDocumentSchemaRequired(),
		},
	}
}

func resourceDataProtectionPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	logGroupName := d.Get(names.AttrLogGroupName).(string)
	input := &cloudwatchlogs.PutDataProtectionPolicyInput{
		LogGroupIdentifier: aws.String(logGroupName),
		PolicyDocument:     aws.String(policy),
	}

	_, err = conn.PutDataProtectionPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Data Protection Policy (%s): %s", logGroupName, err)
	}

	if d.IsNewResource() {
		d.SetId(logGroupName)
	}

	return append(diags, resourceDataProtectionPolicyRead(ctx, d, meta)...)
}

func resourceDataProtectionPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	output, err := findDataProtectionPolicyByLogGroupName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Data Protection Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Data Protection Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.ToString(output.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrLogGroupName, output.LogGroupIdentifier)
	d.Set("policy_document", policyToSet)

	return diags
}

func resourceDataProtectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Logs Data Protection Policy: %s", d.Id())
	_, err := conn.DeleteDataProtectionPolicy(ctx, &cloudwatchlogs.DeleteDataProtectionPolicyInput{
		LogGroupIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Data Protection Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataProtectionPolicyByLogGroupName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*cloudwatchlogs.GetDataProtectionPolicyOutput, error) {
	input := cloudwatchlogs.GetDataProtectionPolicyInput{
		LogGroupIdentifier: aws.String(name),
	}
	output, err := findDataProtectionPolicy(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, err
}

func findDataProtectionPolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetDataProtectionPolicyInput) (*cloudwatchlogs.GetDataProtectionPolicyOutput, error) {
	output, err := conn.GetDataProtectionPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	return output, nil
}
