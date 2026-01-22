// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_express_gateway_service", name="Express Gateway Service")
// @Tags(identifierAttribute="service_arn")
func newExpressGatewayServiceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &expressGatewayServiceResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type expressGatewayServiceResource struct {
	framework.ResourceWithModel[expressGatewayServiceResourceModel]
	framework.WithTimeouts
}

func (r *expressGatewayServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.StringAttribute{
				Optional: true,
				Computed: true,
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
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"health_check_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"infrastructure_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ingress_paths": framework.ResourceComputedListOfObjectsAttribute[ingressPathSummaryModel](ctx),
			"memory": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrNetworkConfiguration: framework.ResourceOptionalComputedListOfObjectsAttribute[expressGatewayServiceNetworkConfigurationModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			"scaling_target":               framework.ResourceOptionalComputedListOfObjectsAttribute[expressGatewayScalingTargetModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			"service_arn":                  framework.ARNAttributeComputedOnly(),
			names.AttrServiceName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"service_revision_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"task_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"wait_for_steady_state": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"primary_container": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[expressGatewayContainerModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"aws_logs_configuration": framework.ResourceOptionalComputedListOfObjectsAttribute[expressGatewayServiceAWSLogsConfigurationModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
						"command": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"container_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"image": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrEnvironment: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyValuePairModel](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[expressGatewayRepositoryCredentialsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_parameter": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"secret": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									"value_from": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
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

func (r *expressGatewayServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan expressGatewayServiceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	if err := checkExpressGatewayServiceExists(ctx, conn, plan.ServiceName, plan.Cluster); err != nil {
		resp.Diagnostics.AddError("Resource Already Exists", err.Error())
		return
	}

	var input ecs.CreateExpressGatewayServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
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

	serviceARN := aws.ToString(out.Service.ServiceArn)
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	var waitOut *awstypes.ECSExpressGatewayService

	if plan.WaitForSteadyState.ValueBool() {
		waitOut, err = waitExpressGatewayServiceStable(ctx, conn, serviceARN, operationTime, createTimeout)
	} else {
		waitOut, err = waitExpressGatewayServiceActive(ctx, conn, serviceARN, createTimeout)
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
		return
	}

	var cluster string
	// Preserve cluster format from state (name vs ARN)
	if waitOut.Cluster != nil {
		cluster = fwflex.StringValueFromFramework(ctx, plan.Cluster)
		if arn.IsARN(cluster) {
			cluster = aws.ToString(waitOut.Cluster)
		} else {
			cluster = clusterNameFromARN(aws.ToString(waitOut.Cluster))
		}
	}

	// Set values for unknowns.
	if len(waitOut.ActiveConfigurations) > 0 {
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, waitOut.ActiveConfigurations[0], &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plan.Cluster = fwflex.StringValueToFramework(ctx, cluster)
	plan.CurrentDeployment = fwflex.StringToFramework(ctx, waitOut.CurrentDeployment)
	plan.ServiceARN = fwflex.StringValueToFramework(ctx, serviceARN)
	plan.ServiceName = fwflex.StringToFramework(ctx, waitOut.ServiceName)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *expressGatewayServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state expressGatewayServiceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	serviceARN := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	out, err := findExpressGatewayServiceByARN(ctx, conn, serviceARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
		return
	}

	if out.Status != nil && (out.Status.StatusCode == awstypes.ExpressGatewayServiceStatusCodeInactive ||
		out.Status.StatusCode == awstypes.ExpressGatewayServiceStatusCodeDraining) {
		resp.State.RemoveResource(ctx)
		return
	}

	if len(out.ActiveConfigurations) > 0 {
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.ActiveConfigurations[0], &state))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set Optional+Computed attributes from API response
	if out.Cluster != nil {
		// Preserve cluster format from state (name vs ARN)
		cluster := state.Cluster.ValueString()
		if cluster == "" {
			// If state cluster is empty, use default cluster name
			state.Cluster = types.StringValue("default")
		} else if arn.IsARN(cluster) {
			state.Cluster = fwflex.StringToFramework(ctx, out.Cluster)
		} else {
			// Always preserve the cluster name format from state
			state.Cluster = fwflex.StringToFramework(ctx, aws.String(clusterNameFromARN(aws.ToString(out.Cluster))))
		}
	} else {
		// If API doesn't return cluster, preserve existing state value
		if state.Cluster.IsNull() || state.Cluster.ValueString() == "" {
			state.Cluster = types.StringValue("default")
		}
	}
	state.CurrentDeployment = fwflex.StringToFramework(ctx, out.CurrentDeployment)
	state.InfrastructureRoleARN = fwflex.StringToFrameworkARN(ctx, out.InfrastructureRoleArn)
	state.ServiceName = fwflex.StringToFramework(ctx, out.ServiceName)

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *expressGatewayServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state expressGatewayServiceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state, fwflex.WithIgnoredField("active_configurations"), fwflex.WithIgnoredField("current_deployment"),
		fwflex.WithIgnoredField("scaling_target"), fwflex.WithIgnoredField(names.AttrTags), fwflex.WithIgnoredField(names.AttrTags),
		fwflex.WithIgnoredField(names.AttrTagsAll))
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceARN := fwflex.StringValueFromFramework(ctx, plan.ServiceARN)
	var operationTime time.Time
	var waitOut *awstypes.ECSExpressGatewayService

	if diff.HasChanges() {
		var input ecs.UpdateExpressGatewayServiceInput

		// ServiceArn is required for the update operation
		input.ServiceArn = aws.String(serviceARN)

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		operationTime = time.Now().UTC()

		out, err := retryExpressGatewayServiceUpdate(ctx, conn, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
			return
		}
		if out == nil || out.Service == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, serviceARN)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)

		if plan.WaitForSteadyState.ValueBool() {
			waitOut, err = waitExpressGatewayServiceStable(ctx, conn, serviceARN, operationTime, updateTimeout)
		} else {
			waitOut, err = waitExpressGatewayServiceActive(ctx, conn, serviceARN, updateTimeout)
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
			return
		}
	} else {
		// No changes, just read current state
		var err error
		waitOut, err = findExpressGatewayServiceNoTagsByARN(ctx, conn, serviceARN)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
			return
		}
	}

	// Set values for unknowns.
	if len(waitOut.ActiveConfigurations) > 0 {
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, waitOut.ActiveConfigurations[0], &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set Optional+Computed attributes from API response
	if waitOut.Cluster != nil {
		// Preserve cluster format from state (name vs ARN)
		cluster := plan.Cluster.ValueString()
		if arn.IsARN(cluster) {
			plan.Cluster = fwflex.StringToFramework(ctx, waitOut.Cluster)
		} else {
			plan.Cluster = fwflex.StringToFramework(ctx, aws.String(clusterNameFromARN(aws.ToString(waitOut.Cluster))))
		}
	}

	plan.CurrentDeployment = fwflex.StringToFramework(ctx, waitOut.CurrentDeployment)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *expressGatewayServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state expressGatewayServiceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	serviceARN := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	input := ecs.DeleteExpressGatewayServiceInput{
		ServiceArn: aws.String(serviceARN),
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	// Try to delete the service once
	_, err := conn.DeleteExpressGatewayService(ctx, &input)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") ||
			errs.IsAErrorMessageContains[*awstypes.ServiceNotActiveException](err, "Cannot perform this operation on a service in INACTIVE status") ||
			errs.IsAErrorMessageContains[*awstypes.ServiceNotActiveException](err, "Service is in DRAINING status") {
			// Service was already deleted/inactive/draining - deletion is already in progress or complete
			return
		} else {
			// Real error occurred
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
			return
		}
	}

	_, err = waitExpressGatewayServiceInactive(ctx, conn, serviceARN, deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceARN)
		return
	}
}

func (r *expressGatewayServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_arn"), req, resp)
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
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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

func statusExpressGatewayService(ctx context.Context, conn *ecs.Client, gatewayServiceARN string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findExpressGatewayServiceNoTagsByARN(ctx, conn, gatewayServiceARN)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func statusExpressGatewayServiceForDeletion(ctx context.Context, conn *ecs.Client, gatewayServiceARN string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findExpressGatewayServiceNoTagsByARN(ctx, conn, gatewayServiceARN)
		if err != nil {
			if retry.NotFound(err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") ||
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
			return nil, smarterr.NewError(&sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Resource not found") {
			return nil, smarterr.NewError(&sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Service == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
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
	if retry.NotFound(err) {
		return nil
	}
	return err
}

func statusExpressGatewayServiceWaitForStable(ctx context.Context, conn *ecs.Client, gatewayServiceARN string, operationTime time.Time) sdkretry.StateRefreshFunc {
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

type expressGatewayServiceResourceModel struct {
	framework.WithRegionModel
	Cluster               types.String                                                                    `tfsdk:"cluster"`
	CPU                   types.String                                                                    `tfsdk:"cpu"`
	CurrentDeployment     types.String                                                                    `tfsdk:"current_deployment"`
	ExecutionRoleARN      fwtypes.ARN                                                                     `tfsdk:"execution_role_arn"`
	HealthCheckPath       types.String                                                                    `tfsdk:"health_check_path"`
	InfrastructureRoleARN fwtypes.ARN                                                                     `tfsdk:"infrastructure_role_arn"`
	IngressPaths          fwtypes.ListNestedObjectValueOf[ingressPathSummaryModel]                        `tfsdk:"ingress_paths"`
	Memory                types.String                                                                    `tfsdk:"memory"`
	NetworkConfiguration  fwtypes.ListNestedObjectValueOf[expressGatewayServiceNetworkConfigurationModel] `tfsdk:"network_configuration"`
	PrimaryContainer      fwtypes.ListNestedObjectValueOf[expressGatewayContainerModel]                   `tfsdk:"primary_container"`
	ScalingTarget         fwtypes.ListNestedObjectValueOf[expressGatewayScalingTargetModel]               `tfsdk:"scaling_target"`
	ServiceARN            types.String                                                                    `tfsdk:"service_arn"`
	ServiceName           types.String                                                                    `tfsdk:"service_name"`
	ServiceRevisionARN    types.String                                                                    `tfsdk:"service_revision_arn"`
	Tags                  tftags.Map                                                                      `tfsdk:"tags"`
	TagsAll               tftags.Map                                                                      `tfsdk:"tags_all"`
	TaskRoleARN           fwtypes.ARN                                                                     `tfsdk:"task_role_arn"`
	Timeouts              timeouts.Value                                                                  `tfsdk:"timeouts"`
	WaitForSteadyState    types.Bool                                                                      `tfsdk:"wait_for_steady_state"`
}

type expressGatewayServiceNetworkConfigurationModel struct {
	SecurityGroups fwtypes.SetOfString `tfsdk:"security_groups"`
	Subnets        fwtypes.SetOfString `tfsdk:"subnets"`
}

type expressGatewayContainerModel struct {
	AWSLogsConfiguration  fwtypes.ListNestedObjectValueOf[expressGatewayServiceAWSLogsConfigurationModel] `tfsdk:"aws_logs_configuration"`
	Command               fwtypes.ListOfString                                                            `tfsdk:"command"`
	ContainerPort         types.Int64                                                                     `tfsdk:"container_port"`
	Environment           fwtypes.ListNestedObjectValueOf[keyValuePairModel]                              `tfsdk:"environment"`
	Image                 types.String                                                                    `tfsdk:"image"`
	RepositoryCredentials fwtypes.ListNestedObjectValueOf[expressGatewayRepositoryCredentialsModel]       `tfsdk:"repository_credentials"`
	Secrets               fwtypes.ListNestedObjectValueOf[secretModel]                                    `tfsdk:"secret"`
}

type expressGatewayServiceAWSLogsConfigurationModel struct {
	LogGroup        types.String `tfsdk:"log_group"`
	LogStreamPrefix types.String `tfsdk:"log_stream_prefix"`
}

type keyValuePairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type expressGatewayRepositoryCredentialsModel struct {
	CredentialsParameter fwtypes.ARN `tfsdk:"credentials_parameter"`
}

type secretModel struct {
	Name      types.String `tfsdk:"name"`
	ValueFrom fwtypes.ARN  `tfsdk:"value_from"`
}

type expressGatewayScalingTargetModel struct {
	AutoScalingMetric      fwtypes.StringEnum[awstypes.ExpressGatewayServiceScalingMetric] `tfsdk:"auto_scaling_metric"`
	AutoScalingTargetValue types.Int64                                                     `tfsdk:"auto_scaling_target_value"`
	MaxTaskCount           types.Int64                                                     `tfsdk:"max_task_count"`
	MinTaskCount           types.Int64                                                     `tfsdk:"min_task_count"`
}

type ingressPathSummaryModel struct {
	AccessType fwtypes.StringEnum[awstypes.AccessType] `tfsdk:"access_type"`
	Endpoint   types.String                            `tfsdk:"endpoint"`
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
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Cannot assume role") ||
				errs.IsAErrorMessageContains[*awstypes.ClientException](err, "AWS was not able to validate the provided access credentials") ||
				errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform: sts:AssumeRole") {
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
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Cannot assume role") ||
				errs.IsAErrorMessageContains[*awstypes.ClientException](err, "AWS was not able to validate the provided access credentials") ||
				errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform: sts:AssumeRole") {
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
