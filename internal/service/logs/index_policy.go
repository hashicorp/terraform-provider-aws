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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_index_policy")
func resourceIndexPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIndexPolicyPut,
		ReadWithoutTimeout:   resourceIndexPolicyRead,
		UpdateWithoutTimeout: resourceIndexPolicyPut,
		DeleteWithoutTimeout: resourceIndexPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: resourceIndexPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrLogGroupName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogGroupName,
			},
			"fields": {
				Type:     schema.TypeList,
				Optional: false,
				Elem:     schema.TypeString,
			},
		},
	}
}

func resourceIndexPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := d.Get(names.AttrName).(string)
	logGroupName := d.Get(names.AttrLogGroupName).(string)
	input := &cloudwatchlogs.PutIndexPolicyInput{
		LogGroupIdentifier: aws.String(logGroupName),
		PolicyDocument:     aws.String(fmt.Sprintf(`{"fields": %s}`, d.Get("fields").([]interface{}))),
	}

	_, err := conn.PutIndexPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Index Policy (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceIndexPolicyRead(ctx, d, meta)...)
}

func resourceIndexPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	input := cloudwatchlogs.DescribeIndexPoliciesInput{
		LogGroupIdentifiers: []string{d.Get(names.AttrLogGroupName).(string)},
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
	d.Set("indexPolicy", ip.IndexPolicies)

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

func resourceIndexPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logGroupName := d.Get(names.AttrLogGroupName).(string)
	d.Set(names.AttrLogGroupName, logGroupName)
	return []*schema.ResourceData{d}, nil
}
