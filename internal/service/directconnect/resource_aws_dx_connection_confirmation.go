package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
)

func ResourceConnectionConfirmation() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectionConfirmationCreate,
		Read:   resourceConnectionConfirmationRead,
		Delete: resourceConnectionConfirmationDelete,

		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectionConfirmationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	connectionID := d.Get("connection_id").(string)
	input := &directconnect.ConfirmConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	log.Printf("[DEBUG] Confirming Direct Connect Connection: %s", input)
	_, err := conn.ConfirmConnection(input)

	if err != nil {
		return fmt.Errorf("error confirming Direct Connection Connection (%s): %w", connectionID, err)
	}

	d.SetId(connectionID)

	if _, err := tfdirectconnect.waitConnectionConfirmed(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Direct Connection Connection (%s) confirm: %w", d.Id(), err)
	}

	return nil
}

func resourceConnectionConfirmationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	_, err := tfdirectconnect.FindConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connection (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceConnectionConfirmationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect connection. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
