// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_capacity_block_reservation", name="Capacity Block Reservation")
// @Tags(identifierAttribute="id")
func ResourceCapacityBlockReservation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapacityBlockReservationCreate,
		ReadWithoutTimeout:   resourceCapacityReservationRead,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_block_offering_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"end_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"end_date_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ephemeral_storage": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"instance_match_criteria": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_platform": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.CapacityReservationInstancePlatform_Values(), false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCapacityBlockReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.PurchaseCapacityBlockInput{
		CapacityBlockOfferingId: aws.String(d.Get("capacity_block_offering_id").(string)),
		InstancePlatform:        aws.String(d.Get("instance_platform").(string)),
		TagSpecifications:       getTagSpecificationsIn(ctx, ec2.ResourceTypeCapacityReservation),
	}

	output, err := conn.PurchaseCapacityBlock(input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Capacity Reservation: %s", err)
	}
	d.SetId(aws.StringValue(output.CapacityReservation.CapacityReservationId))

	if _, err := WaitCapacityReservationActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCapacityReservationRead(ctx, d, meta)...)
}
