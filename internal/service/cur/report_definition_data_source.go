// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cur_report_definition", name="Report Definition")
// @Tags(identifierAttribute="report_name")
func dataSourceReportDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReportDefinitionRead,

		Schema: map[string]*schema.Schema{
			"additional_artifacts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"additional_schema_elements": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"compression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFormat: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"refresh_closed_reports": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"report_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"report_versioning": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrS3Bucket: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"time_unit": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceReportDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CURClient(ctx)

	reportName := d.Get("report_name").(string)
	reportDefinition, err := findReportDefinitionByName(ctx, conn, reportName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost And Usage Report Definition (%s): %s", reportName, err)
	}

	d.SetId(aws.StringValue(reportDefinition.ReportName))
	d.Set("additional_artifacts", reportDefinition.AdditionalArtifacts)
	d.Set("additional_schema_elements", reportDefinition.AdditionalSchemaElements)
	d.Set("compression", reportDefinition.Compression)
	d.Set(names.AttrFormat, reportDefinition.Format)
	d.Set("refresh_closed_reports", reportDefinition.RefreshClosedReports)
	d.Set("report_name", reportDefinition.ReportName)
	d.Set("report_versioning", reportDefinition.ReportVersioning)
	d.Set(names.AttrS3Bucket, reportDefinition.S3Bucket)
	d.Set("s3_prefix", reportDefinition.S3Prefix)
	d.Set("s3_region", reportDefinition.S3Region)
	d.Set("time_unit", reportDefinition.TimeUnit)

	return diags
}
