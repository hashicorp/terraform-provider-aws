package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayVpcAttachments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayVpcAttachmentsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"transit_gateway_vpc_attachments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_support": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv6_support": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"transit_gateway_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tagsSchemaComputed(),
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsEc2TransitGatewayVpcAttachmentsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{}
	var items []*schema.ResourceData

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGatewayVpcAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachments: %s", err)
	}

	if output == nil || len(output.TransitGatewayVpcAttachments) == 0 {
		return errors.New("error reading EC2 Transit Gateway VPC Attachments: no results found")
	}

	for _, transitGatewayVpcAttachment := range output.TransitGatewayVpcAttachments {
		var item *schema.ResourceData
		if transitGatewayVpcAttachment != nil {
			if transitGatewayVpcAttachment.Options == nil {
				return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", transitGatewayVpcAttachment.TransitGatewayAttachmentId)
			}

			item.Set("dns_support", transitGatewayVpcAttachment.Options.DnsSupport)
			item.Set("ipv6_support", transitGatewayVpcAttachment.Options.Ipv6Support)

			if err := item.Set("subnet_ids", aws.StringValueSlice(transitGatewayVpcAttachment.SubnetIds)); err != nil {
				return fmt.Errorf("error setting subnet_ids: %s", err)
			}
			if err := item.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayVpcAttachment.Tags).IgnoreAws().Map()); err != nil {
				return fmt.Errorf("error setting tags: %s", err)
			}

			item.Set("transit_gateway_id", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId))
			item.Set("vpc_id", aws.StringValue(transitGatewayVpcAttachment.VpcId))
			item.Set("vpc_owner_id", aws.StringValue(transitGatewayVpcAttachment.VpcOwnerId))

			item.SetId(aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId))
		}
		items = append(items, item)
	}
	d.Set("transit_gateway_vpc_attachments", items)
	return nil
}
