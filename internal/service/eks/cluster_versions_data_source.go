// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
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
				CustomType: fwtypes.StringEnumType[awstypes.ClusterVersionStatus](),
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
		input.ClusterType = aws.String(data.ClusterType.ValueString())
	}

	input.DefaultOnly = aws.Bool(data.DefaultOnly.ValueBool())

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

	fmt.Printf("Finding cluster versions\n %v %v %v", input.ClusterVersions, aws.ToString(input.ClusterType), input.Status)
	tflog.Debug(ctx, "Finding cluster versions", map[string]interface{}{"input": input})

	pages := eks.NewDescribeClusterVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.ClusterVersions...)
	}

	fmt.Printf("Found cluster versions %v", output)
	tflog.Debug(ctx, "Found cluster versions", map[string]interface{}{"output": output})

	return output, nil
}

type dataSourceClusterVersionsModel struct {
	ClusterType         types.String                                                    `tfsdk:"cluster_type"`
	DefaultOnly         types.Bool                                                      `tfsdk:"default_only"`
	ClusterVersionsOnly fwtypes.ListValueOf[types.String]                               `tfsdk:"cluster_versions_only"`
	Status              fwtypes.StringEnum[awstypes.ClusterVersionStatus]               `tfsdk:"status"`
	ClusterVersions     fwtypes.ListNestedObjectValueOf[customDataSourceClusterVersion] `tfsdk:"cluster_versions"`
}

type customDataSourceClusterVersion struct {
	ClusterType              types.String                                      `tfsdk:"cluster_type"`
	ClusterVersion           types.String                                      `tfsdk:"cluster_version"`
	DefaultPlatformVersion   types.String                                      `tfsdk:"default_platform_version"`
	EndOfExtendedSupportDate timetypes.RFC3339                                 `tfsdk:"end_of_extended_support_date"`
	EndOfStandardSupportDate timetypes.RFC3339                                 `tfsdk:"end_of_standard_support_date"`
	KubernetesPatchVersion   types.String                                      `tfsdk:"kubernetes_patch_version"`
	ReleaseDate              timetypes.RFC3339                                 `tfsdk:"release_date"`
	Status                   fwtypes.StringEnum[awstypes.ClusterVersionStatus] `tfsdk:"status"`
}
