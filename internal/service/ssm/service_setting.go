package ssm

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameServiceSetting = "Service Setting"
)

func ResourceServiceSetting() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceSettingUpdate,
		Read:   resourceServiceSettingRead,
		Update: resourceServiceSettingUpdate,
		Delete: resourceServiceSettingReset,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"setting_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceSettingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] SSM service setting create: %s", d.Get("setting_id").(string))

	updateServiceSettingInput := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(d.Get("setting_id").(string)),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	if _, err := conn.UpdateServiceSetting(updateServiceSettingInput); err != nil {
		return names.Error(names.SSM, names.ErrActionUpdating, ResNameServiceSetting, d.Get("setting_id").(string), err)
	}

	d.SetId(d.Get("setting_id").(string))

	if _, err := waitServiceSettingUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return names.Error(names.SSM, names.ErrActionWaitingForUpdate, ResNameServiceSetting, d.Id(), err)
	}

	return resourceServiceSettingRead(d, meta)
}

func resourceServiceSettingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Reading SSM Activation: %s", d.Id())

	output, err := FindServiceSettingByARN(conn, d.Id())
	if err != nil {
		return names.Error(names.SSM, names.ErrActionReading, ResNameServiceSetting, d.Id(), err)
	}

	// AWS SSM service setting API requires the entire ARN as input,
	// but setting_id in the output is only a part of ARN.
	d.Set("setting_id", output.ARN)
	d.Set("setting_value", output.SettingValue)
	d.Set("arn", output.ARN)
	d.Set("status", output.Status)

	return nil
}

func resourceServiceSettingReset(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())

	resetServiceSettingInput := &ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Get("setting_id").(string)),
	}

	_, err := conn.ResetServiceSetting(resetServiceSettingInput)
	if err != nil {
		return names.Error(names.SSM, names.ErrActionDeleting, ResNameServiceSetting, d.Id(), err)
	}

	if err := waitServiceSettingReset(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return names.Error(names.SSM, names.ErrActionWaitingForDeletion, ResNameServiceSetting, d.Id(), err)
	}

	return nil
}
