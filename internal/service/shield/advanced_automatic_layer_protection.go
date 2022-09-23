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
				ForceNew: true,
			},
		},
	}
}


func resourceAdvancedAutomaticLayerProtectionUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("resourceAdvancedAutomaticLayerProtectionUpdate, not implemented...")
	conn := meta.(*conns.AWSClient).ShieldConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	action := &shield.ResponseAction{}
	switch d. {

	}

	input := &shield.EnableApplicationLayerAutomaticResponseInput{
		Action: &shield.ResponseAction{
			Block: &shield.BlockAction{},
			Count: &shield.CountAction{},
		},
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	resp, err := conn.EnableApplicationLayerAutomaticResponse(input)

	if err != nil {
		return fmt.Errorf("error creating Application Layer Automatic Protection: %s", err)
	}
	d.SetId(aws.StringValue(resp.))
	return resourceProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionRead(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("resourceAdvancedAutomaticLayerProtectionRead, not implemented...")

	conn := meta.(*conns.AWSClient).ShieldConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
		return fmt.Errorf("error reading Shield Protection (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Protection.ProtectionArn)
	d.Set("arn", arn)
	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Shield Protection (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAdvancedAutomaticLayerProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("resourceAdvancedAutomaticLayerProtectionDelete, not implemented...")

	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteProtection(input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Shield Protection (%s): %s", d.Id(), err)
	}
	return nil
}
