package sns

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			ForceNew:         false,
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

	topicAttributeMap = create.AttrMap(map[string]string{
		"application_failure_feedback_role_arn":    TopicAttributeNameApplicationFailureFeedbackRoleArn,
		"application_success_feedback_role_arn":    TopicAttributeNameApplicationSuccessFeedbackRoleArn,
		"application_success_feedback_sample_rate": TopicAttributeNameApplicationSuccessFeedbackSampleRate,
		"arn":                                   TopicAttributeNameTopicArn,
		"content_based_deduplication":           TopicAttributeNameContentBasedDeduplication,
		"delivery_policy":                       TopicAttributeNameDeliveryPolicy,
		"display_name":                          TopicAttributeNameDisplayName,
		"fifo_topic":                            TopicAttributeNameFifoTopic,
		"firehose_failure_feedback_role_arn":    TopicAttributeNameFirehoseFailureFeedbackRoleArn,
		"firehose_success_feedback_role_arn":    TopicAttributeNameFirehoseSuccessFeedbackRoleArn,
		"firehose_success_feedback_sample_rate": TopicAttributeNameFirehoseSuccessFeedbackSampleRate,
		"http_failure_feedback_role_arn":        TopicAttributeNameHTTPFailureFeedbackRoleArn,
		"http_success_feedback_role_arn":        TopicAttributeNameHTTPSuccessFeedbackRoleArn,
		"http_success_feedback_sample_rate":     TopicAttributeNameHTTPSuccessFeedbackSampleRate,
		"kms_master_key_id":                     TopicAttributeNameKmsMasterKeyId,
		"lambda_failure_feedback_role_arn":      TopicAttributeNameLambdaFailureFeedbackRoleArn,
		"lambda_success_feedback_role_arn":      TopicAttributeNameLambdaSuccessFeedbackRoleArn,
		"lambda_success_feedback_sample_rate":   TopicAttributeNameLambdaSuccessFeedbackSampleRate,
		"owner":                                 TopicAttributeNameOwner,
		"policy":                                TopicAttributeNamePolicy,
		"sqs_failure_feedback_role_arn":         TopicAttributeNameSQSFailureFeedbackRoleArn,
		"sqs_success_feedback_role_arn":         TopicAttributeNameSQSSuccessFeedbackRoleArn,
		"sqs_success_feedback_sample_rate":      TopicAttributeNameSQSSuccessFeedbackSampleRate,
	}, topicSchema)
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

	// If FifoTopic is true, then the attribute must be passed into the call to CreateTopic.
	if v, ok := attributes[TopicAttributeNameFifoTopic]; ok {
		input.Attributes = aws.StringMap(map[string]string{
			TopicAttributeNameFifoTopic: v,
		})
		delete(attributes, TopicAttributeNameFifoTopic)
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating SNS Topic: %s", input)
	output, err := conn.CreateTopic(input)

	if err != nil {
		return fmt.Errorf("error creating SNS Topic (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.TopicArn))

	if policy, ok := attributes[TopicAttributeNamePolicy]; ok && policy != "" {
		policyToPut, err := structure.NormalizeJsonString(policy)

		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
		}

		attributes[TopicAttributeNamePolicy] = policyToPut
	}

	err = putTopicAttributes(conn, d.Id(), attributes)

	if err != nil {
		return err
	}

	return resourceTopicRead(d, meta)
}

func resourceTopicUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	// update mutable attributes
	if d.HasChange("application_failure_feedback_role_arn") {
		_, v := d.GetChange("application_failure_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "ApplicationFailureFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_role_arn") {
		_, v := d.GetChange("application_success_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "ApplicationSuccessFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("arn") {
		_, v := d.GetChange("arn")
		if err := updateTopicAttribute(d.Id(), "TopicArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("delivery_policy") {
		_, v := d.GetChange("delivery_policy")
		if err := updateTopicAttribute(d.Id(), "DeliveryPolicy", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("display_name") {
		_, v := d.GetChange("display_name")
		if err := updateTopicAttribute(d.Id(), "DisplayName", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("http_failure_feedback_role_arn") {
		_, v := d.GetChange("http_failure_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "HTTPFailureFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_role_arn") {
		_, v := d.GetChange("http_success_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "HTTPSuccessFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("kms_master_key_id") {
		_, v := d.GetChange("kms_master_key_id")
		if err := updateTopicAttribute(d.Id(), "KmsMasterKeyId", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("content_based_deduplication") {
		_, v := d.GetChange("content_based_deduplication")
		if err := updateTopicAttribute(d.Id(), "ContentBasedDeduplication", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_failure_feedback_role_arn") {
		_, v := d.GetChange("lambda_failure_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "LambdaFailureFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_role_arn") {
		_, v := d.GetChange("lambda_success_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "LambdaSuccessFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}

	if d.HasChange("policy") {
		o, n := d.GetChange("policy")

		if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
			policy, err := structure.NormalizeJsonString(n.(string))

			if err != nil {
				return fmt.Errorf("policy contains an invalid JSON: %s", err)
			}

			if err := updateTopicAttribute(d.Id(), "Policy", policy, conn); err != nil {
				return err
			}
		}
	}

	if d.HasChange("sqs_failure_feedback_role_arn") {
		_, v := d.GetChange("sqs_failure_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "SQSFailureFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_role_arn") {
		_, v := d.GetChange("sqs_success_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "SQSSuccessFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_sample_rate") {
		_, v := d.GetChange("application_success_feedback_sample_rate")
		if err := updateTopicAttribute(d.Id(), "ApplicationSuccessFeedbackSampleRate", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_sample_rate") {
		_, v := d.GetChange("http_success_feedback_sample_rate")
		if err := updateTopicAttribute(d.Id(), "HTTPSuccessFeedbackSampleRate", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_sample_rate") {
		_, v := d.GetChange("lambda_success_feedback_sample_rate")
		if err := updateTopicAttribute(d.Id(), "LambdaSuccessFeedbackSampleRate", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_sample_rate") {
		_, v := d.GetChange("sqs_success_feedback_sample_rate")
		if err := updateTopicAttribute(d.Id(), "SQSSuccessFeedbackSampleRate", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("firehose_failure_feedback_role_arn") {
		_, v := d.GetChange("firehose_failure_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "FirehoseFailureFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("firehose_success_feedback_role_arn") {
		_, v := d.GetChange("firehose_success_feedback_role_arn")
		if err := updateTopicAttribute(d.Id(), "FirehoseSuccessFeedbackRoleArn", v, conn); err != nil {
			return err
		}
	}
	if d.HasChange("firehose_success_feedback_sample_rate") {
		_, v := d.GetChange("firehose_success_feedback_sample_rate")
		if err := updateTopicAttribute(d.Id(), "FirehoseSuccessFeedbackSampleRate", v, conn); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
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

	// Save the policy to set here as ApiAttributesToResourceData will set it to the attribut's value.
	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), attributes[TopicAttributeNamePolicy])

	if err != nil {
		return err
	}

	err = topicAttributeMap.ApiAttributesToResourceData(attributes, d)

	if err != nil {
		return err
	}

	d.Set("policy", policyToSet)

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

	if err != nil {
		return fmt.Errorf("error listing tags for SNS Topic (%s): %w", d.Id(), err)
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

func updateTopicAttribute(topicArn, name string, value interface{}, conn *sns.SNS) error {
	// Ignore an empty policy
	if name == "Policy" && value == "" {
		return nil
	}
	log.Printf("[DEBUG] Updating SNS Topic Attribute: %s", name)

	// Make API call to update attributes
	req := sns.SetTopicAttributesInput{
		TopicArn:       aws.String(topicArn),
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(fmt.Sprintf("%v", value)),
	}

	// Retry the update in the event of an eventually consistent style of
	// error, where say an IAM resource is successfully created but not
	// actually available. See https://github.com/hashicorp/terraform/issues/3660
	_, err := verify.RetryOnAWSCode(sns.ErrCodeInvalidParameterException, func() (interface{}, error) {
		return conn.SetTopicAttributes(&req)
	})

	if err != nil {
		return fmt.Errorf("error setting SNS Topic (%s) attributes: %w", topicArn, err)
	}

	return nil
}

func putTopicAttributes(conn *sns.SNS, arn string, attributes map[string]string) error {
	for name, value := range attributes {
		// Ignore an empty policy.
		if name == TopicAttributeNamePolicy && value == "" {
			continue
		}

		input := &sns.SetTopicAttributesInput{
			AttributeName:  aws.String(name),
			AttributeValue: aws.String(value),
			TopicArn:       aws.String(arn),
		}

		log.Printf("[DEBUG] Setting SNS Topic attribute: %s", input)
		_, err := tfresource.RetryWhenAWSErrCodeEquals(topicCreateTimeout, func() (interface{}, error) {
			return conn.SetTopicAttributes(input)
		}, sns.ErrCodeInvalidParameterException)

		if err != nil {
			return fmt.Errorf("error setting SNS Topic (%s) attribute (%s): %w", arn, name, err)
		}
	}

	return nil
}
