// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_docdb_engine_version")
func DataSourceEngineVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineVersionRead,
		Schema: map[string]*schema.Schema{
			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  engineDocDB,
			},
			"engine_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exportable_log_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"preferred_versions": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{names.AttrVersion},
			},
			"supports_log_exports_to_cloudwatch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"valid_upgrade_targets": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrVersion: {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"preferred_versions"},
			},
			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	input := &docdb.DescribeDBEngineVersionsInput{}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter_group_family"); ok {
		input.DBParameterGroupFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	} else if _, ok := d.GetOk("preferred_versions"); !ok {
		if _, ok := d.GetOk("parameter_group_family"); !ok {
			input.DefaultOnly = aws.Bool(true)
		}
	}

	var engineVersion *awstypes.DBEngineVersion
	var err error
	if preferredVersions := flex.ExpandStringValueList(d.Get("preferred_versions").([]interface{})); len(preferredVersions) > 0 {
		var engineVersions []awstypes.DBEngineVersion

		engineVersions, err = findEngineVersions(ctx, conn, input)

		if err == nil {
		PreferredVersionLoop:
			// Return the first matching version.
			for _, preferredVersion := range preferredVersions {
				for _, v := range engineVersions {
					if preferredVersion == aws.ToString(v.EngineVersion) {
						ev := &v
						engineVersion = ev
						break PreferredVersionLoop
					}
				}
			}

			if engineVersion == nil {
				err = tfresource.NewEmptyResultError(input)
			}
		}
	} else {
		engineVersion, err = findEngineVersion(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("DocumentDB Engine Version", err))
	}

	d.SetId(aws.ToString(engineVersion.EngineVersion))
	d.Set(names.AttrEngine, engineVersion.Engine)
	d.Set("engine_description", engineVersion.DBEngineDescription)
	d.Set("exportable_log_types", engineVersion.ExportableLogTypes)
	d.Set("parameter_group_family", engineVersion.DBParameterGroupFamily)
	d.Set("supports_log_exports_to_cloudwatch", engineVersion.SupportsLogExportsToCloudwatchLogs)
	d.Set("valid_upgrade_targets", tfslices.ApplyToAll(engineVersion.ValidUpgradeTarget, func(v awstypes.UpgradeTarget) string {
		return aws.ToString(v.EngineVersion)
	}))

	d.Set(names.AttrVersion, engineVersion.EngineVersion)
	d.Set("version_description", engineVersion.DBEngineVersionDescription)

	return diags
}

func findEngineVersion(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBEngineVersionsInput) (*awstypes.DBEngineVersion, error) {
	output, err := findEngineVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEngineVersions(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBEngineVersionsInput) ([]awstypes.DBEngineVersion, error) {
	var output []awstypes.DBEngineVersion

	pages := docdb.NewDescribeDBEngineVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.DBEngineVersions...)
	}

	return output, nil
}
