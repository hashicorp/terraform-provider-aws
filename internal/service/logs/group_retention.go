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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudwatch_log_group_retention", name="Log Group Retention")
func resourceGroupRetention() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupRetentionPut,
		ReadWithoutTimeout:   resourceGroupRetentionRead,
		UpdateWithoutTimeout: resourceGroupRetentionPut,
		DeleteWithoutTimeout: resourceGroupRetentionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"log_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogGroupName,
			},
			"retention_in_days": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653}),
			},
		},
	}
}

func resourceGroupRetentionPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	logGroupName := d.Get("log_group_name").(string)
	retentionInDays := d.Get("retention_in_days").(int)

	// Check if the log group exists
	_, err := findLogGroupByName(ctx, conn, logGroupName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", logGroupName, err)
	}

	input := cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    aws.String(logGroupName),
		RetentionInDays: aws.Int32(int32(retentionInDays)),
	}

	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutRetentionPolicy(ctx, &input)
	}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", logGroupName, err)
	}

	d.SetId(logGroupName)

	return append(diags, resourceGroupRetentionRead(ctx, d, meta)...)
}

func resourceGroupRetentionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	lg, err := findLogGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	d.Set("log_group_name", lg.LogGroupName)

	// RetentionInDays can be nil if no retention policy is set
	if lg.RetentionInDays != nil {
		d.Set("retention_in_days", aws.ToInt32(lg.RetentionInDays))
	} else {
		// If no retention policy is set, the resource should be removed from state
		// since this resource specifically manages retention policy
		if !d.IsNewResource() {
			log.Printf("[WARN] CloudWatch Logs Log Group (%s) has no retention policy, removing from state", d.Id())
			d.SetId("")
			return diags
		}
	}

	return diags
}

func resourceGroupRetentionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Log Group retention policy: %s", d.Id())
	input := cloudwatchlogs.DeleteRetentionPolicyInput{
		LogGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteRetentionPolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
	}

	return diags
}
