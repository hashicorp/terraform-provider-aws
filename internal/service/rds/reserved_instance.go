// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
// @Testing(tagsTest=false)
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
			names.AttrARN: {
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
			names.AttrDuration: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"fixed_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrInstanceCount: {
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
			names.AttrStartTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.PurchaseReservedDBInstancesOfferingInput{
		ReservedDBInstancesOfferingId: aws.String(d.Get("offering_id").(string)),
		Tags:                          getTagsIn(ctx),
	}

	if v, ok := d.Get(names.AttrInstanceCount).(int); ok && v > 0 {
		input.DBInstanceCount = aws.Int64(int64(d.Get(names.AttrInstanceCount).(int)))
	}

	if v, ok := d.Get("reservation_id").(string); ok && v != "" {
		input.ReservedDBInstanceId = aws.String(v)
	}

	resp, err := conn.PurchaseReservedDBInstancesOfferingWithContext(ctx, input)
	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionCreating, ResNameReservedInstance, fmt.Sprintf("offering_id: %s, reservation_id: %s", d.Get("offering_id").(string), d.Get("reservation_id").(string)), err)
	}

	d.SetId(aws.StringValue(resp.ReservedDBInstance.ReservedDBInstanceId))

	if err := waitReservedInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionWaitingForCreation, ResNameReservedInstance, d.Id(), err)
	}

	return append(diags, resourceReservedInstanceRead(ctx, d, meta)...)
}

func resourceReservedInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	reservation, err := FindReservedDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id(), err)
	}

	d.Set(names.AttrARN, reservation.ReservedDBInstanceArn)
	d.Set("currency_code", reservation.CurrencyCode)
	d.Set("db_instance_class", reservation.DBInstanceClass)
	d.Set(names.AttrDuration, reservation.Duration)
	d.Set("fixed_price", reservation.FixedPrice)
	d.Set(names.AttrInstanceCount, reservation.DBInstanceCount)
	d.Set("lease_id", reservation.LeaseId)
	d.Set("multi_az", reservation.MultiAZ)
	d.Set("offering_id", reservation.ReservedDBInstancesOfferingId)
	d.Set("offering_type", reservation.OfferingType)
	d.Set("product_description", reservation.ProductDescription)
	d.Set("recurring_charges", flattenRecurringCharges(reservation.RecurringCharges))
	d.Set("reservation_id", reservation.ReservedDBInstanceId)
	d.Set(names.AttrStartTime, (reservation.StartTime).Format(time.RFC3339))
	d.Set(names.AttrState, reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	return diags
}

func resourceReservedInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceReservedInstanceRead(ctx, d, meta)
}

func FindReservedDBInstanceByID(ctx context.Context, conn *rds.RDS, id string) (*rds.ReservedDBInstance, error) {
	input := &rds.DescribeReservedDBInstancesInput{
		ReservedDBInstanceId: aws.String(id),
	}

	output, err := conn.DescribeReservedDBInstancesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeReservedDBInstanceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ReservedDBInstances) == 0 || output.ReservedDBInstances[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ReservedDBInstances); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ReservedDBInstances[0], nil
}

func statusReservedInstance(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReservedDBInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitReservedInstanceCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ReservedInstanceStatePaymentPending,
		},
		Target:         []string{ReservedInstanceStateActive},
		Refresh:        statusReservedInstance(ctx, conn, id),
		NotFoundChecks: 5,
		Timeout:        timeout,
		MinTimeout:     10 * time.Second,
		Delay:          30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func flattenRecurringCharges(recurringCharges []*rds.RecurringCharge) []interface{} {
	if len(recurringCharges) == 0 {
		return []interface{}{}
	}

	var rawRecurringCharges []interface{}
	for _, recurringCharge := range recurringCharges {
		rawRecurringCharge := map[string]interface{}{
			"recurring_charge_amount":    recurringCharge.RecurringChargeAmount,
			"recurring_charge_frequency": aws.StringValue(recurringCharge.RecurringChargeFrequency),
		}

		rawRecurringCharges = append(rawRecurringCharges, rawRecurringCharge)
	}

	return rawRecurringCharges
}
