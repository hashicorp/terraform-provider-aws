package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_network_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	v, err := FindDBSubnetGroupByName(conn, d.Get("name").(string))

	if err != nil {
		return tfresource.SingularDataSourceFindError("RDS DB Subnet Group", err)
	}

	d.SetId(aws.StringValue(v.DBSubnetGroupName))
	d.Set("arn", v.DBSubnetGroupArn)
	d.Set("description", v.DBSubnetGroupDescription)
	d.Set("name", v.DBSubnetGroupName)
	d.Set("status", v.SubnetGroupStatus)
	var subnetIDs []string
	for _, v := range v.Subnets {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetIdentifier))
	}
	d.Set("subnet_ids", subnetIDs)
	d.Set("supported_network_types", aws.StringValueSlice(v.SupportedNetworkTypes))
	d.Set("vpc_id", v.VpcId)

	return nil
}
