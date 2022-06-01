package efs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAccessPoints() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAccessPointsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"file_system_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAccessPointsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	fileSystemID := d.Get("file_system_id").(string)
	input := &efs.DescribeAccessPointsInput{
		FileSystemId: aws.String(fileSystemID),
	}

	output, err := findAccessPointDescriptions(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EFS Access Points: %w", err)
	}

	var accessPointIDs, arns []string

	for _, v := range output {
		accessPointIDs = append(accessPointIDs, aws.StringValue(v.AccessPointId))
		arns = append(arns, aws.StringValue(v.AccessPointArn))
	}

	d.SetId(fileSystemID)
	d.Set("arns", arns)
	d.Set("ids", accessPointIDs)

	return nil
}

func findAccessPointDescriptions(conn *efs.EFS, input *efs.DescribeAccessPointsInput) ([]*efs.AccessPointDescription, error) {
	var output []*efs.AccessPointDescription

	err := conn.DescribeAccessPointsPages(input, func(page *efs.DescribeAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AccessPoints {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
