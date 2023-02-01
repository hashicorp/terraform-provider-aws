package elasticbeanstalk

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceApplicationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"appversion_lifecycle": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_age_in_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"delete_source_from_s3": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	// Get the name and description
	name := d.Get("name").(string)

	resp, err := conn.DescribeApplicationsWithContext(ctx, &elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: []*string{aws.String(name)},
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Applications (%s): %s", name, err)
	}

	if len(resp.Applications) > 1 || len(resp.Applications) < 1 {
		return sdkdiag.AppendErrorf(diags, "%d Applications matched, expected 1", len(resp.Applications))
	}

	app := resp.Applications[0]

	d.SetId(name)
	d.Set("arn", app.ApplicationArn)
	d.Set("name", app.ApplicationName)
	d.Set("description", app.Description)

	if app.ResourceLifecycleConfig != nil {
		d.Set("appversion_lifecycle", flattenResourceLifecycleConfig(app.ResourceLifecycleConfig))
	}

	return diags
}
