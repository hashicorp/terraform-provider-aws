package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmMaintenanceWindowTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmMaintenanceWindowTaskCreate,
		Read:   resourceAwsSsmMaintenanceWindowTaskRead,
		Delete: resourceAwsSsmMaintenanceWindowTaskDelete,

		Schema: map[string]*schema.Schema{
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"max_concurrency": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"max_errors": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"task_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"task_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_role_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"targets": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsSSMMaintenanceWindowTaskName,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},

			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"logging_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_bucket_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"task_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func expandAwsSsmMaintenanceWindowLoggingInfo(config []interface{}) *ssm.LoggingInfo {

	loggingConfig := config[0].(map[string]interface{})

	loggingInfo := &ssm.LoggingInfo{
		S3BucketName: aws.String(loggingConfig["s3_bucket_name"].(string)),
		S3Region:     aws.String(loggingConfig["s3_region"].(string)),
	}

	if s := loggingConfig["s3_bucket_prefix"].(string); s != "" {
		loggingInfo.S3KeyPrefix = aws.String(s)
	}

	return loggingInfo
}

func flattenAwsSsmMaintenanceWindowLoggingInfo(loggingInfo *ssm.LoggingInfo) []interface{} {

	result := make(map[string]interface{})
	result["s3_bucket_name"] = *loggingInfo.S3BucketName
	result["s3_region"] = *loggingInfo.S3Region

	if loggingInfo.S3KeyPrefix != nil {
		result["s3_bucket_prefix"] = *loggingInfo.S3KeyPrefix
	}

	return []interface{}{result}
}

func expandAwsSsmTaskParameters(config []interface{}) map[string]*ssm.MaintenanceWindowTaskParameterValueExpression {
	params := make(map[string]*ssm.MaintenanceWindowTaskParameterValueExpression)
	for _, v := range config {
		paramConfig := v.(map[string]interface{})
		params[paramConfig["name"].(string)] = &ssm.MaintenanceWindowTaskParameterValueExpression{
			Values: expandStringList(paramConfig["values"].([]interface{})),
		}
	}
	return params
}

func flattenAwsSsmTaskParameters(taskParameters map[string]*ssm.MaintenanceWindowTaskParameterValueExpression) []interface{} {
	result := make([]interface{}, 0, len(taskParameters))
	for k, v := range taskParameters {
		taskParam := map[string]interface{}{
			"name":   k,
			"values": flattenStringList(v.Values),
		}
		result = append(result, taskParam)
	}

	return result
}

func resourceAwsSsmMaintenanceWindowTaskCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Registering SSM Maintenance Window Task")

	params := &ssm.RegisterTaskWithMaintenanceWindowInput{
		WindowId:       aws.String(d.Get("window_id").(string)),
		MaxConcurrency: aws.String(d.Get("max_concurrency").(string)),
		MaxErrors:      aws.String(d.Get("max_errors").(string)),
		TaskType:       aws.String(d.Get("task_type").(string)),
		ServiceRoleArn: aws.String(d.Get("service_role_arn").(string)),
		TaskArn:        aws.String(d.Get("task_arn").(string)),
		Targets:        expandAwsSsmTargets(d.Get("targets").([]interface{})),
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("priority"); ok {
		params.Priority = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("logging_info"); ok {
		params.LoggingInfo = expandAwsSsmMaintenanceWindowLoggingInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("task_parameters"); ok {
		params.TaskParameters = expandAwsSsmTaskParameters(v.([]interface{}))
	}

	resp, err := ssmconn.RegisterTaskWithMaintenanceWindow(params)
	if err != nil {
		return err
	}

	d.SetId(*resp.WindowTaskId)

	return resourceAwsSsmMaintenanceWindowTaskRead(d, meta)
}

func resourceAwsSsmMaintenanceWindowTaskRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	windowID := d.Get("window_id").(string)

	params := &ssm.GetMaintenanceWindowTaskInput{
		WindowId:     aws.String(windowID),
		WindowTaskId: aws.String(d.Id()),
	}
	resp, err := ssmconn.GetMaintenanceWindowTask(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		log.Printf("[WARN] Maintenance Window (%s) Task (%s) not found, removing from state", windowID, d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting Maintenance Window (%s) Task (%s): %s", windowID, d.Id(), err)
	}

	d.Set("window_id", resp.WindowId)
	d.Set("max_concurrency", resp.MaxConcurrency)
	d.Set("max_errors", resp.MaxErrors)
	d.Set("task_type", resp.TaskType)
	d.Set("service_role_arn", resp.ServiceRoleArn)
	d.Set("task_arn", resp.TaskArn)
	d.Set("priority", resp.Priority)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)

	if resp.LoggingInfo != nil {
		if err := d.Set("logging_info", flattenAwsSsmMaintenanceWindowLoggingInfo(resp.LoggingInfo)); err != nil {
			return fmt.Errorf("Error setting logging_info error: %#v", err)
		}
	}

	if resp.TaskParameters != nil {
		if err := d.Set("task_parameters", flattenAwsSsmTaskParameters(resp.TaskParameters)); err != nil {
			return fmt.Errorf("Error setting task_parameters error: %#v", err)
		}
	}

	if err := d.Set("targets", flattenAwsSsmTargets(resp.Targets)); err != nil {
		return fmt.Errorf("Error setting targets error: %#v", err)
	}

	return nil
}

func resourceAwsSsmMaintenanceWindowTaskDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deregistering SSM Maintenance Window Task: %s", d.Id())

	params := &ssm.DeregisterTaskFromMaintenanceWindowInput{
		WindowId:     aws.String(d.Get("window_id").(string)),
		WindowTaskId: aws.String(d.Id()),
	}

	_, err := ssmconn.DeregisterTaskFromMaintenanceWindow(params)
	if err != nil {
		return fmt.Errorf("error deregistering SSM Maintenance Window Task (%s): %s", d.Id(), err)
	}

	return nil
}
