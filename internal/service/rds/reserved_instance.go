// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReservedInstance = "Reserved Instance"
)

// @SDKResource("aws_rds_reserved_instance", name="Reserved Instance")
// @Tags(identifierAttribute="arn")
func ResourceReservedInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReservedInstanceCreate,
		ReadWithoutTimeout:   resourceReservedInstanceRead,
		UpdateWithoutTimeout: resourceReservedInstanceUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"currency_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"fixed_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"lease_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"offering_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"offering_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recurring_charges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recurring_charge_amount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"recurring_charge_frequency": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"reservation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReservedInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.PurchaseReservedDBInstancesOfferingInput{
		ReservedDBInstancesOfferingId: aws.String(d.Get("offering_id").(string)),
		Tags:                          getTagsIn(ctx),
	}

	if v, ok := d.Get("instance_count").(int); ok && v > 0 {
		input.DBInstanceCount = aws.Int64(int64(d.Get("instance_count").(int)))
	}

	if v, ok := d.Get("reservation_id").(string); ok && v != "" {
		input.ReservedDBInstanceId = aws.String(v)
	}

	resp, err := conn.PurchaseReservedDBInstancesOfferingWithContext(ctx, input)
	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionCreating, ResNameReservedInstance, fmt.Sprintf("offering_id: %s, reservation_id: %s", d.Get("offering_id").(string), d.Get("reservation_id").(string)), err)
	}

	d.SetId(aws.ToString(resp.ReservedDBInstance.ReservedDBInstanceId))

	if err := waitReservedInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.RDS, create.ErrActionWaitingForCreation, ResNameReservedInstance, d.Id(), err)
	}

	return resourceReservedInstanceRead(ctx, d, meta)
}

func resourceReservedInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	reservation, err := FindReservedDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id(), err)
	}

	d.Set("arn", reservation.ReservedDBInstanceArn)
	d.Set("currency_code", reservation.CurrencyCode)
	d.Set("db_instance_class", reservation.DBInstanceClass)
	d.Set("duration", reservation.Duration)
	d.Set("fixed_price", reservation.FixedPrice)
	d.Set("instance_count", reservation.DBInstanceCount)
	d.Set("lease_id", reservation.LeaseId)
	d.Set("multi_az", reservation.MultiAZ)
	d.Set("offering_id", reservation.ReservedDBInstancesOfferingId)
	d.Set("offering_type", reservation.OfferingType)
	d.Set("product_description", reservation.ProductDescription)
	d.Set("recurring_charges", flattenRecurringCharges(reservation.RecurringCharges))
	d.Set("reservation_id", reservation.ReservedDBInstanceId)
	d.Set("start_time", (reservation.StartTime).Format(time.RFC3339))
	d.Set("state", reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	return nil
}

func resourceReservedInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceReservedInstanceRead(ctx, d, meta)
}

func flattenRecurringCharges(recurringCharges []*rds.RecurringCharge) []interface{} {
	if len(recurringCharges) == 0 {
		return []interface{}{}
	}

	var rawRecurringCharges []interface{}
	for _, recurringCharge := range recurringCharges {
		rawRecurringCharge := map[string]interface{}{
			"recurring_charge_amount":    recurringCharge.RecurringChargeAmount,
			"recurring_charge_frequency": aws.ToString(recurringCharge.RecurringChargeFrequency),
		}

		rawRecurringCharges = append(rawRecurringCharges, rawRecurringCharge)
	}

	return rawRecurringCharges
}
