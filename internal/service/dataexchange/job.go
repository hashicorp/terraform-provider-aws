// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dataexchange_job", name="Job")
func ResourceJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceJob{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameJob = "Job"
)

type resourceJob struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrState: schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.State](),
			},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.Type](),
			},
			"start_on_creation": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"details": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[jobDetailsModel](ctx),
				Blocks: map[string]schema.Block{
					"import_assets_from_s3": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[importAssetsFromS3Model](ctx),
						Attributes: map[string]schema.Attribute{
							"data_set_id": schema.StringAttribute{
								Optional: true,
							},
							"revision_id": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"asset_sources": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[s3AssetSourceModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrBucket: schema.StringAttribute{
											Optional: true,
										},
										names.AttrKey: schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
					"export_assets_to_s3": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[exportAssetsToS3Model](ctx),
						Attributes: map[string]schema.Attribute{
							"data_set_id": schema.StringAttribute{
								Optional: true,
							},
							"revision_id": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"asset_destinations": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[s3AssetDestinationsModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrBucket: schema.StringAttribute{
											Optional: true,
										},
										names.AttrKey: schema.StringAttribute{
											Optional: true,
										},
										"asset_id": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
							"encryption": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[s3EncryptionModel](ctx),
								Attributes: map[string]schema.Attribute{
									names.AttrKMSKeyARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										Validators: []validator.String{
											validators.ARN(),
										},
									},
									names.AttrType: schema.StringAttribute{
										Optional:   true,
										CustomType: fwtypes.StringEnumType[awstypes.ServerSideEncryptionTypes](),
									},
								},
							},
						},
					},
					"import_asset_from_signed_url": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[importAssetFromSignedUrl](ctx),
						Attributes: map[string]schema.Attribute{
							"data_set_id": schema.StringAttribute{
								Optional: true,
							},
							"asset_name": schema.StringAttribute{
								Optional: true,
							},
							"revision_id": schema.StringAttribute{
								Optional: true,
							},
							"md5_hash": schema.StringAttribute{
								Optional: true,
							},
						},
					},
					"export_asset_to_signed_url": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[exportAssetToSignedUrl](ctx),
						Attributes: map[string]schema.Attribute{
							"data_set_id": schema.StringAttribute{
								Optional: true,
							},
							"asset_id": schema.StringAttribute{
								Optional: true,
							},
							"revision_id": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan resourceJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input dataexchange.CreateJobInput
	diags := plan.Expand(ctx, &input)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	out, err := conn.CreateJob(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, "", err),
			err.Error(),
		)
		return
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameJob, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = types.StringPointerValue(out.Id)
	plan.ARN = types.StringPointerValue(out.Arn)
	plan.State = fwtypes.StringEnumValue(out.State)

	if plan.StartOnCreation.ValueBool() {
		err := startJobById(ctx, conn, *out.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		found, err := findJobByID(ctx, conn, *out.Id)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		plan.State = fwtypes.StringEnumValue(found.State)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findJobByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if len(out.Errors) > 0 {
		for _, jobError := range out.Errors {
			err = fmt.Errorf(
				"ResourceType: %s, ResourceId: %s, Code: %s, Message: %s",
				jobError.ResourceType,
				*jobError.ResourceId,
				jobError.Code,
				*jobError.Message,
			)
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
		}
		return
	}

	if out.State == awstypes.StateCancelled {
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan, state resourceJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.StartOnCreation.ValueBool() && plan.StartOnCreation.ValueBool() {
		err := startJobById(ctx, conn, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		out, err := findJobByID(ctx, conn, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		plan.State = fwtypes.StringEnumValue(out.State)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if state.State.ValueEnum() == awstypes.StateWaiting {
		err := cancelJobByID(ctx, conn, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		var input dataexchange.CreateJobInput
		diags := plan.Expand(ctx, &input)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		out, err := conn.CreateJob(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, "", err),
				err.Error(),
			)
			return
		}
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if out == nil || out.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameJob, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ID = types.StringPointerValue(out.Id)
		plan.ARN = types.StringPointerValue(out.Arn)
		plan.State = fwtypes.StringEnumValue(out.State)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
}

func (r *resourceJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.State.ValueEnum() == awstypes.StateWaiting {
		err := cancelJobByID(ctx, conn, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}
}

func (r *resourceJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func startJobById(ctx context.Context, conn *dataexchange.Client, id string) error {
	in := dataexchange.StartJobInput{
		JobId: aws.String(id),
	}

	_, err := conn.StartJob(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return err
	}

	return nil
}

func cancelJobByID(ctx context.Context, conn *dataexchange.Client, id string) error {
	in := dataexchange.CancelJobInput{
		JobId: aws.String(id),
	}

	_, err := conn.CancelJob(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return err
	}

	return nil
}

func findJobByID(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetJobOutput, error) {
	in := &dataexchange.GetJobInput{
		JobId: aws.String(id),
	}

	out, err := conn.GetJob(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type (
	SupportedJobType string

	resourceJobModel struct {
		ARN types.String `tfsdk:"arn"`
		ID  types.String `tfsdk:"id"`

		StartOnCreation types.Bool                         `tfsdk:"start_on_creation"`
		Type            fwtypes.StringEnum[awstypes.Type]  `tfsdk:"type"`
		State           fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`

		Details fwtypes.ObjectValueOf[jobDetailsModel] `tfsdk:"details"`
	}

	jobDetailsModel struct {
		ImportAssetsFromS3       fwtypes.ObjectValueOf[importAssetsFromS3Model]  `tfsdk:"import_assets_from_s3"`
		ExportAssetsToS3         fwtypes.ObjectValueOf[exportAssetsToS3Model]    `tfsdk:"export_assets_to_s3"`
		ImportAssetFromSignedUrl fwtypes.ObjectValueOf[importAssetFromSignedUrl] `tfsdk:"import_asset_from_signed_url"`
		ExportAssetToSignedUrl   fwtypes.ObjectValueOf[exportAssetToSignedUrl]   `tfsdk:"export_asset_to_signed_url"`
	}

	importAssetsFromS3Model struct {
		DataSetId    types.String                                        `tfsdk:"data_set_id"`
		RevisionId   types.String                                        `tfsdk:"revision_id"`
		AssetSources fwtypes.ListNestedObjectValueOf[s3AssetSourceModel] `tfsdk:"asset_sources"`
	}

	exportAssetsToS3Model struct {
		DataSetId         types.String                                              `tfsdk:"data_set_id"`
		RevisionId        types.String                                              `tfsdk:"revision_id"`
		AssetDestinations fwtypes.ListNestedObjectValueOf[s3AssetDestinationsModel] `tfsdk:"asset_destinations"`
		Encryption        fwtypes.ObjectValueOf[s3EncryptionModel]                  `tfsdk:"encryption"`
	}

	importAssetFromSignedUrl struct {
		AssetName  types.String `tfsdk:"asset_name"`
		DataSetId  types.String `tfsdk:"data_set_id"`
		Md5Hash    types.String `tfsdk:"md5_hash"`
		RevisionId types.String `tfsdk:"revision_id"`
	}

	exportAssetToSignedUrl struct {
		AssetId    types.String `tfsdk:"asset_id"`
		DataSetId  types.String `tfsdk:"data_set_id"`
		RevisionId types.String `tfsdk:"revision_id"`
	}

	s3EncryptionModel struct {
		KmsKeyArn types.String                                           `tfsdk:"kms_key_arn"`
		Type      fwtypes.StringEnum[awstypes.ServerSideEncryptionTypes] `tfsdk:"type"`
	}

	s3AssetSourceModel struct {
		Bucket types.String `tfsdk:"bucket"`
		Key    types.String `tfsdk:"key"`
	}

	s3AssetDestinationsModel struct {
		AssetId types.String `tfsdk:"asset_id"`
		Bucket  types.String `tfsdk:"bucket"`
		Key     types.String `tfsdk:"key"`
	}
)

func (m resourceJobModel) Expand(ctx context.Context, v *dataexchange.CreateJobInput) diag.Diagnostics {
	var diags diag.Diagnostics
	v.Details = &awstypes.RequestDetails{}
	diags.Append(flex.Expand(ctx, m.Details, &v.Details)...)
	v.Type = awstypes.Type(m.Type.ValueString())
	return diags
}
