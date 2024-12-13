// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"errors"
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

// @FrameworkResource("aws_dataexchange_event_action", name="Event Action")
func ResourceEventAction(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEventAction{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEventAction = "Event Action"
)

type resourceEventAction struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEventAction) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_dataexchange_event_action"
}

func (r *resourceEventAction) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"action_export_revision_to_s3": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[actionExportRevisionToS3Model](ctx),
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required: true,
					},
					"key_pattern": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("${Revision.CreatedAt}/${Asset.Name}"),
					},
					"s3_encryption_kms_key_arn": schema.StringAttribute{
						CustomType: fwtypes.ARNType,
						Optional:   true,
						Validators: []validator.String{
							validators.ARN(),
						},
					},
					"s3_encryption_type": schema.StringAttribute{
						Optional:   true,
						CustomType: fwtypes.StringEnumType[awstypes.ServerSideEncryptionTypes](),
					},
				},
			},
			"event_revision_published": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[eventRevisionPublishedModel](ctx),
				Attributes: map[string]schema.Attribute{
					"data_set_id": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (r *resourceEventAction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var data resourceEventActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := r.buildCreateInput(ctx, data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	out, err := conn.CreateEventAction(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.ARN = types.StringPointerValue(out.Arn)
	data.ID = types.StringPointerValue(out.Id)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceEventAction) buildCreateInput(ctx context.Context, data resourceEventActionModel) (*dataexchange.CreateEventActionInput, diag.Diagnostics) {
	resDiag := diag.Diagnostics{}

	input := dataexchange.CreateEventActionInput{
		Action: &awstypes.Action{},
		Event:  &awstypes.Event{},
	}

	if !data.ActionExportRevisionToS3.IsNull() {
		actionExportRevisionToS3, diags := data.ActionExportRevisionToS3.ToPtr(ctx)
		if diags.HasError() {
			resDiag.Append(diags...)
			return nil, resDiag
		}
		input.Action.ExportRevisionToS3 = &awstypes.AutoExportRevisionToS3RequestDetails{
			RevisionDestination: &awstypes.AutoExportRevisionDestinationEntry{
				Bucket:     actionExportRevisionToS3.Bucket.ValueStringPointer(),
				KeyPattern: actionExportRevisionToS3.KeyPattern.ValueStringPointer(),
			},
		}

		if !actionExportRevisionToS3.S3EncryptionType.IsNull() {
			input.Action.ExportRevisionToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      actionExportRevisionToS3.S3EncryptionType.ValueEnum(),
				KmsKeyArn: actionExportRevisionToS3.S3EncryptionKmsKeyArn.ValueStringPointer(),
			}
		}
	}

	if !data.EventRevisionPublished.IsNull() {
		eventRevisionPublished, diags := data.EventRevisionPublished.ToPtr(ctx)
		if diags.HasError() {
			return nil, resDiag
		}
		input.Event.RevisionPublished = &awstypes.RevisionPublished{
			DataSetId: eventRevisionPublished.DataSetId.ValueStringPointer(),
		}
	}

	return &input, resDiag
}

func (r *resourceEventAction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceEventActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEventActionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionSetting, ResNameEventAction, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(r.buildModelFromOutput(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEventAction) buildModelFromOutput(ctx context.Context, output *dataexchange.GetEventActionOutput, data *resourceEventActionModel) diag.Diagnostics {
	resDiag := diag.Diagnostics{}
	data.ARN = types.StringPointerValue(output.Arn)

	eventObjectValue, diags := fwtypes.NewObjectValueOf[eventRevisionPublishedModel](ctx, &eventRevisionPublishedModel{
		DataSetId: types.StringPointerValue(output.Event.RevisionPublished.DataSetId),
	})
	if diags.HasError() {
		resDiag.Append(diags...)
		return resDiag
	}

	data.EventRevisionPublished = eventObjectValue

	model := actionExportRevisionToS3Model{
		Bucket:     types.StringPointerValue(output.Action.ExportRevisionToS3.RevisionDestination.Bucket),
		KeyPattern: types.StringPointerValue(output.Action.ExportRevisionToS3.RevisionDestination.KeyPattern),
	}
	if output.Action.ExportRevisionToS3.Encryption != nil {
		model.S3EncryptionType = fwtypes.StringEnumValue(output.Action.ExportRevisionToS3.Encryption.Type)
		model.S3EncryptionKmsKeyArn = types.StringPointerValue(output.Action.ExportRevisionToS3.Encryption.KmsKeyArn)
	}

	actionObjectValue, diags := fwtypes.NewObjectValueOf[actionExportRevisionToS3Model](ctx, &model)
	if diags.HasError() {
		resDiag.Append(diags...)
		return resDiag
	}

	data.ActionExportRevisionToS3 = actionObjectValue

	return resDiag
}

func (r *resourceEventAction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	conn := r.Meta().DataExchangeClient(ctx)

	var data, state resourceEventActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.EventRevisionPublished.Equal(state.EventRevisionPublished) {
		_, err := conn.DeleteEventAction(ctx, &dataexchange.DeleteEventActionInput{
			EventActionId: state.ID.ValueStringPointer(),
		})

		input, diags := r.buildCreateInput(ctx, data)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		out, err := conn.CreateEventAction(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameEventAction, "", nil),
				errors.New("empty output").Error(),
			)
			return
		}
		return
	}

	input, diags := r.buildUpdateInput(ctx, state.ID.ValueString(), data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	out, err := conn.UpdateEventAction(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameEventAction, data.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameEventAction, data.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceEventAction) buildUpdateInput(ctx context.Context, id string, data resourceEventActionModel) (*dataexchange.UpdateEventActionInput, diag.Diagnostics) {
	resDiag := diag.Diagnostics{}

	input := dataexchange.UpdateEventActionInput{
		EventActionId: aws.String(id),
		Action:        &awstypes.Action{},
	}

	if !data.ActionExportRevisionToS3.IsNull() {
		actionExportRevisionToS3, diags := data.ActionExportRevisionToS3.ToPtr(ctx)
		if diags.HasError() {
			resDiag.Append(diags...)
			return nil, resDiag
		}
		input.Action.ExportRevisionToS3 = &awstypes.AutoExportRevisionToS3RequestDetails{
			RevisionDestination: &awstypes.AutoExportRevisionDestinationEntry{
				Bucket:     actionExportRevisionToS3.Bucket.ValueStringPointer(),
				KeyPattern: actionExportRevisionToS3.KeyPattern.ValueStringPointer(),
			},
		}

		if !actionExportRevisionToS3.S3EncryptionType.IsNull() {
			input.Action.ExportRevisionToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      actionExportRevisionToS3.S3EncryptionType.ValueEnum(),
				KmsKeyArn: actionExportRevisionToS3.S3EncryptionKmsKeyArn.ValueStringPointer(),
			}
		}
	}

	return &input, resDiag
}

func (r *resourceEventAction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state resourceEventActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteEventAction(ctx, &dataexchange.DeleteEventActionInput{
		EventActionId: state.ID.ValueStringPointer(),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameEventAction, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEventAction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findEventActionByID(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetEventActionOutput, error) {
	in := &dataexchange.GetEventActionInput{
		EventActionId: aws.String(id),
	}

	out, err := conn.GetEventAction(ctx, in)
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

type resourceEventActionModel struct {
	ARN                      types.String                                         `tfsdk:"arn"`
	ID                       types.String                                         `tfsdk:"id"`
	ActionExportRevisionToS3 fwtypes.ObjectValueOf[actionExportRevisionToS3Model] `tfsdk:"action_export_revision_to_s3"`
	EventRevisionPublished   fwtypes.ObjectValueOf[eventRevisionPublishedModel]   `tfsdk:"event_revision_published"`
}

type actionExportRevisionToS3Model struct {
	Bucket                types.String                                           `tfsdk:"bucket"`
	KeyPattern            types.String                                           `tfsdk:"key_pattern"`
	S3EncryptionKmsKeyArn types.String                                           `tfsdk:"s3_encryption_kms_key_arn"`
	S3EncryptionType      fwtypes.StringEnum[awstypes.ServerSideEncryptionTypes] `tfsdk:"s3_encryption_type"`
}

type eventRevisionPublishedModel struct {
	DataSetId types.String `tfsdk:"data_set_id"`
}
