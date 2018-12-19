package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

var dataPipelineScheduleTypeList = []string{
	"cron",
	"ondemand",
	"timeseries",
}

var dataPipelineActionOnResourceFailureList = []string{
	"retryall",
	"retrynone",
}

var dataPipelineActionOnTaskFailureList = []string{
	"continue",
	"terminate",
}

var dataPipelineParameterObjectTypeList = []string{
	"String",
	"Integer",
	"Double",
	"AWS::S3::ObjectKey",
}

var dataPipelineFailureAndRerunModeList = []string{
	"cascade",
	"none",
}

var dataPipelineS3EncryptionTypeList = []string{
	"NONE",
	"SERVER_SIDE_ENCRYPTION",
}

var dataPipelineS3CompressionTypeList = []string{
	"none",
	"gzip",
}

func resourceAwsDataPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataPipelineCreate,
		Read:   resourceAwsDataPipelineRead,
		Update: resourceAwsDataPipelineUpdate,
		Delete: resourceAwsDataPipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Default
			"default": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},

						"failure_and_rerun_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineFailureAndRerunModeList, false),
							Default:      "none",
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"role": {
							Type:     schema.TypeString,
							Required: true,
						},

						"resource_role": {
							Type:     schema.TypeString,
							Required: true,
						},

						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// Activity
			"copy_activity": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"runs_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"worker_group": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"attempt_status": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"attempt_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"depends_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"failure_and_rerun_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineFailureAndRerunModeList, false),
							Default:      "none",
						},

						"input": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"late_after_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"max_active_instances": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"maximum_retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"on_fail": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_late_action": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_success": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"output": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"precondition": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"report_progress_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_delay": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"schedule_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},
					},
				},
			},

			// Resources
			"ec2_resource": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"action_on_resource_failure": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineActionOnResourceFailureList, false),
						},

						"action_on_task_failure": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineActionOnTaskFailureList, false),
						},

						"associate_public_ip_address": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"attempt_status": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"attempt_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"init_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},

						"key_pair": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"late_after_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"max_active_instances": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"maximum_retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"on_fail": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_late_action": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_success": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"schedule_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},

						"security_group_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"security_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"spot_bid_price": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 20),
						},

						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"terminate_after": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"use_on_demand_on_last_attempt": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			// Databases
			"rds_database": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"username": {
							Type:     schema.TypeString,
							Required: true,
						},

						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},

						"rds_instance_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"database_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"jdbc_driver_jar_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"jdbc_properties": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// Data Node
			"s3_data_node": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"attempt_status": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"attempt_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"compression": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineS3CompressionTypeList, false),
							Default:      "none",
						},

						"data_format": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"depends_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"directory_path": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"failure_and_rerun_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineFailureAndRerunModeList, false),
							Default:      "none",
						},

						"file_path": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"late_after_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"manifest_file_path": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"max_active_instances": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"maximum_retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"on_fail": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_late_action": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_success": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"precondition": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"report_progress_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_delay": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"runs_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"s3_encryption_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineS3EncryptionTypeList, false),
						},

						"schedule_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},

						"worker_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"sql_data_node": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"table": {
							Type:     schema.TypeString,
							Required: true,
						},

						"attempt_status": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"attempt_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"create_table_sql": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"database": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"depends_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"failure_and_rerun_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineFailureAndRerunModeList, false),
							Default:      "none",
						},

						"insert_query": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"late_after_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"max_active_instances": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"maximum_retries": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"on_fail": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_late_action": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_success": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"precondition": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"report_progress_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_delay": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"runs_on": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"schedule_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},

						"schema_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"select_query": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"worker_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"schedule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"period": {
							Type:     schema.TypeString,
							Required: true,
						},

						"start_at": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"schedule.start_date_time"},
						},

						"start_date_time": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"schedule.start_at"},
						},

						"end_date_time": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"occurrences": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"schedule.end_date_time"},
						},

						"parent": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"parameter_object": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "String",
							ValidateFunc: validation.StringInSlice(dataPipelineParameterObjectTypeList, false),
						},
						"optional": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"allowed_values": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"default": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"is_array": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"parameter_value": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"string_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDataPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	name := d.Get("name").(string)
	uniqueID := resource.PrefixedUniqueId(name)

	input := datapipeline.CreatePipelineInput{
		Name:     aws.String(name),
		UniqueId: aws.String(uniqueID),
	}

	if v, ok := d.GetOk("description"); ok && v != nil {
		input.Description = aws.String(v.(string))
	}

	tags := tagsFromMapDataPipeline(d.Get("tags").(map[string]interface{}))
	if len(tags) != 0 && tags != nil {
		input.Tags = tags
	}

	resp, err := conn.CreatePipeline(&input)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] DataPipeline received: %s", resp)
	d.SetId(*resp.PipelineId)
	d.Set("tags", tagsToMapDataPipeline(tags))

	return resourceAwsDataPipelineUpdate(d, meta)
}

func resourceAwsDataPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn
	pipelineID := d.Id()

	parameterObjects := d.Get("parameter_object").([]interface{})
	parameterObjectConfigs := make([]*datapipeline.ParameterObject, 0, len(parameterObjects))
	for _, p := range parameterObjects {
		p := p.(map[string]interface{})
		po, err := buildParameterObject(p)
		if err != nil {
			return err
		}
		parameterObjectConfigs = append(parameterObjectConfigs, po)
	}

	parameterValues := d.Get("parameter_value").([]interface{})
	parameterValuesConfigs := make([]*datapipeline.ParameterValue, 0, len(parameterValues))
	for _, c := range parameterValues {
		pv := &datapipeline.ParameterValue{}
		c := c.(map[string]interface{})
		if val, ok := c["id"].(string); ok && val != "" {
			pv.SetId(val)
		}
		if val, ok := c["string_value"].(string); ok && val != "" {
			pv.SetStringValue(val)
		}
		parameterValuesConfigs = append(parameterValuesConfigs, pv)
	}

	pipelineObjectsConfigs := []*datapipeline.PipelineObject{}

	if v, ok := d.GetOk("default"); ok && v != nil {
		v := v.(map[string]interface{})
		po, err := buildDefaultPipelineObject(v)
		if err != nil {
			return err
		}
		pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
	} else {
		log.Printf("[DEBUG] Not setting default pipeline object for %q", d.Id())
	}

	if v, ok := d.GetOk("copy_activity"); ok && v != nil {
		for _, c := range v.([]interface{}) {
			c := c.(map[string]interface{})
			po, err := buildCommonPipelineObject("CopyActivity", c)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	if v, ok := d.GetOk("ec2_resource"); ok && v != nil {
		for _, e := range v.([]interface{}) {
			e := e.(map[string]interface{})
			po, err := buildCommonPipelineObject("Ec2Resource", e)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	if v, ok := d.GetOk("rds_database"); ok && v != nil {
		for _, r := range v.([]interface{}) {
			r := r.(map[string]interface{})
			po, err := buildCommonPipelineObject("RdsDatabase", r)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	if v, ok := d.GetOk("s3_data_node"); ok && v != nil {
		for _, s := range v.([]interface{}) {
			s := s.(map[string]interface{})
			po, err := buildCommonPipelineObject("S3DataNode", s)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	if v, ok := d.GetOk("sql_data_node"); ok && v != nil {
		for _, s := range v.([]interface{}) {
			s := s.(map[string]interface{})
			po, err := buildCommonPipelineObject("SqlDataNode", s)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	if v, ok := d.GetOk("schedule"); ok && v != nil {
		for _, s := range v.([]interface{}) {
			s := s.(map[string]interface{})
			po, err := buildCommonPipelineObject("Schedule", s)
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, po)
		}
	}

	input := &datapipeline.PutPipelineDefinitionInput{
		PipelineId: aws.String(pipelineID),
	}
	if len(parameterObjectConfigs) > 0 {
		input.ParameterObjects = parameterObjectConfigs
	}

	if len(parameterValuesConfigs) > 0 {
		input.ParameterValues = parameterValuesConfigs
	}

	if len(pipelineObjectsConfigs) > 0 {
		input.PipelineObjects = pipelineObjectsConfigs
	}

	err := input.Validate()
	if err != nil {
		return err
	}

	resp, err := conn.PutPipelineDefinition(input)

	if err != nil {
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error updating DataPipeline (%s): %s", d.Id(), err)
	}

	for _, validationWarn := range resp.ValidationWarnings {
		for _, warn := range validationWarn.Warnings {
			log.Printf("[WARN] %s:  %s", *validationWarn.Id, *warn)
		}
	}

	for _, validationError := range resp.ValidationErrors {
		for _, err := range validationError.Errors {
			log.Printf("[ERROR] %s:  %s", *validationError.Id, *err)
		}
	}

	if err := setTagsDataPipeline(conn, d); err != nil {
		return fmt.Errorf("Error update tags: %s", err)
	}

	return resourceAwsDataPipelineRead(d, meta)
}

func resourceAwsDataPipelineRead(d *schema.ResourceData, meta interface{}) error {

	v, err := resourceAwsDataPipelineRetrieve(d.Id(), meta.(*AWSClient).datapipelineconn)

	if err != nil {
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if v == nil {
		d.SetId("")
		return nil
	}

	d.SetId(*v.PipelineId)
	d.Set("name", v.Name)
	d.Set("description", v.Description)
	d.Set("tags", tagsToMapDataPipeline(v.Tags))

	r, err := resourceAwsDataPipelineDefinitionRetrieve(d.Id(), meta.(*AWSClient).datapipelineconn)
	if err != nil {
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			log.Printf("[WARN] DataPipeline (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	// Parameter Objects
	if err := d.Set("parameter_object", flattenParameterObjects(r.ParameterObjects)); err != nil {
		return fmt.Errorf("error reading DataPipeline \"%s\" parameter objects: %s", d.Id(), err.Error())
	}

	// Parameter Values
	if err := d.Set("parameter_value", flattenParameterValues(r.ParameterValues)); err != nil {
		return fmt.Errorf("error reading DataPipeline \"%s\" parameter values: %s", d.Id(), err.Error())
	}

	if err := flattenPipelineObjects(d, r.PipelineObjects); err != nil {
		return fmt.Errorf("error reading DataPipeline \"%s\" pipeline object: %s", d.Id(), err.Error())
	}

	return nil
}

func resourceAwsDataPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	log.Printf("[DEBUG] DataPipeline  destroy: %s", d.Id())

	opts := datapipeline.DeletePipelineInput{
		PipelineId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] DataPipeline destroy configuration: %v", opts)
	_, err := conn.DeletePipeline(&opts)
	if err != nil {
		return fmt.Errorf("Error deleting Data Pipeline %s: %s", d.Id(), err.Error())
	}

	return waitForDataPipelineDeletion(conn, d.Id())
}

func resourceAwsDataPipelineRetrieve(id string, conn *datapipeline.DataPipeline) (*datapipeline.PipelineDescription, error) {
	opts := datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(id)},
	}

	log.Printf("[DEBUG] Data Pipeline describe configuration: %#v", opts)

	resp, err := conn.DescribePipelines(&opts)
	if err != nil {
		return nil, err
	}

	if len(resp.PipelineDescriptionList) != 1 ||
		*resp.PipelineDescriptionList[0].PipelineId != id {
		if err != nil {
			return nil, nil
		}
	}

	return resp.PipelineDescriptionList[0], nil
}

func resourceAwsDataPipelineDefinitionRetrieve(id string, conn *datapipeline.DataPipeline) (*datapipeline.GetPipelineDefinitionOutput, error) {
	opts := datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(id),
	}
	log.Printf("[DEBUG] Data Pipeline describe configuration: %#v", opts)

	resp, err := conn.GetPipelineDefinition(&opts)
	if err != nil {
		return nil, nil
	}

	return resp, nil
}

func flattenParameterObjects(objects []*datapipeline.ParameterObject) []map[string]interface{} {
	parameterObjects := make([]map[string]interface{}, 0, len(objects))
	for _, object := range objects {
		var obj map[string]interface{}

		obj["id"] = *object.Id
		for _, attribute := range object.Attributes {
			if *attribute.Key == "description" {
				obj["description"] = *attribute.StringValue
			}
			if *attribute.Key == "optional" {
				obj["optional"] = *attribute.StringValue
			}
			if *attribute.Key == "allowed_values" {
				obj["type"] = *attribute.StringValue
			}
			if *attribute.Key == "default" {
				obj["default"] = *attribute.StringValue
			}
			if *attribute.Key == "is_array" {
				obj["is_array"] = *attribute.StringValue
			}
		}
		parameterObjects = append(parameterObjects, obj)
	}

	return parameterObjects
}

func flattenParameterValues(objects []*datapipeline.ParameterValue) []map[string]interface{} {
	parameterValues := make([]map[string]interface{}, 0, len(objects))

	for _, object := range objects {
		var obj map[string]interface{}
		obj["id"] = *object.Id
		obj["string_value"] = *object.StringValue

		parameterValues = append(parameterValues, obj)
	}
	return parameterValues
}

func flattenPipelineObjects(d *schema.ResourceData, pipelineObjects []*datapipeline.PipelineObject) error {
	var copyActivities, ec2Resources, rdsDatabases, s3DataNodes, sqlDataNodes, schedules []map[string]interface{}

	for _, pipelineObject := range pipelineObjects {
		for _, field := range pipelineObject.Fields {
			if *field.Key == "type" {
				switch *field.StringValue {
				case "Default":
					object, err := flattenDefaultPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					d.Set("default", object)
				case "CopyActivity":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					copyActivities = append(copyActivities, object)
				case "Ec2Resource":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					ec2Resources = append(ec2Resources, object)
				case "RdsDatabase":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					rdsDatabases = append(rdsDatabases, object)
				case "S3DataNode":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					s3DataNodes = append(s3DataNodes, object)
				case "SqlDataNode":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					sqlDataNodes = append(sqlDataNodes, object)
				case "Schedule":
					object, err := flattenCommonPipelineObject(pipelineObject)
					if err != nil {
						return err
					}
					schedules = append(schedules, object)
				}
			}
		}
	}

	if len(copyActivities) != 0 {
		d.Set("copy_activity", copyActivities)
	}

	if len(ec2Resources) != 0 {
		d.Set("ec2_resource", ec2Resources)
	}

	if len(rdsDatabases) != 0 {
		d.Set("rds_database", rdsDatabases)
	}

	if len(s3DataNodes) != 0 {
		d.Set("s3_data_node", s3DataNodes)
	}

	if len(sqlDataNodes) != 0 {
		d.Set("sql_data_node", sqlDataNodes)
	}

	if len(schedules) != 0 {
		d.Set("schedule", schedules)
	}

	return nil
}

func waitForDataPipelineDeletion(conn *datapipeline.DataPipeline, pipelineID string) error {
	params := &datapipeline.DescribePipelinesInput{
		PipelineIds: []*string{aws.String(pipelineID)},
	}
	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribePipelines(params)
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			return nil
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("DataPipeline (%s) still exists", pipelineID))
	})
}
