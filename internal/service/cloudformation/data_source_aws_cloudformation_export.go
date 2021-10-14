package cloudformation

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceExport() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceExportRead,

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

func dataSourceExportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn
	var value string
	name := d.Get("name").(string)
	region := meta.(*conns.AWSClient).Region
	d.SetId(fmt.Sprintf("cloudformation-exports-%s-%s", region, name))
	input := &cloudformation.ListExportsInput{}
	err := conn.ListExportsPages(input,
		func(page *cloudformation.ListExportsOutput, lastPage bool) bool {
			for _, e := range page.Exports {
				if name == aws.StringValue(e.Name) {
					value = aws.StringValue(e.Value)
					d.Set("value", e.Value)
					d.Set("exporting_stack_id", e.ExportingStackId)
					return false
				}
			}
			return !lastPage
		})
	if err != nil {
		return fmt.Errorf("Failed listing CloudFormation exports: %w", err)
	}
	if value == "" {
		return fmt.Errorf("%s was not found in CloudFormation Exports for region %s", name, region)
	}
	return nil
}
