package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceIPAMResourceDiscovery() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIPAMResourceDiscoveryRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ipam_resource_discovery_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"operating_regions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_resource_discovery_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceIPAMResourceDiscoveryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeIpamResourceDiscoveriesInput{}

	if v, ok := d.GetOk("ipam_pool_id"); ok {
		input.IpamResourceDiscoveryIds = aws.StringSlice([]string{v.(string)})

	}

	filters, filtersOk := d.GetOk("filter")
	if filtersOk {
		input.Filters = BuildFiltersDataSource(filters.(*schema.Set))
	}

	output, err := conn.DescribeIpamResourceDiscoveries(input)
	var rd *ec2.IpamResourceDiscovery

	if err != nil {
		return err
	}

	if len(output.IpamResourceDiscoveries) == 0 || output.IpamResourceDiscoveries[0] == nil {
		return tfresource.SingularDataSourceFindError("EC2 VPC IPAM resource discovery", tfresource.NewEmptyResultError(input))
	}

	if len(output.IpamResourceDiscoveries) > 1 {
		return fmt.Errorf("multiple IPAM ResourceDiscoverys matched; use additional constraints to reduce matches to a single IPAM pool")
	}

	rd = output.IpamResourceDiscoveries[0]

	d.SetId(aws.StringValue(rd.IpamResourceDiscoveryId))

	d.Set("arn", rd.IpamResourceDiscoveryArn)

	d.Set("description", rd.Description)
	d.Set("is_default", rd.IsDefault)
	d.Set("owner_id", rd.OwnerId)
	d.Set("ipam_resource_discovery_region", rd.IpamResourceDiscoveryRegion)

	if err := d.Set("tags", KeyValueTags(rd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
