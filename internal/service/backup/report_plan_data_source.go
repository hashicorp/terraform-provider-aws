// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_report_plan")
func DataSourceReportPlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReportPlanRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
						names.AttrS3BucketName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrS3KeyPrefix: {
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
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceReportPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	reportPlan, err := FindReportPlanByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Report Plan (%s): %s", name, err)
	}

	d.SetId(aws.ToString(reportPlan.ReportPlanName))

	d.Set(names.AttrARN, reportPlan.ReportPlanArn)
	d.Set(names.AttrCreationTime, reportPlan.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", reportPlan.DeploymentStatus)
	d.Set(names.AttrDescription, reportPlan.ReportPlanDescription)
	d.Set(names.AttrName, reportPlan.ReportPlanName)

	if err := d.Set("report_delivery_channel", flattenReportDeliveryChannel(reportPlan.ReportDeliveryChannel)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_delivery_channel: %s", err)
	}

	if err := d.Set("report_setting", flattenReportSetting(reportPlan.ReportSetting)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_setting: %s", err)
	}

	tags, err := listTags(ctx, conn, aws.ToString(reportPlan.ReportPlanArn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Report Plan (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
