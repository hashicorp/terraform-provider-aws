package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAppautoscalingScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingScheduledActionCreate,
		Read:   resourceAwsAppautoscalingScheduledActionRead,
		Delete: resourceAwsAppautoscalingScheduledActionDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					validServiceNamespaces := []string{"ecs", "elasticmapreduce", "ec2", "appstream", "dynamodb"}
					for _, str := range validServiceNamespaces {
						if value == str {
							return
						}
					}
					errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, validServiceNamespaces, value))
					return
				},
			},
			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_dimension": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					validDimensions := []string{"ecs:service:DesiredCount", "ec2:spot-fleet-request:TargetCapacity",
						"elasticmapreduce:instancegroup:InstanceCount", "appstream:fleet:DesiredCapacity",
						"dynamodb:table:ReadCapacityUnits", "dynamodb:table:WriteCapacityUnits",
						"dynamodb:index:ReadCapacityUnits", "dynamodb:index:WriteCapacityUnits"}
					for _, str := range validDimensions {
						if value == str {
							return
						}
					}
					errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, validDimensions, value))
					return
				},
			},
			"scalable_target_action": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"min_capacity": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"schedule": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"start_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"end_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsAppautoscalingScheduledActionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := applicationautoscaling.PutScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.PutScheduledAction(input)
	if err != nil {
		return err
	}
	return errors.New("error")
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := applicationautoscaling.DescribeScheduledActionsInput{}
	_, err := conn.DescribeScheduledActions(input)
	if err != nil {
		return err
	}
	return errors.New("error")
}

func resourceAwsAppautoscalingScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := applicationautoscaling.DeleteScheduledActionInput{}
	_, err := conn.DeleteScheduledAction(input)
	if err != nil {
		return err
	}
	return errors.New("error")
}
