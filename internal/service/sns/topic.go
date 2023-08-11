// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			Type:                  schema.TypeString,
			Optional:              true,
			Computed:              true,
			ValidateFunc:          validation.StringIsJSON,
			DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
			DiffSuppressOnRefresh: true,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		"signature_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(1, 2),
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
		names.AttrTags:    tftags.TagsSchema(),
		names.AttrTagsAll: tftags.TagsSchemaComputed(),
		"tracing_config": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(TopicTracingConfig_Values(), false),
		},
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
		"signature_version":                     TopicAttributeNameSignatureVersion,
		"sqs_failure_feedback_role_arn":         TopicAttributeNameSQSFailureFeedbackRoleARN,
		"sqs_success_feedback_role_arn":         TopicAttributeNameSQSSuccessFeedbackRoleARN,
		"sqs_success_feedback_sample_rate":      TopicAttributeNameSQSSuccessFeedbackSampleRate,
		"tracing_config":                        TopicAttributeNameTracingConfig,
	}, topicSchema).WithIAMPolicyAttribute("policy").WithMissingSetToNil("*")
)

// @SDKResource("aws_sns_topic", name="Topic")
// @Tags(identifierAttribute="id")
func ResourceTopic() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicCreate,
		ReadWithoutTimeout:   resourceTopicRead,
		UpdateWithoutTimeout: resourceTopicUpdate,
		DeleteWithoutTimeout: resourceTopicDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceTopicCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: topicSchema,
	}
}

func resourceTopicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	var name string
	fifoTopic := d.Get("fifo_topic").(bool)
	if fifoTopic {
		name = create.NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), FIFOTopicNameSuffix)
	} else {
		name = create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	input := &sns.CreateTopicInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	attributes, err := topicAttributeMap.ResourceDataToAPIAttributesCreate(d)

	if err != nil {
		return diag.FromErr(err)
	}

	// The FifoTopic attribute must be passed in the call to CreateTopic.
	if v, ok := attributes[TopicAttributeNameFIFOTopic]; ok {
		input.Attributes = aws.StringMap(map[string]string{
			TopicAttributeNameFIFOTopic: v,
		})

		delete(attributes, TopicAttributeNameFIFOTopic)
	}

	output, err := conn.CreateTopicWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateTopicWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Topic (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.TopicArn))

	// Retry for eventual consistency; if ABAC is in use, this takes some time
	// usually about 10s, presumably for tags really to be there, and we get a
	// permissions error.
	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return nil, putTopicAttributes(ctx, conn, d.Id(), attributes)
	}, sns.ErrCodeAuthorizationErrorException, "no identity-based policy allows")

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceTopicRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting SNS Topic (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTopicRead(ctx, d, meta)...)
}

func resourceTopicRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	attributes, err := FindTopicAttributesByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Topic (%s): %s", d.Id(), err)
	}

	err = topicAttributeMap.APIAttributesToResourceData(attributes, d)

	if err != nil {
		return diag.FromErr(err)
	}

	arn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	name := arn.Resource
	d.Set("name", name)
	if d.Get("fifo_topic").(bool) {
		d.Set("name_prefix", create.NamePrefixFromNameWithSuffix(name, FIFOTopicNameSuffix))
	} else {
		d.Set("name_prefix", create.NamePrefixFromName(name))
	}

	return nil
}

func resourceTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		attributes, err := topicAttributeMap.ResourceDataToAPIAttributesUpdate(d)

		if err != nil {
			return diag.FromErr(err)
		}

		err = putTopicAttributes(ctx, conn, d.Id(), attributes)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceTopicRead(ctx, d, meta)
}

func resourceTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn(ctx)

	log.Printf("[DEBUG] Deleting SNS Topic: %s", d.Id())
	_, err := conn.DeleteTopicWithContext(ctx, &sns.DeleteTopicInput{
		TopicArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SNS Topic (%s): %s", d.Id(), err)
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

func putTopicAttributes(ctx context.Context, conn *sns.SNS, arn string, attributes map[string]string) error {
	for name, value := range attributes {
		// Ignore an empty policy.
		if name == TopicAttributeNamePolicy && value == "" {
			continue
		}

		err := putTopicAttribute(ctx, conn, arn, name, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func putTopicAttribute(ctx context.Context, conn *sns.SNS, arn string, name, value string) error {
	const (
		topicPutAttributeTimeout = 2 * time.Minute
	)
	input := &sns.SetTopicAttributesInput{
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(value),
		TopicArn:       aws.String(arn),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, topicPutAttributeTimeout, func() (interface{}, error) {
		return conn.SetTopicAttributesWithContext(ctx, input)
	}, sns.ErrCodeInvalidParameterException)

	if err != nil {
		return fmt.Errorf("setting SNS Topic (%s) attribute (%s): %w", arn, name, err)
	}

	return nil
}
