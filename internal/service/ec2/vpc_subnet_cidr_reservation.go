// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ec2_subnet_cidr_reservation")
func ResourceSubnetCIDRReservation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetCIDRReservationCreate,
		ReadWithoutTimeout:   resourceSubnetCIDRReservationRead,
		DeleteWithoutTimeout: resourceSubnetCIDRReservationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceSubnetCIDRReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateSubnetCidrReservationInput{
		Cidr:            aws.String(d.Get("cidr_block").(string)),
		ReservationType: aws.String(d.Get("reservation_type").(string)),
		SubnetId:        aws.String(d.Get("subnet_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Subnet CIDR Reservation: %s", input)
	output, err := conn.CreateSubnetCidrReservationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Subnet CIDR Reservation: %s", err)
	}

	d.SetId(aws.StringValue(output.SubnetCidrReservation.SubnetCidrReservationId))

	return append(diags, resourceSubnetCIDRReservationRead(ctx, d, meta)...)
}

func resourceSubnetCIDRReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	output, err := FindSubnetCIDRReservationBySubnetIDAndReservationID(ctx, conn, d.Get("subnet_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Subnet CIDR Reservation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnet CIDR Reservation (%s): %s", d.Id(), err)
	}

	d.Set("cidr_block", output.Cidr)
	d.Set("description", output.Description)
	d.Set("owner_id", output.OwnerId)
	d.Set("reservation_type", output.ReservationType)
	d.Set("subnet_id", output.SubnetId)

	return diags
}

func resourceSubnetCIDRReservationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EC2 Subnet CIDR Reservation: %s", d.Id())
	_, err := conn.DeleteSubnetCidrReservationWithContext(ctx, &ec2.DeleteSubnetCidrReservationInput{
		SubnetCidrReservationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetCIDRReservationIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Subnet CIDR Reservation (%s): %s", d.Id(), err)
	}

	return diags
}
