package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/schema"
)

func generateVariableSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"string_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"double_value": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"dataset_content_version_value": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dataset_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"output_file_uri_value": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateContainerDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image": {
				Type:     schema.TypeString,
				Required: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"resource_configuration": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"volume_size_in_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"variable": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Type:     schema.TypeSet,
						Optional: false,
						Elem:     generateVariableSchema(),
					},
				},
			},
		},
	}
}

func generateQueryFilterSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"delta_time": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offset_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"time_expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateSqlQueryDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sql_query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateQueryFilterSchema(),
			},
		},
	}
}

func generateDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"container_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateContainerDatasetActionSchema(),
			},
			"query_action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateSqlQueryDatasetActionSchema(),
			},
		},
	}
}

func generateS3DestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"glue_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateDatasetContentDeliveryDestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"iot_events_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"s3_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateS3DestinationSchema(),
			},
		},
	}
}

func generateDatasetContentDeliveryRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"entry_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     generateDatasetContentDeliveryDestinationSchema(),
			},
		},
	}
}

func generateRetentionPeriodSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number_of_days": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.number_of_days"},
			},
		},
	}
}

func generateDatasetTriggerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dataset": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type: schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"schedule": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type: schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateVersioningConfigurationSchema() *schema.Resource{
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_version": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.max_version"},
			},		
		}
	}
}

func resourceAwsIotAnalyticsDataset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsDatasetCreate,
		Read:   resourceAwsIotAnalyticsDatasetRead,
		Update: resourceAwsIotAnalyticsDatasetUpdate,
		Delete: resourceAwsIotAnalyticsDatasetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     generateDatasetActionSchema(),
			},
			"content_delivery_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: generateDatasetContentDeliveryRuleSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: generateRetentionPeriodSchema(),
			},
			"trigger": {
				Type: schema.TypeSet,
				Optional: true,
				Elem: generateDatasetTriggerSchema(),
			},
			"versioning_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: generateVersioningConfigurationSchema(),
			},
		},
	}
}

func resourceAwsIotAnalyticsDatasetCreate(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsIotAnalyticsDatasetRead(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsIotAnalyticsDatasetUpdate(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsIotAnalyticsDatasetDelete(d *schema.ResourceData, meta interface{}) error {
}
