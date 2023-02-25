package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceLocations() *schema.Resource {
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

func dataSourceLocationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	locations, err := FindLocations(ctx, conn, &directconnect.DescribeLocationsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect locations: %s", err)
	}

	var locationCodes []*string

	for _, location := range locations {
		locationCodes = append(locationCodes, location.LocationCode)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("location_codes", aws.StringValueSlice(locationCodes))

	return diags
}
