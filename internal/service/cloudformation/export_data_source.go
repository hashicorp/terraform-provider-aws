package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceExport() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExportRead,

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

func dataSourceExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()
	var value string
	name := d.Get("name").(string)
	region := meta.(*conns.AWSClient).Region
	d.SetId(fmt.Sprintf("cloudformation-exports-%s-%s", region, name))
	input := &cloudformation.ListExportsInput{}
	err := conn.ListExportsPagesWithContext(ctx, input,
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
		return sdkdiag.AppendErrorf(diags, "Failed listing CloudFormation exports: %s", err)
	}
	if value == "" {
		return sdkdiag.AppendErrorf(diags, "%s was not found in CloudFormation Exports for region %s", name, region)
	}
	return diags
}
