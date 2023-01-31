package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceLocation() *schema.Resource {
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

func dataSourceLocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()
	locationCode := d.Get("location_code").(string)

	location, err := FindLocationByCode(ctx, conn, locationCode)

	if tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "no Direct Connect location matched; change the search criteria and try again")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect location (%s): %s", locationCode, err)
	}

	d.SetId(locationCode)
	d.Set("available_macsec_port_speeds", aws.StringValueSlice(location.AvailableMacSecPortSpeeds))
	d.Set("available_port_speeds", aws.StringValueSlice(location.AvailablePortSpeeds))
	d.Set("available_providers", aws.StringValueSlice(location.AvailableProviders))
	d.Set("location_code", location.LocationCode)
	d.Set("location_name", location.LocationName)

	return diags
}
