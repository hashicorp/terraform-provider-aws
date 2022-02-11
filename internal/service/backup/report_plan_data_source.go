package backup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceReportPlan() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
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
			"report_delivery_channel": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"formats": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"s3_key_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"report_setting": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"framework_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"number_of_frameworks": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"report_template": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

