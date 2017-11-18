package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
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

	input := &directconnect.AssociateConnectionWithLagInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		LagId:        aws.String(d.Get("lag_id").(string)),
	}
	resp, err := conn.AssociateConnectionWithLag(input)
	if err != nil {
		return err
	}

	d.SetId(*resp.ConnectionId)
	return nil
}

func resourceAwsDxConnectionAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	connectionId := d.Id()
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(connectionId),
	}

	resp, err := conn.DescribeConnections(input)
	if err != nil {
		return err
	}
	if len(resp.Connections) < 1 {
		d.SetId("")
		return nil
	}
	if len(resp.Connections) != 1 {
		return fmt.Errorf("Number of DX Connection (%s) isn't one, got %d", connectionId, len(resp.Connections))
	}
	if d.Id() != *resp.Connections[0].ConnectionId {
		return fmt.Errorf("DX Connection (%s) not found", connectionId)
	}
	if resp.Connections[0].LagId == nil {
		d.SetId("")
		return nil
	}
	if *resp.Connections[0].LagId != d.Get("lag_id").(string) {
		return fmt.Errorf("DX Connection (%s) is associated with unexpected LAG (%s)", connectionId, *resp.Connections[0].LagId)
	}
	return nil
}

func resourceAwsDxConnectionAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.DisassociateConnectionFromLagInput{
		ConnectionId: aws.String(d.Id()),
		LagId:        aws.String(d.Get("lag_id").(string)),
	}

	resp, err := conn.DisassociateConnectionFromLag(input)
	if err != nil {
		return err
	}
	if resp.LagId != nil {
		return fmt.Errorf("DX Connection (%s) is still associated with LAG", d.Id())
	}

	d.SetId("")
	return nil
}
