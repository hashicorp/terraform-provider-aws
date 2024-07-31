// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/attrmap"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
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
		"archive_policy": {
			Type:                  schema.TypeString,
			Optional:              true,
			ValidateFunc:          validation.StringIsJSON,
			DiffSuppressFunc:      verify.SuppressEquivalentJSONWithEmptyDiffs,
			DiffSuppressOnRefresh: true,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		names.AttrARN: {
			Type:     schema.TypeString,
			Computed: true,
		},
		"beginning_archive_time": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"content_based_deduplication": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"delivery_policy": {
			Type:                  schema.TypeString,
			Optional:              true,
			ValidateFunc:          validation.StringIsJSON,
			DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
			DiffSuppressOnRefresh: true,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
		},
		names.AttrDisplayName: {
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
		names.AttrName: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{names.AttrNamePrefix},
		},
		names.AttrNamePrefix: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{names.AttrName},
		},
		names.AttrOwner: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrPolicy: {
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
			ValidateFunc: validation.StringInSlice(topicTracingConfig_Values(), false),
		},
	}

	topicAttributeMap = attrmap.New(map[string]string{
		"application_failure_feedback_role_arn":    topicAttributeNameApplicationFailureFeedbackRoleARN,
		"application_success_feedback_role_arn":    topicAttributeNameApplicationSuccessFeedbackRoleARN,
		"application_success_feedback_sample_rate": topicAttributeNameApplicationSuccessFeedbackSampleRate,
		"archive_policy":                           topicAttributeNameArchivePolicy,
		names.AttrARN:                              topicAttributeNameTopicARN,
		"beginning_archive_time":                   topicAttributeNameBeginningArchiveTime,
		"content_based_deduplication":              topicAttributeNameContentBasedDeduplication,
		"delivery_policy":                          topicAttributeNameDeliveryPolicy,
		names.AttrDisplayName:                      topicAttributeNameDisplayName,
		"fifo_topic":                               topicAttributeNameFIFOTopic,
		"firehose_failure_feedback_role_arn":       topicAttributeNameFirehoseFailureFeedbackRoleARN,
		"firehose_success_feedback_role_arn":       topicAttributeNameFirehoseSuccessFeedbackRoleARN,
		"firehose_success_feedback_sample_rate":    topicAttributeNameFirehoseSuccessFeedbackSampleRate,
		"http_failure_feedback_role_arn":           topicAttributeNameHTTPFailureFeedbackRoleARN,
		"http_success_feedback_role_arn":           topicAttributeNameHTTPSuccessFeedbackRoleARN,
		"http_success_feedback_sample_rate":        topicAttributeNameHTTPSuccessFeedbackSampleRate,
		"kms_master_key_id":                        topicAttributeNameKMSMasterKeyId,
		"lambda_failure_feedback_role_arn":         topicAttributeNameLambdaFailureFeedbackRoleARN,
		"lambda_success_feedback_role_arn":         topicAttributeNameLambdaSuccessFeedbackRoleARN,
		"lambda_success_feedback_sample_rate":      topicAttributeNameLambdaSuccessFeedbackSampleRate,
		names.AttrOwner:                            topicAttributeNameOwner,
		names.AttrPolicy:                           topicAttributeNamePolicy,
		"signature_version":                        topicAttributeNameSignatureVersion,
		"sqs_failure_feedback_role_arn":            topicAttributeNameSQSFailureFeedbackRoleARN,
		"sqs_success_feedback_role_arn":            topicAttributeNameSQSSuccessFeedbackRoleARN,
		"sqs_success_feedback_sample_rate":         topicAttributeNameSQSSuccessFeedbackSampleRate,
		"tracing_config":                           topicAttributeNameTracingConfig,
	}, topicSchema).WithIAMPolicyAttribute(names.AttrPolicy).WithMissingSetToNil("*")
)

// @SDKResource("aws_sns_topic", name="Topic")
// @Tags(identifierAttribute="id")
// @Testing(existsType="map[string]string")
func resourceTopic() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	name := topicName(d)
	input := &sns.CreateTopicInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	attributes, err := topicAttributeMap.ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// The FifoTopic attribute must be passed in the call to CreateTopic.
	if v, ok := attributes[topicAttributeNameFIFOTopic]; ok {
		input.Attributes = map[string]string{
			topicAttributeNameFIFOTopic: v,
		}

		delete(attributes, topicAttributeNameFIFOTopic)
	}

	output, err := conn.CreateTopic(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		input.Tags = nil

		output, err = conn.CreateTopic(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Topic (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TopicArn))

	// Retry for eventual consistency; if ABAC is in use, this takes some time
	// usually about 10s, presumably for tags really to be there, and we get a
	// permissions error.
	_, err = tfresource.RetryWhenIsAErrorMessageContains[*types.AuthorizationErrorException](ctx, propagationTimeout, func() (interface{}, error) {
		return nil, putTopicAttributes(ctx, conn, d.Id(), attributes)
	}, "no identity-based policy allows")

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
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
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := findTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Topic (%s): %s", d.Id(), err)
	}

	err = topicAttributeMap.APIAttributesToResourceData(attributes, d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	arn, err := arn.Parse(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := arn.Resource
	d.Set(names.AttrName, name)
	if d.Get("fifo_topic").(bool) {
		d.Set(names.AttrNamePrefix, create.NamePrefixFromNameWithSuffix(name, fifoTopicNameSuffix))
	} else {
		d.Set(names.AttrNamePrefix, create.NamePrefixFromName(name))
	}

	return diags
}

func resourceTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		attributes, err := topicAttributeMap.ResourceDataToAPIAttributesUpdate(d)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		err = putTopicAttributes(ctx, conn, d.Id(), attributes)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceTopicRead(ctx, d, meta)...)
}

func resourceTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	log.Printf("[DEBUG] Deleting SNS Topic: %s", d.Id())
	_, err := conn.DeleteTopic(ctx, &sns.DeleteTopicInput{
		TopicArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SNS Topic (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTopicCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	fifoTopic := diff.Get("fifo_topic").(bool)
	archivePolicy := diff.Get("archive_policy").(string)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		// Create.
		name := topicName(diff)
		var re *regexp.Regexp

		if fifoTopic {
			re = regexache.MustCompile(`^[0-9A-Za-z_-]{1,251}\.fifo$`)
		} else {
			re = regexache.MustCompile(`^[0-9A-Za-z_-]{1,256}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid topic name: %s", name)
		}
	}

	if !fifoTopic {
		if archivePolicy != "" {
			return errors.New("message archive policy can only be set for FIFO topics")
		}
		if contentBasedDeduplication {
			return errors.New("content-based deduplication can only be set for FIFO topics")
		}
	}

	return nil
}

func putTopicAttributes(ctx context.Context, conn *sns.Client, arn string, attributes map[string]string) error {
	for name, value := range attributes {
		// Ignore an empty policy.
		if name == topicAttributeNamePolicy && value == "" {
			continue
		}

		err := putTopicAttribute(ctx, conn, arn, name, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func putTopicAttribute(ctx context.Context, conn *sns.Client, arn string, name, value string) error {
	const (
		timeout = 2 * time.Minute
	)
	input := &sns.SetTopicAttributesInput{
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(value),
		TopicArn:       aws.String(arn),
	}

	_, err := tfresource.RetryWhenIsA[*types.InvalidParameterException](ctx, timeout, func() (interface{}, error) {
		return conn.SetTopicAttributes(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("setting SNS Topic (%s) attribute (%s): %w", arn, name, err)
	}

	return nil
}

func topicName(d sdkv2.ResourceDiffer) string {
	optFns := []create.NameGeneratorOptionsFunc{create.WithConfiguredName(d.Get(names.AttrName).(string)), create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string))}
	if d.Get("fifo_topic").(bool) {
		optFns = append(optFns, create.WithSuffix(fifoTopicNameSuffix))
	}
	return create.NewNameGenerator(optFns...).Generate()
}

// findTopicAttributesWithValidAWSPrincipalsByARN returns topic attributes, ensuring that any Policy field
// is populated with valid AWS principals, i.e. the principal is either an AWS Account ID or an ARN.
// nosemgrep:ci.aws-in-func-name
func findTopicAttributesWithValidAWSPrincipalsByARN(ctx context.Context, conn *sns.Client, arn string) (map[string]string, error) {
	var attributes map[string]string
	err := tfresource.Retry(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		attributes, err = findTopicAttributesByARN(ctx, conn, arn)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		valid, err := tfiam.PolicyHasValidAWSPrincipals(attributes[topicAttributeNamePolicy])
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if !valid {
			return retry.RetryableError(errors.New("contains invalid principals"))
		}

		return nil
	})

	return attributes, err
}

func findTopicAttributesByARN(ctx context.Context, conn *sns.Client, arn string) (map[string]string, error) {
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(arn),
	}

	output, err := conn.GetTopicAttributes(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Attributes, nil
}
