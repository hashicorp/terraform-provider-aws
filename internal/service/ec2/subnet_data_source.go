package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceSubnet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"available_ip_address_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_for_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"enable_dns64": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_resource_name_dns_aaaa_record_on_launch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_resource_name_dns_a_record_on_launch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_block_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_native": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"map_customer_owned_ip_on_launch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"map_public_ip_on_launch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_hostname_type_on_launch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeSubnetsInput{}

	if id, ok := d.GetOk("id"); ok {
		input.SubnetIds = []*string{aws.String(id.(string))}
	}

	// We specify default_for_az as boolean, but EC2 filters want
	// it to be serialized as a string. Note that setting it to
	// "false" here does not actually filter by it *not* being
	// the default, because Terraform can't distinguish between
	// "false" and "not set".
	defaultForAzStr := ""
	if d.Get("default_for_az").(bool) {
		defaultForAzStr = "true"
	}

	filters := map[string]string{
		"availabilityZone":   d.Get("availability_zone").(string),
		"availabilityZoneId": d.Get("availability_zone_id").(string),
		"defaultForAz":       defaultForAzStr,
		"state":              d.Get("state").(string),
		"vpc-id":             d.Get("vpc_id").(string),
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		filters["cidrBlock"] = v.(string)
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		filters["ipv6-cidr-block-association.ipv6-cidr-block"] = v.(string)
	}

	input.Filters = BuildAttributeFilterList(filters)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	subnet, err := FindSubnet(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Subnet", err)
	}

	d.SetId(aws.StringValue(subnet.SubnetId))

	d.Set("arn", subnet.SubnetArn)
	d.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	d.Set("availability_zone_id", subnet.AvailabilityZoneId)
	d.Set("availability_zone", subnet.AvailabilityZone)
	d.Set("available_ip_address_count", subnet.AvailableIpAddressCount)
	d.Set("cidr_block", subnet.CidrBlock)
	d.Set("customer_owned_ipv4_pool", subnet.CustomerOwnedIpv4Pool)
	d.Set("default_for_az", subnet.DefaultForAz)
	d.Set("enable_dns64", subnet.EnableDns64)
	d.Set("ipv6_native", subnet.Ipv6Native)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop.
	d.Set("ipv6_cidr_block_association_id", nil)
	d.Set("ipv6_cidr_block", nil)

	for _, v := range subnet.Ipv6CidrBlockAssociationSet {
		if v.Ipv6CidrBlockState != nil && aws.StringValue(v.Ipv6CidrBlockState.State) == ec2.VpcCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_cidr_block_association_id", v.AssociationId)
			d.Set("ipv6_cidr_block", v.Ipv6CidrBlock)
		}
	}

	d.Set("map_customer_owned_ip_on_launch", subnet.MapCustomerOwnedIpOnLaunch)
	d.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	d.Set("outpost_arn", subnet.OutpostArn)
	d.Set("owner_id", subnet.OwnerId)
	d.Set("state", subnet.State)

	if subnet.PrivateDnsNameOptionsOnLaunch != nil {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)
		d.Set("enable_resource_name_dns_a_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)
		d.Set("private_dns_hostname_type_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.HostnameType)
	} else {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", nil)
		d.Set("enable_resource_name_dns_a_record_on_launch", nil)
		d.Set("private_dns_hostname_type_on_launch", nil)
	}

	if err := d.Set("tags", KeyValueTags(subnet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("vpc_id", subnet.VpcId)

	return nil
}
