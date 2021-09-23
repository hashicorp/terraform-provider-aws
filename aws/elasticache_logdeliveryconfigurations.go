package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func getAwsElasticacheLogDeliveryConfigurationsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"destination_details": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							ConflictsWith: []string{"log_delivery_configurations.0.destination_details.0.kinesis_firehose"},
						},
						"kinesis_firehose": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delivery_stream": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							ConflictsWith: []string{"log_delivery_configurations.0.destination_details.0.cloudwatch_logs"},
						},
					},
				},
			},
			"destination_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(elasticache.DestinationType_Values(), false),
			},
			"log_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(elasticache.LogFormat_Values(), false),
			},
			"log_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      elasticache.LogTypeSlowLog,
				ValidateFunc: validation.StringInSlice(elasticache.LogType_Values(), false),
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func getAwsElasticacheLogDeliveryConfigurationsComputedSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"destination_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"kinesis_firehose": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delivery_stream": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"destination_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_format": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func expandAwsElasticacheLogDeliveryConfigurations(d *schema.ResourceData) []*elasticache.LogDeliveryConfigurationRequest {
	logDeliveryConfigurationRequest := elasticache.LogDeliveryConfigurationRequest{}

	if _, ok := d.GetOk("log_delivery_configurations"); !ok { // if d.HasChange() removed the block, send a `delete` request to the API.
		logDeliveryConfigurationRequest.SetEnabled(false)
		logDeliveryConfigurationRequest.SetLogType(elasticache.LogTypeSlowLog)
		logDeliveryConfigurations := []*elasticache.LogDeliveryConfigurationRequest{
			&logDeliveryConfigurationRequest,
		}
		return logDeliveryConfigurations
	}

	if _, ok := d.GetOk("log_delivery_configurations.0.destination_details"); ok {
		logDeliveryConfigurationRequest.DestinationDetails = expandAwsElasticacheDestinationDetails(d)
	}
	if v, ok := d.GetOk("log_delivery_configurations.0.destination_type"); ok {
		logDeliveryConfigurationRequest.DestinationType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("log_delivery_configurations.0.log_format"); ok {
		logDeliveryConfigurationRequest.LogFormat = aws.String(v.(string))
	}
	if v, ok := d.GetOk("log_delivery_configurations.0.log_type"); ok {
		logDeliveryConfigurationRequest.LogType = aws.String(v.(string))
	}

	logDeliveryConfigurations := []*elasticache.LogDeliveryConfigurationRequest{
		&logDeliveryConfigurationRequest,
	}
	return logDeliveryConfigurations
}

func expandAwsElasticacheDestinationDetails(d *schema.ResourceData) *elasticache.DestinationDetails {
	destinationDetails := elasticache.DestinationDetails{}
	if v, ok := d.GetOk("log_delivery_configurations.0.destination_details.0.cloudwatch_logs.0.log_group"); ok {
		destinationDetails.CloudWatchLogsDetails = &elasticache.CloudWatchLogsDestinationDetails{
			LogGroup: aws.String(v.(string)),
		}
	}
	if v, ok := d.GetOk("log_delivery_configurations.0.destination_details.0.kinesis_firehose.0.delivery_stream"); ok {
		destinationDetails.KinesisFirehoseDetails = &elasticache.KinesisFirehoseDestinationDetails{
			DeliveryStream: aws.String(v.(string)),
		}
	}
	return &destinationDetails
}

func flattenAwsElasticacheLogDeliveryConfigurations(logDeliveryConfiguration []*elasticache.LogDeliveryConfiguration) []map[string]interface{} {
	if len(logDeliveryConfiguration) == 0 || logDeliveryConfiguration[0] == nil {
		return nil
	}
	config := make(map[string]interface{})
	config["destination_details"] = flattenAwsElasticacheDestinationDetails(logDeliveryConfiguration[0].DestinationDetails)
	config["destination_type"] = aws.StringValue(logDeliveryConfiguration[0].DestinationType)
	config["log_format"] = aws.StringValue(logDeliveryConfiguration[0].LogFormat)
	config["log_type"] = aws.StringValue(logDeliveryConfiguration[0].LogType)
	config["message"] = aws.StringValue(logDeliveryConfiguration[0].Message)
	config["status"] = aws.StringValue(logDeliveryConfiguration[0].Status)
	result := []map[string]interface{}{config}
	return result
}

func flattenAwsElasticacheDestinationDetails(destinationDetails *elasticache.DestinationDetails) []map[string]interface{} {
	if destinationDetails == nil {
		return nil
	}
	config := make(map[string]interface{})
	config["cloudwatch_logs"] = flattenAwsElasticacheCloudWatchLogsDestinationDetails(destinationDetails.CloudWatchLogsDetails)
	config["kinesis_firehose"] = flattenAwsElasticacheKinesisFirehoseDestinationDetails(destinationDetails.KinesisFirehoseDetails)
	result := []map[string]interface{}{config}
	return result
}

func flattenAwsElasticacheCloudWatchLogsDestinationDetails(cloudWatchLogsDestinationDetails *elasticache.CloudWatchLogsDestinationDetails) []map[string]interface{} {
	if cloudWatchLogsDestinationDetails == nil {
		return nil
	}
	config := make(map[string]interface{})
	config["log_group"] = aws.StringValue(cloudWatchLogsDestinationDetails.LogGroup)
	result := []map[string]interface{}{config}
	return result
}

func flattenAwsElasticacheKinesisFirehoseDestinationDetails(kinesisFirehoseDestinationDetails *elasticache.KinesisFirehoseDestinationDetails) []map[string]interface{} {
	if kinesisFirehoseDestinationDetails == nil {
		return nil
	}
	config := make(map[string]interface{})
	config["delivery_stream"] = aws.StringValue(kinesisFirehoseDestinationDetails.DeliveryStream)
	result := []map[string]interface{}{config}
	return result
}
