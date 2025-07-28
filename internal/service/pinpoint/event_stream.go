// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_event_stream", name="Event Stream")
func resourceEventStream() *schema.Resource {
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

func resourceEventStreamUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.WriteEventStream{
		DestinationStreamArn: aws.String(d.Get("destination_stream_arn").(string)),
		RoleArn:              aws.String(d.Get(names.AttrRoleARN).(string)),
	}

	req := pinpoint.PutEventStreamInput{
		ApplicationId:    aws.String(applicationId),
		WriteEventStream: params,
	}

	// Retry for IAM eventual consistency
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.BadRequestException](ctx, propagationTimeout, func() (any, error) {
		return conn.PutEventStream(ctx, &req)
	}, "make sure the IAM Role is configured correctly")

	if tfresource.TimedOut(err) { // nosemgrep:ci.helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutEventStream(ctx, &req)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Pinpoint Event Stream for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceEventStreamRead(ctx, d, meta)...)
}

func resourceEventStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint Event Stream for application %s", d.Id())

	output, err := findEventStreamByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint Event Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint Event Stream (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set("destination_stream_arn", output.DestinationStreamArn)
	d.Set(names.AttrRoleARN, output.RoleArn)

	return diags
}

func resourceEventStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Pinpoint Delete Event Stream: %s", d.Id())
	_, err := conn.DeleteEventStream(ctx, &pinpoint.DeleteEventStreamInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Event Stream for application %s: %s", d.Id(), err)
	}
	return diags
}

func findEventStreamByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.EventStream, error) {
	input := &pinpoint.GetEventStreamInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetEventStream(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.EventStream == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EventStream, nil
}
