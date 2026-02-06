// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_control", name="Control")
// @Tags(identifierAttribute="arn")
func newControlResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &controlResource{}, nil
}

type controlResource struct {
	framework.ResourceWithModel[controlResourceModel]
	framework.WithImportByID
}

func (r *controlResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_plan_instructions": schema.StringAttribute{
				Optional: true,
			},
			"action_plan_title": schema.StringAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"testing_information": schema.StringAttribute{
				Optional: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"control_mapping_sources": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[controlMappingSourceModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_description": schema.StringAttribute{
							Optional: true,
						},
						"source_frequency": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SourceFrequency](),
							Optional:   true,
						},
						"source_id":      framework.IDAttribute(),
						"source_keyword": framework.ResourceOptionalComputedListOfObjectsAttribute[sourceKeywordModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
						"source_name": schema.StringAttribute{
							Required: true,
						},
						"source_set_up_option": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SourceSetUpOption](),
							Required:   true,
						},
						names.AttrSourceType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SourceType](),
							Required:   true,
						},
						"troubleshooting_text": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *controlResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data controlResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input auditmanager.CreateControlInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateControl(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Control (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	control := output.Control
	data.ARN = fwflex.StringToFramework(ctx, control.Arn)
	response.Diagnostics.Append(fwflex.Flatten(ctx, control.ControlMappingSources, &data.ControlMappingSources)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = fwflex.StringToFramework(ctx, control.Id)
	data.Type = fwflex.StringValueToFramework(ctx, control.Type)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *controlResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data controlResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := findControlByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Control (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *controlResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old controlResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	if !new.ActionPlanInstructions.Equal(old.ActionPlanInstructions) ||
		!new.ActionPlanTitle.Equal(old.ActionPlanTitle) ||
		!new.ControlMappingSources.Equal(old.ControlMappingSources) ||
		!new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.TestingInformation.Equal(old.TestingInformation) {
		var input auditmanager.UpdateControlInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ControlId = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateControl(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Audit Manager Control (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *controlResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data controlResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	input := auditmanager.DeleteControlInput{
		ControlId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteControl(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Control (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *controlResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var data controlResourceModel
		response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
		if response.Diagnostics.HasError() {
			return
		}

		controlMappingSources, diags := data.ControlMappingSources.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// A net-new control mapping will be missing source_id, so force a replacement
		//
		// Attribute level plan modifiers are applied before resource modifiers, so ID's
		// previously in state should never be unknown.
		for _, controlMappingSource := range controlMappingSources {
			if controlMappingSource.SourceID.IsUnknown() {
				response.RequiresReplace = []path.Path{path.Root("control_mapping_sources")}
			}
		}
	}
}

func findControlByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Control, error) {
	input := auditmanager.GetControlInput{
		ControlId: aws.String(id),
	}
	output, err := conn.GetControl(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Control == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Control, nil
}

type controlResourceModel struct {
	framework.WithRegionModel
	ActionPlanInstructions types.String                                              `tfsdk:"action_plan_instructions"`
	ActionPlanTitle        types.String                                              `tfsdk:"action_plan_title"`
	ARN                    types.String                                              `tfsdk:"arn"`
	ControlMappingSources  fwtypes.SetNestedObjectValueOf[controlMappingSourceModel] `tfsdk:"control_mapping_sources"`
	Description            types.String                                              `tfsdk:"description"`
	ID                     types.String                                              `tfsdk:"id"`
	Name                   types.String                                              `tfsdk:"name"`
	Tags                   tftags.Map                                                `tfsdk:"tags"`
	TagsAll                tftags.Map                                                `tfsdk:"tags_all"`
	TestingInformation     types.String                                              `tfsdk:"testing_information"`
	Type                   types.String                                              `tfsdk:"type"`
}

type controlMappingSourceModel struct {
	SourceDescription   types.String                                        `tfsdk:"source_description"`
	SourceFrequency     fwtypes.StringEnum[awstypes.SourceFrequency]        `tfsdk:"source_frequency"`
	SourceID            types.String                                        `tfsdk:"source_id"`
	SourceKeyword       fwtypes.ListNestedObjectValueOf[sourceKeywordModel] `tfsdk:"source_keyword"`
	SourceName          types.String                                        `tfsdk:"source_name"`
	SourceSetUpOption   fwtypes.StringEnum[awstypes.SourceSetUpOption]      `tfsdk:"source_set_up_option"`
	SourceType          types.String                                        `tfsdk:"source_type"`
	TroubleshootingText types.String                                        `tfsdk:"troubleshooting_text"`
}

type sourceKeywordModel struct {
	KeywordInputType fwtypes.StringEnum[awstypes.KeywordInputType] `tfsdk:"keyword_input_type"`
	KeywordValue     types.String                                  `tfsdk:"keyword_value"`
}
