// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_subnet_cidr_reservation", name="Subnet CIDR Reservation")
func resourceSubnetCIDRReservation() *schema.Resource {
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

				d.Set(names.AttrSubnetID, subnetID)
				d.SetId(reservationID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrCIDRBlock: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidCIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reservation_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SubnetCidrReservationType](),
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSubnetCIDRReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateSubnetCidrReservationInput{
		Cidr:            aws.String(d.Get(names.AttrCIDRBlock).(string)),
		ReservationType: awstypes.SubnetCidrReservationType(d.Get("reservation_type").(string)),
		SubnetId:        aws.String(d.Get(names.AttrSubnetID).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Subnet CIDR Reservation: %s", aws.ToString(input.SubnetId))
	output, err := conn.CreateSubnetCidrReservation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Subnet CIDR Reservation: %s", err)
	}

	d.SetId(aws.ToString(output.SubnetCidrReservation.SubnetCidrReservationId))

	return append(diags, resourceSubnetCIDRReservationRead(ctx, d, meta)...)
}

func resourceSubnetCIDRReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := findSubnetCIDRReservationBySubnetIDAndReservationID(ctx, conn, d.Get(names.AttrSubnetID).(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Subnet CIDR Reservation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnet CIDR Reservation (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCIDRBlock, output.Cidr)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrOwnerID, output.OwnerId)
	d.Set("reservation_type", output.ReservationType)
	d.Set(names.AttrSubnetID, output.SubnetId)

	return diags
}

func resourceSubnetCIDRReservationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Subnet CIDR Reservation: %s", d.Id())
	_, err := conn.DeleteSubnetCidrReservation(ctx, &ec2.DeleteSubnetCidrReservationInput{
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
