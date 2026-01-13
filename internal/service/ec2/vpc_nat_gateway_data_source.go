// Copyright IBM Corp. 2014, 2026
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

// @SDKDataSource("aws_nat_gateway", name="NAT Gateway")
func dataSourceNATGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNATGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_provision_zones": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_ips": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_address": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"connectivity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_nat_gateway_address": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAssociationID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secondary_allocation_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"secondary_private_ip_address_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"secondary_private_ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrSubnetID: {
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

func dataSourceNATGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	var diags diag.Diagnostics

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	input := &ec2.DescribeNatGatewaysInput{
		Filter: newAttributeFilterList(
			map[string]string{
				names.AttrState: d.Get(names.AttrState).(string),
				"subnet-id":     d.Get(names.AttrSubnetID).(string),
				"vpc-id":        d.Get(names.AttrVPCID).(string),
			},
		),
	}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.NatGatewayIds = []string{v.(string)}
	}

	if tags, ok := d.GetOk(names.AttrTags); ok {
		input.Filter = append(input.Filter, newTagFilterList(
			svcTags(tftags.New(ctx, tags.(map[string]any))),
		)...)
	}

	input.Filter = append(input.Filter, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filter = nil
	}

	ngw, err := findNATGateway(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 NAT Gateway", err))
	}

	d.SetId(aws.ToString(ngw.NatGatewayId))

	d.Set("availability_mode", ngw.AvailabilityMode)
	d.Set("connectivity_type", ngw.ConnectivityType)
	d.Set(names.AttrState, ngw.State)
	d.Set(names.AttrVPCID, ngw.VpcId)

	switch ngw.AvailabilityMode {
	case awstypes.AvailabilityModeZonal:
		var secondaryAllocationIDs, secondaryPrivateIPAddresses []string

		for _, address := range ngw.NatGatewayAddresses {
			// Length check guarantees the attributes are always set (#30865).
			if isPrimary := aws.ToBool(address.IsPrimary); isPrimary || len(ngw.NatGatewayAddresses) == 1 {
				d.Set("allocation_id", address.AllocationId)
				d.Set(names.AttrAssociationID, address.AssociationId)
				d.Set(names.AttrNetworkInterfaceID, address.NetworkInterfaceId)
				d.Set("private_ip", address.PrivateIp)
				d.Set("public_ip", address.PublicIp)
			} else if !isPrimary {
				if allocationID := aws.ToString(address.AllocationId); allocationID != "" {
					secondaryAllocationIDs = append(secondaryAllocationIDs, allocationID)
				}
				if privateIP := aws.ToString(address.PrivateIp); privateIP != "" {
					secondaryPrivateIPAddresses = append(secondaryPrivateIPAddresses, privateIP)
				}
			}
		}
		d.Set("secondary_allocation_ids", secondaryAllocationIDs)
		d.Set("secondary_private_ip_address_count", len(secondaryPrivateIPAddresses))
		d.Set("secondary_private_ip_addresses", secondaryPrivateIPAddresses)
		d.Set(names.AttrSubnetID, ngw.SubnetId)

	case awstypes.AvailabilityModeRegional:
		d.Set("auto_provision_zones", ngw.AutoProvisionZones)
		d.Set("auto_scaling_ips", ngw.AutoScalingIps)
		if ngw.AutoProvisionZones == awstypes.AutoProvisionZonesStateEnabled {
			d.Set("availability_zone_address", nil)
		} else if err := d.Set("availability_zone_address", flattenNATGatewayAvailabilityZoneAddresses(ngw.NatGatewayAddresses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting availability_zone_address: %s", err)
		}
		if err := d.Set("regional_nat_gateway_address", flattenRegionalNATGatewayAddress(ngw.NatGatewayAddresses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting regional_nat_gateway_address: %s", err)
		}
		d.Set("route_table_id", ngw.RouteTableId)
	}

	if err := d.Set(names.AttrTags, keyValueTags(ctx, ngw.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
