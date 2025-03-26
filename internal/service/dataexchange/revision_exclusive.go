// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dataexchange_revision_exclusive", name="Revision Exclusive")
// @Tags(identifierAttribute="arn")
func newResourceRevisionExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRevisionExclusive{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	// r.SetDefaultUpdateTimeout(30 * time.Minute)
	// r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRevisionExclusive = "Revision Exclusive"
)

type resourceRevisionExclusive struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpUpdate[resourceRevisionExclusive]
}

func (r *resourceRevisionExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 16_348),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"asset": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[assetModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						names.AttrCreatedAt: schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrID: framework.IDAttribute(),
						names.AttrName: schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"updated_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"import_assets_from_s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[importAssetsFromS3Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"asset_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[assetSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												names.AttrKey: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceRevisionExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan resourceRevisionExclusiveModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input dataexchange.CreateRevisionInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRevision(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assets := make([]assetModel, len(plan.Assets.Elements()))
	existingAssetIDs := make([]string, 0, len(plan.Assets.Elements()))
	for i, asset := range nestedObjectCollectionAllMust(ctx, plan.Assets) {
		importAssetsFromS3, d := asset.ImportAssetsFromS3.ToPtr(ctx)
		resp.Diagnostics.Append(d...)
		if d.HasError() {
			return
		}

		var importAssetsFromS3RequestDetails awstypes.ImportAssetsFromS3RequestDetails
		resp.Diagnostics.Append(flex.Expand(ctx, importAssetsFromS3, &importAssetsFromS3RequestDetails)...)
		if resp.Diagnostics.HasError() {
			return
		}
		importAssetsFromS3RequestDetails.DataSetId = plan.DataSetID.ValueStringPointer()
		importAssetsFromS3RequestDetails.RevisionId = plan.ID.ValueStringPointer()

		requestDetails := awstypes.RequestDetails{
			ImportAssetsFromS3: &importAssetsFromS3RequestDetails,
		}
		createJobInput := dataexchange.CreateJobInput{
			Type:    awstypes.TypeImportAssetsFromS3,
			Details: &requestDetails,
		}
		createJobOutput, err := conn.CreateJob(ctx, &createJobInput)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", err),
				err.Error(),
			)
			return
		}
		if createJobOutput == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", nil),
				errors.New("empty output").Error(),
			)
			return
		}

		startJobInput := dataexchange.StartJobInput{
			JobId: createJobOutput.Id,
		}
		_, err = conn.StartJob(ctx, &startJobInput)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", err),
				err.Error(),
			)
			return
		}

		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		_, err = waitJobCompleted(ctx, conn, aws.ToString(createJobOutput.Id), createTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", err),
				err.Error(),
			)
			return
		}

		listAssetsInput := dataexchange.ListRevisionAssetsInput{
			DataSetId:  plan.DataSetID.ValueStringPointer(),
			RevisionId: plan.ID.ValueStringPointer(),
		}
		var newAsset awstypes.AssetEntry
		paginator := dataexchange.NewListRevisionAssetsPaginator(conn, &listAssetsInput)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", err),
					err.Error(),
				)
				return
			}

			for _, v := range page.Assets {
				if !slices.Contains(existingAssetIDs, aws.ToString(v.Id)) {
					newAsset = v
				}
			}
		}
		if newAsset.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionExclusive, "XXX", nil),
				"missing new asset",
			)
			return
		}
		resp.Diagnostics.Append(flex.Flatten(ctx, newAsset, asset)...)
		if resp.Diagnostics.HasError() {
			return
		}
		assets[i] = *asset // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	}

	assetsVal, d := fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, assets)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	plan.Assets = assetsVal

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRevisionExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceRevisionExclusiveModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRevisionByID(ctx, conn, state.DataSetID.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionReading, ResNameRevisionExclusive, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRevisionExclusive) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceRevisionExclusiveModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dataexchange.DeleteRevisionInput{
		DataSetId:  state.DataSetID.ValueStringPointer(),
		RevisionId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteRevision(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameRevisionExclusive, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// func (r *resourceRevisionExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
// }

func findRevisionByID(ctx context.Context, conn *dataexchange.Client, dataSetId, revisionId string) (*dataexchange.GetRevisionOutput, error) {
	input := dataexchange.GetRevisionInput{
		DataSetId:  aws.String(dataSetId),
		RevisionId: aws.String(revisionId),
	}
	output, err := conn.GetRevision(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type resourceRevisionExclusiveModel struct {
	ARN       types.String                               `tfsdk:"arn"`
	Assets    fwtypes.SetNestedObjectValueOf[assetModel] `tfsdk:"asset"`
	Comment   types.String                               `tfsdk:"comment"`
	CreatedAt timetypes.RFC3339                          `tfsdk:"created_at"`
	DataSetID types.String                               `tfsdk:"data_set_id"`
	ID        types.String                               `tfsdk:"id"`
	UpdatedAt timetypes.RFC3339                          `tfsdk:"updated_at"`
	Tags      tftags.Map                                 `tfsdk:"tags"`
	TagsAll   tftags.Map                                 `tfsdk:"tags_all"`
	Timeouts  timeouts.Value                             `tfsdk:"timeouts"`
}

type assetModel struct {
	ARN                types.String                                             `tfsdk:"arn"`
	CreatedAt          timetypes.RFC3339                                        `tfsdk:"created_at"`
	ID                 types.String                                             `tfsdk:"id"`
	ImportAssetsFromS3 fwtypes.ListNestedObjectValueOf[importAssetsFromS3Model] `tfsdk:"import_assets_from_s3"`
	Name               types.String                                             `tfsdk:"name"`
	UpdatedAt          timetypes.RFC3339                                        `tfsdk:"updated_at"`
}

type importAssetsFromS3Model struct {
	AssetSources fwtypes.ListNestedObjectValueOf[assetSourceModel] `tfsdk:"asset_source"`
}

type assetSourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

func sweepRevisions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := dataexchange.ListDataSetRevisionsInput{}
	conn := client.DataExchangeClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := dataexchange.NewListDataSetRevisionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Revisions {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceRevisionExclusive, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.Id))),
			)
		}
	}

	return sweepResources, nil
}

func waitJobCompleted(ctx context.Context, conn *dataexchange.Client, jobID string, timeout time.Duration) (*dataexchange.GetJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.StateWaiting, awstypes.StateInProgress),
		Target:       enum.Slice(awstypes.StateCompleted),
		Refresh:      statusJob(ctx, conn, jobID),
		Timeout:      timeout,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dataexchange.GetJobOutput); ok {
		if output.State == awstypes.StateError {
			return output, jobError(output.Errors)
		}

		return output, err
	}

	return nil, err
}

func jobError(errs []awstypes.JobError) error {
	return errors.Join(tfslices.ApplyToAll(errs, func(e awstypes.JobError) error {
		return fmt.Errorf("%s: %s", e.Code, *e.Message)
	})...)
}

func statusJob(ctx context.Context, conn *dataexchange.Client, jobID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findJobByID(ctx, conn, jobID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func findJobByID(ctx context.Context, conn *dataexchange.Client, jobID string) (*dataexchange.GetJobOutput, error) {
	input := dataexchange.GetJobInput{
		JobId: aws.String(jobID),
	}

	out, err := conn.GetJob(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out, nil
}

func nestedObjectCollectionAll[T any](ctx context.Context, v nestedObjectCollectionValue[T]) (iter.Seq2[int, *T], diag.Diagnostics) {
	s, diags := v.ToSlice(ctx)
	if diags.HasError() {
		return nil, diags
	}

	return func(yield func(int, *T) bool) {
		for i, e := range slices.All(s) {
			if !yield(i, e) {
				return
			}
		}
	}, diags
}

func nestedObjectCollectionAllMust[T any](ctx context.Context, v nestedObjectCollectionValue[T]) iter.Seq2[int, *T] {
	return fwdiag.Must(nestedObjectCollectionAll(ctx, v))
}

type nestedObjectCollectionValue[T any] interface {
	ToSlice(context.Context) ([]*T, diag.Diagnostics)
}
