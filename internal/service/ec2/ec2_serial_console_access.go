package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSerialConsoleAccessCreate,
		ReadWithoutTimeout:   resourceSerialConsoleAccessRead,
		UpdateWithoutTimeout: resourceSerialConsoleAccessUpdate,
		DeleteWithoutTimeout: resourceSerialConsoleAccessDelete,

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

func resourceSerialConsoleAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return diag.Errorf("error setting EC2 Serial Console Access (%t): %s", enabled, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return resourceSerialConsoleAccessRead(ctx, d, meta)
}

func resourceSerialConsoleAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	output, err := conn.GetSerialConsoleAccessStatusWithContext(ctx, &ec2.GetSerialConsoleAccessStatusInput{})

	if err != nil {
		return diag.Errorf("error reading EC2 Serial Console Access: %s", err)
	}

	d.Set("enabled", output.SerialConsoleAccessEnabled)

	return nil
}

func resourceSerialConsoleAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	enabled := d.Get("enabled").(bool)
	if err := setSerialConsoleAccess(ctx, conn, enabled); err != nil {
		return diag.Errorf("error updating EC2 Serial Console Access (%t): %s", enabled, err)
	}

	return resourceSerialConsoleAccessRead(ctx, d, meta)
}

func resourceSerialConsoleAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Removing the resource disables serial console access.
	if err := setSerialConsoleAccess(ctx, conn, false); err != nil {
		return diag.Errorf("error disabling EC2 Serial Console Access: %s", err)
	}

	return nil
}

func setSerialConsoleAccess(ctx context.Context, conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableSerialConsoleAccessWithContext(ctx, &ec2.EnableSerialConsoleAccessInput{})
	} else {
		_, err = conn.DisableSerialConsoleAccessWithContext(ctx, &ec2.DisableSerialConsoleAccessInput{})
	}

	return err
}
