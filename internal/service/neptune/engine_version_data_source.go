// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_neptune_engine_version")
func DataSourceEngineVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineVersionRead,
		Schema: map[string]*schema.Schema{
			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  engineNeptune,
			},
			"engine_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exportable_log_types": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
			"supported_timezones": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supports_log_exports_to_cloudwatch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_read_replica": {
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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	input := &neptune.DescribeDBEngineVersionsInput{}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameter_group_family"); ok {
		input.DBParameterGroupFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	} else if _, ok := d.GetOk("preferred_versions"); !ok {
		input.DefaultOnly = aws.Bool(true)
	}

	var engineVersion *neptune.DBEngineVersion
	var err error
	if preferredVersions := flex.ExpandStringValueList(d.Get("preferred_versions").([]interface{})); len(preferredVersions) > 0 {
		var engineVersions []*neptune.DBEngineVersion

		engineVersions, err = findEngineVersions(ctx, conn, input)

		if err == nil {
		PreferredVersionLoop:
			// Return the first matching version.
			for _, preferredVersion := range preferredVersions {
				for _, v := range engineVersions {
					if preferredVersion == aws.StringValue(v.EngineVersion) {
						engineVersion = v
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
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Neptune Engine Version", err))
	}

	d.SetId(aws.StringValue(engineVersion.EngineVersion))
	d.Set(names.AttrEngine, engineVersion.Engine)
	d.Set("engine_description", engineVersion.DBEngineDescription)
	d.Set("exportable_log_types", aws.StringValueSlice(engineVersion.ExportableLogTypes))
	d.Set("parameter_group_family", engineVersion.DBParameterGroupFamily)
	d.Set("supported_timezones", tfslices.ApplyToAll(engineVersion.SupportedTimezones, func(v *neptune.Timezone) string {
		return aws.StringValue(v.TimezoneName)
	}))
	d.Set("supports_log_exports_to_cloudwatch", engineVersion.SupportsLogExportsToCloudwatchLogs)
	d.Set("supports_read_replica", engineVersion.SupportsReadReplica)
	d.Set("valid_upgrade_targets", tfslices.ApplyToAll(engineVersion.ValidUpgradeTarget, func(v *neptune.UpgradeTarget) string {
		return aws.StringValue(v.EngineVersion)
	}))

	d.Set(names.AttrVersion, engineVersion.EngineVersion)
	d.Set("version_description", engineVersion.DBEngineVersionDescription)

	return diags
}

func findEngineVersion(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBEngineVersionsInput) (*neptune.DBEngineVersion, error) {
	output, err := findEngineVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findEngineVersions(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBEngineVersionsInput) ([]*neptune.DBEngineVersion, error) {
	var output []*neptune.DBEngineVersion

	err := conn.DescribeDBEngineVersionsPagesWithContext(ctx, input, func(page *neptune.DescribeDBEngineVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBEngineVersions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
