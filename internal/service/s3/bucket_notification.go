// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_notification", name="Bucket Notification")
func resourceBucketNotification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketNotificationPut,
		ReadWithoutTimeout:   resourceBucketNotificationRead,
		UpdateWithoutTimeout: resourceBucketNotificationPut,
		DeleteWithoutTimeout: resourceBucketNotificationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"eventbridge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"lambda_function": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"lambda_function_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"queue": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"queue_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"topic": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrTopicARN: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceBucketNotificationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		filterRulesSliceStartLen = 2
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	bucket := d.Get(names.AttrBucket).(string)

	var eventbridgeConfig *types.EventBridgeConfiguration
	if d.Get("eventbridge").(bool) {
		eventbridgeConfig = &types.EventBridgeConfiguration{}
	}

	lambdaFunctionNotifications := d.Get("lambda_function").([]interface{})
	lambdaConfigs := make([]types.LambdaFunctionConfiguration, 0, len(lambdaFunctionNotifications))
	for i, c := range lambdaFunctionNotifications {
		lc := types.LambdaFunctionConfiguration{}

		c := c.(map[string]interface{})

		if val, ok := c[names.AttrID].(string); ok && val != "" {
			lc.Id = aws.String(val)
		} else {
			lc.Id = aws.String(id.PrefixedUniqueId("tf-s3-lambda-"))
		}

		if val, ok := c["lambda_function_arn"].(string); ok {
			lc.LambdaFunctionArn = aws.String(val)
		}

		lc.Events = flex.ExpandStringyValueSet[types.Event](d.Get(fmt.Sprintf("lambda_function.%d.events", i)).(*schema.Set))

		filterRules := make([]types.FilterRule, 0, filterRulesSliceStartLen)
		if val, ok := c["filter_prefix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNamePrefix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if val, ok := c["filter_suffix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNameSuffix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if len(filterRules) > 0 {
			lc.Filter = &types.NotificationConfigurationFilter{
				Key: &types.S3KeyFilter{
					FilterRules: filterRules,
				},
			}
		}
		lambdaConfigs = append(lambdaConfigs, lc)
	}

	queueNotifications := d.Get("queue").([]interface{})
	queueConfigs := make([]types.QueueConfiguration, 0, len(queueNotifications))
	for i, c := range queueNotifications {
		qc := types.QueueConfiguration{}

		c := c.(map[string]interface{})

		if val, ok := c[names.AttrID].(string); ok && val != "" {
			qc.Id = aws.String(val)
		} else {
			qc.Id = aws.String(id.PrefixedUniqueId("tf-s3-queue-"))
		}

		if val, ok := c["queue_arn"].(string); ok {
			qc.QueueArn = aws.String(val)
		}

		qc.Events = flex.ExpandStringyValueSet[types.Event](d.Get(fmt.Sprintf("queue.%d.events", i)).(*schema.Set))

		filterRules := make([]types.FilterRule, 0, filterRulesSliceStartLen)
		if val, ok := c["filter_prefix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNamePrefix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if val, ok := c["filter_suffix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNameSuffix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if len(filterRules) > 0 {
			qc.Filter = &types.NotificationConfigurationFilter{
				Key: &types.S3KeyFilter{
					FilterRules: filterRules,
				},
			}
		}
		queueConfigs = append(queueConfigs, qc)
	}

	topicNotifications := d.Get("topic").([]interface{})
	topicConfigs := make([]types.TopicConfiguration, 0, len(topicNotifications))
	for i, c := range topicNotifications {
		tc := types.TopicConfiguration{}

		c := c.(map[string]interface{})

		if val, ok := c[names.AttrID].(string); ok && val != "" {
			tc.Id = aws.String(val)
		} else {
			tc.Id = aws.String(id.PrefixedUniqueId("tf-s3-topic-"))
		}

		if val, ok := c[names.AttrTopicARN].(string); ok {
			tc.TopicArn = aws.String(val)
		}

		tc.Events = flex.ExpandStringyValueSet[types.Event](d.Get(fmt.Sprintf("topic.%d.events", i)).(*schema.Set))

		filterRules := make([]types.FilterRule, 0, filterRulesSliceStartLen)
		if val, ok := c["filter_prefix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNamePrefix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if val, ok := c["filter_suffix"].(string); ok && val != "" {
			filterRule := types.FilterRule{
				Name:  types.FilterRuleNameSuffix,
				Value: aws.String(val),
			}
			filterRules = append(filterRules, filterRule)
		}
		if len(filterRules) > 0 {
			tc.Filter = &types.NotificationConfigurationFilter{
				Key: &types.S3KeyFilter{
					FilterRules: filterRules,
				},
			}
		}
		topicConfigs = append(topicConfigs, tc)
	}

	notificationConfiguration := &types.NotificationConfiguration{}
	if eventbridgeConfig != nil {
		notificationConfiguration.EventBridgeConfiguration = eventbridgeConfig
	}
	if len(lambdaConfigs) > 0 {
		notificationConfiguration.LambdaFunctionConfigurations = lambdaConfigs
	}
	if len(queueConfigs) > 0 {
		notificationConfiguration.QueueConfigurations = queueConfigs
	}
	if len(topicConfigs) > 0 {
		notificationConfiguration.TopicConfigurations = topicConfigs
	}
	input := &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(bucket),
		NotificationConfiguration: notificationConfiguration,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketNotificationConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "NotificationConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Notification: %s", bucket, err)
	}

	if d.IsNewResource() {
		d.SetId(bucket)

		_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
			return findBucketNotificationConfiguration(ctx, conn, d.Id(), "")
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Notification (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketNotificationRead(ctx, d, meta)...)
}

func resourceBucketNotificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	output, err := findBucketNotificationConfiguration(ctx, conn, d.Id(), "")

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Notification (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Notification (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())
	d.Set("eventbridge", output.EventBridgeConfiguration != nil)
	if err := d.Set("lambda_function", flattenLambdaFunctionConfigurations(output.LambdaFunctionConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_function: %s", err)
	}
	if err := d.Set("queue", flattenQueueConfigurations(output.QueueConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting queue: %s", err)
	}
	if err := d.Set("topic", flattenTopicConfigurations(output.TopicConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting queue: %s", err)
	}

	return diags
}

func resourceBucketNotificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	input := &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(d.Id()),
		NotificationConfiguration: &types.NotificationConfiguration{},
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Notification: %s", d.Id())
	_, err := conn.PutBucketNotificationConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Notification (%s): %s", d.Id(), err)
	}

	// Don't wait for the notification configuration to disappear as it still exists after update.

	return diags
}

func findBucketNotificationConfiguration(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketNotificationConfigurationOutput, error) {
	input := &s3.GetBucketNotificationConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketNotificationConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
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

	return output, nil
}

func flattenNotificationConfigurationFilter(filter *types.NotificationConfigurationFilter) map[string]interface{} {
	filterRules := map[string]interface{}{}
	if filter.Key == nil || filter.Key.FilterRules == nil {
		return filterRules
	}

	for _, f := range filter.Key.FilterRules {
		name := strings.ToLower(string(f.Name))
		if name == string(types.FilterRuleNamePrefix) {
			filterRules["filter_prefix"] = aws.ToString(f.Value)
		} else if name == string(types.FilterRuleNameSuffix) {
			filterRules["filter_suffix"] = aws.ToString(f.Value)
		}
	}
	return filterRules
}

func flattenTopicConfigurations(configs []types.TopicConfiguration) []map[string]interface{} {
	topicNotifications := make([]map[string]interface{}, 0, len(configs))
	for _, notification := range configs {
		var conf map[string]interface{}
		if filter := notification.Filter; filter != nil {
			conf = flattenNotificationConfigurationFilter(filter)
		} else {
			conf = map[string]interface{}{}
		}

		conf[names.AttrID] = aws.ToString(notification.Id)
		conf["events"] = notification.Events
		conf[names.AttrTopicARN] = aws.ToString(notification.TopicArn)
		topicNotifications = append(topicNotifications, conf)
	}

	return topicNotifications
}

func flattenQueueConfigurations(configs []types.QueueConfiguration) []map[string]interface{} {
	queueNotifications := make([]map[string]interface{}, 0, len(configs))
	for _, notification := range configs {
		var conf map[string]interface{}
		if filter := notification.Filter; filter != nil {
			conf = flattenNotificationConfigurationFilter(filter)
		} else {
			conf = map[string]interface{}{}
		}

		conf[names.AttrID] = aws.ToString(notification.Id)
		conf["events"] = notification.Events
		conf["queue_arn"] = aws.ToString(notification.QueueArn)
		queueNotifications = append(queueNotifications, conf)
	}

	return queueNotifications
}

func flattenLambdaFunctionConfigurations(configs []types.LambdaFunctionConfiguration) []map[string]interface{} {
	lambdaFunctionNotifications := make([]map[string]interface{}, 0, len(configs))
	for _, notification := range configs {
		var conf map[string]interface{}
		if filter := notification.Filter; filter != nil {
			conf = flattenNotificationConfigurationFilter(filter)
		} else {
			conf = map[string]interface{}{}
		}

		conf[names.AttrID] = aws.ToString(notification.Id)
		conf["events"] = notification.Events
		conf["lambda_function_arn"] = aws.ToString(notification.LambdaFunctionArn)
		lambdaFunctionNotifications = append(lambdaFunctionNotifications, conf)
	}

	return lambdaFunctionNotifications
}
