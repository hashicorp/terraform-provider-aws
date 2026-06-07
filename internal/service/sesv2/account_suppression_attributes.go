// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sesv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_account_suppression_attributes", name="Account Suppression Attributes")
func newAccountSuppressionAttributesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountSuppressionAttributesResource{}

	return r, nil
}

type accountSuppressionAttributesResource struct {
	framework.ResourceWithModel[accountSuppressionAttributesResourceModel]
	framework.WithImportByID
}

func (r *accountSuppressionAttributesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"suppressed_reasons": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.SuppressionListReason](),
				Required:    true,
				ElementType: fwtypes.StringEnumType[awstypes.SuppressionListReason](),
			},
		},
		Blocks: map[string]schema.Block{
			"validation_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[suppressionValidationAttributesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"condition_threshold": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[suppressionConditionThresholdModel](ctx, fwtypes.WithSemanticEqualityFunc(suppressionConditionThresholdSemanticEquals)),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"condition_threshold_enabled": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FeatureStatus](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"overall_confidence_threshold": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[suppressionConfidenceThresholdModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"confidence_verdict_threshold": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.SuppressionConfidenceVerdictThreshold](),
													Required:   true,
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

func (r *accountSuppressionAttributesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	id := r.Meta().AccountID(ctx)
	input := &sesv2.PutAccountSuppressionAttributesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutAccountSuppressionAttributes(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SESv2 Account Suppression Attributes (%s)", id), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSuppressionAttributesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	suppressionAttributes, err := findAccountSuppressionAttributes(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SESv2 Account Suppression Attributes (%s)", data.ID.ValueString()), err.Error())

		return
	}

	priorState := data

	response.Diagnostics.Append(fwflex.Flatten(ctx, suppressionAttributes, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	normalizeAccountSuppressionAttributesState(ctx, &data, &priorState, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSuppressionAttributesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	input := &sesv2.PutAccountSuppressionAttributesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutAccountSuppressionAttributes(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating SESv2 Account Suppression Attributes (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accountSuppressionAttributesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accountSuppressionAttributesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	// Reset singleton settings to the service defaults on Delete.
	input := &sesv2.PutAccountSuppressionAttributesInput{
		SuppressedReasons: []awstypes.SuppressionListReason{
			awstypes.SuppressionListReasonBounce,
			awstypes.SuppressionListReasonComplaint,
		},
		ValidationAttributes: &awstypes.SuppressionValidationAttributes{
			ConditionThreshold: &awstypes.SuppressionConditionThreshold{
				ConditionThresholdEnabled: awstypes.FeatureStatusDisabled,
			},
		},
	}

	_, err := conn.PutAccountSuppressionAttributes(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("resetting SESv2 Account Suppression Attributes (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAccountSuppressionAttributes(ctx context.Context, conn *sesv2.Client) (*awstypes.SuppressionAttributes, error) {
	output, err := findAccount(ctx, conn)

	if err != nil {
		return nil, err
	}

	if output.SuppressionAttributes == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.SuppressionAttributes, nil
}

type accountSuppressionAttributesResourceModel struct {
	framework.WithRegionModel
	ID                   types.String                                                          `tfsdk:"id"`
	SuppressedReasons    fwtypes.SetOfStringEnum[awstypes.SuppressionListReason]               `tfsdk:"suppressed_reasons"`
	ValidationAttributes fwtypes.ListNestedObjectValueOf[suppressionValidationAttributesModel] `tfsdk:"validation_attributes"`
}

type suppressionValidationAttributesModel struct {
	ConditionThreshold fwtypes.ListNestedObjectValueOf[suppressionConditionThresholdModel] `tfsdk:"condition_threshold"`
}

type suppressionConditionThresholdModel struct {
	ConditionThresholdEnabled  fwtypes.StringEnum[awstypes.FeatureStatus]                           `tfsdk:"condition_threshold_enabled"`
	OverallConfidenceThreshold fwtypes.ListNestedObjectValueOf[suppressionConfidenceThresholdModel] `tfsdk:"overall_confidence_threshold"`
}

type suppressionConfidenceThresholdModel struct {
	ConfidenceVerdictThreshold fwtypes.StringEnum[awstypes.SuppressionConfidenceVerdictThreshold] `tfsdk:"confidence_verdict_threshold"`
}

// When condition_threshold_enabled is set to "DISABLED", and overall_confidence_threshold block is not set in the prior state
// the overall_confidence_threshold block in the state is set to an empty list to prevent drift.
func normalizeAccountSuppressionAttributesState(ctx context.Context, data, priorState *accountSuppressionAttributesResourceModel, diags *diag.Diagnostics) {
	validationAttributes, d := data.ValidationAttributes.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || validationAttributes == nil {
		return
	}

	conditionThreshold, d := validationAttributes.ConditionThreshold.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || conditionThreshold == nil {
		return
	}

	if conditionThreshold.ConditionThresholdEnabled.IsNull() || conditionThreshold.ConditionThresholdEnabled.IsUnknown() {
		return
	}

	if conditionThreshold.ConditionThresholdEnabled.ValueEnum() == awstypes.FeatureStatusDisabled {
		priorOverallConfidenceThreshold, d := priorStateOverallConfidenceThreshold(ctx, priorState)
		diags.Append(d...)
		if diags.HasError() || priorOverallConfidenceThreshold != nil {
			return
		}

		conditionThreshold.OverallConfidenceThreshold = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*suppressionConfidenceThresholdModel{})
		validationAttributes.ConditionThreshold = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, conditionThreshold, fwtypes.WithSemanticEqualityFunc(suppressionConditionThresholdSemanticEquals))
		data.ValidationAttributes = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, validationAttributes)
	}
}

// Extracts overall_confidence_threshold from the prior state, if it exists.
// This is used to determine whether to set the overall_confidence_threshold to an empty list when condition_threshold_enabled is set to "DISABLED".
func priorStateOverallConfidenceThreshold(ctx context.Context, priorState *accountSuppressionAttributesResourceModel) (*suppressionConfidenceThresholdModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	validationAttributes, d := priorState.ValidationAttributes.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || validationAttributes == nil {
		return nil, diags
	}

	conditionThreshold, d := validationAttributes.ConditionThreshold.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || conditionThreshold == nil {
		return nil, diags
	}

	overallConfidenceThreshold, d := conditionThreshold.OverallConfidenceThreshold.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return overallConfidenceThreshold, diags
}

func suppressionConditionThresholdSemanticEquals(ctx context.Context, current, new fwtypes.NestedCollectionValue[suppressionConditionThresholdModel]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	currentValue, d := current.ToPtr(ctx)
	diags.Append(d...)
	newValue, d := new.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || currentValue == nil || newValue == nil {
		return false, diags
	}

	if currentValue.ConditionThresholdEnabled.IsNull() || currentValue.ConditionThresholdEnabled.IsUnknown() ||
		newValue.ConditionThresholdEnabled.IsNull() || newValue.ConditionThresholdEnabled.IsUnknown() {
		return false, diags
	}

	if currentValue.ConditionThresholdEnabled.ValueEnum() != newValue.ConditionThresholdEnabled.ValueEnum() {
		return false, diags
	}

	return newValue.ConditionThresholdEnabled.ValueEnum() == awstypes.FeatureStatusDisabled, diags
}
