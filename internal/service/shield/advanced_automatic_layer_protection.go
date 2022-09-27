package shield

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:     schema.TypeString,
				Required: true,
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

	if d.Get("action").(string) == "block" {
		action.Block = &shield.BlockAction{}
	}

	if d.Get("action").(string) == "count" {
		action.Count = &shield.CountAction{}
	}

	input := &shield.UpdateApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	_, err := conn.UpdateApplicationLayerAutomaticResponseRequest(input)
	if err != nil {
		return fmt.Errorf("error updating Application Layer Automatic Protection: %s", err)
	}

	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	action := &shield.ResponseAction{}

	if d.Get("action").(string) == "block" {
		action.Block = &shield.BlockAction{}
	}

	if d.Get("action").(string) == "count" {
		action.Count = &shield.CountAction{}
	}

	enableAutomaticResponseInput := &shield.EnableApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	describeProtectionInput := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	res, err := conn.DescribeProtection(describeProtectionInput)
	if err != nil {
		return fmt.Errorf("error reading Protection: %s", err)
	}

	_, err = conn.EnableApplicationLayerAutomaticResponse(enableAutomaticResponseInput)
	if err != nil {
		return fmt.Errorf("error creating Application Layer Automatic Protection: %s", err)
	}

	d.SetId(aws.StringValue(res.Protection.Id))
	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
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

	a := resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Block.String()
	b := resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Count.String()
	var action string

	if a != "" {
		action = a
	}
	if b != "" {
		action = b
	}

	arn := aws.StringValue(resp.Protection.ProtectionArn)
	d.Set("arn", arn)
	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)
	d.Set("action", action)

	return nil
}

func resourceAdvancedAutomaticLayerProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DisableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	output, _ := conn.DisableApplicationLayerAutomaticResponseRequest(input)

	if tfawserr.ErrCodeEquals(output.Error, shield.ErrCodeResourceNotFoundException) {
		return nil
	}

	if output.Error != nil {
		return fmt.Errorf("error deleting Application Layer Automatic Protection (%s): %s", d.Id(), output.Error)
	}
	return nil
}
