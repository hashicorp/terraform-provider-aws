// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_event_stream")
func ResourceEventStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventStreamUpsert,
		ReadWithoutTimeout:   resourceEventStreamRead,
		UpdateWithoutTimeout: resourceEventStreamUpsert,
		DeleteWithoutTimeout: resourceEventStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceEventStreamUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &pinpoint.WriteEventStream{
		DestinationStreamArn: aws.String(d.Get("destination_stream_arn").(string)),
		RoleArn:              aws.String(d.Get(names.AttrRoleARN).(string)),
	}

	req := pinpoint.PutEventStreamInput{
		ApplicationId:    aws.String(applicationId),
		WriteEventStream: params,
	}

	// Retry for IAM eventual consistency
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.PutEventStreamWithContext(ctx, &req)

		if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeBadRequestException, "make sure the IAM Role is configured correctly") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutEventStreamWithContext(ctx, &req)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Pinpoint Event Stream for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceEventStreamRead(ctx, d, meta)...)
}

func resourceEventStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[INFO] Reading Pinpoint Event Stream for application %s", d.Id())

	output, err := conn.GetEventStreamWithContext(ctx, &pinpoint.GetEventStreamInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint Event Stream for application %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "getting Pinpoint Event Stream for application %s: %s", d.Id(), err)
	}

	res := output.EventStream
	d.Set(names.AttrApplicationID, res.ApplicationId)
	d.Set("destination_stream_arn", res.DestinationStreamArn)
	d.Set(names.AttrRoleARN, res.RoleArn)

	return diags
}

func resourceEventStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Pinpoint Delete Event Stream: %s", d.Id())
	_, err := conn.DeleteEventStreamWithContext(ctx, &pinpoint.DeleteEventStreamInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Event Stream for application %s: %s", d.Id(), err)
	}
	return diags
}
