package ec2

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceIPAMPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIPAMPoolRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allocation_default_netmask_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allocation_max_netmask_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allocation_min_netmask_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allocation_resource_tags": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_import": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aws_service": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_depth": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"publicly_advertisable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"source_ipam_pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceIPAMPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeIpamPoolsInput{}

	if v, ok := d.GetOk("ipam_pool_id"); ok {
		input.IpamPoolIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	pool, err := FindIPAMPool(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("IPAM Pool", err)
	}

	d.SetId(aws.StringValue(pool.IpamPoolId))
	d.Set("address_family", pool.AddressFamily)
	d.Set("allocation_default_netmask_length", pool.AllocationDefaultNetmaskLength)
	d.Set("allocation_max_netmask_length", pool.AllocationMaxNetmaskLength)
	d.Set("allocation_min_netmask_length", pool.AllocationMinNetmaskLength)
	d.Set("allocation_resource_tags", KeyValueTags(tagsFromIPAMAllocationTags(pool.AllocationResourceTags)).Map())
	d.Set("arn", pool.IpamPoolArn)
	d.Set("auto_import", pool.AutoImport)
	d.Set("aws_service", pool.AwsService)
	d.Set("description", pool.Description)
	scopeID := strings.Split(aws.StringValue(pool.IpamScopeArn), "/")[1]
	d.Set("ipam_scope_id", scopeID)
	d.Set("ipam_scope_type", pool.IpamScopeType)
	d.Set("locale", pool.Locale)
	d.Set("pool_depth", pool.PoolDepth)
	d.Set("publicly_advertisable", pool.PubliclyAdvertisable)
	d.Set("source_ipam_pool_id", pool.SourceIpamPoolId)
	d.Set("state", pool.State)

	if err := d.Set("tags", KeyValueTags(pool.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
