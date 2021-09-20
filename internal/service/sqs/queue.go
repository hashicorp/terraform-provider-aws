package sqs

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	sqsQueueSchema = map[string]*schema.Schema{
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"content_based_deduplication": {
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},

		"deduplication_scope": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(DeduplicationScope_Values(), false),
		},

		"delay_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      DefaultQueueDelaySeconds,
			ValidateFunc: validation.IntBetween(0, 900),
		},

		"fifo_queue": {
			Type:     schema.TypeBool,
			Default:  false,
			ForceNew: true,
			Optional: true,
		},

		"fifo_throughput_limit": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(FIFOThroughputLimit_Values(), false),
		},

		"kms_data_key_reuse_period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(60, 86_400),
		},

		"kms_master_key_id": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"max_message_size": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      DefaultQueueMaximumMessageSize,
			ValidateFunc: validation.IntBetween(1024, 262_144),
		},

		"message_retention_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      DefaultQueueMessageRetentionPeriod,
			ValidateFunc: validation.IntBetween(60, 1_209_600),
		},

		"name": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"name_prefix"},
		},

		"name_prefix": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"name"},
		},

		"policy": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
		},

		"receive_wait_time_seconds": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  DefaultQueueReceiveMessageWaitTimeSeconds,
		},

		"redrive_policy": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringIsJSON,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},

		"url": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"visibility_timeout_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      DefaultQueueVisibilityTimeout,
			ValidateFunc: validation.IntBetween(0, 43_200),
		},

		"tags":     tftags.TagsSchema(),
		"tags_all": tftags.TagsSchemaComputed(),
	}

	sqsQueueAttributeMap = NewAttrMap(map[string]string{
		"delay_seconds":                     sqs.QueueAttributeNameDelaySeconds,
		"max_message_size":                  sqs.QueueAttributeNameMaximumMessageSize,
		"message_retention_seconds":         sqs.QueueAttributeNameMessageRetentionPeriod,
		"receive_wait_time_seconds":         sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds,
		"visibility_timeout_seconds":        sqs.QueueAttributeNameVisibilityTimeout,
		"policy":                            sqs.QueueAttributeNamePolicy,
		"redrive_policy":                    sqs.QueueAttributeNameRedrivePolicy,
		"arn":                               sqs.QueueAttributeNameQueueArn,
		"fifo_queue":                        sqs.QueueAttributeNameFifoQueue,
		"content_based_deduplication":       sqs.QueueAttributeNameContentBasedDeduplication,
		"kms_master_key_id":                 sqs.QueueAttributeNameKmsMasterKeyId,
		"kms_data_key_reuse_period_seconds": sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds,
		"deduplication_scope":               sqs.QueueAttributeNameDeduplicationScope,
		"fifo_throughput_limit":             sqs.QueueAttributeNameFifoThroughputLimit,
	}, sqsQueueSchema)
)

// A number of these are marked as computed because if you don't
// provide a value, SQS will provide you with defaults (which are the
// default values specified below)
func ResourceQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceQueueCreate,
		Read:   resourceQueueRead,
		Update: resourceQueueUpdate,
		Delete: resourceQueueDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			resourceAwsSqsQueueCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: sqsQueueSchema,
	}
}

func resourceQueueCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	fifoQueue := d.Get("fifo_queue").(bool)

	if fifoQueue {
		name = create.NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), FIFOQueueNameSuffix)
	} else {
		name = create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	input := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
	}

	attributes, err := sqsQueueAttributeMap.ResourceDataToApiAttributesCreate(d)

	if err != nil {
		return err
	}

	input.Attributes = aws.StringMap(attributes)

	// Tag-on-create is currently only supported in AWS Commercial
	if len(tags) > 0 && meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		input.Tags = tags.IgnoreAws().SqsTags()
	}

	log.Printf("[DEBUG] Creating SQS Queue: %s", input)
	var output *sqs.CreateQueueOutput
	err = resource.Retry(queueCreatedTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateQueue(input)

		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDeletedRecently) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateQueue(input)
	}

	if err != nil {
		return fmt.Errorf("error creating SQS Queue (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.QueueUrl))

	err = waitQueueAttributesPropagated(conn, d.Id(), attributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue (%s) attributes to create: %w", d.Id(), err)
	}

	// Tag-on-create is currently only supported in AWS Commercial
	if len(tags) > 0 && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID {
		if err := tftags.SqsUpdateTags(conn, d.Id(), nil, tags); err != nil {
			return fmt.Errorf("error updating SQS Queue (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceQueueRead(d, meta)
}

func resourceQueueRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindQueueAttributesByURL(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SQS Queue (%s): %w", d.Id(), err)
	}

	name, err := QueueNameFromURL(d.Id())

	if err != nil {
		return err
	}

	err = sqsQueueAttributeMap.ApiAttributesToResourceData(output, d)

	if err != nil {
		return err
	}

	// Backwards compatibility: https://github.com/hashicorp/terraform-provider-aws/issues/19786.
	if d.Get("kms_data_key_reuse_period_seconds").(int) == 0 {
		d.Set("kms_data_key_reuse_period_seconds", DefaultQueueKMSDataKeyReusePeriodSeconds)
	}

	d.Set("name", name)
	if d.Get("fifo_queue").(bool) {
		d.Set("name_prefix", create.NamePrefixFromNameWithSuffix(name, FIFOQueueNameSuffix))
	} else {
		d.Set("name_prefix", create.NamePrefixFromName(name))
	}
	d.Set("url", d.Id())

	tags, err := tftags.SqsListTags(conn, d.Id())

	if err != nil {
		// Non-standard partitions (e.g. US Gov) and some local development
		// solutions do not yet support this API call. Depending on the
		// implementation it may return InvalidAction or AWS.SimpleQueueService.UnsupportedOperation
		if !tfawserr.ErrCodeEquals(err, ErrCodeInvalidAction) && !tfawserr.ErrCodeEquals(err, sqs.ErrCodeUnsupportedOperation) {
			return fmt.Errorf("error listing tags for SQS Queue (%s): %w", d.Id(), err)
		}
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	if d.HasChangesExcept("tags", "tags_all") {
		attributes, err := sqsQueueAttributeMap.ResourceDataToApiAttributesUpdate(d)

		if err != nil {
			return err
		}

		input := &sqs.SetQueueAttributesInput{
			Attributes: aws.StringMap(attributes),
			QueueUrl:   aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating SQS Queue: %s", input)
		_, err = conn.SetQueueAttributes(input)

		if err != nil {
			return fmt.Errorf("error updating SQS Queue (%s) attributes: %w", d.Id(), err)
		}

		err = waitQueueAttributesPropagated(conn, d.Id(), attributes)

		if err != nil {
			return fmt.Errorf("error waiting for SQS Queue (%s) attributes to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.SqsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating SQS Queue (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceQueueRead(d, meta)
}

func resourceQueueDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	log.Printf("[DEBUG] Deleting SQS Queue: %s", d.Id())
	_, err := conn.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SQS Queue (%s): %w", d.Id(), err)
	}

	err = waitQueueDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSqsQueueCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	fifoQueue := diff.Get("fifo_queue").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		// Create.

		var name string

		if fifoQueue {
			name = create.NameWithSuffix(diff.Get("name").(string), diff.Get("name_prefix").(string), FIFOQueueNameSuffix)
		} else {
			name = create.Name(diff.Get("name").(string), diff.Get("name_prefix").(string))
		}

		var re *regexp.Regexp

		if fifoQueue {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid queue name: %s", name)
		}

	}

	if !fifoQueue && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
	}

	return nil
}
