// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_capacity_reservation", name="Capacity Reservation")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceCapacityReservation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapacityReservationCreate,
		ReadWithoutTimeout:   resourceCapacityReservationRead,
		UpdateWithoutTimeout: resourceCapacityReservationUpdate,
		DeleteWithoutTimeout: resourceCapacityReservationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.EndDateTypeUnlimited,
				ValidateDiagFunc: enum.Validate[awstypes.EndDateType](),
			},
			"ephemeral_storage": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			names.AttrInstanceCount: {
				Type:     schema.TypeInt,
				Required: true,
			},
			"instance_match_criteria": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.InstanceMatchCriteriaOpen,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceMatchCriteria](),
			},
			"instance_platform": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationInstancePlatform](),
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrOwnerID: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.CapacityReservationTenancyDefault,
				ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationTenancy](),
			},
		},
	}
}

func resourceCapacityReservationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateCapacityReservationInput{
		AvailabilityZone:  aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		ClientToken:       aws.String(id.UniqueId()),
		EndDateType:       awstypes.EndDateType(d.Get("end_date_type").(string)),
		InstanceCount:     aws.Int32(int32(d.Get(names.AttrInstanceCount).(int))),
		InstancePlatform:  awstypes.CapacityReservationInstancePlatform(d.Get("instance_platform").(string)),
		InstanceType:      aws.String(d.Get(names.AttrInstanceType).(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeCapacityReservation),
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
		input.InstanceMatchCriteria = awstypes.InstanceMatchCriteria(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("placement_group_arn"); ok {
		input.PlacementGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		input.Tenancy = awstypes.CapacityReservationTenancy(v.(string))
	}

	output, err := conn.CreateCapacityReservation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Capacity Reservation: %s", err)
	}

	d.SetId(aws.ToString(output.CapacityReservation.CapacityReservationId))

	if _, err := waitCapacityReservationActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCapacityReservationRead(ctx, d, meta)...)
}

func resourceCapacityReservationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	reservation, err := findCapacityReservationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Capacity Reservation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Capacity Reservation (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, reservation.CapacityReservationArn)
	d.Set(names.AttrAvailabilityZone, reservation.AvailabilityZone)
	d.Set("ebs_optimized", reservation.EbsOptimized)
	if reservation.EndDate != nil {
		d.Set("end_date", aws.ToTime(reservation.EndDate).Format(time.RFC3339))
	} else {
		d.Set("end_date", nil)
	}
	d.Set("end_date_type", reservation.EndDateType)
	d.Set("ephemeral_storage", reservation.EphemeralStorage)
	d.Set(names.AttrInstanceCount, reservation.TotalInstanceCount)
	d.Set("instance_match_criteria", reservation.InstanceMatchCriteria)
	d.Set("instance_platform", reservation.InstancePlatform)
	d.Set(names.AttrInstanceType, reservation.InstanceType)
	d.Set("outpost_arn", reservation.OutpostArn)
	d.Set(names.AttrOwnerID, reservation.OwnerId)
	d.Set("placement_group_arn", reservation.PlacementGroupArn)
	d.Set("tenancy", reservation.Tenancy)

	setTagsOut(ctx, reservation.Tags)

	return diags
}

func resourceCapacityReservationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyCapacityReservationInput{
			CapacityReservationId: aws.String(d.Id()),
			EndDateType:           awstypes.EndDateType(d.Get("end_date_type").(string)),
			InstanceCount:         aws.Int32(int32(d.Get(names.AttrInstanceCount).(int))),
		}

		if v, ok := d.GetOk("end_date"); ok {
			v, _ := time.Parse(time.RFC3339, v.(string))

			input.EndDate = aws.Time(v)
		}

		_, err := conn.ModifyCapacityReservation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Capacity Reservation (%s): %s", d.Id(), err)
		}

		if _, err := waitCapacityReservationActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapacityReservationRead(ctx, d, meta)...)
}

func resourceCapacityReservationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Capacity Reservation: %s", d.Id())
	_, err := conn.CancelCapacityReservation(ctx, &ec2.CancelCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Capacity Reservation (%s): %s", d.Id(), err)
	}

	if _, err := waitCapacityReservationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Capacity Reservation (%s) delete: %s", d.Id(), err)
	}

	return diags
}
