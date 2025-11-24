// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_express_gateway_service", name="Express Gateway Service")
// @Tags(identifierAttribute="service_arn")
func newResourceExpressGatewayService(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceExpressGatewayService{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameExpressGatewayService = "Express Gateway Service"
)

type resourceExpressGatewayService struct {
	framework.ResourceWithModel[resourceExpressGatewayServiceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceExpressGatewayService) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"active_configurations": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[activeConfigurationModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"service_revision_arn":     types.StringType,
						names.AttrExecutionRoleARN: types.StringType,
						"task_role_arn":            types.StringType,
						"cpu":                      types.StringType,
						"memory":                   types.StringType,
						"health_check_path":        types.StringType,
						names.AttrCreatedAt:        types.StringType,
						names.AttrNetworkConfiguration: types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									names.AttrSecurityGroups: types.SetType{ElemType: types.StringType},
									names.AttrSubnets:        types.SetType{ElemType: types.StringType},
								},
							},
						},
						"primary_container": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"image":          types.StringType,
									"container_port": types.Int64Type,
								},
							},
						},
						"scaling_target": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"min_task_count":            types.Int64Type,
									"max_task_count":            types.Int64Type,
									"auto_scaling_metric":       types.StringType,
									"auto_scaling_target_value": types.Int64Type,
								},
							},
						},
						"ingress_paths": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"access_type":      types.StringType,
									names.AttrEndpoint: types.StringType,
								},
							},
						},
					},
				},
			},
			"cluster": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.StringAttribute{
				Optional: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"current_deployment": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			names.AttrExecutionRoleARN: schema.StringAttribute{
				Required: true,
			},
			"health_check_path": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"infrastructure_role_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"memory": schema.StringAttribute{
				Optional: true,
			},
			"service_arn": framework.ARNAttributeComputedOnly(),
			names.AttrServiceName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[statusModel](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						names.AttrStatusCode:   types.StringType,
						names.AttrStatusReason: types.StringType,
					},
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"task_role_arn": schema.StringAttribute{
				Optional: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_for_steady_state": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroups: schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrSubnets: schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"primary_container": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[primaryContainerModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"command": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"container_port": schema.Int64Attribute{
							Optional: true,
						},
						"image": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"aws_logs_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[awsLogsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group": schema.StringAttribute{
										Required: true,
									},
									"log_stream_prefix": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						names.AttrEnvironment: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[environmentModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"repository_credentials": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[repositoryCredentialsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_parameter": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"secrets": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									"value_from": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"scaling_target": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scalingTargetModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_scaling_metric": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.ExpressGatewayServiceScalingMetric](),
							},
						},
						"auto_scaling_target_value": schema.Int64Attribute{
							Optional: true,
						},
						"max_task_count": schema.Int64Attribute{
							Optional: true,
						},
						"min_task_count": schema.Int64Attribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceExpressGatewayService) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ECSClient(ctx)

	var plan resourceExpressGatewayServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if err := checkExpressGatewayServiceExists(ctx, conn, plan.ServiceName, plan.Cluster); err != nil {
		resp.Diagnostics.AddError("Resource Already Exists", err.Error())
		return
	}

	var input ecs.CreateExpressGatewayServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	operationTime := time.Now().UTC()

	out, err := retryExpressGatewayServiceCreate(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServiceName.String())
		return
	}
	if out == nil || out.Service == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ServiceName.String())
		return
	}

	plan.ServiceArn = flex.StringToFramework(ctx, out.Service.ServiceArn)
	plan.ID = flex.StringToFramework(ctx, out.Service.ServiceArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	var waitOut *awstypes.ECSExpressGatewayService

	if plan.WaitForSteadyState.ValueBool() {
		waitOut, err = waitExpressGatewayServiceStable(ctx, conn, *out.Service.ServiceArn, operationTime, createTimeout)
	} else {
		waitOut, err = waitExpressGatewayServiceActive(ctx, conn, plan.ServiceArn.ValueString(), createTimeout)
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServiceArn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, waitOut.ActiveConfigurations, &plan.ActiveConfigurations))

	// Set Optional+Computed attributes from API response
	if waitOut.Cluster != nil {
		// Preserve cluster format from state (name vs ARN)
		cluster := plan.Cluster.ValueString()
		if arn.IsARN(cluster) {
			plan.Cluster = flex.StringToFramework(ctx, waitOut.Cluster)
		} else {
			plan.Cluster = flex.StringToFramework(ctx, aws.String(clusterNameFromARN(aws.ToString(waitOut.Cluster))))
		}
	}

	if waitOut.CreatedAt != nil {
		plan.CreatedAt = timetypes.NewRFC3339TimeValue(*waitOut.CreatedAt)
	}
	if waitOut.UpdatedAt != nil {
		plan.UpdatedAt = timetypes.NewRFC3339TimeValue(*waitOut.UpdatedAt)
	}

	if waitOut.CurrentDeployment != nil {
		plan.CurrentDeployment = flex.StringToFramework(ctx, waitOut.CurrentDeployment)
	} else {
		plan.CurrentDeployment = types.StringNull()
	}

	plan.Region = types.StringValue(r.Meta().Region(ctx))

	if waitOut.ServiceName != nil {
		plan.ServiceName = flex.StringToFramework(ctx, waitOut.ServiceName)
	}

	if waitOut.Status != nil {
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, []awstypes.ExpressGatewayServiceStatus{*waitOut.Status}, &plan.Status))
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceExpressGatewayService) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ECSClient(ctx)

	var state resourceExpressGatewayServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	serviceArn := state.ServiceArn.ValueString()
	if serviceArn == "" {
		serviceArn = state.ID.ValueString()
	}
	out, err := findExpressGatewayServiceByARN(ctx, conn, serviceArn)
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	if out.Status != nil && out.Status.StatusCode == awstypes.ExpressGatewayServiceStatusCodeInactive {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ServiceArn = flex.StringToFramework(ctx, out.ServiceArn)
	state.ID = flex.StringToFramework(ctx, out.ServiceArn)

	if len(out.ActiveConfigurations) > 0 {
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.ActiveConfigurations, &state.ActiveConfigurations))
	}

	if out.CreatedAt != nil {
		state.CreatedAt = timetypes.NewRFC3339TimeValue(*out.CreatedAt)
	}
	if out.UpdatedAt != nil {
		state.UpdatedAt = timetypes.NewRFC3339TimeValue(*out.UpdatedAt)
	}

	// Set Optional+Computed attributes from API response
	if out.Cluster != nil {
		// Preserve cluster format from state (name vs ARN)
		cluster := state.Cluster.ValueString()
		if cluster == "" {
			// If state cluster is empty, use default cluster name
			state.Cluster = types.StringValue("default")
		} else if arn.IsARN(cluster) {
			state.Cluster = flex.StringToFramework(ctx, out.Cluster)
		} else {
			// Always preserve the cluster name format from state
			state.Cluster = flex.StringToFramework(ctx, aws.String(clusterNameFromARN(aws.ToString(out.Cluster))))
		}
	} else {
		// If API doesn't return cluster, preserve existing state value
		if state.Cluster.IsNull() || state.Cluster.ValueString() == "" {
			state.Cluster = types.StringValue("default")
		}
	}

	if out.CurrentDeployment != nil {
		state.CurrentDeployment = flex.StringToFramework(ctx, out.CurrentDeployment)
	} else {
		state.CurrentDeployment = types.StringNull()
	}

	if out.InfrastructureRoleArn != nil {
		state.InfrastructureRoleArn = flex.StringToFramework(ctx, out.InfrastructureRoleArn)
	}

	if out.ServiceName != nil {
		state.ServiceName = flex.StringToFramework(ctx, out.ServiceName)
	}

	if out.Status != nil {
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, []awstypes.ExpressGatewayServiceStatus{*out.Status}, &state.Status))
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceExpressGatewayService) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ECSClient(ctx)

	var plan, state resourceExpressGatewayServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state, flex.WithIgnoredField("active_configurations"), flex.WithIgnoredField("current_deployment"),
		flex.WithIgnoredField("scaling_target"), flex.WithIgnoredField(names.AttrTags), flex.WithIgnoredField(names.AttrTags),
		flex.WithIgnoredField(names.AttrTagsAll))
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	var operationTime time.Time
	var waitOut *awstypes.ECSExpressGatewayService

	if diff.HasChanges() {
		var input ecs.UpdateExpressGatewayServiceInput

		// ServiceArn is required for the update operation
		input.ServiceArn = plan.ServiceArn.ValueStringPointer()

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		operationTime = time.Now().UTC()

		out, err := retryExpressGatewayServiceUpdate(ctx, conn, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServiceArn.String())
			return
		}
		if out == nil || out.Service == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ServiceArn.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)

		if plan.WaitForSteadyState.ValueBool() {
			waitOut, err = waitExpressGatewayServiceStable(ctx, conn, plan.ServiceArn.ValueString(), operationTime, updateTimeout)
		} else {
			waitOut, err = waitExpressGatewayServiceActive(ctx, conn, plan.ServiceArn.ValueString(), updateTimeout)
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServiceArn.String())
			return
		}
	} else {
		// No changes, just read current state
		var err error
		waitOut, err = findExpressGatewayServiceNoTagsByARN(ctx, conn, plan.ServiceArn.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ServiceArn.String())
			return
		}
	}

	// Set Computed attributes from updated service state
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, waitOut.ActiveConfigurations, &plan.ActiveConfigurations))
	if waitOut.CreatedAt != nil {
		plan.CreatedAt = timetypes.NewRFC3339TimeValue(*waitOut.CreatedAt)
	}
	if waitOut.UpdatedAt != nil {
		plan.UpdatedAt = timetypes.NewRFC3339TimeValue(*waitOut.UpdatedAt)
	}

	// Set Optional+Computed attributes from API response
	if waitOut.Cluster != nil {
		// Preserve cluster format from state (name vs ARN)
		cluster := plan.Cluster.ValueString()
		if arn.IsARN(cluster) {
			plan.Cluster = flex.StringToFramework(ctx, waitOut.Cluster)
		} else {
			plan.Cluster = flex.StringToFramework(ctx, aws.String(clusterNameFromARN(aws.ToString(waitOut.Cluster))))
		}
	}

	if waitOut.CurrentDeployment != nil {
		plan.CurrentDeployment = flex.StringToFramework(ctx, waitOut.CurrentDeployment)
	} else {
		plan.CurrentDeployment = types.StringNull()
	}

	if waitOut.ServiceName != nil {
		plan.ServiceName = flex.StringToFramework(ctx, waitOut.ServiceName)
	}

	if waitOut.Status != nil {
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, []awstypes.ExpressGatewayServiceStatus{*waitOut.Status}, &plan.Status))
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceExpressGatewayService) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ECSClient(ctx)

	var state resourceExpressGatewayServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ecs.DeleteExpressGatewayServiceInput{
		ServiceArn: state.ServiceArn.ValueStringPointer(),
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err := tfresource.RetryWhen(ctx, deleteTimeout,
		func(ctx context.Context) (any, error) {
			return conn.DeleteExpressGatewayService(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.InvalidParameterException](err) || errs.IsA[*awstypes.ServiceNotActiveException](err) {
				return false, err
			}
			return false, err
		},
	)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") ||
			errs.IsAErrorMessageContains[*awstypes.ServiceNotActiveException](err, "Cannot perform this operation on a service in INACTIVE status") {
			// Service was already deleted/inactive
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	_, err = waitExpressGatewayServiceInactive(ctx, conn, state.ServiceArn.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

const (
	gatewayServiceStatusActive   = string(awstypes.ExpressGatewayServiceStatusCodeActive)
	gatewayServiceStatusDraining = string(awstypes.ExpressGatewayServiceStatusCodeDraining)
	gatewayServiceStatusInactive = string(awstypes.ExpressGatewayServiceStatusCodeInactive)

	// Non-standard statuses for statusExpressGatewayServiceWaitForStable().
	gatewayServiceStatusPending = "tfPENDING"
	gatewayServiceStatusStable  = "tfSTABLE"
)

func waitExpressGatewayServiceActive(ctx context.Context, conn *ecs.Client, ARN string, timeout time.Duration) (*awstypes.ECSExpressGatewayService, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{gatewayServiceStatusInactive, gatewayServiceStatusDraining},
		Target:  []string{gatewayServiceStatusActive},
		Refresh: statusExpressGatewayService(ctx, conn, ARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ECSExpressGatewayService); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitExpressGatewayServiceStable(ctx context.Context, conn *ecs.Client, gatewayServiceARN string, operationTime time.Time, timeout time.Duration) (*awstypes.ECSExpressGatewayService, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{gatewayServiceStatusInactive, gatewayServiceStatusDraining, gatewayServiceStatusPending},
		Target:  []string{gatewayServiceStatusActive, gatewayServiceStatusStable},
		Refresh: statusExpressGatewayServiceWaitForStable(ctx, conn, gatewayServiceARN, operationTime),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.ECSExpressGatewayService); ok {
		return output, err
	}

	return nil, err
}

func waitExpressGatewayServiceInactive(ctx context.Context, conn *ecs.Client, id string, timeout time.Duration) (*awstypes.ECSExpressGatewayService, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{gatewayServiceStatusActive},
		Target:     []string{gatewayServiceStatusInactive, gatewayServiceStatusDraining},
		Refresh:    statusExpressGatewayServiceForDeletion(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ECSExpressGatewayService); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusExpressGatewayService(ctx context.Context, conn *ecs.Client, gatewayServiceARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findExpressGatewayServiceNoTagsByARN(ctx, conn, gatewayServiceARN)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func statusExpressGatewayServiceForDeletion(ctx context.Context, conn *ecs.Client, gatewayServiceARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findExpressGatewayServiceNoTagsByARN(ctx, conn, gatewayServiceARN)
		if err != nil {
			if tfresource.NotFound(err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") ||
				errs.IsAErrorMessageContains[*awstypes.ServiceNotActiveException](err, "Cannot perform this operation on a service in INACTIVE status") {
				mockService := &awstypes.ECSExpressGatewayService{
					ServiceArn: aws.String(gatewayServiceARN),
					Status: &awstypes.ExpressGatewayServiceStatus{
						StatusCode: awstypes.ExpressGatewayServiceStatusCodeInactive,
					},
				}
				return mockService, gatewayServiceStatusInactive, nil
			}
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func findExpressGatewayServiceByARN(ctx context.Context, conn *ecs.Client, ARN string) (*awstypes.ECSExpressGatewayService, error) {
	input := &ecs.DescribeExpressGatewayServiceInput{
		ServiceArn: aws.String(ARN),
		Include:    []awstypes.ExpressGatewayServiceInclude{awstypes.ExpressGatewayServiceIncludeTags},
	}

	output, err := findExpressGatewayService(ctx, conn, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(partitionFromConn(conn), err) {
		input.Include = nil

		output, err = findExpressGatewayService(ctx, conn, input)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findExpressGatewayServiceNoTagsByARN(ctx context.Context, conn *ecs.Client, ARN string) (*awstypes.ECSExpressGatewayService, error) {
	input := &ecs.DescribeExpressGatewayServiceInput{
		ServiceArn: aws.String(ARN),
	}

	return findExpressGatewayService(ctx, conn, input)
}

func findExpressGatewayService(ctx context.Context, conn *ecs.Client, input *ecs.DescribeExpressGatewayServiceInput) (*awstypes.ECSExpressGatewayService, error) {
	out, err := conn.DescribeExpressGatewayService(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Service == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(input))
	}

	return out.Service, nil
}

func checkExpressGatewayServiceExists(ctx context.Context, conn *ecs.Client, serviceName, cluster types.String) error {
	if serviceName.IsNull() || serviceName.IsUnknown() {
		return nil
	}

	clusterName := cluster.ValueString()
	if clusterName == "" {
		clusterName = "default"
	}

	_, err := findServiceNoTagsByTwoPartKey(ctx, conn, serviceName.ValueString(), clusterName)
	if err == nil {
		return fmt.Errorf("Express Gateway Service %s already exists in cluster %s", serviceName.ValueString(), clusterName)
	}
	if tfresource.NotFound(err) {
		return nil
	}
	return err
}

func statusExpressGatewayServiceWaitForStable(ctx context.Context, conn *ecs.Client, gatewayServiceARN string, operationTime time.Time) retry.StateRefreshFunc {
	var deploymentArn *string

	return func() (any, string, error) {
		outputRaw, serviceStatus, err := statusExpressGatewayService(ctx, conn, gatewayServiceARN)()
		if err != nil {
			return nil, "", err
		}

		if serviceStatus != gatewayServiceStatusActive {
			return outputRaw, serviceStatus, nil
		}

		output := outputRaw.(*awstypes.ECSExpressGatewayService)

		if deploymentArn == nil && output.CurrentDeployment != nil {
			deploymentArn = output.CurrentDeployment
		} else {
			input := &ecs.ListServiceDeploymentsInput{
				Cluster: output.Cluster,
				Service: output.ServiceName,
				CreatedAt: &awstypes.CreatedAt{
					After: &operationTime,
				},
			}

			listServiceDeploymentsOutput, err := conn.ListServiceDeployments(ctx, input)
			if err != nil {
				return nil, "Error getting latest deployment.", err
			}

			if len(listServiceDeploymentsOutput.ServiceDeployments) > 0 {
				deploymentArn = listServiceDeploymentsOutput.ServiceDeployments[0].ServiceDeploymentArn
			}
		}

		if deploymentArn != nil {
			deploymentStatus, err := findDeploymentStatus(ctx, conn, *deploymentArn)
			if err != nil {
				return nil, "", err
			}
			return output, deploymentStatus, nil
		}

		return output, gatewayServiceStatusPending, nil
	}
}

type resourceExpressGatewayServiceModel struct {
	framework.WithRegionModel
	ActiveConfigurations fwtypes.ListNestedObjectValueOf[activeConfigurationModel] `tfsdk:"active_configurations"`
	Cluster              types.String                                              `tfsdk:"cluster"`
	CPU                  types.String                                              `tfsdk:"cpu"`
	CreatedAt            timetypes.RFC3339                                         `tfsdk:"created_at"`
	CurrentDeployment    types.String                                              `tfsdk:"current_deployment"`

	ExecutionRoleArn      types.String                                               `tfsdk:"execution_role_arn"`
	HealthCheckPath       types.String                                               `tfsdk:"health_check_path"`
	ID                    types.String                                               `tfsdk:"id"`
	InfrastructureRoleArn types.String                                               `tfsdk:"infrastructure_role_arn"`
	Memory                types.String                                               `tfsdk:"memory"`
	NetworkConfiguration  fwtypes.ListNestedObjectValueOf[networkConfigurationModel] `tfsdk:"network_configuration"`
	PrimaryContainer      fwtypes.ListNestedObjectValueOf[primaryContainerModel]     `tfsdk:"primary_container"`
	ScalingTarget         fwtypes.ListNestedObjectValueOf[scalingTargetModel]        `tfsdk:"scaling_target"`
	ServiceArn            types.String                                               `tfsdk:"service_arn"`
	ServiceName           types.String                                               `tfsdk:"service_name"`
	Status                fwtypes.ListNestedObjectValueOf[statusModel]               `tfsdk:"status"`
	Tags                  tftags.Map                                                 `tfsdk:"tags"`
	TagsAll               tftags.Map                                                 `tfsdk:"tags_all"`
	TaskRoleArn           types.String                                               `tfsdk:"task_role_arn"`
	Timeouts              timeouts.Value                                             `tfsdk:"timeouts"`
	UpdatedAt             timetypes.RFC3339                                          `tfsdk:"updated_at"`
	WaitForSteadyState    types.Bool                                                 `tfsdk:"wait_for_steady_state"`
}

type networkConfigurationModel struct {
	SecurityGroups fwtypes.SetValueOf[types.String] `tfsdk:"security_groups"`
	Subnets        fwtypes.SetValueOf[types.String] `tfsdk:"subnets"`
}

type primaryContainerModel struct {
	AwsLogsConfiguration  fwtypes.ListNestedObjectValueOf[awsLogsConfigurationModel]  `tfsdk:"aws_logs_configuration"`
	Command               fwtypes.ListValueOf[types.String]                           `tfsdk:"command"`
	ContainerPort         types.Int64                                                 `tfsdk:"container_port"`
	Environment           fwtypes.ListNestedObjectValueOf[environmentModel]           `tfsdk:"environment"`
	Image                 types.String                                                `tfsdk:"image"`
	RepositoryCredentials fwtypes.ListNestedObjectValueOf[repositoryCredentialsModel] `tfsdk:"repository_credentials"`
	Secrets               fwtypes.ListNestedObjectValueOf[secretModel]                `tfsdk:"secrets"`
}

type awsLogsConfigurationModel struct {
	LogGroup        types.String `tfsdk:"log_group"`
	LogStreamPrefix types.String `tfsdk:"log_stream_prefix"`
}

type environmentModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type repositoryCredentialsModel struct {
	CredentialsParameter types.String `tfsdk:"credentials_parameter"`
}

type secretModel struct {
	Name      types.String `tfsdk:"name"`
	ValueFrom types.String `tfsdk:"value_from"`
}

type scalingTargetModel struct {
	AutoScalingMetric      types.String `tfsdk:"auto_scaling_metric"`
	AutoScalingTargetValue types.Int64  `tfsdk:"auto_scaling_target_value"`
	MaxTaskCount           types.Int64  `tfsdk:"max_task_count"`
	MinTaskCount           types.Int64  `tfsdk:"min_task_count"`
}

type activeConfigurationModel struct {
	ServiceRevisionArn   types.String                                                     `tfsdk:"service_revision_arn"`
	ExecutionRoleArn     types.String                                                     `tfsdk:"execution_role_arn"`
	TaskRoleArn          types.String                                                     `tfsdk:"task_role_arn"`
	Cpu                  types.String                                                     `tfsdk:"cpu"`
	Memory               types.String                                                     `tfsdk:"memory"`
	HealthCheckPath      types.String                                                     `tfsdk:"health_check_path"`
	CreatedAt            timetypes.RFC3339                                                `tfsdk:"created_at"`
	NetworkConfiguration fwtypes.ListNestedObjectValueOf[configNetworkConfigurationModel] `tfsdk:"network_configuration"`
	PrimaryContainer     fwtypes.ListNestedObjectValueOf[configPrimaryContainerModel]     `tfsdk:"primary_container"`
	ScalingTarget        fwtypes.ListNestedObjectValueOf[configScalingTargetModel]        `tfsdk:"scaling_target"`
	IngressPaths         fwtypes.ListNestedObjectValueOf[ingressPathModel]                `tfsdk:"ingress_paths"`
}

type configNetworkConfigurationModel struct {
	SecurityGroups fwtypes.SetValueOf[types.String] `tfsdk:"security_groups"`
	Subnets        fwtypes.SetValueOf[types.String] `tfsdk:"subnets"`
}

type configPrimaryContainerModel struct {
	Image         types.String `tfsdk:"image"`
	ContainerPort types.Int64  `tfsdk:"container_port"`
}

type configScalingTargetModel struct {
	MinTaskCount           types.Int64  `tfsdk:"min_task_count"`
	MaxTaskCount           types.Int64  `tfsdk:"max_task_count"`
	AutoScalingMetric      types.String `tfsdk:"auto_scaling_metric"`
	AutoScalingTargetValue types.Int64  `tfsdk:"auto_scaling_target_value"`
}

type ingressPathModel struct {
	AccessType types.String `tfsdk:"access_type"`
	Endpoint   types.String `tfsdk:"endpoint"`
}

type statusModel struct {
	StatusCode   types.String `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func retryExpressGatewayServiceCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateExpressGatewayServiceInput) (*ecs.CreateExpressGatewayServiceOutput, error) {
	const (
		serviceCreateTimeout = 2 * time.Minute
		timeout              = propagationTimeout + serviceCreateTimeout
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateExpressGatewayService(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Cannot assume role") {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		return nil, err
	}
	return outputRaw.(*ecs.CreateExpressGatewayServiceOutput), nil
}

func retryExpressGatewayServiceUpdate(ctx context.Context, conn *ecs.Client, input *ecs.UpdateExpressGatewayServiceInput) (*ecs.UpdateExpressGatewayServiceOutput, error) {
	const (
		serviceUpdateTimeout = 2 * time.Minute
		timeout              = propagationTimeout + serviceUpdateTimeout
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.UpdateExpressGatewayService(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Cannot assume role") {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		return nil, err
	}
	return outputRaw.(*ecs.UpdateExpressGatewayServiceOutput), nil
}
