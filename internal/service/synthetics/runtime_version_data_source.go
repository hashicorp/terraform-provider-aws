// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_synthetics_runtime_version", name="Runtime Version")
func newDataSourceRuntimeVersion(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRuntimeVersion{}, nil
}

const (
	DSNameRuntimeVersion = "Runtime Version Data Source"
)

type dataSourceRuntimeVersion struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRuntimeVersion) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"deprecation_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"release_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"latest": schema.BoolAttribute{
				Optional: true,
				Validators: []validator.Bool{
					boolvalidator.Equals(true),
					boolvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("latest"),
						path.MatchRoot(names.AttrVersion),
					}...),
				},
			},
			names.AttrPrefix: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^.*[^-]$`), "must not end with a hyphen"),
				},
			},
			names.AttrVersion: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^\d.*$`), "must start with a digit"),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("latest"),
						path.MatchRoot(names.AttrVersion),
					}...),
				},
			},
			"version_name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceRuntimeVersion) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SyntheticsClient(ctx)

	var data dataSourceRuntimeVersionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	latest := data.Latest.ValueBool()
	prefix := data.Prefix.ValueString()
	version := data.Version.ValueString()

	out, err := findRuntimeVersions(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Synthetics, create.ErrActionReading, DSNameRuntimeVersion, "", err),
			err.Error(),
		)
		return
	}

	var runtimeVersion *awstypes.RuntimeVersion
	var latestReleaseDate *time.Time

	for _, v := range out {
		if strings.HasPrefix(aws.ToString(v.VersionName), prefix) {
			if latest {
				if latestReleaseDate == nil || aws.ToTime(v.ReleaseDate).After(*latestReleaseDate) {
					latestReleaseDate = v.ReleaseDate
					runtimeVersion = &v
				}
			} else {
				if aws.ToString(v.VersionName) == fmt.Sprintf("%s-%s", prefix, version) {
					runtimeVersion = &v
					break
				}
			}
		}
	}

	if runtimeVersion == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Synthetics, create.ErrActionReading, DSNameRuntimeVersion, "", err),
			"Query returned no results.",
		)
		return
	}

	data.ID = flex.StringToFramework(ctx, runtimeVersion.VersionName)
	resp.Diagnostics.Append(flex.Flatten(ctx, runtimeVersion, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRuntimeVersionModel struct {
	DeprecationDate timetypes.RFC3339 `tfsdk:"deprecation_date"`
	Description     types.String      `tfsdk:"description"`
	ID              types.String      `tfsdk:"id"`
	Latest          types.Bool        `tfsdk:"latest"`
	Prefix          types.String      `tfsdk:"prefix"`
	ReleaseDate     timetypes.RFC3339 `tfsdk:"release_date"`
	Version         types.String      `tfsdk:"version"`
	VersionName     types.String      `tfsdk:"version_name"`
}
