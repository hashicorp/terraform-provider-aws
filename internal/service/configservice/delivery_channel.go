// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_delivery_channel")
func ResourceDeliveryChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeliveryChannelPut,
		ReadWithoutTimeout:   resourceDeliveryChannelRead,
		UpdateWithoutTimeout: resourceDeliveryChannelPut,
		DeleteWithoutTimeout: resourceDeliveryChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "default",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"s3_kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"snapshot_delivery_properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delivery_frequency": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validExecutionFrequency(),
						},
					},
				},
			},
		},
	}
}

func resourceDeliveryChannelPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	name := d.Get("name").(string)
	channel := configservice.DeliveryChannel{
		Name:         aws.String(name),
		S3BucketName: aws.String(d.Get("s3_bucket_name").(string)),
	}

	if v, ok := d.GetOk("s3_key_prefix"); ok {
		channel.S3KeyPrefix = aws.String(v.(string))
	}
	if v, ok := d.GetOk("s3_kms_key_arn"); ok {
		channel.S3KmsKeyArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("sns_topic_arn"); ok {
		channel.SnsTopicARN = aws.String(v.(string))
	}

	if p, ok := d.GetOk("snapshot_delivery_properties"); ok {
		propertiesBlocks := p.([]interface{})
		block := propertiesBlocks[0].(map[string]interface{})

		if v, ok := block["delivery_frequency"]; ok {
			channel.ConfigSnapshotDeliveryProperties = &configservice.ConfigSnapshotDeliveryProperties{
				DeliveryFrequency: aws.String(v.(string)),
			}
		}
	}

	input := configservice.PutDeliveryChannelInput{DeliveryChannel: &channel}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.PutDeliveryChannelWithContext(ctx, &input)
		if err == nil {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, "InsufficientDeliveryPolicyException") {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutDeliveryChannelWithContext(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating Delivery Channel failed: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceDeliveryChannelRead(ctx, d, meta)...)
}

func resourceDeliveryChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	input := configservice.DescribeDeliveryChannelsInput{
		DeliveryChannelNames: []*string{aws.String(d.Id())},
	}

	out, err := conn.DescribeDeliveryChannelsWithContext(ctx, &input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchDeliveryChannelException) {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameDeliveryChannel, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameDeliveryChannel, d.Id(), err)
	}

	if !d.IsNewResource() && len(out.DeliveryChannels) < 1 {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameDeliveryChannel, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && len(out.DeliveryChannels) < 1 {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameDeliveryChannel, d.Id(), errors.New("not found after creation"))
	}

	if len(out.DeliveryChannels) > 1 {
		return sdkdiag.AppendErrorf(diags, "Received %d delivery channels under %s (expected exactly 1): %s",
			len(out.DeliveryChannels), d.Id(), out.DeliveryChannels)
	}

	channel := out.DeliveryChannels[0]

	d.Set("name", channel.Name)
	d.Set("s3_bucket_name", channel.S3BucketName)
	d.Set("s3_key_prefix", channel.S3KeyPrefix)
	d.Set("s3_kms_key_arn", channel.S3KmsKeyArn)
	d.Set("sns_topic_arn", channel.SnsTopicARN)

	if channel.ConfigSnapshotDeliveryProperties != nil {
		d.Set("snapshot_delivery_properties", flattenSnapshotDeliveryProperties(channel.ConfigSnapshotDeliveryProperties))
	}

	return diags
}

func resourceDeliveryChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)
	input := configservice.DeleteDeliveryChannelInput{
		DeliveryChannelName: aws.String(d.Id()),
	}

	err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, err := conn.DeleteDeliveryChannelWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, configservice.ErrCodeLastDeliveryChannelDeleteFailedException, "there is a running configuration recorder") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDeliveryChannelWithContext(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Unable to delete delivery channel: %s", err)
	}

	return diags
}
