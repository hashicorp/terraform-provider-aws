package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSfnStateMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSfnStateMachineCreate,
		Read:   resourceAwsSfnStateMachineRead,
		Update: resourceAwsSfnStateMachineUpdate,
		Delete: resourceAwsSfnStateMachineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"definition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024*1024), // 1048576
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSfnStateMachineName,
			},

			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  sfn.StateMachineTypeStandard,
				ValidateFunc: validation.StringInSlice([]string{
					sfn.StateMachineTypeStandard,
					sfn.StateMachineTypeExpress,
				}, false),
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destinations": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_log_group_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"include_execution_data": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"level": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								sfn.LogLevelAll,
								sfn.LogLevelError,
								sfn.LogLevelFatal,
								sfn.LogLevelOff,
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsSfnStateMachineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Print("[DEBUG] Creating Step Function State Machine")

	params := &sfn.CreateStateMachineInput{
		Definition: aws.String(d.Get("definition").(string)),
		Name:       aws.String(d.Get("name").(string)),
		RoleArn:    aws.String(d.Get("role_arn").(string)),
		Type:       aws.String(d.Get("type").(string)),
		Tags:       keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().SfnTags(),
	}

	if v, ok := d.GetOk("logging_configuration"); ok {
		params.LoggingConfiguration = expandSfnLoggingConfig(v.([]interface{}))
	}

	var stateMachine *sfn.CreateStateMachineOutput

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		stateMachine, err = conn.CreateStateMachine(params)

		if err != nil {
			// Note: the instance may be in a deleting mode, hence the retry
			// when creating the step function. This can happen when we are
			// updating the resource (since there is no update API call).
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == sfn.ErrCodeStateMachineDeleting {
				return resource.RetryableError(err)
			}
			//This is done to deal with IAM eventual consistency
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "AccessDeniedException" {
				return resource.RetryableError(err)
			}
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == sfn.ErrCodeInvalidLoggingConfiguration {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		stateMachine, err = conn.CreateStateMachine(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating Step Function State Machine: %s", err)
	}

	d.SetId(*stateMachine.StateMachineArn)

	return resourceAwsSfnStateMachineRead(d, meta)
}

func resourceAwsSfnStateMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Printf("[DEBUG] Reading Step Function State Machine: %s", d.Id())

	sm, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	})
	if err != nil {

		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() == sfn.ErrCodeStateMachineDoesNotExist {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("definition", sm.Definition)
	d.Set("name", sm.Name)
	d.Set("role_arn", sm.RoleArn)
	d.Set("status", sm.Status)
	d.Set("type", sm.Type)

	loggingConfig := flattenSfnLoggingConfig(sm.LoggingConfiguration)
	if err := d.Set("logging_configuration", loggingConfig); err != nil {
		log.Printf("[DEBUG] Error setting logging_configuration: %s", err)
	}

	if err := d.Set("creation_date", sm.CreationDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting creation_date: %s", err)
	}

	tags, err := keyvaluetags.SfnListTags(conn, d.Id())

	if err != nil && !isAWSErr(err, "UnknownOperationException", "") {
		return fmt.Errorf("error listing tags for SFN State Machine (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSfnStateMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn

	params := &sfn.UpdateStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
		Definition:      aws.String(d.Get("definition").(string)),
		RoleArn:         aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("logging_configuration"); ok {
		params.LoggingConfiguration = expandSfnLoggingConfig(v.([]interface{}))
	}

	_, err := conn.UpdateStateMachine(params)

	log.Printf("[DEBUG] Updating Step Function State Machine: %#v", params)

	if err != nil {
		if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "State Machine Does Not Exist") {
			return fmt.Errorf("Error updating Step Function State Machine: %s", err)
		}
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SfnUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSfnStateMachineRead(d, meta)
}

func resourceAwsSfnStateMachineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Printf("[DEBUG] Deleting Step Function State Machine: %s", d.Id())

	input := &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteStateMachine(input)

		if err == nil {
			return nil
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteStateMachine(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting SFN state machine: %s", err)
	}
	return nil
}

func expandSfnLoggingConfig(config []interface{}) *sfn.LoggingConfiguration {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	m := config[0].(map[string]interface{})
	loggingConfiguration := &sfn.LoggingConfiguration{
		Destinations:         expandSfnDestinations(m["destinations"].([]interface{})),
		IncludeExecutionData: aws.Bool(m["include_execution_data"].(bool)),
		Level:                aws.String(m["level"].(string)),
	}

	return loggingConfiguration
}

func flattenSfnLoggingConfig(loggingConfiguration *sfn.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"destinations":           flattenSfnDestinations(loggingConfiguration.Destinations[0]),
		"include_execution_data": aws.BoolValue(loggingConfiguration.IncludeExecutionData),
		"level":                  aws.StringValue(loggingConfiguration.Level),
	}

	return []interface{}{m}
}

func expandSfnDestinations(l []interface{}) []*sfn.LogDestination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	if m["cloudwatch_log_group_arn"] == nil {
		return nil
	}

	logGroup := &sfn.CloudWatchLogsLogGroup{
		LogGroupArn: aws.String(m["cloudwatch_log_group_arn"].(string)),
	}

	dest := &sfn.LogDestination{
		CloudWatchLogsLogGroup: logGroup,
	}

	logDestinations := []*sfn.LogDestination{dest}

	return logDestinations
}

func flattenSfnDestinations(destinations *sfn.LogDestination) []interface{} {
	if destinations == nil {
		return []interface{}{}
	}

	logGroup := destinations.CloudWatchLogsLogGroup.LogGroupArn
	m := map[string]interface{}{
		"cloudwatch_log_group_arn": aws.StringValue(logGroup),
	}

	return []interface{}{m}
}
