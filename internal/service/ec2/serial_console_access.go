package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceSerialConsoleAccessCreate,
		Read:   resourceSerialConsoleAccessRead,
		Update: resourceSerialConsoleAccessUpdate,
		Delete: resourceSerialConsoleAccessDelete,
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

func resourceSerialConsoleAccessCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(conn, enabled); err != nil {
		return fmt.Errorf("error creating serial console access (%t): %s", enabled, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return resourceSerialConsoleAccessRead(d, meta)
}

func resourceSerialConsoleAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	resp, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})
	if err != nil {
		return fmt.Errorf("error reading serial console access: %s", err)
	}

	d.Set("enabled", resp.SerialConsoleAccessEnabled)

	return nil
}

func resourceSerialConsoleAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(conn, enabled); err != nil {
		return fmt.Errorf("error updating serial console access (%t): %s", enabled, err)
	}

	return resourceSerialConsoleAccessRead(d, meta)
}

func resourceSerialConsoleAccessDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Removing the resource disables default encryption.
	if err := setSerialConsoleAccess(conn, false); err != nil {
		return fmt.Errorf("error disabling serial console access: %s", err)
	}

	return nil
}

func setSerialConsoleAccess(conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableSerialConsoleAccess(&ec2.EnableSerialConsoleAccessInput{})
	} else {
		_, err = conn.DisableSerialConsoleAccess(&ec2.DisableSerialConsoleAccessInput{})
	}

	return err
}
