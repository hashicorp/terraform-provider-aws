package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAccountSettingDefault() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountSettingDefaultCreate,
		Read:   resourceAccountSettingDefaultRead,
		Update: resourceAccountSettingDefaultUpdate,
		Delete: resourceAccountSettingDefaultDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAccountSettingDefaultImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"serviceLongArnFormat", "taskLongArnFormat", "containerInstanceLongArnFormat", "awsvpcTrunking", "containerInsights"}, false),
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAccountSettingDefaultImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   "ecs",
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAccountSettingDefaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	settingName := d.Get("name").(string)
	settingValue := d.Get("value").(string)
	log.Printf("[DEBUG] Setting Account Default %s", settingName)

	input := ecs.PutAccountSettingDefaultInput{
		Name:  aws.String(settingName),
		Value: aws.String(settingValue),
	}

	out, err := conn.PutAccountSettingDefault(&input)

	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Account Setting Default %s set", aws.StringValue(out.Setting.Value))

	d.SetId(aws.StringValue(out.Setting.Value))
	d.Set("principal_arn", out.Setting.PrincipalArn)

	return resourceAccountSettingDefaultRead(d, meta)
}

func resourceAccountSettingDefaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	accountSettingName := aws.String(d.Get("name").(string))

	input := &ecs.ListAccountSettingsInput{
		Name:              accountSettingName,
		EffectiveSettings: aws.Bool(true),
	}

	log.Printf("[DEBUG] Reading Default Account Settings: %s", input)
	resp, err := conn.ListAccountSettings(input)

	if err != nil {
		return err
	}

	if len(resp.Settings) == 0 {
		log.Printf("[WARN] Default Account Setting for #{accountSettingName} not set. Removing from state")
		d.SetId("")
		return nil
	}

	for _, r := range resp.Settings {
		d.SetId(aws.StringValue(r.PrincipalArn))
		d.Set("principal_arn", r.PrincipalArn)
		d.Set("name", r.Name)
		d.Set("value", r.Value)
	}

	return nil
}

func resourceAccountSettingDefaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	settingName := d.Get("name").(string)
	settingValue := d.Get("value").(string)

	if d.HasChange("value") {
		input := ecs.PutAccountSettingDefaultInput{
			Name:  aws.String(settingName),
			Value: aws.String(settingValue),
		}

		_, err := conn.PutAccountSettingDefault(&input)
		if err != nil {
			return fmt.Errorf("Error Updating Default Account settings (%s): %s", settingName, err)
		}
	}

	return nil
}

func resourceAccountSettingDefaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	settingName := d.Get("name").(string)

	log.Printf("[WARN] Disabling ECS Default Account Setting %s", settingName)
	input := ecs.PutAccountSettingDefaultInput{
		Name:  aws.String(settingName),
		Value: aws.String("disabled"),
	}

	_, err := conn.PutAccountSettingDefault(&input)

	if err != nil {
		return fmt.Errorf("Error disabling ECS Account Default setting: %s", err)
	}

	log.Printf("[DEBUG] ECS Account default setting %q disabled", settingName)
	return nil
}
