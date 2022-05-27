package directconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceConnectionAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectionAssociationCreate,
		Read:   resourceConnectionAssociationRead,
		Delete: resourceConnectionAssociationDelete,

		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectionAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	connectionID := d.Get("connection_id").(string)
	lagID := d.Get("lag_id").(string)
	input := &directconnect.AssociateConnectionWithLagInput{
		ConnectionId: aws.String(connectionID),
		LagId:        aws.String(lagID),
	}

	log.Printf("[DEBUG] Creating Direct Connect Connection LAG Association: %s", input)
	output, err := conn.AssociateConnectionWithLag(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Connection (%s) LAG (%s) Association: %w", connectionID, lagID, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	return nil
}

func resourceConnectionAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	lagID := d.Get("lag_id").(string)
	err := FindConnectionAssociationExists(conn, d.Id(), lagID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) LAG (%s) Association not found, removing from state", d.Id(), lagID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connection (%s) LAG (%s) Association: %w", d.Id(), lagID, err)
	}

	return nil
}

func resourceConnectionAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	return deleteConnectionLAGAssociation(conn, d.Id(), d.Get("lag_id").(string))
}

func deleteConnectionLAGAssociation(conn *directconnect.DirectConnect, connectionID, lagID string) error {
	input := &directconnect.DisassociateConnectionFromLagInput{
		ConnectionId: aws.String(connectionID),
		LagId:        aws.String(lagID),
	}

	_, err := tfresource.RetryWhen(
		connectionDisassociatedTimeout,
		func() (interface{}, error) {
			return conn.DisassociateConnectionFromLag(input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Connection does not exist") ||
				tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Lag does not exist") {
				return false, nil
			}

			if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "is in a transitioning state") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Connection (%s) LAG (%s) Association: %w", connectionID, lagID, err)
	}

	return err
}
