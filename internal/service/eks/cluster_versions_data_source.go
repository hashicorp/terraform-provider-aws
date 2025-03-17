// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_eks_cluster_versions", name="Cluster Versions")
func newDataSourceClusterVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceClusterVersions{}, nil
}

const (
	DSNameClusterVersions = "Cluster Versions Data Source"
)

type dataSourceClusterVersions struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceClusterVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_type": schema.StringAttribute{
				Optional: true,
			},
			"cluster_versions": framework.DataSourceComputedListOfObjectAttribute[customDataSourceClusterVersion](ctx),
			"cluster_versions_only": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
			},
			"default_only": schema.BoolAttribute{
				Optional: true,
			},
			"include_all": schema.BoolAttribute{
				Optional: true,
			},
			"version_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VersionStatus](),
				Optional:   true,
			},
		},
	}
}

func (d *dataSourceClusterVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EKSClient(ctx)

	var data dataSourceClusterVersionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := eks.DescribeClusterVersionsInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findClusterVersions(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionReading, DSNameClusterVersions, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.ClusterVersions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findClusterVersions(ctx context.Context, conn *eks.Client, input *eks.DescribeClusterVersionsInput) ([]awstypes.ClusterVersionInformation, error) {
	out := make([]awstypes.ClusterVersionInformation, 0)

	pages := eks.NewDescribeClusterVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		out = append(out, page.ClusterVersions...)
	}

	return out, nil
}

type dataSourceClusterVersionsModel struct {
	ClusterType         types.String                                                    `tfsdk:"cluster_type"`
	ClusterVersions     fwtypes.ListNestedObjectValueOf[customDataSourceClusterVersion] `tfsdk:"cluster_versions"`
	ClusterVersionsOnly fwtypes.ListValueOf[types.String]                               `tfsdk:"cluster_versions_only"`
	DefaultOnly         types.Bool                                                      `tfsdk:"default_only"`
	IncludeAll          types.Bool                                                      `tfsdk:"include_all"`
	VersionStatus       fwtypes.StringEnum[awstypes.VersionStatus]                      `tfsdk:"version_status"`
}

type customDataSourceClusterVersion struct {
	ClusterType              types.String                               `tfsdk:"cluster_type"`
	ClusterVersion           types.String                               `tfsdk:"cluster_version"`
	DefaultPlatformVersion   types.String                               `tfsdk:"default_platform_version"`
	DefaultVersion           types.Bool                                 `tfsdk:"default_version"`
	EndOfExtendedSupportDate timetypes.RFC3339                          `tfsdk:"end_of_extended_support_date"`
	EndOfStandardSupportDate timetypes.RFC3339                          `tfsdk:"end_of_standard_support_date"`
	KubernetesPatchVersion   types.String                               `tfsdk:"kubernetes_patch_version"`
	ReleaseDate              timetypes.RFC3339                          `tfsdk:"release_date"`
	VersionStatus            fwtypes.StringEnum[awstypes.VersionStatus] `tfsdk:"version_status"`
}
