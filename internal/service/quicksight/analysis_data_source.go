// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_quicksight_analysis", name="Analysis")
// @Tags(identifierAttribute="arn")
func dataSourceAnalysis() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAnalysisRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"analysis_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				names.AttrCreatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"definition": quicksightschema.AnalysisDefinitionDataSourceSchema(),
				"last_published_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrLastUpdatedTime: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrPermissions: quicksightschema.PermissionsDataSourceSchema(),
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"theme_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceAnalysisRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	analysisID := d.Get("analysis_id").(string)
	id := analysisCreateResourceID(awsAccountID, analysisID)

	analysis, err := findAnalysisByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("analysis_id", analysis.AnalysisId)
	d.Set(names.AttrARN, analysis.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrCreatedTime, analysis.CreatedTime.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedTime, analysis.LastUpdatedTime.Format(time.RFC3339))
	d.Set(names.AttrName, analysis.Name)
	d.Set(names.AttrStatus, analysis.Status)
	d.Set("theme_arn", analysis.ThemeArn)

	definition, err := findAnalysisDefinitionByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s) definition: %s", d.Id(), err)
	}

	if err := d.Set("definition", quicksightschema.FlattenAnalysisDefinition(definition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting definition: %s", err)
	}

	permissions, err := findAnalysisPermissionsByTwoPartKey(ctx, conn, awsAccountID, analysisID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Analysis (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}
