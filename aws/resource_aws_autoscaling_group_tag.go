package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tagresource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAutoscalingGroupTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAutoscalingGroupTagCreate,
		Read:   resourceAwsAutoscalingGroupTagRead,
		Update: resourceAwsAutoscalingGroupTagUpdate,
		Delete: resourceAwsAutoscalingGroupTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tag": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsAutoscalingGroupTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	identifier := d.Get("autoscaling_group_name").(string)
	tags := d.Get("tag").([]interface{})
	key := tags[0].(map[string]interface{})["key"].(string)

	if err := keyvaluetags.AutoscalingUpdateTags(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, nil, tags); err != nil {
		return fmt.Errorf("error creating AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	d.SetId(tagresource.SetResourceID(identifier, key))

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	identifier, key, err := tagresource.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	value, err := keyvaluetags.AutoscalingGetTag(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, key)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AutoScaling Group (%s) tag (%s), removing from state", identifier, key)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	d.Set("autoscaling_group_name", identifier)

	if err := d.Set("tag", []map[string]interface{}{{
		"key":                 key,
		"value":               value.Value,
		"propagate_at_launch": value.AdditionalBoolFields["PropagateAtLaunch"],
	}}); err != nil {
		return fmt.Errorf("error setting tag: %w", err)
	}

	return nil
}

func resourceAwsAutoscalingGroupTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	identifier, key, err := tagresource.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, nil, d.Get("tag")); err != nil {
		return fmt.Errorf("error updating AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	identifier, key, err := tagresource.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, identifier, tfautoscaling.TagResourceTypeAutoScalingGroup, d.Get("tag"), nil); err != nil {
		return fmt.Errorf("error deleting AutoScaling Group (%s) tag (%s): %w", identifier, key, err)
	}

	return nil
}
