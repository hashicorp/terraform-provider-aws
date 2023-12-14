// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_capacity_reservation", name="Capacity Reservation")
// @Tags(identifierAttribute="id")
func ResourceCapacityReservation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapacityReservationCreate,
		ReadWithoutTimeout:   resourceCapacityReservationRead,
		UpdateWithoutTimeout: resourceCapacityReservationUpdate,
		DeleteWithoutTimeout: resourceCapacityReservationDelete,
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
				Required: true,
				ForceNew: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"end_date_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.EndDateTypeUnlimited,
				ValidateFunc: validation.StringInSlice(ec2.EndDateType_Values(), false),
			},
			"ephemeral_storage": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"instance_match_criteria": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.InstanceMatchCriteriaOpen,
				ValidateFunc: validation.StringInSlice(ec2.InstanceMatchCriteria_Values(), false),
			},
			"instance_platform": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.CapacityReservationInstancePlatform_Values(), false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tenancy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.CapacityReservationTenancyDefault,
				ValidateFunc: validation.StringInSlice(ec2.CapacityReservationTenancy_Values(), false),
			},
		},
	}
}

func resourceCapacityReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateCapacityReservationInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		EndDateType:       aws.String(d.Get("end_date_type").(string)),
		InstanceCount:     aws.Int64(int64(d.Get("instance_count").(int))),
		InstancePlatform:  aws.String(d.Get("instance_platform").(string)),
		InstanceType:      aws.String(d.Get("instance_type").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeCapacityReservation),
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		input.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("end_date"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))

		input.EndDate = aws.Time(v)
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok {
		input.EphemeralStorage = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_match_criteria"); ok {
		input.InstanceMatchCriteria = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("placement_group_arn"); ok {
		input.PlacementGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		input.Tenancy = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Capacity Reservation: %s", input)
	output, err := conn.CreateCapacityReservationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Capacity Reservation: %s", err)
	}

	d.SetId(aws.StringValue(output.CapacityReservation.CapacityReservationId))

	if _, err := WaitCapacityReservationActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCapacityReservationRead(ctx, d, meta)...)
}

func resourceCapacityReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	reservation, err := FindCapacityReservationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Capacity Reservation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Capacity Reservation (%s): %s", d.Id(), err)
	}

	d.Set("arn", reservation.CapacityReservationArn)
	d.Set("availability_zone", reservation.AvailabilityZone)
	d.Set("ebs_optimized", reservation.EbsOptimized)
	if reservation.EndDate != nil {
		d.Set("end_date", aws.TimeValue(reservation.EndDate).Format(time.RFC3339))
	} else {
		d.Set("end_date", nil)
	}
	d.Set("end_date_type", reservation.EndDateType)
	d.Set("ephemeral_storage", reservation.EphemeralStorage)
	d.Set("instance_count", reservation.TotalInstanceCount)
	d.Set("instance_match_criteria", reservation.InstanceMatchCriteria)
	d.Set("instance_platform", reservation.InstancePlatform)
	d.Set("instance_type", reservation.InstanceType)
	d.Set("outpost_arn", reservation.OutpostArn)
	d.Set("owner_id", reservation.OwnerId)
	d.Set("placement_group_arn", reservation.PlacementGroupArn)
	d.Set("tenancy", reservation.Tenancy)

	setTagsOut(ctx, reservation.Tags)

	return diags
}

func resourceCapacityReservationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyCapacityReservationInput{
			CapacityReservationId: aws.String(d.Id()),
			EndDateType:           aws.String(d.Get("end_date_type").(string)),
			InstanceCount:         aws.Int64(int64(d.Get("instance_count").(int))),
		}

		if v, ok := d.GetOk("end_date"); ok {
			v, _ := time.Parse(time.RFC3339, v.(string))

			input.EndDate = aws.Time(v)
		}

		log.Printf("[DEBUG] Updating EC2 Capacity Reservation: %s", input)
		_, err := conn.ModifyCapacityReservationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Capacity Reservation (%s): %s", d.Id(), err)
		}

		if _, err := WaitCapacityReservationActive(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityReservationRead(ctx, d, meta)...)
}

func resourceCapacityReservationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Capacity Reservation: %s", d.Id())
	_, err := conn.CancelCapacityReservationWithContext(ctx, &ec2.CancelCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Capacity Reservation (%s): %s", d.Id(), err)
	}

	if _, err := WaitCapacityReservationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) delete: %s", d.Id(), err)
	}

	return diags
}
