package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
				Elem:     generateVariableSchema(),
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
				MaxItems: 1,
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
				MaxItems: 1,
				Elem:     generateContainerDatasetActionSchema(),
			},
			"query_action": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
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
			"iotevents_destination": {
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
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateVersioningConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_versions": {
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
		},
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
				Elem:     generateDatasetContentDeliveryRuleSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
			"trigger": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateDatasetTriggerSchema(),
			},
			"versioning_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateVersioningConfigurationSchema(),
			},
		},
	}
}

func parseVariable(rawAction map[string]interface{}) *iotanalytics.Variable {
	variable := &iotanalytics.Variable{
		Name: aws.String(rawAction["name"].(string)),
	}

	if v, ok := rawAction["string_value"]; ok {
		variable.StringValue = aws.String(v.(string))
	}

	if v, ok := rawAction["double_value"]; ok {
		variable.DoubleValue = aws.Float64(v.(float64))
	}

	rawDatasetContentVersionValueSet := rawAction["dataset_content_version_value"].(*schema.Set).List()
	if len(rawDatasetContentVersionValueSet) > 0 {
		rawDatasetContentVersionValue := rawDatasetContentVersionValueSet[0].(map[string]interface{})
		datasetContentVersionValue := &iotanalytics.DatasetContentVersionValue{
			DatasetName: aws.String(rawDatasetContentVersionValue["dataset_name"].(string)),
		}
		variable.DatasetContentVersionValue = datasetContentVersionValue
	}

	rawOutputFileUriValueSet := rawAction["output_file_uri_value"].(*schema.Set).List()
	if len(rawOutputFileUriValueSet) > 0 {
		rawOutputFileUriValue := rawOutputFileUriValueSet[0].(map[string]interface{})
		outputFileUriValue := &iotanalytics.OutputFileUriValue{
			FileName: aws.String(rawOutputFileUriValue["file_name"].(string)),
		}
		variable.OutputFileUriValue = outputFileUriValue
	}

	return variable
}

func parseContainerAction(rawContainerAction map[string]interface{}) *iotanalytics.ContainerDatasetAction {
	containerAction := &iotanalytics.ContainerDatasetAction{
		Image:            aws.String(rawContainerAction["image"].(string)),
		ExecutionRoleArn: aws.String(rawContainerAction["execution_role_arn"].(string)),
	}

	rawResourceConfiguration := rawContainerAction["resource_configuration"].(*schema.Set).List()[0].(map[string]interface{})
	containerAction.ResourceConfiguration = &iotanalytics.ResourceConfiguration{
		ComputeType:    aws.String(rawResourceConfiguration["compute_type"].(string)),
		VolumeSizeInGB: aws.Int64(int64(rawResourceConfiguration["volume_size_in_gb"].(int))),
	}

	variables := make([]*iotanalytics.Variable, 0)
	rawVariables := rawContainerAction["variable"].(*schema.Set).List()
	for _, rawVar := range rawVariables {
		variable := parseVariable(rawVar.(map[string]interface{}))
		variables = append(variables, variable)
	}
	containerAction.Variables = variables

	return containerAction
}

func parseQueryFilter(rawQueryFilter map[string]interface{}) *iotanalytics.QueryFilter {
	rawDeltaTime := rawQueryFilter["delta_time"].(*schema.Set).List()[0].(map[string]interface{})
	deltaTime := &iotanalytics.DeltaTime{
		OffsetSeconds:  aws.Int64(int64(rawDeltaTime["offset_seconds"].(int))),
		TimeExpression: aws.String(rawDeltaTime["time_expression"].(string)),
	}
	queryFilter := &iotanalytics.QueryFilter{
		DeltaTime: deltaTime,
	}
	return queryFilter
}

func parseSqlQueryAction(rawSqlQueryAction map[string]interface{}) *iotanalytics.SqlQueryDatasetAction {
	sqlQueryAction := &iotanalytics.SqlQueryDatasetAction{
		SqlQuery: aws.String(rawSqlQueryAction["sql_query"].(string)),
	}

	filters := make([]*iotanalytics.QueryFilter, 0)
	rawFilters := rawSqlQueryAction["filter"].(*schema.Set).List()
	for _, rawFilter := range rawFilters {
		filter := parseQueryFilter(rawFilter.(map[string]interface{}))
		filters = append(filters, filter)
	}
	sqlQueryAction.Filters = filters
	return sqlQueryAction

}

func parseAction(rawAction map[string]interface{}) *iotanalytics.DatasetAction {
	action := &iotanalytics.DatasetAction{
		ActionName: aws.String(rawAction["name"].(string)),
	}

	rawContainerActionSet := rawAction["container_action"].(*schema.Set).List()
	if len(rawContainerActionSet) > 0 {
		rawContainerAction := rawContainerActionSet[0].(map[string]interface{})
		action.ContainerAction = parseContainerAction(rawContainerAction)
	}

	rawQueryActionSet := rawAction["container_action"].(*schema.Set).List()
	if len(rawQueryActionSet) > 0 {
		rawQueryAction := rawQueryActionSet[0].(map[string]interface{})
		action.QueryAction = parseSqlQueryAction(rawQueryAction)
	}

	return action
}

func parseS3Destination(rawS3Destination map[string]interface{}) *iotanalytics.S3DestinationConfiguration {
	s3Destination := &iotanalytics.S3DestinationConfiguration{
		Bucket:  aws.String(rawS3Destination["bucket"].(string)),
		Key:     aws.String(rawS3Destination["key"].(string)),
		RoleArn: aws.String(rawS3Destination["role_arn"].(string)),
	}

	rawGlueConfigurationSet := rawS3Destination["glue_configuration"].(*schema.Set).List()
	if len(rawGlueConfigurationSet) > 0 {
		rawGlueConfiguration := rawGlueConfigurationSet[0].(map[string]interface{})
		s3Destination.GlueConfiguration = &iotanalytics.GlueConfiguration{
			DatabaseName: aws.String(rawGlueConfiguration["database_name"].(string)),
			TableName:    aws.String(rawGlueConfiguration["table_name"].(string)),
		}
	}

	return s3Destination
}

func parseIotEventsDestination(rawIotEventsDestination map[string]interface{}) *iotanalytics.IotEventsDestinationConfiguration {
	return &iotanalytics.IotEventsDestinationConfiguration{
		InputName: aws.String(rawIotEventsDestination["input_name"].(string)),
		RoleArn:   aws.String(rawIotEventsDestination["role_arn"].(string)),
	}
}

func parseDestination(rawDestination map[string]interface{}) *iotanalytics.DatasetContentDeliveryDestination {
	destination := &iotanalytics.DatasetContentDeliveryDestination{}

	rawIotEventsDestinationSet := rawDestination["iotevents_destination"].(*schema.Set).List()
	if len(rawIotEventsDestinationSet) > 0 {
		rawIotEventsDestination := rawIotEventsDestinationSet[0].(map[string]interface{})
		destination.IotEventsDestinationConfiguration = parseIotEventsDestination(rawIotEventsDestination)
	}

	rawS3DestinationSet := rawDestination["s3_destination"].(*schema.Set).List()
	if len(rawS3DestinationSet) > 0 {
		rawS3Destination := rawS3DestinationSet[0].(map[string]interface{})
		destination.S3DestinationConfiguration = parseS3Destination(rawS3Destination)
	}

	return destination
}

func parseContentDeliveryRule(rawContentDeliveryRule map[string]interface{}) *iotanalytics.DatasetContentDeliveryRule {
	rawDestination := rawContentDeliveryRule["destination"].(*schema.Set).List()[0].(map[string]interface{})
	datasetContentDeliveryRule := &iotanalytics.DatasetContentDeliveryRule{
		Destination: parseDestination(rawDestination),
	}

	if rawEntryPoint, ok := rawContentDeliveryRule["entry_name"]; ok {
		datasetContentDeliveryRule.EntryName = aws.String(rawEntryPoint.(string))
	}

	return datasetContentDeliveryRule
}

func parseRetentionPeriod(rawRetentionPeriod map[string]interface{}) *iotanalytics.RetentionPeriod {

	var numberOfDays *int64
	if v, ok := rawRetentionPeriod["number_of_days"]; ok && int64(v.(int)) > 1 {
		numberOfDays = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawRetentionPeriod["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.RetentionPeriod{
		NumberOfDays: numberOfDays,
		Unlimited:    unlimited,
	}
}

func parseTrigger(rawTrigger map[string]interface{}) *iotanalytics.DatasetTrigger {
	trigger := &iotanalytics.DatasetTrigger{}

	rawDatasetSet := rawTrigger["dataset"].(*schema.Set).List()
	if len(rawDatasetSet) > 0 {
		rawDataset := rawDatasetSet[0].(map[string]interface{})
		trigger.Dataset = &iotanalytics.TriggeringDataset{
			Name: aws.String(rawDataset["name"].(string)),
		}
	}

	rawScheduleSet := rawTrigger["schedule"].(*schema.Set).List()
	if len(rawScheduleSet) > 0 {
		rawSchedule := rawScheduleSet[0].(map[string]interface{})
		trigger.Schedule = &iotanalytics.Schedule{
			Expression: aws.String(rawSchedule["expression"].(string)),
		}
	}

	return trigger
}

func parseVersioningConfiguration(rawVersioningConfiguration map[string]interface{}) *iotanalytics.VersioningConfiguration {
	var maxVersion *int64
	if v, ok := rawVersioningConfiguration["max_versions"]; ok && int64(v.(int)) > 1 {
		maxVersion = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawVersioningConfiguration["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.VersioningConfiguration{
		MaxVersions: maxVersion,
		Unlimited:   unlimited,
	}
}

func resourceAwsIotAnalyticsDatasetCreate(d *schema.ResourceData, meta interface{}) error {
	// TODO: make function that return structure of ready-to-use fields to fill
	// CreateDatasetInput and UpdateDatasetInput structures
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.CreateDatasetInput{
		DatasetName: aws.String(name),
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.CreateDataset(params)

	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsIotAnalyticsDatasetRead(d, meta)
}

func resourceAwsIotAnalyticsDatasetRead(d *schema.ResourceData, meta interface{}) error {
}

func resourceAwsIotAnalyticsDatasetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.UpdateDatasetInput{
		DatasetName: aws.String(name),
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.UpdateDataset(params)

	if err != nil {
		return err
	}

	return resourceAwsIotAnalyticsDatasetRead(d, meta)

}

func resourceAwsIotAnalyticsDatasetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteDatasetInput{
		DatasetName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Dataset: %s", params)
	_, err := conn.DeleteDataset(params)

	return err
}
