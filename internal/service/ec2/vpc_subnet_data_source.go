// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_subnet")
func dataSourceSubnet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubnetRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
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
			names.AttrCIDRBlock: {
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
			"enable_lni_at_device_index": {
				Type:     schema.TypeInt,
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
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
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
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_hostname_type_on_launch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeSubnetsInput{}

	if id, ok := d.GetOk(names.AttrID); ok {
		input.SubnetIds = []string{id.(string)}
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
		"availabilityZone":   d.Get(names.AttrAvailabilityZone).(string),
		"availabilityZoneId": d.Get("availability_zone_id").(string),
		"defaultForAz":       defaultForAzStr,
		names.AttrState:      d.Get(names.AttrState).(string),
		"vpc-id":             d.Get(names.AttrVPCID).(string),
	}

	if v, ok := d.GetOk(names.AttrCIDRBlock); ok {
		filters["cidrBlock"] = v.(string)
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		filters["ipv6-cidr-block-association.ipv6-cidr-block"] = v.(string)
	}

	input.Filters = newAttributeFilterList(filters)

	if tags, tagsOk := d.GetOk(names.AttrTags); tagsOk {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	subnet, err := findSubnet(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Subnet", err))
	}

	d.SetId(aws.ToString(subnet.SubnetId))

	d.Set(names.AttrARN, subnet.SubnetArn)
	d.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	d.Set("availability_zone_id", subnet.AvailabilityZoneId)
	d.Set(names.AttrAvailabilityZone, subnet.AvailabilityZone)
	d.Set("available_ip_address_count", subnet.AvailableIpAddressCount)
	d.Set(names.AttrCIDRBlock, subnet.CidrBlock)
	d.Set("customer_owned_ipv4_pool", subnet.CustomerOwnedIpv4Pool)
	d.Set("default_for_az", subnet.DefaultForAz)
	d.Set("enable_dns64", subnet.EnableDns64)
	d.Set("enable_lni_at_device_index", subnet.EnableLniAtDeviceIndex)
	d.Set("ipv6_native", subnet.Ipv6Native)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop.
	d.Set("ipv6_cidr_block_association_id", nil)
	d.Set("ipv6_cidr_block", nil)

	for _, v := range subnet.Ipv6CidrBlockAssociationSet {
		if v.Ipv6CidrBlockState != nil && v.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_cidr_block_association_id", v.AssociationId)
			d.Set("ipv6_cidr_block", v.Ipv6CidrBlock)
		}
	}

	d.Set("map_customer_owned_ip_on_launch", subnet.MapCustomerOwnedIpOnLaunch)
	d.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	d.Set("outpost_arn", subnet.OutpostArn)
	d.Set(names.AttrOwnerID, subnet.OwnerId)
	d.Set(names.AttrState, subnet.State)

	if subnet.PrivateDnsNameOptionsOnLaunch != nil {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)
		d.Set("enable_resource_name_dns_a_record_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)
		d.Set("private_dns_hostname_type_on_launch", subnet.PrivateDnsNameOptionsOnLaunch.HostnameType)
	} else {
		d.Set("enable_resource_name_dns_aaaa_record_on_launch", nil)
		d.Set("enable_resource_name_dns_a_record_on_launch", nil)
		d.Set("private_dns_hostname_type_on_launch", nil)
	}

	if err := d.Set(names.AttrTags, keyValueTags(ctx, subnet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.Set(names.AttrVPCID, subnet.VpcId)

	return diags
}
