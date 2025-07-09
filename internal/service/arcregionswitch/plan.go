// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func FindPlanByARN(ctx context.Context, conn *arcregionswitch.Client, arn string) (*awstypes.Plan, error) {
	input := arcregionswitch.GetPlanInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetPlan(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Plan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Plan, nil
}

func ListTags(ctx context.Context, conn *arcregionswitch.Client, identifier string) (tftags.KeyValueTags, error) {
	input := arcregionswitch.ListTagsForResourceInput{
		Arn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return tftags.New(ctx, output.ResourceTags), nil
}

func UpdateTags(ctx context.Context, conn *arcregionswitch.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := arcregionswitch.UntagResourceInput{
			Arn:             aws.String(identifier),
			ResourceTagKeys: removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := arcregionswitch.TagResourceInput{
			Arn:  aws.String(identifier),
			Tags: updatedTags.Map(),
		}

		_, err := conn.TagResource(ctx, &input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// @FrameworkResource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func newResourcePlan(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePlan{}
	return r, nil
}

type resourcePlan struct {
	framework.ResourceWithConfigure
}

func (r *resourcePlan) ValidateModel(ctx context.Context, schema *fwschema.Schema) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	// Basic validation is handled by the schema validators
	return diags
}

type resourcePlanModel struct {
	ARN                          types.String `tfsdk:"arn"`
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	ExecutionRole                types.String `tfsdk:"execution_role"`
	RecoveryApproach             types.String `tfsdk:"recovery_approach"`
	Regions                      types.List   `tfsdk:"regions"`
	Description                  types.String `tfsdk:"description"`
	PrimaryRegion                types.String `tfsdk:"primary_region"`
	RecoveryTimeObjectiveMinutes types.Int64  `tfsdk:"recovery_time_objective_minutes"`
	AssociatedAlarms             types.Set    `tfsdk:"associated_alarms"`
	Workflow                     types.List   `tfsdk:"workflow"`
	Region                       types.String `tfsdk:"region"`
	Tags                         tftags.Map   `tfsdk:"tags"`
	TagsAll                      tftags.Map   `tfsdk:"tags_all"`
}

type associatedAlarmModel struct {
	Name               types.String `tfsdk:"name"`
	AlarmType          types.String `tfsdk:"alarm_type"`
	ResourceIdentifier types.String `tfsdk:"resource_identifier"`
	CrossAccountRole   types.String `tfsdk:"cross_account_role"`
	ExternalId         types.String `tfsdk:"external_id"`
}

type workflowModel struct {
	WorkflowTargetAction types.String `tfsdk:"workflow_target_action"`
	WorkflowTargetRegion types.String `tfsdk:"workflow_target_region"`
	WorkflowDescription  types.String `tfsdk:"workflow_description"`
	Step                 types.List   `tfsdk:"step"`
}

type route53HealthCheckConfigModel struct {
	HostedZoneId     types.String `tfsdk:"hosted_zone_id"`
	RecordName       types.String `tfsdk:"record_name"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
	TimeoutMinutes   types.Int64  `tfsdk:"timeout_minutes"`
	RecordSets       types.List   `tfsdk:"record_sets"`
}

type recordSetModel struct {
	RecordSetIdentifier types.String `tfsdk:"record_set_identifier"`
	Region              types.String `tfsdk:"region"`
}

type stepModel struct {
	Name                         types.String `tfsdk:"name"`
	ExecutionBlockType           types.String `tfsdk:"execution_block_type"`
	Description                  types.String `tfsdk:"description"`
	ExecutionApprovalConfig      types.List   `tfsdk:"execution_approval_config"`
	Route53HealthCheckConfig     types.List   `tfsdk:"route53_health_check_config"`
	CustomActionLambdaConfig     types.List   `tfsdk:"custom_action_lambda_config"`
	GlobalAuroraConfig           types.List   `tfsdk:"global_aurora_config"`
	Ec2AsgCapacityIncreaseConfig types.List   `tfsdk:"ec2_asg_capacity_increase_config"`
	EcsCapacityIncreaseConfig    types.List   `tfsdk:"ecs_capacity_increase_config"`
	EksResourceScalingConfig     types.List   `tfsdk:"eks_resource_scaling_config"`
	ArcRoutingControlConfig      types.List   `tfsdk:"arc_routing_control_config"`
	ParallelConfig               types.List   `tfsdk:"parallel_config"`
}

type executionApprovalConfigModel struct {
	ApprovalRole   types.String `tfsdk:"approval_role"`
	TimeoutMinutes types.Int64  `tfsdk:"timeout_minutes"`
}

type customActionLambdaConfigModel struct {
	RegionToRun          types.String  `tfsdk:"region_to_run"`
	RetryIntervalMinutes types.Float64 `tfsdk:"retry_interval_minutes"`
	TimeoutMinutes       types.Int64   `tfsdk:"timeout_minutes"`
	Lambda               types.List    `tfsdk:"lambda"`
	Ungraceful           types.List    `tfsdk:"ungraceful"`
}

type lambdaModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type ungracefulModel struct {
	Behavior types.String `tfsdk:"behavior"`
}

// Global Aurora Configuration Models
type globalAuroraConfigModel struct {
	Behavior                types.String `tfsdk:"behavior"`
	GlobalClusterIdentifier types.String `tfsdk:"global_cluster_identifier"`
	DatabaseClusterArns     types.List   `tfsdk:"database_cluster_arns"`
	CrossAccountRole        types.String `tfsdk:"cross_account_role"`
	ExternalId              types.String `tfsdk:"external_id"`
	TimeoutMinutes          types.Int64  `tfsdk:"timeout_minutes"`
	Ungraceful              types.List   `tfsdk:"ungraceful"`
}

type globalAuroraUngracefulModel struct {
	Ungraceful types.String `tfsdk:"ungraceful"`
}

// EC2 ASG Configuration Models
type ec2AsgCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	Asgs                       types.List   `tfsdk:"asgs"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type asgModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type ec2UngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ECS Configuration Models
type ecsCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	Services                   types.List   `tfsdk:"services"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type serviceModel struct {
	ClusterArn       types.String `tfsdk:"cluster_arn"`
	ServiceArn       types.String `tfsdk:"service_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type ecsUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// EKS Configuration Models
type eksResourceScalingConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	KubernetesResourceType     types.List   `tfsdk:"kubernetes_resource_type"`
	EksClusters                types.List   `tfsdk:"eks_clusters"`
	ScalingResources           types.List   `tfsdk:"scaling_resources"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type kubernetesResourceTypeModel struct {
	ApiVersion types.String `tfsdk:"api_version"`
	Kind       types.String `tfsdk:"kind"`
}

type eksClusterModel struct {
	ClusterArn       types.String `tfsdk:"cluster_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type scalingResourcesModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Resources types.List   `tfsdk:"resources"`
}

type kubernetesScalingResourceModel struct {
	ResourceName types.String `tfsdk:"resource_name"`
	Name         types.String `tfsdk:"name"`
	Namespace    types.String `tfsdk:"namespace"`
	HpaName      types.String `tfsdk:"hpa_name"`
}

type eksUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ARC Routing Control Configuration Models
type arcRoutingControlConfigModel struct {
	CrossAccountRole         types.String `tfsdk:"cross_account_role"`
	ExternalId               types.String `tfsdk:"external_id"`
	TimeoutMinutes           types.Int64  `tfsdk:"timeout_minutes"`
	RegionAndRoutingControls types.List   `tfsdk:"region_and_routing_controls"`
}

type regionAndRoutingControlsModel struct {
	Region             types.String `tfsdk:"region"`
	RoutingControlArns types.List   `tfsdk:"routing_control_arns"`
}

// Parallel Configuration Models
type parallelConfigModel struct {
	Step types.List `tfsdk:"step"`
}

type parallelStepModel struct {
	Name                     types.String `tfsdk:"name"`
	ExecutionBlockType       types.String `tfsdk:"execution_block_type"`
	Description              types.String `tfsdk:"description"`
	ExecutionApprovalConfig  types.List   `tfsdk:"execution_approval_config"`
	CustomActionLambdaConfig types.List   `tfsdk:"custom_action_lambda_config"`
}

func (r *resourcePlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	var regions []string
	resp.Diagnostics.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := arcregionswitch.CreatePlanInput{
		Name:             aws.String(plan.Name.ValueString()),
		ExecutionRole:    aws.String(plan.ExecutionRole.ValueString()),
		RecoveryApproach: awstypes.RecoveryApproach(plan.RecoveryApproach.ValueString()),
		Regions:          regions,
	}

	if !plan.Description.IsNull() {
		input.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.PrimaryRegion.IsNull() {
		input.PrimaryRegion = aws.String(plan.PrimaryRegion.ValueString())
	}

	if !plan.RecoveryTimeObjectiveMinutes.IsNull() {
		input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(plan.RecoveryTimeObjectiveMinutes.ValueInt64()))
	}

	// Handle workflows - API requires this field
	if !plan.Workflow.IsNull() {
		var workflows []workflowModel
		resp.Diagnostics.Append(plan.Workflow.ElementsAs(ctx, &workflows, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Workflows = expandWorkflowsFromFramework(workflows)
	}

	// Handle associated alarms
	if !plan.AssociatedAlarms.IsNull() && !plan.AssociatedAlarms.IsUnknown() {
		var alarms []associatedAlarmModel
		resp.Diagnostics.Append(plan.AssociatedAlarms.ElementsAs(ctx, &alarms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.AssociatedAlarms = expandAssociatedAlarmsFromFramework(alarms)
	}

	// Handle tags - use getTagsIn to get all tags including provider defaults
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating ARC Region Switch Plan", err.Error())
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Plan.Arn))
	plan.ID = types.StringValue(aws.ToString(output.Plan.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func getAssociatedAlarmObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                types.StringType,
			"alarm_type":          types.StringType,
			"resource_identifier": types.StringType,
			"cross_account_role":  types.StringType,
			"external_id":         types.StringType,
		},
	}
}

func getWorkflowObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"workflow_target_action": types.StringType,
			"workflow_target_region": types.StringType,
			"workflow_description":   types.StringType,
			"step":                   types.ListType{ElemType: getStepObjectType()},
		},
	}
}

func getStepObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                             types.StringType,
			"execution_block_type":             types.StringType,
			"description":                      types.StringType,
			"execution_approval_config":        types.ListType{ElemType: getExecutionApprovalConfigObjectType()},
			"route53_health_check_config":      types.ListType{ElemType: getRoute53HealthCheckConfigObjectType()},
			"custom_action_lambda_config":      types.ListType{ElemType: getCustomActionLambdaConfigObjectType()},
			"global_aurora_config":             types.ListType{ElemType: getGlobalAuroraConfigObjectType()},
			"ec2_asg_capacity_increase_config": types.ListType{ElemType: getEc2AsgCapacityIncreaseConfigObjectType()},
			"ecs_capacity_increase_config":     types.ListType{ElemType: getEcsCapacityIncreaseConfigObjectType()},
			"eks_resource_scaling_config":      types.ListType{ElemType: getEksResourceScalingConfigObjectType()},
			"arc_routing_control_config":       types.ListType{ElemType: getArcRoutingControlConfigObjectType()},
			"parallel_config":                  types.ListType{ElemType: getParallelConfigObjectType()},
		},
	}
}

func getRoute53HealthCheckConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"hosted_zone_id":     types.StringType,
			"record_name":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
			"timeout_minutes":    types.Int64Type,
			"record_sets":        types.ListType{ElemType: getRecordSetObjectType()},
		},
	}
}

func getRecordSetObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"record_set_identifier": types.StringType,
			"region":                types.StringType,
		},
	}
}

func getExecutionApprovalConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"approval_role":   types.StringType,
			"timeout_minutes": types.Int64Type,
		},
	}
}

func getCustomActionLambdaConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"region_to_run":          types.StringType,
			"retry_interval_minutes": types.Float64Type,
			"timeout_minutes":        types.Int64Type,
			"lambda":                 types.ListType{ElemType: getLambdaObjectType()},
			"ungraceful":             types.ListType{ElemType: getUngracefulObjectType()},
		},
	}
}

func getGlobalAuroraConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"behavior":                  types.StringType,
			"global_cluster_identifier": types.StringType,
			"database_cluster_arns":     types.ListType{ElemType: types.StringType},
			"cross_account_role":        types.StringType,
			"external_id":               types.StringType,
			"timeout_minutes":           types.Int64Type,
			"ungraceful":                types.ListType{ElemType: getGlobalAuroraUngracefulObjectType()},
		},
	}
}

func getGlobalAuroraUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ungraceful": types.StringType,
		},
	}
}

func getEc2AsgCapacityIncreaseConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"asgs":                         types.ListType{ElemType: getAsgObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEc2UngracefulObjectType()},
		},
	}
}

func getAsgObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"arn":                types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getEc2UngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getEcsCapacityIncreaseConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"services":                     types.ListType{ElemType: getServiceObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEcsUngracefulObjectType()},
		},
	}
}

func getServiceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cluster_arn":        types.StringType,
			"service_arn":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getEcsUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getEksResourceScalingConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"kubernetes_resource_type":     types.ListType{ElemType: getKubernetesResourceTypeObjectType()},
			"eks_clusters":                 types.ListType{ElemType: getEksClusterObjectType()},
			"scaling_resources":            types.ListType{ElemType: getScalingResourcesObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEksUngracefulObjectType()},
		},
	}
}

func getKubernetesResourceTypeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"api_version": types.StringType,
			"kind":        types.StringType,
		},
	}
}

func getEksClusterObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cluster_arn":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getScalingResourcesObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"namespace": types.StringType,
			"resources": types.ListType{ElemType: getKubernetesScalingResourceObjectType()},
		},
	}
}

func getKubernetesScalingResourceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"resource_name": types.StringType,
			"name":          types.StringType,
			"namespace":     types.StringType,
			"hpa_name":      types.StringType,
		},
	}
}

func getEksUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getArcRoutingControlConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cross_account_role":          types.StringType,
			"external_id":                 types.StringType,
			"timeout_minutes":             types.Int64Type,
			"region_and_routing_controls": types.ListType{ElemType: getRegionAndRoutingControlsObjectType()},
		},
	}
}

func getRegionAndRoutingControlsObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"region":               types.StringType,
			"routing_control_arns": types.ListType{ElemType: types.StringType},
		},
	}
}

func getParallelConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"step": types.ListType{ElemType: getParallelStepObjectType()},
		},
	}
}

func getParallelStepObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                        types.StringType,
			"execution_block_type":        types.StringType,
			"description":                 types.StringType,
			"execution_approval_config":   types.ListType{ElemType: getExecutionApprovalConfigObjectType()},
			"custom_action_lambda_config": types.ListType{ElemType: getCustomActionLambdaConfigObjectType()},
		},
	}
}

func getLambdaObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"arn":                types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"behavior": types.StringType,
		},
	}
}

func (r *resourcePlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	plan, err := FindPlanByARN(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading ARC Region Switch Plan", err.Error())
		return
	}

	state.ARN = types.StringValue(aws.ToString(plan.Arn))
	state.Name = types.StringValue(aws.ToString(plan.Name))
	state.ExecutionRole = types.StringValue(aws.ToString(plan.ExecutionRole))
	state.RecoveryApproach = types.StringValue(string(plan.RecoveryApproach))

	regions, diags := types.ListValueFrom(ctx, types.StringType, plan.Regions)
	resp.Diagnostics.Append(diags...)
	state.Regions = regions

	if plan.Description != nil {
		state.Description = types.StringValue(aws.ToString(plan.Description))
	} else {
		state.Description = types.StringNull()
	}

	if plan.PrimaryRegion != nil {
		state.PrimaryRegion = types.StringValue(aws.ToString(plan.PrimaryRegion))
	} else {
		state.PrimaryRegion = types.StringNull()
	}

	if plan.RecoveryTimeObjectiveMinutes != nil {
		state.RecoveryTimeObjectiveMinutes = types.Int64Value(int64(aws.ToInt32(plan.RecoveryTimeObjectiveMinutes)))
	} else {
		state.RecoveryTimeObjectiveMinutes = types.Int64Null()
	}

	// Handle associated alarms
	if len(plan.AssociatedAlarms) > 0 {
		alarmElements := make([]attr.Value, 0, len(plan.AssociatedAlarms))
		for alarmName, alarm := range plan.AssociatedAlarms {
			alarmAttrs := map[string]attr.Value{
				"name":                types.StringValue(alarmName),
				"alarm_type":          types.StringValue(string(alarm.AlarmType)),
				"resource_identifier": types.StringValue(aws.ToString(alarm.ResourceIdentifier)),
			}

			if alarm.CrossAccountRole != nil {
				alarmAttrs["cross_account_role"] = types.StringValue(aws.ToString(alarm.CrossAccountRole))
			} else {
				alarmAttrs["cross_account_role"] = types.StringNull()
			}

			if alarm.ExternalId != nil {
				alarmAttrs["external_id"] = types.StringValue(aws.ToString(alarm.ExternalId))
			} else {
				alarmAttrs["external_id"] = types.StringNull()
			}

			alarmObj, alarmDiags := types.ObjectValue(getAssociatedAlarmObjectType().AttrTypes, alarmAttrs)
			resp.Diagnostics.Append(alarmDiags...)
			alarmElements = append(alarmElements, alarmObj)
		}

		alarmSet, alarmSetDiags := types.SetValue(getAssociatedAlarmObjectType(), alarmElements)
		resp.Diagnostics.Append(alarmSetDiags...)
		state.AssociatedAlarms = alarmSet
	} else {
		state.AssociatedAlarms = types.SetNull(getAssociatedAlarmObjectType())
	}

	// Handle workflows
	if len(plan.Workflows) > 0 {
		workflows, diags := flattenWorkflowsToFramework(ctx, plan.Workflows)
		resp.Diagnostics.Append(diags...)
		state.Workflow = workflows
	} else {
		state.Workflow = types.ListNull(getWorkflowObjectType())
	}

	// Handle tags
	tags, err := ListTags(ctx, r.Meta().ARCRegionSwitchClient(ctx), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("listing tags for ARC Region Switch Plan", err.Error())
		return
	}
	setTagsOut(ctx, tags.Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourcePlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	// Convert workflows from Framework List to slice
	var workflows []workflowModel
	resp.Diagnostics.Append(plan.Workflow.ElementsAs(ctx, &workflows, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert associated alarms from Framework Set to slice
	var alarms []associatedAlarmModel
	resp.Diagnostics.Append(plan.AssociatedAlarms.ElementsAs(ctx, &alarms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := arcregionswitch.UpdatePlanInput{
		Arn:              aws.String(state.ID.ValueString()),
		ExecutionRole:    aws.String(plan.ExecutionRole.ValueString()),
		Workflows:        expandWorkflowsFromFramework(workflows),
		AssociatedAlarms: expandAssociatedAlarmsFromFramework(alarms),
	}

	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			input.Description = aws.String("")
		} else {
			input.Description = aws.String(plan.Description.ValueString())
		}
	}

	if !plan.RecoveryTimeObjectiveMinutes.Equal(state.RecoveryTimeObjectiveMinutes) {
		if plan.RecoveryTimeObjectiveMinutes.IsNull() {
			input.RecoveryTimeObjectiveMinutes = aws.Int32(0)
		} else {
			input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(plan.RecoveryTimeObjectiveMinutes.ValueInt64()))
		}
	}

	_, err := conn.UpdatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating ARC Region Switch Plan", err.Error())
		return
	}

	// Handle tags update
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := UpdateTags(ctx, conn, plan.ID.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError("updating tags", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	input := arcregionswitch.DeletePlanInput{
		Arn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeletePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("deleting ARC Region Switch Plan", err.Error())
		return
	}
}

func (r *resourcePlan) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourcePlan) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Basic validation is handled by the schema validators
}

func (r *resourcePlan) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_arcregionswitch_plan"
}

func (r *resourcePlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"execution_role": fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"recovery_approach": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("activeActive", "activePassive"),
				},
			},
			"regions": fwschema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"primary_region": fwschema.StringAttribute{
				Optional: true,
			},
			"recovery_time_objective_minutes": fwschema.Int64Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"associated_alarms": fwschema.SetNestedBlock{
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						names.AttrName: fwschema.StringAttribute{
							Required: true,
						},
						"alarm_type": fwschema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("applicationHealth", "trigger"),
							},
						},
						"resource_identifier": fwschema.StringAttribute{
							Required: true,
						},
						"cross_account_role": fwschema.StringAttribute{
							Optional: true,
						},
						"external_id": fwschema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"workflow": fwschema.ListNestedBlock{
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"workflow_target_action": fwschema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("activate", "deactivate"),
							},
						},
						"workflow_target_region": fwschema.StringAttribute{
							Optional: true,
						},
						"workflow_description": fwschema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"step": fwschema.ListNestedBlock{
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									names.AttrName: fwschema.StringAttribute{
										Required: true,
									},
									"execution_block_type": fwschema.StringAttribute{
										Required: true,
									},
									names.AttrDescription: fwschema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]fwschema.Block{
									"execution_approval_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"approval_role": fwschema.StringAttribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
										},
									},
									"route53_health_check_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"hosted_zone_id": fwschema.StringAttribute{
													Required: true,
												},
												"record_name": fwschema.StringAttribute{
													Required: true,
												},
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												"external_id": fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"record_sets": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"record_set_identifier": fwschema.StringAttribute{
																Required: true,
															},
															"region": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"custom_action_lambda_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"region_to_run": fwschema.StringAttribute{
													Required: true,
												},
												"retry_interval_minutes": fwschema.Float64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"lambda": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrARN: fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"behavior": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"global_aurora_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"behavior": fwschema.StringAttribute{
													Required: true,
												},
												"global_cluster_identifier": fwschema.StringAttribute{
													Required: true,
												},
												"database_cluster_arns": fwschema.ListAttribute{
													ElementType: types.StringType,
													Required:    true,
												},
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												names.AttrExternalID: fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"ungraceful": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"ec2_asg_capacity_increase_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"asgs": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrARN: fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"ecs_capacity_increase_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"services": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"cluster_arn": fwschema.StringAttribute{
																Required: true,
															},
															"service_arn": fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"eks_resource_scaling_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"kubernetes_resource_type": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"api_version": fwschema.StringAttribute{
																Required: true,
															},
															"kind": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"eks_clusters": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"cluster_arn": fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"scaling_resources": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrNamespace: fwschema.StringAttribute{
																Required: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"resources": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"resource_name": fwschema.StringAttribute{
																			Required: true,
																		},
																		names.AttrName: fwschema.StringAttribute{
																			Required: true,
																		},
																		names.AttrNamespace: fwschema.StringAttribute{
																			Required: true,
																		},
																		"hpa_name": fwschema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"arc_routing_control_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												names.AttrExternalID: fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"region_and_routing_controls": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"region": fwschema.StringAttribute{
																Required: true,
															},
															"routing_control_arns": fwschema.ListAttribute{
																ElementType: types.StringType,
																Required:    true,
															},
														},
													},
												},
											},
										},
									},
									"parallel_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Blocks: map[string]fwschema.Block{
												"step": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrName: fwschema.StringAttribute{
																Required: true,
															},
															"execution_block_type": fwschema.StringAttribute{
																Required: true,
															},
															names.AttrDescription: fwschema.StringAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"execution_approval_config": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"approval_role": fwschema.StringAttribute{
																			Required: true,
																		},
																		"timeout_minutes": fwschema.Int64Attribute{
																			Optional: true,
																		},
																	},
																},
															},
															"custom_action_lambda_config": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"region_to_run": fwschema.StringAttribute{
																			Required: true,
																		},
																		"retry_interval_minutes": fwschema.Float64Attribute{
																			Required: true,
																		},
																		"timeout_minutes": fwschema.Int64Attribute{
																			Optional: true,
																		},
																	},
																	Blocks: map[string]fwschema.Block{
																		"lambda": fwschema.ListNestedBlock{
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					names.AttrARN: fwschema.StringAttribute{
																						Required: true,
																					},
																					"cross_account_role": fwschema.StringAttribute{
																						Optional: true,
																					},
																					names.AttrExternalID: fwschema.StringAttribute{
																						Optional: true,
																					},
																				},
																			},
																		},
																		"ungraceful": fwschema.ListNestedBlock{
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					"behavior": fwschema.StringAttribute{
																						Required: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
