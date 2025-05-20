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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dsql_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dsql;dsql.GetClusterOutput")
func newResourceCluster(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCluster{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCluster = "Cluster"
)

type resourceCluster struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:        framework.ARNAttributeComputedOnly(),
			names.AttrIdentifier: framework.IDAttribute(),
			names.AttrTags:       tftags.TagsAttribute(),
			names.AttrTagsAll:    tftags.TagsAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"multi_region_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiRegionPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"clusters": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Computed:    true,
							Optional:    true,
						},
						"witness_region": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
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

func (r *resourceCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var plan resourceClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input dsql.CreateClusterInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameCluster, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Identifier == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameCluster, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Identifier = types.StringPointerValue(out.Identifier)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitClusterCreated(ctx, conn, plan.Identifier.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForCreation, ResNameCluster, plan.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	if err := createTags(ctx, conn, plan.ARN.ValueString(), getTagsIn(ctx)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionCreating, ResNameCluster, plan.Identifier.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var state resourceClusterModel
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
			create.ProblemStandardMessage(names.DSQL, create.ErrActionReading, ResNameCluster, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	if sourceClusterARN := out.Arn; sourceClusterARN != nil && out.MultiRegionProperties != nil {
		// Remove the current cluster from the list of clusters in the multi-region properties
		// This is needed because one of the ARNs of the clusters in the multi-region properties is
		// the same as the ARN of this specific cluster, and we need to remove it from the
		// list of clusters to avoid a conflict when updating the resource

		clusters := out.MultiRegionProperties.Clusters
		clusters = slices.DeleteFunc(clusters, func(s string) bool {
			return strings.EqualFold(s, aws.ToString(sourceClusterARN))
		})
		out.MultiRegionProperties.Clusters = clusters
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var plan, state resourceClusterModel
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
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateCluster(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DSQL, create.ErrActionUpdating, ResNameCluster, plan.Identifier.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DSQL, create.ErrActionUpdating, ResNameCluster, plan.Identifier.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitClusterUpdated(ctx, conn, plan.Identifier.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForUpdate, ResNameCluster, plan.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DSQLClient(ctx)

	var state resourceClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dsql.DeleteClusterInput{
		Identifier: state.Identifier.ValueStringPointer(),
	}

	_, err := conn.DeleteCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionDeleting, ResNameCluster, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitClusterDeleted(ctx, conn, state.Identifier.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DSQL, create.ErrActionWaitingForDeletion, ResNameCluster, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrIdentifier), req, resp)
}

func waitClusterCreated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusCreating),
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

func waitClusterUpdated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusUpdating),
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

func waitClusterDeleted(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClusterStatusDeleting, awstypes.ClusterStatusPendingDelete),
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func statusCluster(ctx context.Context, conn *dsql.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findClusterByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findClusterByID(ctx context.Context, conn *dsql.Client, id string) (*dsql.GetClusterOutput, error) {
	input := dsql.GetClusterInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceClusterModel struct {
	ARN                       types.String                                                `tfsdk:"arn"`
	Identifier                types.String                                                `tfsdk:"identifier"`
	DeletionProtectionEnabled types.Bool                                                  `tfsdk:"deletion_protection_enabled"`
	Timeouts                  timeouts.Value                                              `tfsdk:"timeouts"`
	MultiRegionProperties     fwtypes.ListNestedObjectValueOf[multiRegionPropertiesModel] `tfsdk:"multi_region_properties"`
	Tags                      tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                  `tfsdk:"tags_all"`
}

type multiRegionPropertiesModel struct {
	Clusters      fwtypes.SetOfString `tfsdk:"clusters"`
	WitnessRegion types.String        `tfsdk:"witness_region"`
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := dsql.ListClustersInput{}
	conn := client.DSQLClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := dsql.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCluster, client,
				sweepfw.NewAttribute(names.AttrIdentifier, aws.ToString(v.Identifier))),
			)
		}
	}

	return sweepResources, nil
}
