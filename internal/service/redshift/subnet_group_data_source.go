package redshift

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	subnetgroup, err := FindSubnetGroupByName(conn, d.Get("name").(string))

	if err != nil {
		return fmt.Errorf("reading Redshift Subnet Group (%s): %w", d.Id(), err)
	}

	d.SetId(aws.StringValue(subnetgroup.ClusterSubnetGroupName))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("subnetgroup:%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", subnetgroup.Description)
	d.Set("name", subnetgroup.ClusterSubnetGroupName)
	d.Set("subnet_ids", subnetIdsToSlice(subnetgroup.Subnets))

	//lintignore:AWSR002
	if err := d.Set("tags", KeyValueTags(subnetgroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
