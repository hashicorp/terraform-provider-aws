package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsCloudFormationExport() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFormationExportRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exporting_stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudFormationExportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn
	var name, value string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	}
	region := meta.(*AWSClient).region
	d.SetId(fmt.Sprintf("cloudformation-exports-%s-%s", region, name))
	input := &cloudformation.ListExportsInput{}
	err := conn.ListExportsPages(input,
		func(page *cloudformation.ListExportsOutput, lastPage bool) bool {
			for _, e := range page.Exports {
				if name == *e.Name {
					value = *e.Value
					d.Set("value", *e.Value)
					d.Set("exporting_stack_id", *e.ExportingStackId)
					return false
				}
			}
			if page.NextToken != nil {
				return true
			} else {
				return false
			}
		})
	if err != nil {
		return fmt.Errorf("Failed listing CloudFormation exports: %s", err)
	}
	if "" == value {
		return fmt.Errorf("%s was not found in CloudFormation Exports for region %s", name, region)
	}
	return nil
}
