// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_eks_cluster_versions", name="Cluster Versions")
func newClusterVersionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &clusterVersionsDataSource{}, nil
}

type clusterVersionsDataSource struct {
	framework.DataSourceWithModel[clusterVersionsDataSourceModel]
}

func (d *clusterVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *clusterVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EKSClient(ctx)

	var data clusterVersionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := eks.DescribeClusterVersionsInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findClusterVersions(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"reading EKS Cluster Versions",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data.ClusterVersions)...)
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

type clusterVersionsDataSourceModel struct {
	framework.WithRegionModel
	ClusterType         types.String                                                    `tfsdk:"cluster_type"`
	ClusterVersions     fwtypes.ListNestedObjectValueOf[customDataSourceClusterVersion] `tfsdk:"cluster_versions"`
	ClusterVersionsOnly fwtypes.ListValueOf[types.String]                               `tfsdk:"cluster_versions_only"`
	DefaultOnly         types.Bool                                                      `tfsdk:"default_only"`
	IncludeAll          types.Bool                                                      `tfsdk:"include_all"`
	VersionStatus       fwtypes.StringEnum[awstypes.VersionStatus]                      `tfsdk:"version_status"`
}

type customDataSourceClusterVersion struct {
	ClusterType                 types.String                                                      `tfsdk:"cluster_type"`
	ClusterVersion              types.String                                                      `tfsdk:"cluster_version"`
	ControlPlaneComponentConfig fwtypes.ListNestedObjectValueOf[controlPlaneConfigInfoModel]      `tfsdk:"control_plane_component_config"`
	ControlPlaneScalingTiers    fwtypes.ListNestedObjectValueOf[controlPlaneScalingTierInfoModel] `tfsdk:"control_plane_scaling_tiers"`
	DefaultPlatformVersion      types.String                                                      `tfsdk:"default_platform_version"`
	DefaultVersion              types.Bool                                                        `tfsdk:"default_version"`
	EndOfExtendedSupportDate    timetypes.RFC3339                                                 `tfsdk:"end_of_extended_support_date"`
	EndOfStandardSupportDate    timetypes.RFC3339                                                 `tfsdk:"end_of_standard_support_date"`
	KubernetesPatchVersion      types.String                                                      `tfsdk:"kubernetes_patch_version"`
	ReleaseDate                 timetypes.RFC3339                                                 `tfsdk:"release_date"`
	VersionStatus               fwtypes.StringEnum[awstypes.VersionStatus]                        `tfsdk:"version_status"`
}

type controlPlaneConfigInfoModel struct {
	KubeApiServerConfig         fwtypes.ListNestedObjectValueOf[kubeAPIServerVersionConfigModel]         `tfsdk:"kube_api_server_config"`
	KubeControllerManagerConfig fwtypes.ListNestedObjectValueOf[kubeControllerManagerVersionConfigModel] `tfsdk:"kube_controller_manager_config"`
	KubeSchedulerConfig         fwtypes.ListNestedObjectValueOf[kubeSchedulerVersionConfigModel]         `tfsdk:"kube_scheduler_config"`
}

type kubeAPIServerVersionConfigModel struct {
	EventTtl             fwtypes.ListNestedObjectValueOf[durationParameterConfigModel]  `tfsdk:"event_ttl"`
	ServiceNodePortRange fwtypes.ListNestedObjectValueOf[portRangeParameterConfigModel] `tfsdk:"service_node_port_range"`
}

type kubeControllerManagerVersionConfigModel struct {
	HorizontalPodAutoscalerControllerConfig fwtypes.ListNestedObjectValueOf[horizontalPodAutoscalerControllerVersionConfigModel] `tfsdk:"horizontal_pod_autoscaler_controller_config"`
	PodGcControllerConfig                   fwtypes.ListNestedObjectValueOf[podGcControllerVersionConfigModel]                   `tfsdk:"pod_gc_controller_config"`
}

type horizontalPodAutoscalerControllerVersionConfigModel struct {
	HorizontalPodAutoscalerSyncPeriod fwtypes.ListNestedObjectValueOf[durationParameterConfigModel] `tfsdk:"horizontal_pod_autoscaler_sync_period"`
}

type podGcControllerVersionConfigModel struct {
	TerminatedPodGcThreshold fwtypes.ListNestedObjectValueOf[integerParameterConfigModel] `tfsdk:"terminated_pod_gc_threshold"`
}

type integerParameterConfigModel struct {
	Constraints  fwtypes.ListNestedObjectValueOf[integerConstraintsModel] `tfsdk:"constraints"`
	DefaultValue types.Int32                                              `tfsdk:"default_value"`
}

type integerConstraintsModel struct {
	Max types.Int32 `tfsdk:"max"`
	Min types.Int32 `tfsdk:"min"`
}

type kubeSchedulerVersionConfigModel struct {
	NodeResourcesFit fwtypes.ListNestedObjectValueOf[nodeResourcesFitVersionConfigModel] `tfsdk:"node_resources_fit"`
}

type nodeResourcesFitVersionConfigModel struct {
	ScoringStrategy fwtypes.ListNestedObjectValueOf[scoringStrategyConfigModel] `tfsdk:"scoring_strategy"`
}

type scoringStrategyConfigModel struct {
	Constraints  fwtypes.ListNestedObjectValueOf[scoringStrategyConstraintsModel] `tfsdk:"constraints"`
	DefaultValue fwtypes.ListNestedObjectValueOf[scoringStrategyModel]            `tfsdk:"default_value"`
}

type scoringStrategyModel struct {
	Resources fwtypes.ListNestedObjectValueOf[resourceWeightModel] `tfsdk:"resources"`
	Type      fwtypes.StringEnum[awstypes.ScoringStrategyType]     `tfsdk:"type"`
}

type resourceWeightModel struct {
	Name   types.String `tfsdk:"name"`
	Weight types.Int32  `tfsdk:"weight"`
}

type scoringStrategyConstraintsModel struct {
	Resources       fwtypes.ListNestedObjectValueOf[resourceConstraintsModel]     `tfsdk:"resources"`
	ScoringStrategy fwtypes.ListNestedObjectValueOf[allowedValuesConstraintModel] `tfsdk:"scoring_strategy"`
}

type resourceConstraintsModel struct {
	Name   fwtypes.ListNestedObjectValueOf[allowedValuesConstraintModel] `tfsdk:"name"`
	Weight fwtypes.ListNestedObjectValueOf[integerRangeConstraintModel]  `tfsdk:"weight"`
}

type allowedValuesConstraintModel struct {
	AllowedValues fwtypes.ListValueOf[types.String] `tfsdk:"allowed_values"`
}

type integerRangeConstraintModel struct {
	Max types.Int32 `tfsdk:"max"`
	Min types.Int32 `tfsdk:"min"`
}

type durationParameterConfigModel struct {
	Constraints  fwtypes.ListNestedObjectValueOf[durationConstraintsModel] `tfsdk:"constraints"`
	DefaultValue types.String                                              `tfsdk:"default_value"`
}

type durationConstraintsModel struct {
	Max types.String `tfsdk:"max"`
	Min types.String `tfsdk:"min"`
}

type portRangeParameterConfigModel struct {
	Constraints  fwtypes.ListNestedObjectValueOf[portRangeConstraintsModel] `tfsdk:"constraints"`
	DefaultValue fwtypes.ListNestedObjectValueOf[serviceNodePortRangeModel] `tfsdk:"default_value"`
}

type portRangeConstraintsModel struct {
	MaxPort fwtypes.ListNestedObjectValueOf[integerRangeConstraintModel] `tfsdk:"max_port"`
	MinPort fwtypes.ListNestedObjectValueOf[integerRangeConstraintModel] `tfsdk:"min_port"`
}

type serviceNodePortRangeModel struct {
	MaxPort types.Int32 `tfsdk:"max_port"`
	MinPort types.Int32 `tfsdk:"min_port"`
}

type controlPlaneScalingTierInfoModel struct {
	ApiRequestConcurrency                types.Int32                                                  `tfsdk:"api_request_concurrency"`
	ClusterDatabaseSizeGb                types.Int32                                                  `tfsdk:"cluster_database_size_gb"`
	ControlPlaneComponentConfigOverrides fwtypes.ListNestedObjectValueOf[controlPlaneConfigInfoModel] `tfsdk:"control_plane_component_config_overrides"`
	PodSchedulingRatePerSecond           types.Int32                                                  `tfsdk:"pod_scheduling_rate_per_second"`
	TierName                             types.String                                                 `tfsdk:"tier_name"`
}
