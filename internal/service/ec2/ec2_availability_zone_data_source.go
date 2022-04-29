package ec2

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceAvailabilityZone() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAvailabilityZoneRead,

		Schema: map[string]*schema.Schema{
			"all_availability_zones": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"filter": CustomFiltersSchema(),
			"group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_border_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAvailabilityZoneRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeAvailabilityZonesInput{}

	if v, ok := d.GetOk("all_availability_zones"); ok {
		input.AllAvailabilityZones = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("zone_id"); ok {
		input.ZoneIds = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("name"); ok {
		input.ZoneNames = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = BuildAttributeFilterList(
		map[string]string{
			"state": d.Get("state").(string),
		},
	)

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	az, err := FindAvailabilityZone(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Availability Zone", err)
	}

	// As a convenience when working with AZs generically, we expose
	// the AZ suffix alone, without the region name.
	// This can be used e.g. to create lookup tables by AZ letter that
	// work regardless of region.
	nameSuffix := aws.StringValue(az.ZoneName)[len(aws.StringValue(az.RegionName)):]
	// For Local and Wavelength zones, remove any leading "-".
	nameSuffix = strings.TrimLeft(nameSuffix, "-")

	d.SetId(aws.StringValue(az.ZoneName))
	d.Set("group_name", az.GroupName)
	d.Set("name", az.ZoneName)
	d.Set("name_suffix", nameSuffix)
	d.Set("network_border_group", az.NetworkBorderGroup)
	d.Set("opt_in_status", az.OptInStatus)
	d.Set("parent_zone_id", az.ParentZoneId)
	d.Set("parent_zone_name", az.ParentZoneName)
	d.Set("region", az.RegionName)
	d.Set("state", az.State)
	d.Set("zone_id", az.ZoneId)
	d.Set("zone_type", az.ZoneType)

	return nil
}
