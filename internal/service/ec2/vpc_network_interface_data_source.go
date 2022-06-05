package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkInterfaceRead,
		Schema: map[string]*schema.Schema{
			"arn": {
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
						"association_id": {
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
						"instance_id": {
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
			"availability_zone": {
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
			"owner_id": {
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
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeNetworkInterfacesInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.NetworkInterfaceIds = []*string{aws.String(v.(string))}
	}

	eni, err := FindNetworkInterface(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface: %w", err)
	}

	d.SetId(aws.StringValue(eni.NetworkInterfaceId))
	ownerID := aws.StringValue(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-interface/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if eni.Association != nil {
		if err := d.Set("association", []interface{}{flattenNetworkInterfaceAssociation(eni.Association)}); err != nil {
			return fmt.Errorf("error setting association: %w", err)
		}
	} else {
		d.Set("association", nil)
	}
	if eni.Attachment != nil {
		if err := d.Set("attachment", []interface{}{flattenNetworkInterfaceAttachmentForDataSource(eni.Attachment)}); err != nil {
			return fmt.Errorf("error setting attachment: %w", err)
		}
	} else {
		d.Set("attachment", nil)
	}
	d.Set("availability_zone", eni.AvailabilityZone)
	d.Set("description", eni.Description)
	d.Set("security_groups", FlattenGroupIdentifiers(eni.Groups))
	d.Set("interface_type", eni.InterfaceType)
	d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses))
	d.Set("mac_address", eni.MacAddress)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set("owner_id", ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)
	d.Set("private_ips", FlattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses))
	d.Set("requester_id", eni.RequesterId)
	d.Set("subnet_id", eni.SubnetId)
	d.Set("vpc_id", eni.VpcId)

	if err := d.Set("tags", KeyValueTags(eni.TagSet).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func flattenNetworkInterfaceAttachmentForDataSource(apiObject *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttachmentId; v != nil {
		tfMap["attachment_id"] = aws.StringValue(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.Int64Value(v)
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance_id"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceOwnerId; v != nil {
		tfMap["instance_owner_id"] = aws.StringValue(v)
	}

	return tfMap
}
