package ec2

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCPeeringConnectionOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCPeeringConnectionOptionsCreate,
		Read:   resourceVPCPeeringConnectionOptionsRead,
		Update: resourceVPCPeeringConnectionOptionsUpdate,
		Delete: resourceVPCPeeringConnectionOptionsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accepter":  vpcPeeringConnectionOptionsSchema,
			"requester": vpcPeeringConnectionOptionsSchema,
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCPeeringConnectionOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcPeeringConnectionID := d.Get("vpc_peering_connection_id").(string)
	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, vpcPeeringConnectionID)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection (%s): %w", vpcPeeringConnectionID, err)
	}

	d.SetId(vpcPeeringConnectionID)

	if err := modifyVPCPeeringConnectionOptions(conn, d, vpcPeeringConnection, false); err != nil {
		return err
	}

	return resourceVPCPeeringConnectionOptionsRead(d, meta)
}

func resourceVPCPeeringConnectionOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Peering Connection Options %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection Options (%s): %w", d.Id(), err)
	}

	d.Set("vpc_peering_connection_id", vpcPeeringConnection.VpcPeeringConnectionId)

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)}); err != nil {
			return fmt.Errorf("error setting accepter: %w", err)
		}
	} else {
		d.Set("accepter", nil)
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)}); err != nil {
			return fmt.Errorf("error setting requester: %w", err)
		}
	} else {
		d.Set("requester", nil)
	}

	return nil
}

func resourceVPCPeeringConnectionOptionsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection (%s): %w", d.Id(), err)
	}

	if err := modifyVPCPeeringConnectionOptions(conn, d, vpcPeeringConnection, false); err != nil {
		return err
	}

	return resourceVPCPeeringConnectionOptionsRead(d, meta)
}

func resourceVPCPeeringConnectionOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	// Don't do anything with the underlying VPC Peering Connection.
	return nil
}
