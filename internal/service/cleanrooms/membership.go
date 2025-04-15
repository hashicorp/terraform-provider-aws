// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameMembership = "Membership"
)

// @FrameworkResource("aws_cleanrooms_membership",name="Membership")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cleanrooms;cleanrooms.GetMembershipOutput")
// @Testing(checkDestroyNoop=true)
func newResourceMembership(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceMembership{}

	return r, nil
}

type resourceMembership struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceMembership) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"collaboration_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collaboration_creator_account_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collaboration_creator_display_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collaboration_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collaboration_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreateTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"member_abilities": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"query_log_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MembershipQueryLogStatus](),
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"update_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"default_result_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[defaultResultConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"output_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[outputConfiguration](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3ModelData](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
												},
												"key_prefix": schema.StringAttribute{
													Optional: true,
												},
												"result_format": schema.StringAttribute{
													Required: true,
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
			"payment_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[paymentConfiguration](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"query_compute": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[queryComputeData](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"is_responsible": schema.BoolAttribute{
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

	response.Schema = s
}

func (r *resourceMembership) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceMembershipData
	conn := r.Meta().CleanRoomsClient(ctx)

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := cleanrooms.CreateMembershipInput{
		CollaborationIdentifier: data.CollaborationID.ValueStringPointer(),
		Tags:                    getTagsIn(ctx),
	}

	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateMembership(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CleanRooms, create.ErrActionCreating, ResNameMembership, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Membership, &data, fwflex.WithIgnoredFieldNamesAppend("PaymentConfiguration"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceMembership) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceMembershipData
	conn := r.Meta().CleanRoomsClient(ctx)

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findMembershipByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CleanRooms, create.ErrActionReading, ResNameMembership, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Membership, &data, fwflex.WithIgnoredFieldNamesAppend("PaymentConfiguration"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceMembership) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resourceMembershipData
	conn := r.Meta().CleanRoomsClient(ctx)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := cleanrooms.UpdateMembershipInput{
			MembershipIdentifier: plan.ID.ValueStringPointer(),
		}

		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateMembership(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CleanRooms, create.ErrActionUpdating, ResNameMembership, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output.Membership, &plan, fwflex.WithIgnoredFieldNamesAppend("PaymentConfiguration"))...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceMembership) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceMembershipData
	conn := r.Meta().CleanRoomsClient(ctx)
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting CleanRooms Membership", map[string]any{
		names.AttrID: data.ID.ValueString(),
	})

	input := cleanrooms.DeleteMembershipInput{
		MembershipIdentifier: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteMembership(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	// Occurs when the membership was created with the same account as the collaboration.
	// There is no way to force delete the membership as it is default to the collaboration itself.
	if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Cannot delete membership for the creator of the collaboration while their membership is still ACTIVE") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CleanRooms, create.ErrActionDeleting, ResNameMembership, data.ID.String(), err),
			err.Error(),
		)
	}
}

type resourceMembershipData struct {
	ARN                             types.String                                                `tfsdk:"arn"`
	CollaborationARN                types.String                                                `tfsdk:"collaboration_arn"`
	CollaborationCreatorAccountID   types.String                                                `tfsdk:"collaboration_creator_account_id"`
	CollaborationCreatorDisplayName types.String                                                `tfsdk:"collaboration_creator_display_name"`
	CollaborationID                 types.String                                                `tfsdk:"collaboration_id"`
	CollaborationName               types.String                                                `tfsdk:"collaboration_name"`
	CreateTime                      timetypes.RFC3339                                           `tfsdk:"create_time"`
	DefaultResultConfiguration      fwtypes.ListNestedObjectValueOf[defaultResultConfiguration] `tfsdk:"default_result_configuration"`
	ID                              types.String                                                `tfsdk:"id"`
	MemberAbilities                 fwtypes.ListValueOf[types.String]                           `tfsdk:"member_abilities"`
	PaymentConfiguration            fwtypes.ListNestedObjectValueOf[paymentConfiguration]       `tfsdk:"payment_configuration"`
	QueryLogStatus                  fwtypes.StringEnum[awstypes.MembershipQueryLogStatus]       `tfsdk:"query_log_status"`
	Status                          types.String                                                `tfsdk:"status"`
	Tags                            tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                         tftags.Map                                                  `tfsdk:"tags_all"`
	UpdateTime                      timetypes.RFC3339                                           `tfsdk:"update_time"`
}

type defaultResultConfiguration struct {
	OutputConfiguration fwtypes.ListNestedObjectValueOf[outputConfiguration] `tfsdk:"output_configuration"`
	RoleARN             fwtypes.ARN                                          `tfsdk:"role_arn"`
}

type paymentConfiguration struct {
	QueryCompute fwtypes.ListNestedObjectValueOf[queryComputeData] `tfsdk:"query_compute"`
}

type queryComputeData struct {
	IsResponsible types.Bool `tfsdk:"is_responsible"`
}

type s3ModelData struct {
	Bucket       types.String `tfsdk:"bucket"`
	KeyPrefix    types.String `tfsdk:"key_prefix"`
	ResultFormat types.String `tfsdk:"result_format"`
}

var (
	_ fwflex.Expander  = outputConfiguration{}
	_ fwflex.Flattener = (*outputConfiguration)(nil)
)

type outputConfiguration struct {
	S3 fwtypes.ListNestedObjectValueOf[s3ModelData] `tfsdk:"s3"`
}

func (m outputConfiguration) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.S3.IsNull():
		s3Data, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var s awstypes.MembershipProtectedQueryOutputConfigurationMemberS3
		diags.Append(fwflex.Expand(ctx, s3Data, &s.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &s, diags
	}

	return nil, diags
}

func (m *outputConfiguration) Flatten(ctx context.Context, input any) (diags diag.Diagnostics) {
	switch t := input.(type) {
	case awstypes.MembershipProtectedQueryOutputConfigurationMemberS3:
		var model s3ModelData
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	}

	return diags
}

func findMembershipByID(ctx context.Context, conn *cleanrooms.Client, id string) (*cleanrooms.GetMembershipOutput, error) {
	in := &cleanrooms.GetMembershipInput{
		MembershipIdentifier: aws.String(id),
	}

	out, err := conn.GetMembership(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Membership == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
