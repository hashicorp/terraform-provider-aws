// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_dx_location", name="Location")
func dataSourceLocation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocationRead,

		Schema: map[string]*schema.Schema{
			"available_macsec_port_speeds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"available_port_speeds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"available_providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"location_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLocationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	input := &directconnect.DescribeLocationsInput{}
	locationCode := d.Get("location_code").(string)
	location, err := findLocation(ctx, conn, input, func(v *awstypes.Location) bool {
		return aws.ToString(v.LocationCode) == locationCode
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Direct Connect Location", err))
	}

	d.SetId(locationCode)
	d.Set("available_macsec_port_speeds", location.AvailableMacSecPortSpeeds)
	d.Set("available_port_speeds", location.AvailablePortSpeeds)
	d.Set("available_providers", location.AvailableProviders)
	d.Set("location_code", location.LocationCode)
	d.Set("location_name", location.LocationName)

	return diags
}

func findLocation(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLocationsInput, filter tfslices.Predicate[*awstypes.Location]) (*awstypes.Location, error) {
	output, err := findLocations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocations(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLocationsInput, filter tfslices.Predicate[*awstypes.Location]) ([]awstypes.Location, error) {
	output, err := conn.DescribeLocations(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.Locations, tfslices.PredicateValue(filter)), nil
}
