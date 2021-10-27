package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSecurityGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecurityGroupsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecurityGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	req := &ec2.DescribeSecurityGroupsInput{}

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")

	if !filtersOk && !tagsOk {
		return fmt.Errorf("One of filters or tags must be assigned")
	}

	if filtersOk {
		req.Filters = append(req.Filters,
			buildFiltersDataSource(filters.(*schema.Set))...)
	}
	if tagsOk {
		req.Filters = append(req.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	log.Printf("[DEBUG] Reading Security Groups with request: %s", req)

	var ids, vpcIds, arns []string
	for {
		resp, err := conn.DescribeSecurityGroups(req)
		if err != nil {
			return fmt.Errorf("error reading security groups: %w", err)
		}

		for _, sg := range resp.SecurityGroups {
			ids = append(ids, aws.StringValue(sg.GroupId))
			vpcIds = append(vpcIds, aws.StringValue(sg.VpcId))

			arn := arn.ARN{
				Partition: meta.(*conns.AWSClient).Partition,
				Service:   ec2.ServiceName,
				Region:    meta.(*conns.AWSClient).Region,
				AccountID: aws.StringValue(sg.OwnerId),
				Resource:  fmt.Sprintf("security-group/%s", aws.StringValue(sg.GroupId)),
			}.String()

			arns = append(arns, arn)
		}

		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}

	if len(ids) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	log.Printf("[DEBUG] Found %d security groups via given filter: %s", len(ids), req)

	d.SetId(meta.(*conns.AWSClient).Region)

	err := d.Set("ids", ids)
	if err != nil {
		return err
	}

	if err = d.Set("vpc_ids", vpcIds); err != nil {
		return fmt.Errorf("error setting vpc_ids: %s", err)
	}

	if err = d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %s", err)
	}

	return nil
}
