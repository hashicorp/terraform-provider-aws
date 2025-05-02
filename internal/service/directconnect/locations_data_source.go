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
)

// @SDKDataSource("aws_dx_locations", name="Locations")
func dataSourceLocations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocationsRead,

		Schema: map[string]*schema.Schema{
			"location_codes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceLocationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	input := &directconnect.DescribeLocationsInput{}
	locations, err := findLocations(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Location]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Locations: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("location_codes", tfslices.ApplyToAll(locations, func(v awstypes.Location) string {
		return aws.ToString(v.LocationCode)
	}))

	return diags
}
