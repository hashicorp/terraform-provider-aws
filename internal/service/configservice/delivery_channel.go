// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_delivery_channel", name="Delivery Channel")
func resourceDeliveryChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeliveryChannelPut,
		ReadWithoutTimeout:   resourceDeliveryChannelRead,
		UpdateWithoutTimeout: resourceDeliveryChannelPut,
		DeleteWithoutTimeout: resourceDeliveryChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultDeliveryChannelName,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrS3BucketName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrS3KeyPrefix: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"s3_kms_key_arn": {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.MaximumExecutionFrequency](),
						},
					},
				},
			},
			names.AttrSNSTopicARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceDeliveryChannelPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &configservice.PutDeliveryChannelInput{
		DeliveryChannel: &types.DeliveryChannel{
			Name:         aws.String(name),
			S3BucketName: aws.String(d.Get(names.AttrS3BucketName).(string)),
		},
	}

	if v, ok := d.GetOk(names.AttrS3KeyPrefix); ok {
		input.DeliveryChannel.S3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_kms_key_arn"); ok {
		input.DeliveryChannel.S3KmsKeyArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_delivery_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		if v, ok := tfMap["delivery_frequency"]; ok {
			input.DeliveryChannel.ConfigSnapshotDeliveryProperties = &types.ConfigSnapshotDeliveryProperties{
				DeliveryFrequency: types.MaximumExecutionFrequency(v.(string)),
			}
		}
	}

	if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
		input.DeliveryChannel.SnsTopicARN = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*types.InsufficientDeliveryPolicyException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutDeliveryChannel(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Delivery Channel (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceDeliveryChannelRead(ctx, d, meta)...)
}

func resourceDeliveryChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	channel, err := findDeliveryChannelByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Delivery Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Delivery Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, channel.Name)
	d.Set(names.AttrS3BucketName, channel.S3BucketName)
	d.Set(names.AttrS3KeyPrefix, channel.S3KeyPrefix)
	d.Set("s3_kms_key_arn", channel.S3KmsKeyArn)
	if channel.ConfigSnapshotDeliveryProperties != nil {
		d.Set("snapshot_delivery_properties", flattenSnapshotDeliveryProperties(channel.ConfigSnapshotDeliveryProperties))
	}
	d.Set(names.AttrSNSTopicARN, channel.SnsTopicARN)

	return diags
}

func resourceDeliveryChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 30 * time.Second
	)
	log.Printf("[DEBUG] Deleting ConfigService Delivery Channel: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.LastDeliveryChannelDeleteFailedException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteDeliveryChannel(ctx, &configservice.DeleteDeliveryChannelInput{
			DeliveryChannelName: aws.String(d.Id()),
		})
	}, "there is a running configuration recorder")

	if errs.IsA[*types.NoSuchDeliveryChannelException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Delivery Channel (%s): %s", d.Id(), err)
	}

	return diags
}

func findDeliveryChannelByName(ctx context.Context, conn *configservice.Client, name string) (*types.DeliveryChannel, error) {
	input := &configservice.DescribeDeliveryChannelsInput{
		DeliveryChannelNames: []string{name},
	}

	return findDeliveryChannel(ctx, conn, input)
}

func findDeliveryChannel(ctx context.Context, conn *configservice.Client, input *configservice.DescribeDeliveryChannelsInput) (*types.DeliveryChannel, error) {
	output, err := findDeliveryChannels(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDeliveryChannels(ctx context.Context, conn *configservice.Client, input *configservice.DescribeDeliveryChannelsInput) ([]types.DeliveryChannel, error) {
	output, err := conn.DescribeDeliveryChannels(ctx, input)

	if errs.IsA[*types.NoSuchDeliveryChannelException](err) {
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

	return output.DeliveryChannels, nil
}

func flattenSnapshotDeliveryProperties(apiObject *types.ConfigSnapshotDeliveryProperties) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"delivery_frequency": apiObject.DeliveryFrequency,
	}

	return []interface{}{tfMap}
}
