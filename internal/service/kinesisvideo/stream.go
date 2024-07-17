// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisvideo

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_video_stream", name="Stream")
// @Tags(identifierAttribute="id")
func ResourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"data_retention_in_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},

			names.AttrDeviceName: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"media_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn(ctx)

	input := &kinesisvideo.CreateStreamInput{
		StreamName:           aws.String(d.Get(names.AttrName).(string)),
		DataRetentionInHours: aws.Int64(int64(d.Get("data_retention_in_hours").(int))),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDeviceName); ok {
		input.DeviceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("media_type"); ok {
		input.MediaType = aws.String(v.(string))
	}

	resp, err := conn.CreateStreamWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Video Stream: %s", err)
	}

	arn := aws.StringValue(resp.StreamARN)
	d.SetId(arn)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusCreating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(ctx, conn, arn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for creating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn(ctx)

	descOpts := &kinesisvideo.DescribeStreamInput{
		StreamARN: aws.String(d.Id()),
	}

	resp, err := conn.DescribeStreamWithContext(ctx, descOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.StreamInfo.StreamName)
	d.Set("data_retention_in_hours", resp.StreamInfo.DataRetentionInHours)
	d.Set(names.AttrDeviceName, resp.StreamInfo.DeviceName)
	d.Set(names.AttrKMSKeyID, resp.StreamInfo.KmsKeyId)
	d.Set("media_type", resp.StreamInfo.MediaType)
	d.Set(names.AttrARN, resp.StreamInfo.StreamARN)
	if err := d.Set(names.AttrCreationTime, resp.StreamInfo.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}
	d.Set(names.AttrVersion, resp.StreamInfo.Version)

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn(ctx)

	updateOpts := &kinesisvideo.UpdateStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get(names.AttrVersion).(string)),
	}

	if v, ok := d.GetOk(names.AttrDeviceName); ok {
		updateOpts.DeviceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("media_type"); ok {
		updateOpts.MediaType = aws.String(v.(string))
	}

	if _, err := conn.UpdateStreamWithContext(ctx, updateOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusUpdating},
		Target:     []string{kinesisvideo.StatusActive},
		Refresh:    StreamStateRefresh(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for updating Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoConn(ctx)

	if _, err := conn.DeleteStreamWithContext(ctx, &kinesisvideo.DeleteStreamInput{
		StreamARN:      aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get(names.AttrVersion).(string)),
	}); err != nil {
		if tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{kinesisvideo.StatusDeleting},
		Target:     []string{"DELETED"},
		Refresh:    StreamStateRefresh(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	return diags
}

func StreamStateRefresh(ctx context.Context, conn *kinesisvideo.KinesisVideo, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &kinesisvideo.DescribeStreamInput{}

		resp, err := conn.DescribeStreamWithContext(ctx, &kinesisvideo.DescribeStreamInput{
			StreamARN: aws.String(arn),
		})
		if tfawserr.ErrCodeEquals(err, kinesisvideo.ErrCodeResourceNotFoundException) {
			return emptyResp, "DELETED", nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.StreamInfo.Status), nil
	}
}
