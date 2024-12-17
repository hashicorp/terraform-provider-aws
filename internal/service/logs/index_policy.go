// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameIndexPolicy = "Index Policy"
)

// @SDKResource("aws_cloudwatch_log_index_policy")
func resourceIndexPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIndexPolicyPut,
		ReadWithoutTimeout:   resourceIndexPolicyRead,
		UpdateWithoutTimeout: resourceIndexPolicyPut,
		DeleteWithoutTimeout: resourceIndexPolicyDelete,

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
			"policy_document": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceIndexPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	logGroupName := d.Get(names.AttrLogGroupName).(string)

	policyDocument, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyDocument, err)
	}

	input := &cloudwatchlogs.PutIndexPolicyInput{
		LogGroupIdentifier: aws.String(logGroupName),
		PolicyDocument:     aws.String(policyDocument),
	}

	output, err := conn.PutIndexPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Index Policy (%s): %s", d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%s:index-policy", *output.IndexPolicy.LogGroupIdentifier))

	return append(diags, resourceIndexPolicyRead(ctx, d, meta)...)
}

func resourceIndexPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	logGroupName := d.Id()
	input := cloudwatchlogs.DescribeIndexPoliciesInput{
		LogGroupIdentifiers: []string{logGroupName},
	}

	ip, err := conn.DescribeIndexPolicies(ctx, &input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Index Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Index Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrLogGroupName, ip.IndexPolicies[0].LogGroupIdentifier)
	d.Set("policy_document", ip.IndexPolicies[0].PolicyDocument)

	return diags
}

func resourceIndexPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Index Policy: %s", d.Id())
	_, err := conn.DeleteIndexPolicy(ctx, &cloudwatchlogs.DeleteIndexPolicyInput{
		LogGroupIdentifier: aws.String(d.Get(names.AttrLogGroupName).(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Index Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findIndexPolicyByLogGroupName(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName string) ([]types.IndexPolicy, error) {
	input := cloudwatchlogs.DescribeIndexPoliciesInput{
		LogGroupIdentifiers: []string{logGroupName},
	}

	ip, err := conn.DescribeIndexPolicies(ctx, &input)
	if err != nil {
		return nil, err
	}

	return ip.IndexPolicies, nil
}
