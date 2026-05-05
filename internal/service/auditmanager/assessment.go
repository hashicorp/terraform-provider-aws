// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

const iamPropagationTimeout = 2 * time.Minute

// @FrameworkResource("aws_auditmanager_assessment", name="Assessment")
// @Tags(identifierAttribute="arn")
func newAssessmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assessmentResource{}, nil
}

type assessmentResource struct {
	framework.ResourceWithModel[assessmentResourceModel]
	framework.WithImportByID
}

func (r *assessmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"framework_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			// The roles attribute is split into "roles" and "roles_all" to account for roles
			// that are given access to assessments by default. It isn't possible for this attribute
			// to be both Required (CreateAssessment and UpdateAssessment both require non-empty
			// values) and Computed (capturing roles with access by default and returned in
			// the response output). "roles" stores the items specifically added by the practitioner,
			// while "roles_all" will track everything with access to the assessment.
			//
			"roles_all": framework.ResourceComputedListOfObjectsAttribute[roleModel](ctx, listplanmodifier.UseStateForUnknown()),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"assessment_reports_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[assessmentReportsDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDestination: schema.StringAttribute{
							Required: true,
						},
						"destination_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssessmentReportDestinationType](),
							Required:   true,
						},
					},
				},
			},
			"roles": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[roleModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"role_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RoleType](),
							Required:   true,
						},
					},
				},
			},
			names.AttrScope: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scopeModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_accounts": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[awsAccountModel](ctx),
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
						"aws_services": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[awsServiceModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrServiceName: schema.StringAttribute{
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

func (r *assessmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data assessmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input auditmanager.CreateAssessmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	// Include retry handling to allow for IAM propagation
	//
	// Example:
	//   ResourceNotFoundException: The operation tried to access a nonexistent resource. The resource
	//   might not be specified correctly, or its status might not be active. Check and try again.
	outputRaw, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceNotFoundException](ctx, iamPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreateAssessment(ctx, &input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Assessment (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	assessment := outputRaw.(*auditmanager.CreateAssessmentOutput).Assessment
	data.ARN = fwflex.StringToFramework(ctx, assessment.Arn)
	data.ID = fwflex.StringToFramework(ctx, assessment.Metadata.Id)
	response.Diagnostics.Append(fwflex.Flatten(ctx, assessment.Metadata.Roles, &data.RolesAll)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Status = fwflex.StringValueToFramework(ctx, assessment.Metadata.Status)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *assessmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data assessmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := findAssessmentByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Assessment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	save := data.Roles
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Metadata, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	if output.Framework != nil {
		data.FrameworkID = fwflex.StringToFramework(ctx, output.Framework.Id)
	}
	data.Roles = save
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Metadata.Roles, &data.RolesAll)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *assessmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old assessmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	if !new.AssessmentReportsDestination.Equal(old.AssessmentReportsDestination) ||
		!new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.Roles.Equal(old.Roles) ||
		!new.Scope.Equal(old.Scope) {
		var input auditmanager.UpdateAssessmentInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.AssessmentDescription = fwflex.StringFromFramework(ctx, new.Description)
		input.AssessmentId = fwflex.StringFromFramework(ctx, new.ID)
		input.AssessmentName = fwflex.StringFromFramework(ctx, new.Name)

		output, err := conn.UpdateAssessment(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Audit Manager Assessment (%s)", new.ID.ValueString()), err.Error())

			return
		}

		new.Status = fwflex.StringValueToFramework(ctx, output.Assessment.Metadata.Status)
	} else {
		new.Status = old.Status
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *assessmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data assessmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	input := auditmanager.DeleteAssessmentInput{
		AssessmentId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteAssessment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Assessment (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAssessmentByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Assessment, error) {
	input := auditmanager.GetAssessmentInput{
		AssessmentId: aws.String(id),
	}
	output, err := conn.GetAssessment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Assessment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Assessment, nil
}

type assessmentResourceModel struct {
	framework.WithRegionModel
	ARN                          types.String                                                       `tfsdk:"arn"`
	AssessmentReportsDestination fwtypes.ListNestedObjectValueOf[assessmentReportsDestinationModel] `tfsdk:"assessment_reports_destination"`
	Description                  types.String                                                       `tfsdk:"description"`
	ID                           types.String                                                       `tfsdk:"id"`
	FrameworkID                  types.String                                                       `tfsdk:"framework_id"`
	Name                         types.String                                                       `tfsdk:"name"`
	Roles                        fwtypes.SetNestedObjectValueOf[roleModel]                          `tfsdk:"roles"`
	RolesAll                     fwtypes.ListNestedObjectValueOf[roleModel]                         `tfsdk:"roles_all"`
	Scope                        fwtypes.ListNestedObjectValueOf[scopeModel]                        `tfsdk:"scope"`
	Status                       types.String                                                       `tfsdk:"status"`
	Tags                         tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                         `tfsdk:"tags_all"`
}

type assessmentReportsDestinationModel struct {
	Destination     types.String                                                 `tfsdk:"destination"`
	DestinationType fwtypes.StringEnum[awstypes.AssessmentReportDestinationType] `tfsdk:"destination_type"`
}

type roleModel struct {
	RoleARN  fwtypes.ARN                           `tfsdk:"role_arn"`
	RoleType fwtypes.StringEnum[awstypes.RoleType] `tfsdk:"role_type"`
}

type scopeModel struct {
	AWSAccounts fwtypes.SetNestedObjectValueOf[awsAccountModel] `tfsdk:"aws_accounts"`
	AWSServices fwtypes.SetNestedObjectValueOf[awsServiceModel] `tfsdk:"aws_services"`
}

type awsAccountModel struct {
	ID types.String `tfsdk:"id"`
}

type awsServiceModel struct {
	ServiceName types.String `tfsdk:"service_name"`
}
