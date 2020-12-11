package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLakeFormationResource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLakeFormationResourceRead,

		Schema: map[string]*schema.Schema{
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLakeFormationResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.DescribeResourceInput{}

	if v, ok := d.GetOk("resource_arn"); ok {
		input.ResourceArn = aws.String(v.(string))
	}

	output, err := conn.DescribeResource(input)

	if err != nil {
		return fmt.Errorf("error reading data source, Lake Formation Resource (resource_arn: %s): %w", aws.StringValue(input.ResourceArn), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return fmt.Errorf("error reading data source, Lake Formation Resource: empty response")
	}

	d.SetId(aws.StringValue(input.ResourceArn))
	// d.Set("resource_arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set("role_arn", output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return nil
}
