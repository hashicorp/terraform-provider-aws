package ssm

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"setting_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceSettingUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] SSM service setting create: %s", d.Id())

	updateServiceSettingInput := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(d.Get("setting_id").(string)),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	if _, err := ssmconn.UpdateServiceSetting(updateServiceSettingInput); err != nil {
		return fmt.Errorf("Error updating SSM service setting: %s", err)
	}

	d.SetId(d.Get("setting_id").(string))

	return resourceServiceSettingRead(d, meta)
}

func resourceServiceSettingRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Reading SSM Activation: %s", d.Id())

	params := &ssm.GetServiceSettingInput{
		SettingId: aws.String(d.Id()),
	}

	resp, err := ssmconn.GetServiceSetting(params)

	if err != nil {
		return fmt.Errorf("Error reading SSM service setting: %s", err)
	}

	serviceSetting := resp.ServiceSetting
	// AWS SSM service setting API requires the entire ARN as input,
	// but setting_id in the output is only a part of ARN.
	d.Set("setting_id", serviceSetting.ARN)
	d.Set("setting_value", serviceSetting.SettingValue)
	d.Set("arn", serviceSetting.ARN)
	d.Set("last_modified_date", aws.TimeValue(serviceSetting.LastModifiedDate).Format(time.RFC3339))
	d.Set("last_modified_user", serviceSetting.LastModifiedUser)
	d.Set("status", serviceSetting.Status)

	return nil
}

func resourceServiceSettingReset(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())

	resetServiceSettingInput := &ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Get("setting_id").(string)),
	}

	_, err := ssmconn.ResetServiceSetting(resetServiceSettingInput)

	if err != nil {
		return fmt.Errorf("Error deleting SSM Service Setting: %s", err)
	}

	return nil
}
