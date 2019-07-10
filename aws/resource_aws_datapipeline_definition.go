package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

var dataPipelineScheduleTypeList = []string{
	"cron",
	"ondemand",
	"timeseries",
}

var dataPipelineFailureAndRerunModeList = []string{
	"cascade",
	"none",
}

var dataPipelineParameterObjectTypeList = []string{
	"String",
	"Integer",
	"Double",
	"AWS::S3::ObjectKey",
}

var dataPipelineParameterObjectStartAtList = []string{
	"FIRST_ACTIVATION_DATE_TIME",
}

func resourceAwsDataPipelineDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataPipelineDefinitionPut,
		Read:   resourceAwsDataPipelineDefinitionRead,
		Update: resourceAwsDataPipelineDefinitionPut,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDataPipelineDefinitionImport,
		},

		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"default": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},

						"role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},

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

						"schedule": {
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
							ValidateFunc:  validation.StringInSlice(dataPipelineParameterObjectStartAtList, false),
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
							Default:  false,
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
		},
	}
}

func resourceAwsDataPipelineDefinitionPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.PutPipelineDefinitionInput{
		PipelineId: aws.String(pipelineID),
	}

	pipelineObjectsConfigs := []*datapipeline.PipelineObject{}

	if v, ok := d.GetOk("default"); ok {
		pipelineObject, err := expandDataPipelineDefaultPipelineObject(v.([]interface{}))
		if err != nil {
			return err
		}
		pipelineObjectsConfigs = append(pipelineObjectsConfigs, pipelineObject)
	}

	if v, ok := d.GetOk("schedule"); ok {
		for _, s := range v.([]interface{}) {
			pipelineObject, err := expandDataPipelinePipelineObject("Schedule", s.(map[string]interface{}))
			if err != nil {
				return err
			}
			pipelineObjectsConfigs = append(pipelineObjectsConfigs, pipelineObject)
		}
	}

	if len(pipelineObjectsConfigs) > 0 {
		input.PipelineObjects = pipelineObjectsConfigs
	}

	parameterObjectsConfigs := []*datapipeline.ParameterObject{}
	if v, ok := d.GetOk("parameter_object"); ok {
		for _, p := range v.([]interface{}) {
			parameterObject, err := expandDataPipelineParameterObject(p.(map[string]interface{}))
			if err != nil {
				return err
			}
			parameterObjectsConfigs = append(parameterObjectsConfigs, parameterObject)
		}
	}

	if len(parameterObjectsConfigs) > 0 {
		input.ParameterObjects = parameterObjectsConfigs
	}

	parameterValuesConfigs := []*datapipeline.ParameterValue{}
	if v, ok := d.GetOk("parameter_value"); ok {
		for _, p := range v.([]interface{}) {
			parameterValue, err := expandDataPipelineParameterValue(p.(map[string]interface{}))
			if err != nil {
				return err
			}
			parameterValuesConfigs = append(parameterValuesConfigs, parameterValue)
		}
	}

	if len(parameterValuesConfigs) > 0 {
		input.ParameterValues = parameterValuesConfigs
	}

	_, err := conn.PutPipelineDefinition(input)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
		log.Printf("[WARN] DataPipeline (%s) not found, removing from state", pipelineID)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error putting DataPipeline (%s) Definition: %s", pipelineID, err)
	}

	d.SetId(pipelineID)
	return resourceAwsDataPipelineDefinitionRead(d, meta)
}

func resourceAwsDataPipelineDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	v, err := resourceAwsDataPipelineDefinitionRetrieve(d.Id(), conn)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
		log.Printf("[WARN] DataPipeline (%s) Definition not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting DataPipeline (%s) Definition: %s", d.Id(), err)
	}

	// Parameter Objects
	if err := d.Set("parameter_object", flattenDataPipelineParameterObjects(v.ParameterObjects)); err != nil {
		return fmt.Errorf("error setting DataPipeline (%s) parameter objects: %s", d.Id(), err)
	}

	// Parameter Values
	if err := d.Set("parameter_value", flattenDataPipelineParameterValues(v.ParameterValues)); err != nil {
		return fmt.Errorf("error setting DataPipeline (%s) parameter values: %s", d.Id(), err)
	}

	if err := flattenDataPipelinePipelineObjects(d, v.PipelineObjects); err != nil {
		return err
	}

	return nil
}

func resourceAwsDataPipelineDefinitionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("pipeline_id", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsDataPipelineDefinitionRetrieve(id string, conn *datapipeline.DataPipeline) (*datapipeline.GetPipelineDefinitionOutput, error) {
	opts := datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(id),
	}
	log.Printf("[DEBUG] Data Pipeline describe configuration: %#v", opts)

	resp, err := conn.GetPipelineDefinition(&opts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
