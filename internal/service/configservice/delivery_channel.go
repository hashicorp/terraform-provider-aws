package configservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceDeliveryChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeliveryChannelPut,
		Read:   resourceDeliveryChannelRead,
		Update: resourceDeliveryChannelPut,
		Delete: resourceDeliveryChannelDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceDeliveryChannelPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

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

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutDeliveryChannel(&input)
		if err == nil {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, "InsufficientDeliveryPolicyException") {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutDeliveryChannel(&input)
	}
	if err != nil {
		return fmt.Errorf("Creating Delivery Channel failed: %s", err)
	}

	d.SetId(name)

	return resourceDeliveryChannelRead(d, meta)
}

func resourceDeliveryChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	input := configservice.DescribeDeliveryChannelsInput{
		DeliveryChannelNames: []*string{aws.String(d.Id())},
	}

	out, err := conn.DescribeDeliveryChannels(&input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchDeliveryChannelException) {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Delivery Channel", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.ConfigService, names.ErrActionReading, "Delivery Channel", d.Id(), err)
	}

	if !d.IsNewResource() && len(out.DeliveryChannels) < 1 {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Delivery Channel", d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && len(out.DeliveryChannels) < 1 {
		return names.Error(names.ConfigService, names.ErrActionReading, "Delivery Channel", d.Id(), errors.New("not found after creation"))
	}

	if len(out.DeliveryChannels) > 1 {
		return fmt.Errorf("Received %d delivery channels under %s (expected exactly 1): %s",
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

	return nil
}

func resourceDeliveryChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	input := configservice.DeleteDeliveryChannelInput{
		DeliveryChannelName: aws.String(d.Id()),
	}

	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		_, err := conn.DeleteDeliveryChannel(&input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, configservice.ErrCodeLastDeliveryChannelDeleteFailedException, "there is a running configuration recorder") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDeliveryChannel(&input)
	}
	if err != nil {
		return fmt.Errorf("Unable to delete delivery channel: %s", err)
	}

	return nil
}
