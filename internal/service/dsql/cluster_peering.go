// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dsql

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dsql_cluster_peering", name="Cluster Peering")
func newResourceClusterPeering(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceClusterPeering{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameClusterPeering = "Cluster Peering"
)

type resourceClusterPeering struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceClusterPeering) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
			},
			"clusters": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			"witness_region": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceClusterPeering) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var plan resourceClusterPeeringModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var multiRegionProperties awstypes.MultiRegionProperties
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &multiRegionProperties, flex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := dsql.UpdateClusterInput{
		Identifier:            plan.Identifier.ValueStringPointer(),
		MultiRegionProperties: &multiRegionProperties,
	}

	out, err := conn.UpdateCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameClusterPeering, plan.Identifier.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Identifier == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameClusterPeering, plan.Identifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	result, err := waitClusterPeeringCreated(ctx, conn, plan.Identifier.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForCreation, ResNameClusterPeering, plan.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	if result == nil || result.MultiRegionProperties == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameClusterPeering, plan.Identifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	properties := prepareMultiRegionProperties(result)
	resp.Diagnostics.Append(flex.Flatten(ctx, &properties, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceClusterPeering) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var state resourceClusterPeeringModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findClusterByID(ctx, conn, state.Identifier.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionReading, ResNameClusterPeering, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	if out.MultiRegionProperties == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionReading, ResNameClusterPeering, state.Identifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	properties := prepareMultiRegionProperties(out)
	resp.Diagnostics.Append(flex.Flatten(ctx, &properties, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceClusterPeering) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var plan, state resourceClusterPeeringModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input dsql.UpdateClusterInput

		input.Identifier = plan.Identifier.ValueStringPointer()
		input.MultiRegionProperties = &awstypes.MultiRegionProperties{} // TODO: validate the need for this
		resp.Diagnostics.Append(flex.Expand(ctx, plan, input.MultiRegionProperties, flex.WithFieldNamePrefix("Test"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateCluster(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DSQL, create.ErrActionUpdating, ResNameClusterPeering, plan.Identifier.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DSQL, create.ErrActionUpdating, ResNameClusterPeering, plan.Identifier.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitClusterPeeringUpdated(ctx, conn, plan.Identifier.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForUpdate, ResNameClusterPeering, plan.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceClusterPeering) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var state resourceClusterPeeringModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dsql.UpdateClusterInput{
		Identifier:            state.Identifier.ValueStringPointer(),
		MultiRegionProperties: nil,
	}

	_, err := conn.UpdateCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionDeleting, ResNameClusterPeering, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitClusterPeeringDeleted(ctx, conn, state.Identifier.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForDeletion, ResNameClusterPeering, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceClusterPeering) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrIdentifier), req, resp)
}

func waitClusterPeeringCreated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusUpdating, awstypes.ClusterStatusPendingSetup, awstypes.ClusterStatusCreating),
		Target:                    enum.Slice(awstypes.ClusterStatusActive),
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitClusterPeeringUpdated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusUpdating, awstypes.ClusterStatusPendingSetup),
		Target:                    enum.Slice(awstypes.ClusterStatusActive),
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitClusterPeeringDeleted(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusUpdating),
		Target:                    enum.Slice(awstypes.ClusterStatusActive, awstypes.ClusterStatusPendingSetup),
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return out, err
	}

	return nil, err
}

type resourceClusterPeeringModel struct {
	Identifier    types.String        `tfsdk:"identifier"`
	Clusters      fwtypes.SetOfString `tfsdk:"clusters"`
	WitnessRegion types.String        `tfsdk:"witness_region"`
	Timeouts      timeouts.Value      `tfsdk:"timeouts"`
}

func prepareMultiRegionProperties(out *dsql.GetClusterOutput) (properties awstypes.MultiRegionProperties) {
	if out == nil || out.MultiRegionProperties == nil {
		return properties
	}
	properties = *out.MultiRegionProperties // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	if len(properties.Clusters) > 0 {
		clusters := properties.Clusters
		if sourceClusterARN := out.Arn; sourceClusterARN != nil {
			clusters = slices.DeleteFunc(clusters, func(s string) bool {
				return strings.EqualFold(s, aws.ToString(sourceClusterARN))
			})
		}
		properties.Clusters = clusters
	}
	return properties
}
