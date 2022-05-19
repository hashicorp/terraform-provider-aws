package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSEncryptionByDefaultCreate,
		Read:   resourceEBSEncryptionByDefaultRead,
		Update: resourceEBSEncryptionByDefaultUpdate,
		Delete: resourceEBSEncryptionByDefaultDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceEBSEncryptionByDefaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setEBSEncryptionByDefault(conn, enabled); err != nil {
		return fmt.Errorf("error creating EBS encryption by default (%t): %s", enabled, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return resourceEBSEncryptionByDefaultRead(d, meta)
}

func resourceEBSEncryptionByDefaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	resp, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return fmt.Errorf("error reading EBS encryption by default: %s", err)
	}

	d.Set("enabled", resp.EbsEncryptionByDefault)

	return nil
}

func resourceEBSEncryptionByDefaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setEBSEncryptionByDefault(conn, enabled); err != nil {
		return fmt.Errorf("error updating EBS encryption by default (%t): %s", enabled, err)
	}

	return resourceEBSEncryptionByDefaultRead(d, meta)
}

func resourceEBSEncryptionByDefaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Removing the resource disables default encryption.
	if err := setEBSEncryptionByDefault(conn, false); err != nil {
		return fmt.Errorf("error disabling EBS encryption by default: %s", err)
	}

	return nil
}

func setEBSEncryptionByDefault(conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableEbsEncryptionByDefault(&ec2.EnableEbsEncryptionByDefaultInput{})
	} else {
		_, err = conn.DisableEbsEncryptionByDefault(&ec2.DisableEbsEncryptionByDefaultInput{})
	}

	return err
}
