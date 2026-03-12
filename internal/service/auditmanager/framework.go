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

// @FrameworkResource("aws_auditmanager_framework", name="Framework")
// @Tags(identifierAttribute="arn")
func newFrameworkResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &frameworkResource{}, nil
}

type frameworkResource struct {
	framework.ResourceWithModel[frameworkResourceModel]
	framework.WithImportByID
}

func (r *frameworkResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compliance_type": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"framework_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"control_sets": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[controlSetModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: framework.IDAttribute(),
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[frameworkControlModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrID: schema.StringAttribute{
										Required: true,
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

func (r *frameworkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data frameworkResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input auditmanager.CreateAssessmentFrameworkInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAssessmentFramework(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Framework (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	framework := output.Framework
	data.ARN = fwflex.StringToFramework(ctx, framework.Arn)
	response.Diagnostics.Append(fwflex.Flatten(ctx, framework.ControlSets, &data.ControlSets)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = fwflex.StringToFramework(ctx, framework.Id)
	data.Type = fwflex.StringValueToFramework(ctx, framework.Type)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *frameworkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data frameworkResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := findFrameworkByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Framework (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *frameworkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old frameworkResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	if !new.ControlSets.Equal(old.ControlSets) ||
		!new.ComplianceType.Equal(old.ComplianceType) ||
		!new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) {
		var input auditmanager.UpdateAssessmentFrameworkInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.FrameworkId = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateAssessmentFramework(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Audit Manager Framework (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *frameworkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data frameworkResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	input := auditmanager.DeleteAssessmentFrameworkInput{
		FrameworkId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteAssessmentFramework(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Framework (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *frameworkResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var data frameworkResourceModel
		response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
		if response.Diagnostics.HasError() {
			return
		}

		controlSets, diags := data.ControlSets.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// A net-new control set will be missing id, so force a replacement
		//
		// Attribute level plan modifiers are applied before resource modifiers, so ID's
		// previously in state should never be unknown.
		for _, controlSet := range controlSets {
			if controlSet.ID.IsUnknown() {
				response.RequiresReplace = []path.Path{path.Root("control_sets")}
			}
		}
	}
}

func findFrameworkByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Framework, error) {
	input := auditmanager.GetAssessmentFrameworkInput{
		FrameworkId: aws.String(id),
	}
	output, err := conn.GetAssessmentFramework(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Framework == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Framework, nil
}

type frameworkResourceModel struct {
	framework.WithRegionModel
	ARN            types.String                                    `tfsdk:"arn"`
	ComplianceType types.String                                    `tfsdk:"compliance_type"`
	ControlSets    fwtypes.SetNestedObjectValueOf[controlSetModel] `tfsdk:"control_sets"`
	Description    types.String                                    `tfsdk:"description"`
	ID             types.String                                    `tfsdk:"id"`
	Name           types.String                                    `tfsdk:"name"`
	Tags           tftags.Map                                      `tfsdk:"tags"`
	TagsAll        tftags.Map                                      `tfsdk:"tags_all"`
	Type           types.String                                    `tfsdk:"framework_type"`
}

type controlSetModel struct {
	Controls fwtypes.SetNestedObjectValueOf[frameworkControlModel] `tfsdk:"controls"`
	ID       types.String                                          `tfsdk:"id"`
	Name     types.String                                          `tfsdk:"name"`
}

type frameworkControlModel struct {
	ID types.String `tfsdk:"id"`
}
