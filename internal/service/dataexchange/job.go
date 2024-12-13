// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

func (r *resourceJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_dataexchange_job"
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
			"data_set_id": schema.StringAttribute{
				Required: true,
			},
			"revision_id": schema.StringAttribute{
				Optional: true,
			},
			"asset_id": schema.StringAttribute{
				Optional: true,
			},
			"start_on_creation": schema.BoolAttribute{
				Optional: true,
			},
			"s3_asset_destination_encryption_kms_key_arn": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validators.ARN(),
				},
			},
			"s3_asset_destination_encryption_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ServerSideEncryptionTypes](),
				Optional:   true,
			},
			"signed_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"signed_url_expires_at": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"s3_event_action_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"url_asset_name": schema.StringAttribute{
				Optional: true,
			},
			"url_asset_md5_hash": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"s3_asset_destinations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3AssetDestinationsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"asset_id": schema.StringAttribute{
							Required: true,
						},
						"bucket": schema.StringAttribute{
							Required: true,
						},
						"key": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"s3_revision_destinations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3RevisionDestinationsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"revision_id": schema.StringAttribute{
							Required: true,
						},
						"bucket": schema.StringAttribute{
							Required: true,
						},
						"key_pattern": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"s3_asset_sources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3AssetSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bucket": schema.StringAttribute{
							Required: true,
						},
						"key": schema.StringAttribute{
							Required: true,
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

	out, diags := r.createJob(ctx, conn, plan)
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

func (r *resourceJob) buildCreateInput(ctx context.Context, data resourceJobModel) (*dataexchange.CreateJobInput, diag.Diagnostics) {
	resDiag := diag.Diagnostics{}
	input := &dataexchange.CreateJobInput{
		Type:    data.Type.ValueEnum(),
		Details: &awstypes.RequestDetails{},
	}

	switch data.Type.ValueEnum() {
	case awstypes.TypeExportAssetsToS3:
		input.Details.ExportAssetsToS3 = &awstypes.ExportAssetsToS3RequestDetails{
			DataSetId:  data.DataSetId.ValueStringPointer(),
			RevisionId: data.RevisionId.ValueStringPointer(),
		}

		if !data.S3AssetDestinations.IsNull() {
			s3AssetDestinations, diags := data.S3AssetDestinations.ToSlice(ctx)
			if diags.HasError() {
				resDiag.Append(diags...)
				return nil, resDiag
			}

			input.Details.ExportAssetsToS3.AssetDestinations = make([]awstypes.AssetDestinationEntry, len(s3AssetDestinations))
			for i, value := range s3AssetDestinations {
				input.Details.ExportAssetsToS3.AssetDestinations[i] = awstypes.AssetDestinationEntry{
					AssetId: value.AssetId.ValueStringPointer(),
					Bucket:  value.Bucket.ValueStringPointer(),
					Key:     value.Key.ValueStringPointer(),
				}
			}
		}

		if !data.S3AssetDestinationEncryptionType.IsNull() {
			input.Details.ExportAssetsToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      data.S3AssetDestinationEncryptionType.ValueEnum(),
				KmsKeyArn: data.S3AssetDestinationEncryptionKmsKeyArn.ValueStringPointer(),
			}
		}
		break
	case awstypes.TypeExportAssetToSignedUrl:
		input.Details.ExportAssetToSignedUrl = &awstypes.ExportAssetToSignedUrlRequestDetails{
			AssetId:    data.AssetId.ValueStringPointer(),
			DataSetId:  data.DataSetId.ValueStringPointer(),
			RevisionId: data.RevisionId.ValueStringPointer(),
		}
		break
	case awstypes.TypeExportRevisionsToS3:
		input.Details.ExportRevisionsToS3 = &awstypes.ExportRevisionsToS3RequestDetails{
			DataSetId: data.DataSetId.ValueStringPointer(),
		}

		if !data.S3RevisionDestinations.IsNull() {
			s3RevisionDestinations, diags := data.S3RevisionDestinations.ToSlice(ctx)
			if diags.HasError() {
				resDiag.Append(diags...)
				return nil, resDiag
			}

			input.Details.ExportRevisionsToS3.RevisionDestinations = make([]awstypes.RevisionDestinationEntry, len(s3RevisionDestinations))
			for i, values := range s3RevisionDestinations {
				input.Details.ExportRevisionsToS3.RevisionDestinations[i] = awstypes.RevisionDestinationEntry{
					RevisionId: values.RevisionId.ValueStringPointer(),
					Bucket:     values.Bucket.ValueStringPointer(),
					KeyPattern: values.KeyPattern.ValueStringPointer(),
				}
			}
		}

		if !data.S3AssetDestinationEncryptionType.IsNull() {
			input.Details.ExportRevisionsToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      data.S3AssetDestinationEncryptionType.ValueEnum(),
				KmsKeyArn: data.S3AssetDestinationEncryptionKmsKeyArn.ValueStringPointer(),
			}
		}
		break
	case awstypes.TypeImportAssetsFromS3:
		input.Details.ImportAssetsFromS3 = &awstypes.ImportAssetsFromS3RequestDetails{
			DataSetId:  data.DataSetId.ValueStringPointer(),
			RevisionId: data.RevisionId.ValueStringPointer(),
		}

		if !data.S3AssetSources.IsNull() {
			s3AssetSources, diags := data.S3AssetSources.ToSlice(ctx)
			if diags.HasError() {
				resDiag.Append(diags...)
				return nil, resDiag
			}

			input.Details.ImportAssetsFromS3.AssetSources = make([]awstypes.AssetSourceEntry, len(s3AssetSources))
			for i, source := range s3AssetSources {
				input.Details.ImportAssetsFromS3.AssetSources[i] = awstypes.AssetSourceEntry{
					Bucket: source.Bucket.ValueStringPointer(),
					Key:    source.Key.ValueStringPointer(),
				}
			}
		}
		break
	case awstypes.TypeImportAssetFromSignedUrl:
		input.Details.ImportAssetFromSignedUrl = &awstypes.ImportAssetFromSignedUrlRequestDetails{
			AssetName:  data.UrlAssetName.ValueStringPointer(),
			DataSetId:  data.DataSetId.ValueStringPointer(),
			Md5Hash:    data.UrlAssetMd5Hash.ValueStringPointer(),
			RevisionId: data.RevisionId.ValueStringPointer(),
		}
		break
	default:
		resDiag.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, data.ID.String(), errors.New("unsupported type")),
			fmt.Sprintf("Type %s not supported", data.Type.ValueString()),
		)
		return nil, resDiag
	}

	return input, resDiag
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

	resp.Diagnostics.Append(r.buildModelFromOutput(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceJob) buildModelFromOutput(ctx context.Context, out *dataexchange.GetJobOutput, data *resourceJobModel) diag.Diagnostics {
	resDiag := diag.Diagnostics{}
	data.ARN = types.StringPointerValue(out.Arn)
	data.State = fwtypes.StringEnumValue(out.State)
	data.Type = fwtypes.StringEnumValue(out.Type)

	if out.Details != nil {
		switch out.Type {
		case awstypes.TypeExportAssetsToS3:
			if out.Details.ExportAssetsToS3 != nil {
				data.DataSetId = types.StringPointerValue(out.Details.ExportAssetsToS3.DataSetId)
				if out.Details.ExportAssetsToS3.Encryption != nil {
					data.S3AssetDestinationEncryptionType = fwtypes.StringEnumValue(out.Details.ExportAssetsToS3.Encryption.Type)
					data.S3AssetDestinationEncryptionKmsKeyArn = types.StringPointerValue(out.Details.ExportAssetsToS3.Encryption.KmsKeyArn)
				}
				dest := make([]s3AssetDestinationsModel, len(out.Details.ExportAssetsToS3.AssetDestinations))
				for i, destination := range out.Details.ExportAssetsToS3.AssetDestinations {
					dest[i] = s3AssetDestinationsModel{
						AssetId: types.StringPointerValue(destination.AssetId),
						Bucket:  types.StringPointerValue(destination.Bucket),
						Key:     types.StringPointerValue(destination.Key),
					}
				}
				data.S3AssetDestinations, _ = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, dest)
			}
		case awstypes.TypeExportAssetToSignedUrl:
			if out.Details.ExportAssetToSignedUrl != nil {
				data.AssetId = types.StringPointerValue(out.Details.ExportAssetToSignedUrl.AssetId)
				data.SignedUrl = types.StringPointerValue(out.Details.ExportAssetToSignedUrl.SignedUrl)
				if out.Details.ExportAssetToSignedUrl.SignedUrlExpiresAt != nil {
					data.SignedUrlExpiresAt = types.StringValue(out.Details.ExportAssetToSignedUrl.SignedUrlExpiresAt.String())
				}
				data.DataSetId = types.StringPointerValue(out.Details.ExportAssetToSignedUrl.DataSetId)
				data.RevisionId = types.StringPointerValue(out.Details.ExportAssetToSignedUrl.RevisionId)
			}
			break
		case awstypes.TypeExportRevisionsToS3:
			if out.Details.ExportRevisionsToS3 != nil {
				data.DataSetId = types.StringPointerValue(out.Details.ExportRevisionsToS3.DataSetId)
				if out.Details.ExportRevisionsToS3.Encryption != nil {
					data.S3AssetDestinationEncryptionType = fwtypes.StringEnumValue(out.Details.ExportRevisionsToS3.Encryption.Type)
					data.S3AssetDestinationEncryptionKmsKeyArn = types.StringPointerValue(out.Details.ExportRevisionsToS3.Encryption.KmsKeyArn)
				}
				data.S3EventActionArn = types.StringPointerValue(out.Details.ExportRevisionsToS3.EventActionArn)

				dest := make([]s3RevisionDestinationsModel, len(out.Details.ExportRevisionsToS3.RevisionDestinations))
				for i, destination := range out.Details.ExportRevisionsToS3.RevisionDestinations {
					dest[i] = s3RevisionDestinationsModel{
						RevisionId: types.StringPointerValue(destination.RevisionId),
						Bucket:     types.StringPointerValue(destination.Bucket),
						KeyPattern: types.StringPointerValue(destination.KeyPattern),
					}
				}
				data.S3RevisionDestinations, _ = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, dest)
			}
			break
		case awstypes.TypeImportAssetFromSignedUrl:
			if out.Details.ImportAssetFromSignedUrl != nil {
				data.UrlAssetName = types.StringPointerValue(out.Details.ImportAssetFromSignedUrl.AssetName)
				data.SignedUrl = types.StringPointerValue(out.Details.ImportAssetFromSignedUrl.SignedUrl)
				if out.Details.ImportAssetFromSignedUrl.SignedUrlExpiresAt != nil {
					data.SignedUrlExpiresAt = types.StringValue(out.Details.ImportAssetFromSignedUrl.SignedUrlExpiresAt.String())
				}
				data.UrlAssetMd5Hash = types.StringPointerValue(out.Details.ImportAssetFromSignedUrl.Md5Hash)
				data.DataSetId = types.StringPointerValue(out.Details.ImportAssetFromSignedUrl.DataSetId)
				data.RevisionId = types.StringPointerValue(out.Details.ImportAssetFromSignedUrl.RevisionId)
			}
			break
		case awstypes.TypeImportAssetsFromS3:
			if out.Details.ImportAssetsFromS3 != nil {
				data.DataSetId = types.StringPointerValue(out.Details.ImportAssetsFromS3.DataSetId)
				data.RevisionId = types.StringPointerValue(out.Details.ImportAssetsFromS3.RevisionId)
				dest := make([]s3AssetSourceModel, len(out.Details.ImportAssetsFromS3.AssetSources))
				for i, destination := range out.Details.ImportAssetsFromS3.AssetSources {
					dest[i] = s3AssetSourceModel{
						Bucket: types.StringPointerValue(destination.Bucket),
						Key:    types.StringPointerValue(destination.Key),
					}
				}
				data.S3AssetSources, _ = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, dest)
			}
			break
		default:

		}
	}

	return resDiag
}

func (r *resourceJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan, state resourceJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.StartOnCreation.ValueBool() == false && plan.StartOnCreation.ValueBool() == true {
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

		out, diags := r.createJob(ctx, conn, plan)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceJob) createJob(ctx context.Context, conn *dataexchange.Client, data resourceJobModel) (*dataexchange.CreateJobOutput, diag.Diagnostics) {
	in, diags := r.buildCreateInput(ctx, data)
	if diags.HasError() {
		return nil, diags
	}

	resDiag := diag.Diagnostics{}
	out, err := conn.CreateJob(ctx, in)
	if err != nil {
		resDiag.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameJob, "", err),
			err.Error(),
		)
		return nil, resDiag
	}

	return out, nil
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

type resourceJobModel struct {
	ARN types.String `tfsdk:"arn"`
	ID  types.String `tfsdk:"id"`

	StartOnCreation types.Bool                         `tfsdk:"start_on_creation"`
	Type            fwtypes.StringEnum[awstypes.Type]  `tfsdk:"type"`
	State           fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`

	DataSetId  types.String `tfsdk:"data_set_id"`
	RevisionId types.String `tfsdk:"revision_id"`
	AssetId    types.String `tfsdk:"asset_id"`

	S3AssetDestinations                   fwtypes.ListNestedObjectValueOf[s3AssetDestinationsModel] `tfsdk:"s3_asset_destinations"`
	S3AssetDestinationEncryptionType      fwtypes.StringEnum[awstypes.ServerSideEncryptionTypes]    `tfsdk:"s3_asset_destination_encryption_type"`
	S3AssetDestinationEncryptionKmsKeyArn types.String                                              `tfsdk:"s3_asset_destination_encryption_kms_key_arn"`

	SignedUrl          types.String `tfsdk:"signed_url"`
	SignedUrlExpiresAt types.String `tfsdk:"signed_url_expires_at"`

	S3RevisionDestinations fwtypes.ListNestedObjectValueOf[s3RevisionDestinationsModel] `tfsdk:"s3_revision_destinations"`
	S3EventActionArn       types.String                                                 `tfsdk:"s3_event_action_arn"`

	UrlAssetName    types.String `tfsdk:"url_asset_name"`
	UrlAssetMd5Hash types.String `tfsdk:"url_asset_md5_hash"`

	S3AssetSources fwtypes.ListNestedObjectValueOf[s3AssetSourceModel] `tfsdk:"s3_asset_sources"`
}

type s3AssetSourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

type s3AssetDestinationsModel struct {
	AssetId types.String `tfsdk:"asset_id"`
	Bucket  types.String `tfsdk:"bucket"`
	Key     types.String `tfsdk:"key"`
}

type s3RevisionDestinationsModel struct {
	RevisionId types.String `tfsdk:"revision_id"`
	Bucket     types.String `tfsdk:"bucket"`
	KeyPattern types.String `tfsdk:"key_pattern"`
}
