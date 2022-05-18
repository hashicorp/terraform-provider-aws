package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
				parts := strings.Split(d.Id(), ":")
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected SUBNET_ID:RESERVATION_ID", d.Id())
				}
				subnetID := parts[0]
				reservationID := parts[1]

				d.Set("subnet_id", subnetID)
				d.SetId(reservationID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reservation_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.SubnetCidrReservationType_Values(), false),
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSubnetCIDRReservationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.CreateSubnetCidrReservationInput{
		Cidr:            aws.String(d.Get("cidr_block").(string)),
		ReservationType: aws.String(d.Get("reservation_type").(string)),
		SubnetId:        aws.String(d.Get("subnet_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Subnet CIDR Reservation: %s", input)
	output, err := conn.CreateSubnetCidrReservation(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Subnet CIDR Reservation: %w", err)
	}

	d.SetId(aws.StringValue(output.SubnetCidrReservation.SubnetCidrReservationId))

	return resourceSubnetCIDRReservationRead(d, meta)
}

func resourceSubnetCIDRReservationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	output, err := FindSubnetCIDRReservationBySubnetIDAndReservationID(conn, d.Get("subnet_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Subnet CIDR Reservation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Subnet CIDR Reservation (%s): %w", d.Id(), err)
	}

	d.Set("cidr_block", output.Cidr)
	d.Set("description", output.Description)
	d.Set("owner_id", output.OwnerId)
	d.Set("reservation_type", output.ReservationType)
	d.Set("subnet_id", output.SubnetId)

	return nil
}

func resourceSubnetCIDRReservationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Subnet CIDR Reservation: %s", d.Id())
	_, err := conn.DeleteSubnetCidrReservation(&ec2.DeleteSubnetCidrReservationInput{
		SubnetCidrReservationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetCIDRReservationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Subnet CIDR Reservation (%s): %w", d.Id(), err)
	}

	return err
}
