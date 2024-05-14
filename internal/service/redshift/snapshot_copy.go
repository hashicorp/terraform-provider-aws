// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Snapshot Copy")
func newResourceSnapshotCopy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSnapshotCopy{}, nil
}

const (
	ResNameSnapshotCopy = "Snapshot Copy"
)

type resourceSnapshotCopy struct {
	framework.ResourceWithConfigure
}

func (r *resourceSnapshotCopy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_redshift_snapshot_copy"
}

func (r *resourceSnapshotCopy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClusterIdentifier: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"manual_snapshot_retention_period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRetentionPeriod: schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_copy_grant_name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceSnapshotCopy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan resourceSnapshotCopyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(plan.ClusterIdentifier.ValueString())

	in := &redshift.EnableSnapshotCopyInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.EnableSnapshotCopy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameSnapshotCopy, plan.ClusterIdentifier.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Cluster == nil || out.Cluster.ClusterSnapshotCopyStatus == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameSnapshotCopy, plan.ClusterIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.Cluster.ClusterSnapshotCopyStatus, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSnapshotCopy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state resourceSnapshotCopyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSnapshotCopyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionSetting, ResNameSnapshotCopy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSnapshotCopy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan, state resourceSnapshotCopyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RetentionPeriod.Equal(state.RetentionPeriod) {
		in := &redshift.ModifySnapshotCopyRetentionPeriodInput{
			ClusterIdentifier: aws.String(plan.ClusterIdentifier.ValueString()),
			RetentionPeriod:   aws.Int32(int32(plan.RetentionPeriod.ValueInt64())),
		}

		out, err := conn.ModifySnapshotCopyRetentionPeriod(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionUpdating, ResNameSnapshotCopy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Cluster == nil || out.Cluster.ClusterSnapshotCopyStatus == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionUpdating, ResNameSnapshotCopy, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.Cluster.ClusterSnapshotCopyStatus, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSnapshotCopy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state resourceSnapshotCopyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshift.DisableSnapshotCopyInput{
		ClusterIdentifier: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DisableSnapshotCopy(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			return
		} else if errs.IsA[*awstypes.SnapshotCopyAlreadyDisabledFault](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionDeleting, ResNameSnapshotCopy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceSnapshotCopy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrClusterIdentifier), req.ID)...)
}

func findSnapshotCopyByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.ClusterSnapshotCopyStatus, error) {
	in := &redshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(id),
	}

	out, err := conn.DescribeClusters(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	// API should return a ClusterNotFound fault in this case, but check length for
	// extra safety
	if len(out.Clusters) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   errors.New("not found"),
			LastRequest: in,
		}
	}
	if out.Clusters[0].ClusterSnapshotCopyStatus == nil {
		return nil, &retry.NotFoundError{
			LastError:   errors.New("snapshot copy not enabled"),
			LastRequest: in,
		}
	}

	return out.Clusters[0].ClusterSnapshotCopyStatus, nil
}

type resourceSnapshotCopyData struct {
	ID                            types.String `tfsdk:"id"`
	ClusterIdentifier             types.String `tfsdk:"cluster_identifier"`
	DestinationRegion             types.String `tfsdk:"destination_region"`
	ManualSnapshotRetentionPeriod types.Int64  `tfsdk:"manual_snapshot_retention_period"`
	RetentionPeriod               types.Int64  `tfsdk:"retention_period"`
	SnapshotCopyGrantName         types.String `tfsdk:"snapshot_copy_grant_name"`
}
