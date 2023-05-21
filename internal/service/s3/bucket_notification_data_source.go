package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_s3_bucket_notification")
func DataSourceBucketNotification() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketNotification,
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_function_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lambda_function_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"events": {
							Type:     schema.TypeList,
							Computed: true,
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"queue_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"events": {
							Type:     schema.TypeList,
							Computed: true,
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"topic_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"topic_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"events": {
							Type:     schema.TypeList,
							Computed: true,
						},
						"filter_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"filter_suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBucketNotification(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket := d.Get("bucket").(string)
	input := &s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(bucket),
	}

	notification, err := conn.GetBucketNotificationConfigurationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Failed getting S3 bucket (%s): %s", bucket, err)
	}

	d.Set("lambda_function_configurations", getLambdaConfigurations(notification.LambdaFunctionConfigurations))
	d.Set("queue_configurations", getQueueConfigurations(notification.QueueConfigurations))
	d.Set("topic_configurations", getTopicConfigurations(notification.TopicConfigurations))

	return nil
}

func getLambdaConfigurations(configurations []*s3.LambdaFunctionConfiguration) []map[string]interface{} {
	lambdaFunctionConfigurations := make([]map[string]interface{}, 0)
	for _, lambdaFunctionConf := range configurations {
		lambdaFilterRules := lambdaFunctionConf.Filter.Key.FilterRules
		lambdaFunctionConfigurations = append(lambdaFunctionConfigurations, map[string]interface{}{
			"id":                  aws.StringValue(lambdaFunctionConf.Id),
			"lambda_function_arn": aws.StringValue(lambdaFunctionConf.LambdaFunctionArn),
			"events":              aws.StringValueSlice(lambdaFunctionConf.Events),
			"filter_prefix":       getValue(lambdaFilterRules, "Prefix"),
			"filter_suffix":       getValue(lambdaFilterRules, "Suffix"),
		})
	}
	return lambdaFunctionConfigurations
}

func getQueueConfigurations(configurations []*s3.QueueConfiguration) []map[string]interface{} {
	queueConfigurations := make([]map[string]interface{}, 0)
	for _, queueConf := range configurations {
		queueFilterRules := queueConf.Filter.Key.FilterRules
		queueConfigurations = append(queueConfigurations, map[string]interface{}{
			"id":            aws.StringValue(queueConf.Id),
			"queue_arn":     aws.StringValue(queueConf.QueueArn),
			"events":        aws.StringValueSlice(queueConf.Events),
			"filter_prefix": getValue(queueFilterRules, "Prefix"),
			"filter_suffix": getValue(queueFilterRules, "Suffix"),
		})
	}
	return queueConfigurations
}

func getTopicConfigurations(configurations []*s3.TopicConfiguration) []map[string]interface{} {
	topicConfigurations := make([]map[string]interface{}, 0)
	for _, topicConf := range configurations {
		topicFilterRules := topicConf.Filter.Key.FilterRules
		topicConfigurations = append(topicConfigurations, map[string]interface{}{
			"id":            aws.StringValue(topicConf.Id),
			"topic_arn":     aws.StringValue(topicConf.TopicArn),
			"events":        aws.StringValueSlice(topicConf.Events),
			"filter_prefix": getValue(topicFilterRules, "Prefix"),
			"filter_suffix": getValue(topicFilterRules, "Suffix"),
		})
	}
	return topicConfigurations
}

func getValue(filterRules []*s3.FilterRule, key string) string {
	for _, rule := range filterRules {
		if *rule.Name == key {
			return aws.StringValue(rule.Value)
		}
	}
	return ""
}
