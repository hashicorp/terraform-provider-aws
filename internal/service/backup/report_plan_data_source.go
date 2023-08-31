// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_backup_report_plan")
func DataSourceReportPlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReportPlanRead,

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
						"accounts": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
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
						"organization_units": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"regions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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

func dataSourceReportPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	reportPlan, err := FindReportPlanByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Report Plan (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(reportPlan.ReportPlanName))

	d.Set("arn", reportPlan.ReportPlanArn)
	d.Set("creation_time", reportPlan.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", reportPlan.DeploymentStatus)
	d.Set("description", reportPlan.ReportPlanDescription)
	d.Set("name", reportPlan.ReportPlanName)

	if err := d.Set("report_delivery_channel", flattenReportDeliveryChannel(reportPlan.ReportDeliveryChannel)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_delivery_channel: %s", err)
	}

	if err := d.Set("report_setting", flattenReportSetting(reportPlan.ReportSetting)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_setting: %s", err)
	}

	tags, err := listTags(ctx, conn, aws.StringValue(reportPlan.ReportPlanArn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Report Plan (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
