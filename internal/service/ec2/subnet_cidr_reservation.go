package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubnetCIDRReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetCIDRReservationCreate,
		Read:   resourceSubnetCIDRReservationRead,
		Delete: resourceSubnetCIDRReservationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected SUBNET_ID:RESERVATION_ID", d.Id())
				}
				subnetId := idParts[0]
				reservationId := idParts[1]

				d.Set("subnet_id", subnetId)
				d.SetId(reservationId)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidCIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"reservation_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.SubnetCidrReservationType_Values(), false),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetCIDRReservationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	createOpts := &ec2.CreateSubnetCidrReservationInput{
		Cidr:            aws.String(d.Get("cidr_block").(string)),
		SubnetId:        aws.String(d.Get("subnet_id").(string)),
		ReservationType: aws.String(d.Get("reservation_type").(string)),
	}
	if description := d.Get("description").(string); description != "" {
		createOpts.Description = aws.String(description)
	}

	resp, err := conn.CreateSubnetCidrReservation(createOpts)

	if err != nil {
		return fmt.Errorf("error creating subnet CIDR reservation: %w", err)
	}

	reservation := resp.SubnetCidrReservation
	reservationId := aws.StringValue(reservation.SubnetCidrReservationId)
	d.SetId(reservationId)
	log.Printf("[INFO] Subnet Reservation ID: %s", reservationId)

	return resourceSubnetCIDRReservationRead(d, meta)
}

func resourceSubnetCIDRReservationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	reservation, err := FindSubnetCidrReservationById(conn, d.Id(), d.Get("subnet_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Subnet CIDR reservation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading EC2 subnet CIDR reservation (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] Setting cidr %s for id %s", aws.StringValue(reservation.Cidr), d.Id())
	d.Set("cidr_block", reservation.Cidr)
	d.Set("description", reservation.Description)
	d.Set("owner_id", reservation.OwnerId)
	d.Set("reservation_type", reservation.ReservationType)
	d.Set("subnet_id", reservation.SubnetId)

	return nil
}

func resourceSubnetCIDRReservationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	req := &ec2.DeleteSubnetCidrReservationInput{
		SubnetCidrReservationId: aws.String(d.Id()),
	}
	_, err := conn.DeleteSubnetCidrReservation(req)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetCidrReservationIDNotFound) {
		return nil
	}
	return err
}
