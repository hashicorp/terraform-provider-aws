// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReservedInstanceOffering = "Reserved Instance Offering"
)

// @SDKDataSource("aws_rds_reserved_instance_offering")
func DataSourceReservedOffering() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReservedOfferingRead,
		Schema: map[string]*schema.Schema{
			"currency_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"fixed_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"offering_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"offering_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Partial Upfront",
					"All Upfront",
					"No Upfront",
				}, false),
			},
			"product_description": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceReservedOfferingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeReservedDBInstancesOfferingsInput{
		DBInstanceClass:    aws.String(d.Get("db_instance_class").(string)),
		Duration:           aws.String(fmt.Sprint(d.Get("duration").(int))),
		MultiAZ:            aws.Bool(d.Get("multi_az").(bool)),
		OfferingType:       aws.String(d.Get("offering_type").(string)),
		ProductDescription: aws.String(d.Get("product_description").(string)),
	}

	resp, err := conn.DescribeReservedDBInstancesOfferingsWithContext(ctx, input)
	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionReading, ResNameReservedInstanceOffering, "unknown", err)
	}

	if len(resp.ReservedDBInstancesOfferings) == 0 {
		return diag.Errorf("no %s %s found matching criteria; try different search", names.RDS, ResNameReservedInstanceOffering)
	}

	if len(resp.ReservedDBInstancesOfferings) > 1 {
		return diag.Errorf("More than one %s %s found matching criteria; try different search", names.RDS, ResNameReservedInstanceOffering)
	}

	offering := resp.ReservedDBInstancesOfferings[0]

	d.SetId(aws.ToString(offering.ReservedDBInstancesOfferingId))
	d.Set("currency_code", offering.CurrencyCode)
	d.Set("db_instance_class", offering.DBInstanceClass)
	d.Set("duration", offering.Duration)
	d.Set("fixed_price", offering.FixedPrice)
	d.Set("multi_az", offering.MultiAZ)
	d.Set("offering_type", offering.OfferingType)
	d.Set("product_description", offering.ProductDescription)
	d.Set("offering_id", offering.ReservedDBInstancesOfferingId)

	return nil
}
