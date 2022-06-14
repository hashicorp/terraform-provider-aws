package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
				ValidateFunc: validation.StringInSlice(ecs.SettingName_Values(), false),
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
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
		Service:   ecs.ServiceName,
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

	input := &ecs.ListAccountSettingsInput{
		Name:              aws.String(d.Get("name").(string)),
		EffectiveSettings: aws.Bool(true),
	}

	log.Printf("[DEBUG] Reading Default Account Settings: %s", input)
	resp, err := conn.ListAccountSettings(input)

	if err != nil {
		return err
	}

	if len(resp.Settings) == 0 {
		log.Printf("[WARN] Account Setting Default not set. Removing from state")
		d.SetId("")
		return nil
	}

	for _, r := range resp.Settings {
		d.SetId(aws.StringValue(r.PrincipalArn))
		d.Set("name", r.Name)
		d.Set("principal_arn", r.PrincipalArn)
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
			return fmt.Errorf("updating ECS Account Setting Default (%s): %w", settingName, err)
		}
	}

	return nil
}

func resourceAccountSettingDefaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	settingName := d.Get("name").(string)

	log.Printf("[WARN] Disabling ECS Account Setting Default %s", settingName)
	input := ecs.PutAccountSettingDefaultInput{
		Name:  aws.String(settingName),
		Value: aws.String("disabled"),
	}

	_, err := conn.PutAccountSettingDefault(&input)

	if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "You can no longer disable") {
		log.Printf("[DEBUG] ECS Account Setting Default (%q) could not be disabled: %s", settingName, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("disabling ECS Account Setting Default: %s", err)
	}

	log.Printf("[DEBUG] ECS Account Setting Default (%q) disabled", settingName)
	return nil
}
