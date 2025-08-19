// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_reserved_instance", name="Reserved Instance")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceReservedInstance() *schema.Resource {
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
	}
}

func resourceReservedInstanceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.PurchaseReservedDBInstancesOfferingInput{
		ReservedDBInstancesOfferingId: aws.String(d.Get("offering_id").(string)),
		Tags:                          getTagsIn(ctx),
	}

	if v, ok := d.Get(names.AttrInstanceCount).(int); ok && v > 0 {
		input.DBInstanceCount = aws.Int32(int32(d.Get(names.AttrInstanceCount).(int)))
	}

	if v, ok := d.Get("reservation_id").(string); ok && v != "" {
		input.ReservedDBInstanceId = aws.String(v)
	}

	output, err := conn.PurchaseReservedDBInstancesOffering(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Reserved Instance: %s", err)
	}

	d.SetId(aws.ToString(output.ReservedDBInstance.ReservedDBInstanceId))

	if _, err := waitReservedInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Reserved Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReservedInstanceRead(ctx, d, meta)...)
}

func resourceReservedInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	reservation, err := findReservedDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Reserved Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Reserved Instance (%s): %s", d.Id(), err)
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
	d.Set(names.AttrStartTime, reservation.StartTime.Format(time.RFC3339))
	d.Set(names.AttrState, reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	return diags
}

func resourceReservedInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceReservedInstanceRead(ctx, d, meta)
}

func findReservedDBInstanceByID(ctx context.Context, conn *rds.Client, id string) (*types.ReservedDBInstance, error) {
	input := &rds.DescribeReservedDBInstancesInput{
		ReservedDBInstanceId: aws.String(id),
	}
	output, err := findReservedDBInstance(ctx, conn, input, tfslices.PredicateTrue[*types.ReservedDBInstance]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ReservedDBInstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findReservedDBInstance(ctx context.Context, conn *rds.Client, input *rds.DescribeReservedDBInstancesInput, filter tfslices.Predicate[*types.ReservedDBInstance]) (*types.ReservedDBInstance, error) {
	output, err := findReservedDBInstances(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReservedDBInstances(ctx context.Context, conn *rds.Client, input *rds.DescribeReservedDBInstancesInput, filter tfslices.Predicate[*types.ReservedDBInstance]) ([]types.ReservedDBInstance, error) {
	var output []types.ReservedDBInstance

	pages := rds.NewDescribeReservedDBInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ReservedDBInstanceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReservedDBInstances {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusReservedInstance(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findReservedDBInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func waitReservedInstanceCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.ReservedDBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{reservedInstanceStatePaymentPending},
		Target:         []string{reservedInstanceStateActive},
		Refresh:        statusReservedInstance(ctx, conn, id),
		NotFoundChecks: 5,
		Timeout:        timeout,
		MinTimeout:     10 * time.Second,
		Delay:          30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ReservedDBInstance); ok {
		return output, err
	}

	return nil, err
}

func flattenRecurringCharges(apiObjects []types.RecurringCharge) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"recurring_charge_amount":    aws.ToFloat64(apiObject.RecurringChargeAmount),
			"recurring_charge_frequency": aws.ToString(apiObject.RecurringChargeFrequency),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
