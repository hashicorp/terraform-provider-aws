package backup

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceReportPlan() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceReportPlanRead,

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

func dataSourceReportPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	resp, err := conn.DescribeReportPlan(&backup.DescribeReportPlanInput{
		ReportPlanName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("Error getting Backup Report Plan: %w", err)
	}

	d.SetId(aws.StringValue(resp.ReportPlan.ReportPlanName))

	d.Set("arn", resp.ReportPlan.ReportPlanArn)
	d.Set("deployment_status", resp.ReportPlan.DeploymentStatus)
	d.Set("description", resp.ReportPlan.ReportPlanDescription)
	d.Set("name", resp.ReportPlan.ReportPlanName)

	if err := d.Set("creation_time", resp.ReportPlan.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	if err := d.Set("report_delivery_channel", flattenReportDeliveryChannel(resp.ReportPlan.ReportDeliveryChannel)); err != nil {
		return fmt.Errorf("error setting report_delivery_channel: %w", err)
	}

	if err := d.Set("report_setting", flattenReportSetting(resp.ReportPlan.ReportSetting)); err != nil {
		return fmt.Errorf("error setting report_delivery_channel: %w", err)
	}

	tags, err := ListTags(conn, aws.StringValue(resp.ReportPlan.ReportPlanArn))

	if err != nil {
		return fmt.Errorf("error listing tags for Backup Report Plan (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
