// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisvideo

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisvideo"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisvideo/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_video_stream", name="Stream")
// @Tags(identifierAttribute="id")
func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kinesisvideo.CreateStreamInput{
		DataRetentionInHours: aws.Int32(int32(d.Get("data_retention_in_hours").(int))),
		StreamName:           aws.String(name),
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

	output, err := conn.CreateStream(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Video Stream (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.StreamARN))

	if _, err := waitStreamCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Video Stream (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoClient(ctx)

	stream, err := findStreamByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Video Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, stream.StreamARN)
	d.Set(names.AttrCreationTime, stream.CreationTime.Format(time.RFC3339))
	d.Set("data_retention_in_hours", stream.DataRetentionInHours)
	d.Set(names.AttrDeviceName, stream.DeviceName)
	d.Set(names.AttrKMSKeyID, stream.KmsKeyId)
	d.Set("media_type", stream.MediaType)
	d.Set(names.AttrName, stream.StreamName)
	d.Set(names.AttrVersion, stream.Version)

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &kinesisvideo.UpdateStreamInput{
			CurrentVersion: aws.String(d.Get(names.AttrVersion).(string)),
			StreamARN:      aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDeviceName); ok {
			input.DeviceName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("media_type"); ok {
			input.MediaType = aws.String(v.(string))
		}

		_, err := conn.UpdateStream(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Video Stream (%s): %s", d.Id(), err)
		}

		if _, err := waitStreamUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Video Stream (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisVideoClient(ctx)

	log.Printf("[DEBUG] Deleting Kinesis Video Stream: %s", d.Id())
	_, err := conn.DeleteStream(ctx, &kinesisvideo.DeleteStreamInput{
		CurrentVersion: aws.String(d.Get(names.AttrVersion).(string)),
		StreamARN:      aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Video Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitStreamDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Video Stream (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStreamByARN(ctx context.Context, conn *kinesisvideo.Client, arn string) (*awstypes.StreamInfo, error) {
	input := &kinesisvideo.DescribeStreamInput{
		StreamARN: aws.String(arn),
	}

	return findStream(ctx, conn, input)
}

func findStream(ctx context.Context, conn *kinesisvideo.Client, input *kinesisvideo.DescribeStreamInput) (*awstypes.StreamInfo, error) {
	output, err := conn.DescribeStream(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StreamInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StreamInfo, nil
}

func statusStream(ctx context.Context, conn *kinesisvideo.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findStreamByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *kinesisvideo.Client, arn string, timeout time.Duration) (*awstypes.StreamInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusCreating),
		Target:     enum.Slice(awstypes.StatusActive),
		Refresh:    statusStream(ctx, conn, arn),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.StreamInfo); ok {
		return output, err
	}

	return nil, err
}

func waitStreamUpdated(ctx context.Context, conn *kinesisvideo.Client, arn string, timeout time.Duration) (*awstypes.StreamInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusUpdating),
		Target:     enum.Slice(awstypes.StatusActive),
		Refresh:    statusStream(ctx, conn, arn),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.StreamInfo); ok {
		return output, err
	}

	return nil, err
}

func waitStreamDeleted(ctx context.Context, conn *kinesisvideo.Client, arn string, timeout time.Duration) (*awstypes.StreamInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusDeleting),
		Target:     []string{},
		Refresh:    statusStream(ctx, conn, arn),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.StreamInfo); ok {
		return output, err
	}

	return nil, err
}
