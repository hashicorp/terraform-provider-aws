// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_endpoint", name="Endpoint")
func dataSourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCEndpointRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDNSName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_record_ip_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_dns_only_for_inbound_resolver_endpoint": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"requester_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpcEndpointsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"vpc-endpoint-state": d.Get(names.AttrState).(string),
				"vpc-id":             d.Get(names.AttrVPCID).(string),
				"service-name":       d.Get(names.AttrServiceName).(string),
			},
		),
	}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.VpcEndpointIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)
	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	vpce, err := findVPCEndpoint(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 VPC Endpoint", err))
	}

	d.SetId(aws.ToString(vpce.VpcEndpointId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.ToString(vpce.OwnerId),
		Resource:  fmt.Sprintf("vpc-endpoint/%s", d.Id()),
	}.String()
	serviceName := aws.ToString(vpce.ServiceName)

	d.Set(names.AttrARN, arn)
	if err := d.Set("dns_entry", flattenDNSEntries(vpce.DnsEntries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dns_entry: %s", err)
	}
	if vpce.DnsOptions != nil {
		if err := d.Set("dns_options", []interface{}{flattenDNSOptions(vpce.DnsOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_options: %s", err)
		}
	} else {
		d.Set("dns_options", nil)
	}
	d.Set(names.AttrIPAddressType, vpce.IpAddressType)
	d.Set("network_interface_ids", vpce.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, vpce.OwnerId)
	d.Set("private_dns_enabled", vpce.PrivateDnsEnabled)
	d.Set("requester_managed", vpce.RequesterManaged)
	d.Set("route_table_ids", vpce.RouteTableIds)
	d.Set(names.AttrSecurityGroupIDs, flattenSecurityGroupIdentifiers(vpce.Groups))
	d.Set(names.AttrServiceName, serviceName)
	d.Set(names.AttrState, vpce.State)
	d.Set(names.AttrSubnetIDs, vpce.SubnetIds)
	// VPC endpoints don't have types in GovCloud, so set type to default if empty
	if v := string(vpce.VpcEndpointType); v == "" {
		d.Set("vpc_endpoint_type", awstypes.VpcEndpointTypeGateway)
	} else {
		d.Set("vpc_endpoint_type", v)
	}
	d.Set(names.AttrVPCID, vpce.VpcId)

	if pl, err := findPrefixListByName(ctx, conn, serviceName); err != nil {
		if tfresource.NotFound(err) {
			d.Set("cidr_blocks", nil)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Prefix List (%s): %s", serviceName, err)
		}
	} else {
		d.Set("cidr_blocks", pl.Cidrs)
		d.Set("prefix_list_id", pl.PrefixListId)
	}

	policy, err := structure.NormalizeJsonString(aws.ToString(vpce.PolicyDocument))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy contains invalid JSON: %s", err)
	}

	d.Set(names.AttrPolicy, policy)

	if err := d.Set(names.AttrTags, keyValueTags(ctx, vpce.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
