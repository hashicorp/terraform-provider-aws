// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_network_interface", name="Network Interface")
// @Tags
// @Testing(tagsTest=false)
func dataSourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNetworkInterfaceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association": {
				Type:     schema.TypeList,
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
						"carrier_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_owned_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"attachment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrInstanceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"interface_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"mac_address": {
				Type:     schema.TypeString,
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
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"requester_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeNetworkInterfacesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.NetworkInterfaceIds = []string{v.(string)}
	}

	eni, err := findNetworkInterface(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Network Interface", err))
	}

	d.SetId(aws.ToString(eni.NetworkInterfaceId))
	ownerID := aws.ToString(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ec2",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  "network-interface/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	if eni.Association != nil {
		if err := d.Set("association", []interface{}{flattenNetworkInterfaceAssociation(eni.Association)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting association: %s", err)
		}
	} else {
		d.Set("association", nil)
	}
	if eni.Attachment != nil {
		if err := d.Set("attachment", []interface{}{flattenNetworkInterfaceAttachmentForDataSource(eni.Attachment)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting attachment: %s", err)
		}
	} else {
		d.Set("attachment", nil)
	}
	d.Set(names.AttrAvailabilityZone, eni.AvailabilityZone)
	d.Set(names.AttrDescription, eni.Description)
	d.Set(names.AttrSecurityGroups, flattenGroupIdentifiers(eni.Groups))
	d.Set("interface_type", eni.InterfaceType)
	d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses))
	d.Set("mac_address", eni.MacAddress)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set(names.AttrOwnerID, ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)
	d.Set("private_ips", flattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses))
	d.Set("requester_id", eni.RequesterId)
	d.Set(names.AttrSubnetID, eni.SubnetId)
	d.Set(names.AttrVPCID, eni.VpcId)

	setTagsOut(ctx, eni.TagSet)

	return diags
}

func flattenNetworkInterfaceAttachmentForDataSource(apiObject *types.NetworkInterfaceAttachment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttachmentId; v != nil {
		tfMap["attachment_id"] = aws.ToString(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.ToInt32(v)
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap[names.AttrInstanceID] = aws.ToString(v)
	}

	if v := apiObject.InstanceOwnerId; v != nil {
		tfMap["instance_owner_id"] = aws.ToString(v)
	}

	return tfMap
}
