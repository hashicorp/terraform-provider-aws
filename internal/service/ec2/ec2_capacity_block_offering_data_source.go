// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
	"time"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

const (
	ResNameCapacityBlockOffering = "Capacity Block Offering"
)

// @SDKDataSource("aws_ec2_capacity_block_offering")
func DataSourceCapacityBlockOffering() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCapacityBlockOfferingRead,
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_duration": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"currency_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"start_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upfront_fee": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCapacityBlockOfferingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeCapacityBlockOfferingsInput{
		CapacityDurationHours: aws.Int64(int64(d.Get("capacity_duration").(int))),
		InstanceCount:         aws.Int64(int64(d.Get("instance_count").(int))),
		InstanceType:          aws.String(d.Get("instance_type").(string)),
	}

	if v, ok := d.GetOk("start_date"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.StartDateRange = aws.Time(v)
	}

	if v, ok := d.GetOk("end_date"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.EndDateRange = aws.Time(v)
	}

	output, err := conn.DescribeCapacityBlockOfferingsWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameCapacityBlockOffering, "unknown", err)
	}

	if len(output.CapacityBlockOfferings) == 0 {
		return diag.Errorf("no %s %s found matching criteria; try different search", names.EC2, ResNameCapacityBlockOffering)
	}

	if len(output.CapacityBlockOfferings) > 1 {
		return diag.Errorf("More than one %s %s found matching criteria; try different search", names.EC2, ResNameCapacityBlockOffering)
	}

	if err != nil {

		return sdkdiag.AppendErrorf(diags, "creating EC2 Capacity Reservation: %s", err)
	}

	cbo := output.CapacityBlockOfferings[0]
	{
		d.SetId(aws_sdkv2.ToString(cbo.CapacityBlockOfferingId))
		d.Set("availability_zone", cbo.AvailabilityZone)
		d.Set("capacity_duration", cbo.CapacityBlockDurationHours)
		d.Set("currency_code", cbo.CurrencyCode)
		if cbo.EndDate != nil {
			d.Set("end_date", aws.TimeValue(cbo.EndDate).Format(time.RFC3339))
		} else {
			d.Set("end_date", nil)
		}
		d.Set("instance_count", cbo.InstanceCount)
		d.Set("instance_type", cbo.InstanceType)
		if cbo.StartDate != nil {
			d.Set("start_date", aws.TimeValue(cbo.StartDate).Format(time.RFC3339))
		} else {
			d.Set("start_date", nil)
		}
		d.Set("tenancy", cbo.Tenancy)
		d.Set("upfront_fee", cbo.UpfrontFee)
	}

	return nil
}
