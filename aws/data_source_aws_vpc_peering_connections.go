package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsVpcPeeringConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcPeeringConnectionsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"peer_cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"filter": ec2CustomFiltersSchema(),
			"tags":   tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsVpcPeeringConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading VPC Peering Connections.")

	req := &ec2.DescribeVpcPeeringConnectionsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.VpcPeeringConnectionIds = aws.StringSlice([]string{id.(string)})
	}

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)
	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading VPC Peering Connections: %s", req)
	resp, err := conn.DescribeVpcPeeringConnections(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.VpcPeeringConnections) == 0 {
		return fmt.Errorf("no matching VPC peering connections found")
	}

	var ids, cidr_blocks, peer_cidr_blocks []string
	for _, pcx := range resp.VpcPeeringConnections {
		ids = append(ids, aws.StringValue(pcx.VpcPeeringConnectionId))
		cidr_blocks = append(cidr_blocks, *pcx.RequesterVpcInfo.CidrBlock)
		peer_cidr_blocks = append(peer_cidr_blocks, *pcx.AccepterVpcInfo.CidrBlock)
	}

	if len(ids) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	log.Printf("[DEBUG] Found %d peering connections via given filter: %s", len(ids), req)

	d.SetId(resource.UniqueId())
	err = d.Set("ids", ids)
	if err != nil {
		return err
	}

	err = d.Set("cidr_blocks", cidr_blocks)
	if err != nil {
		return err
	}

	err = d.Set("peer_cidr_blocks", peer_cidr_blocks)
	if err != nil {
		return err
	}
	return nil
}
