package aws

import (
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
