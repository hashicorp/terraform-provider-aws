package aws

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAppautoscalingScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingScheduledActionCreate,
		Read:   resourceAwsAppautoscalingScheduledActionRead,
		Delete: resourceAwsAppautoscalingScheduledActionDelete,

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsAppautoscalingScheduledActionCreate(d *schema.ResourceData, meta interface{}) error {
	return errors.New("error")
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	return errors.New("error")
}

func resourceAwsAppautoscalingScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	return errors.New("error")
}
