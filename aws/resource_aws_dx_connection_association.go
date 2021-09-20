package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directconnect/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDxConnectionAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxConnectionAssociationCreate,
		Read:   resourceAwsDxConnectionAssociationRead,
		Delete: resourceAwsDxConnectionAssociationDelete,

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

func resourceAwsDxConnectionAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

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

func resourceAwsDxConnectionAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	lagID := d.Get("lag_id").(string)
	err := finder.ConnectionAssociationExists(conn, d.Id(), lagID)

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

func resourceAwsDxConnectionAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	lagID := d.Get("lag_id").(string)
	input := &directconnect.DisassociateConnectionFromLagInput{
		ConnectionId: aws.String(d.Id()),
		LagId:        aws.String(lagID),
	}

	_, err := tfresource.RetryWhen(
		waiter.ConnectionDisassociatedTimeout,
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
		return fmt.Errorf("error deleting Direct Connect Connection (%s) LAG (%s) Association: %w", d.Id(), lagID, err)
	}

	return err
}
