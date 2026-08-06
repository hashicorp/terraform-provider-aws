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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_endpoints", name="Endpoints")
func dataSourceVPCEndpoints() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCEndpointsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"service_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_endpoint_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_endpoint_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VpcEndpointType](),
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
						names.AttrID: {
							Type:     schema.TypeString,
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
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
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
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVPCEndpointsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	input := &ec2.DescribeVpcEndpointsInput{}

	if v, ok := d.GetOk("vpc_endpoint_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcEndpointIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	input.Filters = append(input.Filters, newAttributeFilterList(
		map[string]string{
			"ip-address-type":    d.Get(names.AttrIPAddressType).(string),
			"service-name":       d.Get(names.AttrServiceName).(string),
			"service-region":     d.Get("service_region").(string),
			"vpc-endpoint-state": d.Get(names.AttrState).(string),
			"vpc-endpoint-type":  d.Get("vpc_endpoint_type").(string),
			"vpc-id":             d.Get(names.AttrVPCID).(string),
		},
	)...)

	input.Filters = append(input.Filters, newTagFilterList(
		svcTags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]any))),
	)...)
	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	vpces, err := findVPCEndpoints(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoints: %s", err)
	}

	var vpcEndpointIDs []string
	var vpcEndpoints []map[string]any

	for _, vpce := range vpces {
		vpcEndpointID := aws.ToString(vpce.VpcEndpointId)
		vpcEndpointIDs = append(vpcEndpointIDs, vpcEndpointID)

		ownerID := aws.ToString(vpce.OwnerId)
		vpcEndpointData := map[string]any{
			names.AttrARN:              vpcEndpointARN(ctx, meta.(*conns.AWSClient), ownerID, vpcEndpointID),
			"dns_entry":                flattenDNSEntries(vpce.DnsEntries),
			names.AttrID:               vpcEndpointID,
			names.AttrIPAddressType:    vpce.IpAddressType,
			"network_interface_ids":    vpce.NetworkInterfaceIds,
			names.AttrOwnerID:          ownerID,
			"private_dns_enabled":      vpce.PrivateDnsEnabled,
			"requester_managed":        vpce.RequesterManaged,
			"route_table_ids":          vpce.RouteTableIds,
			names.AttrSecurityGroupIDs: flattenSecurityGroupIdentifiers(vpce.Groups),
			names.AttrServiceName:      aws.ToString(vpce.ServiceName),
			names.AttrState:            vpce.State,
			names.AttrSubnetIDs:        vpce.SubnetIds,
			names.AttrVPCID:            aws.ToString(vpce.VpcId),
		}

		if vpce.DnsOptions != nil {
			vpcEndpointData["dns_options"] = []any{flattenDNSOptions(vpce.DnsOptions)}
		} else {
			vpcEndpointData["dns_options"] = nil
		}

		// VPC endpoints don't have types in GovCloud, so set type to default if empty
		if v := string(vpce.VpcEndpointType); v == "" {
			vpcEndpointData["vpc_endpoint_type"] = awstypes.VpcEndpointTypeGateway
		} else {
			vpcEndpointData["vpc_endpoint_type"] = v
		}

		serviceName := aws.ToString(vpce.ServiceName)
		if pl, err := findPrefixListByName(ctx, conn, serviceName); err != nil {
			if !retry.NotFound(err) {
				return sdkdiag.AppendErrorf(diags, "reading EC2 Prefix List (%s): %s", serviceName, err)
			}
			vpcEndpointData["cidr_blocks"] = nil
			vpcEndpointData["prefix_list_id"] = nil
		} else {
			vpcEndpointData["cidr_blocks"] = pl.Cidrs
			vpcEndpointData["prefix_list_id"] = aws.ToString(pl.PrefixListId)
		}

		if policy, err := structure.NormalizeJsonString(aws.ToString(vpce.PolicyDocument)); err != nil {
			return sdkdiag.AppendErrorf(diags, "policy contains invalid JSON: %s", err)
		} else {
			vpcEndpointData[names.AttrPolicy] = policy
		}

		vpcEndpointData[names.AttrTags] = keyValueTags(ctx, vpce.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()

		vpcEndpoints = append(vpcEndpoints, vpcEndpointData)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrIDs, vpcEndpointIDs)
	d.Set("vpc_endpoints", vpcEndpoints)

	return diags
}
