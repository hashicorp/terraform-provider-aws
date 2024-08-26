// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_reserved_instance_offering", name="Reserved Instance Offering")
func dataSourceReservedOffering() *schema.Resource {
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
			names.AttrDuration: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeReservedDBInstancesOfferingsInput{
		DBInstanceClass:    aws.String(d.Get("db_instance_class").(string)),
		Duration:           aws.String(fmt.Sprint(d.Get(names.AttrDuration).(int))),
		MultiAZ:            aws.Bool(d.Get("multi_az").(bool)),
		OfferingType:       aws.String(d.Get("offering_type").(string)),
		ProductDescription: aws.String(d.Get("product_description").(string)),
	}

	offering, err := findReservedDBInstancesOffering(ctx, conn, input, tfslices.PredicateTrue[*types.ReservedDBInstancesOffering]())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("RDS Reserved Instance Offering", err))
	}

	offeringID := aws.ToString(offering.ReservedDBInstancesOfferingId)
	d.SetId(offeringID)
	d.Set("currency_code", offering.CurrencyCode)
	d.Set("db_instance_class", offering.DBInstanceClass)
	d.Set(names.AttrDuration, offering.Duration)
	d.Set("fixed_price", offering.FixedPrice)
	d.Set("multi_az", offering.MultiAZ)
	d.Set("offering_id", offeringID)
	d.Set("offering_type", offering.OfferingType)
	d.Set("product_description", offering.ProductDescription)

	return diags
}

func findReservedDBInstancesOffering(ctx context.Context, conn *rds.Client, input *rds.DescribeReservedDBInstancesOfferingsInput, filter tfslices.Predicate[*types.ReservedDBInstancesOffering]) (*types.ReservedDBInstancesOffering, error) {
	output, err := findReservedDBInstancesOfferings(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReservedDBInstancesOfferings(ctx context.Context, conn *rds.Client, input *rds.DescribeReservedDBInstancesOfferingsInput, filter tfslices.Predicate[*types.ReservedDBInstancesOffering]) ([]types.ReservedDBInstancesOffering, error) {
	var output []types.ReservedDBInstancesOffering

	pages := rds.NewDescribeReservedDBInstancesOfferingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ReservedDBInstancesOfferingNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReservedDBInstancesOfferings {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
