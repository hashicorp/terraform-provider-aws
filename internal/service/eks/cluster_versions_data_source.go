// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"

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

func (d *dataSourceClusterVersions) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_eks_cluster_versions"
}

func (d *dataSourceClusterVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_type": schema.StringAttribute{
				Optional: true,
			},
			"default_only": schema.BoolAttribute{
				Optional: true,
			},
			"cluster_versions_only": schema.ListAttribute{
				Optional:   true,
				CustomType: fwtypes.ListOfStringType,
			},
			names.AttrStatus: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[clusterVersionAWSStatus](),
			},
			"cluster_versions": framework.DataSourceComputedListOfObjectAttribute[customDataSourceClusterVersion](ctx),
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

	input := &eks.DescribeClusterVersionsInput{}

	if data.ClusterType.String() != "" {
		input.ClusterType = data.ClusterType.ValueStringPointer()
	}

	input.DefaultOnly = data.DefaultOnly.ValueBoolPointer()

	if len(data.ClusterVersionsOnly.Elements()) > 0 && !data.ClusterVersions.IsNull() {
		clVersions := make([]string, 0, len(data.ClusterVersionsOnly.Elements()))
		for _, v := range data.ClusterVersionsOnly.Elements() {
			clVersions = append(clVersions, v.String())
		}

		input.ClusterVersions = clVersions
	}

	if data.Status.String() != "" {
		input.Status = awstypes.ClusterVersionStatus(data.Status.ValueString())
	}

	// TIP: -- 3. Get information about a resource from AWS
	out, err := findClusterVersions(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprint(names.EKS, create.ErrActionReading, DSNameClusterVersions, err), err.Error())
		return
	}

	output := &eks.DescribeClusterVersionsOutput{
		ClusterVersions: out,
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findClusterVersions(ctx context.Context, conn *eks.Client, input *eks.DescribeClusterVersionsInput) ([]awstypes.ClusterVersionInformation, error) {
	output := make([]awstypes.ClusterVersionInformation, 0)

	pages := eks.NewDescribeClusterVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.ClusterVersions...)
	}

	return output, nil
}

type clusterVersionAWSStatus string

// Values returns all known values for ClusterVersionStatus. Note that this can be
// expanded in the future, and so it is only as up to date as the client.
//
// The ordering of this slice is not guaranteed to be stable across updates.
func (clusterVersionAWSStatus) Values() []clusterVersionAWSStatus {
	return []clusterVersionAWSStatus{
		"UNSUPPORTED",
		"STANDARD_SUPPORT",
		"EXTENDED_SUPPORT",
	}
}

type dataSourceClusterVersionsModel struct {
	ClusterType         types.String                                                    `tfsdk:"cluster_type"`
	DefaultOnly         types.Bool                                                      `tfsdk:"default_only"`
	ClusterVersionsOnly fwtypes.ListValueOf[types.String]                               `tfsdk:"cluster_versions_only"`
	Status              fwtypes.StringEnum[clusterVersionAWSStatus]                     `tfsdk:"status"`
	ClusterVersions     fwtypes.ListNestedObjectValueOf[customDataSourceClusterVersion] `tfsdk:"cluster_versions"`
}

type customDataSourceClusterVersion struct {
	ClusterType              types.String                                `tfsdk:"cluster_type"`
	ClusterVersion           types.String                                `tfsdk:"cluster_version"`
	DefaultPlatformVersion   types.String                                `tfsdk:"default_platform_version"`
	EndOfExtendedSupportDate timetypes.RFC3339                           `tfsdk:"end_of_extended_support_date"`
	EndOfStandardSupportDate timetypes.RFC3339                           `tfsdk:"end_of_standard_support_date"`
	KubernetesPatchVersion   types.String                                `tfsdk:"kubernetes_patch_version"`
	ReleaseDate              timetypes.RFC3339                           `tfsdk:"release_date"`
	DefaultVersion           types.Bool                                  `tfsdk:"default_version"`
	Status                   fwtypes.StringEnum[clusterVersionAWSStatus] `tfsdk:"status"`
}
