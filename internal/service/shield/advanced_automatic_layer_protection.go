package shield

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"log"
)

func ResourceAdvancedAutomaticLayerProtection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdvancedAutomaticLayerProtectionCreate,
		Update: resourceAdvancedAutomaticLayerProtectionUpdate,
		Read:   resourceAdvancedAutomaticLayerProtectionRead,
		Delete: resourceAdvancedAutomaticLayerProtectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"block", "count"}, false),
			},
		},
	}
}

func resourceAdvancedAutomaticLayerProtectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	if !d.HasChange("action") {
		return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
	}

	action := &shield.ResponseAction{}
	switch d.Get("action").(string) {
	case "block":
		action.Block = &shield.BlockAction{}
	case "count":
		action.Count = &shield.CountAction{}
	}

	input := &shield.UpdateApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Id()),
	}

	_, err := conn.UpdateApplicationLayerAutomaticResponse(input)
	if err != nil {
		return fmt.Errorf("error updating Application Layer Automatic Protection: %s", err)
	}

	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	action := &shield.ResponseAction{}
	switch d.Get("action").(string) {
	case "block":
		action.Block = &shield.BlockAction{}
	case "count":
		action.Count = &shield.CountAction{}
	}

	enableAutomaticResponseInput := &shield.EnableApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	_, err := conn.EnableApplicationLayerAutomaticResponse(enableAutomaticResponseInput)
	if err != nil {
		return fmt.Errorf("error creating Application Layer Automatic Protection: %s", err)
	}

	d.SetId(d.Get("resource_arn").(string))
	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtection(input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Application Layer Automatic Protection (%s): %s", d.Id(), err)
	}

	var action string
	if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action != nil {
		if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Block != nil {
			action = "block"
		}
		if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Count != nil {
			action = "count"
		}
	}

	d.Set("action", action)

	return nil
}

func resourceAdvancedAutomaticLayerProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	shieldConn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DisableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(d.Id()),
	}

	_, err := shieldConn.DisableApplicationLayerAutomaticResponse(input)
	if err != nil {
		return fmt.Errorf("error deleting Application Layer Automatic Protection (%s): %s", d.Id(), err)
	}

	return nil
}