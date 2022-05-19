package sns

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	topicSchema = map[string]*schema.Schema{
		"application_failure_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"application_success_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"application_success_feedback_sample_rate": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"content_based_deduplication": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"delivery_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"display_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"fifo_topic": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},
		"firehose_failure_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"firehose_success_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"firehose_success_feedback_sample_rate": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"http_failure_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"http_success_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"http_success_feedback_sample_rate": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"kms_master_key_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"lambda_failure_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"lambda_success_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"lambda_success_feedback_sample_rate": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 100),
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
		"owner": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"policy": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"sqs_failure_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"sqs_success_feedback_role_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
		},
		"sqs_success_feedback_sample_rate": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 100),
		},
		"tags":     tftags.TagsSchema(),
		"tags_all": tftags.TagsSchemaComputed(),
	}

	topicAttributeMap = attrmap.New(map[string]string{
		"application_failure_feedback_role_arn":    TopicAttributeNameApplicationFailureFeedbackRoleARN,
		"application_success_feedback_role_arn":    TopicAttributeNameApplicationSuccessFeedbackRoleARN,
		"application_success_feedback_sample_rate": TopicAttributeNameApplicationSuccessFeedbackSampleRate,
		"arn":                                   TopicAttributeNameTopicARN,
		"content_based_deduplication":           TopicAttributeNameContentBasedDeduplication,
		"delivery_policy":                       TopicAttributeNameDeliveryPolicy,
		"display_name":                          TopicAttributeNameDisplayName,
		"fifo_topic":                            TopicAttributeNameFIFOTopic,
		"firehose_failure_feedback_role_arn":    TopicAttributeNameFirehoseFailureFeedbackRoleARN,
		"firehose_success_feedback_role_arn":    TopicAttributeNameFirehoseSuccessFeedbackRoleARN,
		"firehose_success_feedback_sample_rate": TopicAttributeNameFirehoseSuccessFeedbackSampleRate,
		"http_failure_feedback_role_arn":        TopicAttributeNameHTTPFailureFeedbackRoleARN,
		"http_success_feedback_role_arn":        TopicAttributeNameHTTPSuccessFeedbackRoleARN,
		"http_success_feedback_sample_rate":     TopicAttributeNameHTTPSuccessFeedbackSampleRate,
		"kms_master_key_id":                     TopicAttributeNameKMSMasterKeyId,
		"lambda_failure_feedback_role_arn":      TopicAttributeNameLambdaFailureFeedbackRoleARN,
		"lambda_success_feedback_role_arn":      TopicAttributeNameLambdaSuccessFeedbackRoleARN,
		"lambda_success_feedback_sample_rate":   TopicAttributeNameLambdaSuccessFeedbackSampleRate,
		"owner":                                 TopicAttributeNameOwner,
		"policy":                                TopicAttributeNamePolicy,
		"sqs_failure_feedback_role_arn":         TopicAttributeNameSQSFailureFeedbackRoleARN,
		"sqs_success_feedback_role_arn":         TopicAttributeNameSQSSuccessFeedbackRoleARN,
		"sqs_success_feedback_sample_rate":      TopicAttributeNameSQSSuccessFeedbackSampleRate,
	}, topicSchema).WithIAMPolicyAttribute("policy")
)

func ResourceTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceTopicCreate,
		Read:   resourceTopicRead,
		Update: resourceTopicUpdate,
		Delete: resourceTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			resourceTopicCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: topicSchema,
	}
}

func resourceTopicCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	fifoTopic := d.Get("fifo_topic").(bool)
	if fifoTopic {
		name = create.NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), FIFOTopicNameSuffix)
	} else {
		name = create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	input := &sns.CreateTopicInput{
		Name: aws.String(name),
	}

	attributes, err := topicAttributeMap.ResourceDataToApiAttributesCreate(d)

	if err != nil {
		return err
	}

	// The FifoTopic attribute must be passed in the call to CreateTopic.
	if v, ok := attributes[TopicAttributeNameFIFOTopic]; ok {
		input.Attributes = aws.StringMap(map[string]string{
			TopicAttributeNameFIFOTopic: v,
		})

		delete(attributes, TopicAttributeNameFIFOTopic)
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating SNS Topic: %s", input)
	output, err := conn.CreateTopic(input)

	// Some partitions may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed creating SNS Topic (%s) with tags: %s. Trying create without tags.", name, err)
		input.Tags = nil
		output, err = conn.CreateTopic(input)
	}

	if err != nil {
		return fmt.Errorf("failed creating SNS Topic (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.TopicArn))

	err = putTopicAttributes(conn, d.Id(), attributes)

	if err != nil {
		return err
	}

	// Post-create tagging supported in some partitions
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, d.Id(), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			// if default tags only, log and continue (i.e., should error if explicitly setting tags and they can't be)
			log.Printf("[WARN] failed adding tags after create for SNS Topic (%s): %s", d.Id(), err)
			return resourceTopicRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for SNS Topic (%s): %w", d.Id(), err)
		}
	}

	return resourceTopicRead(d, meta)
}

func resourceTopicRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	attributes, err := FindTopicAttributesByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SNS Topic (%s): %w", d.Id(), err)
	}

	err = topicAttributeMap.ApiAttributesToResourceData(attributes, d)

	if err != nil {
		return err
	}

	arn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing ARN (%s): %w", d.Id(), err)
	}

	name := arn.Resource
	d.Set("name", name)
	if d.Get("fifo_topic").(bool) {
		d.Set("name_prefix", create.NamePrefixFromNameWithSuffix(name, FIFOTopicNameSuffix))
	} else {
		d.Set("name_prefix", create.NamePrefixFromName(name))
	}

	tags, err := ListTags(conn, d.Id())

	if verify.CheckISOErrorTagsUnsupported(err) {
		// ISO partitions may not support tagging, giving error
		log.Printf("[WARN] failed listing tags for SNS Topic (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed listing tags for SNS Topic (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceTopicUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	if d.HasChangesExcept("tags", "tags_all") {
		attributes, err := topicAttributeMap.ResourceDataToApiAttributesUpdate(d)

		if err != nil {
			return err
		}

		err = putTopicAttributes(conn, d.Id(), attributes)

		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Id(), o, n)

		if verify.CheckISOErrorTagsUnsupported(err) {
			// ISO partitions may not support tagging, giving error
			log.Printf("[WARN] failed updating tags for SNS Topic (%s): %s", d.Id(), err)
			return resourceTopicRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed updating tags for SNS Topic (%s): %w", d.Id(), err)
		}
	}

	return resourceTopicRead(d, meta)
}

func resourceTopicDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	log.Printf("[DEBUG] Deleting SNS Topic: %s", d.Id())
	_, err := conn.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SNS Topic (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceTopicCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	fifoTopic := diff.Get("fifo_topic").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		// Create.

		var name string

		if fifoTopic {
			name = create.NameWithSuffix(diff.Get("name").(string), diff.Get("name_prefix").(string), FIFOTopicNameSuffix)
		} else {
			name = create.Name(diff.Get("name").(string), diff.Get("name_prefix").(string))
		}

		var re *regexp.Regexp

		if fifoTopic {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,251}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,256}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid topic name: %s", name)
		}

	}

	if !fifoTopic && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO topics")
	}

	return nil
}

func putTopicAttributes(conn *sns.SNS, arn string, attributes map[string]string) error {
	for name, value := range attributes {
		// Ignore an empty policy.
		if name == TopicAttributeNamePolicy && value == "" {
			continue
		}

		err := putTopicAttribute(conn, arn, name, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func putTopicAttribute(conn *sns.SNS, arn string, name, value string) error {
	input := &sns.SetTopicAttributesInput{
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(value),
		TopicArn:       aws.String(arn),
	}

	log.Printf("[DEBUG] Setting SNS Topic attribute: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(topicPutAttributeTimeout, func() (interface{}, error) {
		return conn.SetTopicAttributes(input)
	}, sns.ErrCodeInvalidParameterException)

	if err != nil {
		return fmt.Errorf("error setting SNS Topic (%s) attribute (%s): %w", arn, name, err)
	}

	return nil
}
