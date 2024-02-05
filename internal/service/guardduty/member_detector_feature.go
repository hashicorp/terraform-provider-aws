// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Member Detector Feature")
func newResourceMemberDetectorFeature(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceMemberDetectorFeature{}
	return r, nil
}

const (
	memberDetectorFeatureResourceIDPartCount = 3
	memberDetectorFeatureResourceName        = "Member Detector Feature"
	memberDetectorFeatureResourceTypeName    = "aws_guardduty_member_detector_feature"
)

type resourceMemberDetectorFeature struct {
	framework.ResourceWithConfigure
}

func (r *resourceMemberDetectorFeature) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = memberDetectorFeatureResourceTypeName
}

func (r *resourceMemberDetectorFeature) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"detector_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.DetectorFeature](),
				},
			},
			"status": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.FeatureStatus](),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"additional_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[additionalConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.OrgFeatureAdditionalConfiguration](),
							},
						},
						"status": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.FeatureStatus](),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceMemberDetectorFeature) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GuardDutyClient(ctx)

	var plan resourceMemberDetectorFeatureModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := createUpdateMemberDetectorsInput(plan)

	if !plan.AdditionalConfiguration.IsNull() {
		resp.Diagnostics.Append(fwflex.Expand(ctx, &plan.AdditionalConfiguration, &in.Features[0].AdditionalConfiguration)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, err := updateMemberDetectorFeature(ctx, conn, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GuardDuty, create.ErrActionCreating, memberDetectorFeatureResourceName, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set the ID and save the state
	plan.setID()

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceMemberDetectorFeature) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceMemberDetectorFeatureModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().GuardDutyClient(ctx)

	output, err := FindMemberDetectorFeatureByThreePartKey(ctx, conn, data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GuardDuty, create.ErrActionReading, memberDetectorFeatureResourceName, data.ID.ValueString(), err),
			err.Error(),
		)

		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.AdditionalConfiguration, &data.AdditionalConfiguration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceMemberDetectorFeature) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceMemberDetectorFeatureModel

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GuardDutyClient(ctx)

	in := createUpdateMemberDetectorsInput(new)

	if !new.AdditionalConfiguration.IsNull() {
		resp.Diagnostics.Append(fwflex.Expand(ctx, new.AdditionalConfiguration, &in.Features[0].AdditionalConfiguration)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, err := updateMemberDetectorFeature(ctx, conn, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GuardDuty, create.ErrActionUpdating, memberDetectorFeatureResourceName, new.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceMemberDetectorFeature) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op
}

// ==== HELPERS ====
func createUpdateMemberDetectorsInput(plan resourceMemberDetectorFeatureModel) *guardduty.UpdateMemberDetectorsInput {
	in := &guardduty.UpdateMemberDetectorsInput{
		AccountIds: []string{plan.AccountID.ValueString()},
		DetectorId: aws.String(plan.DetectorID.ValueString()),
		Features: []awstypes.MemberFeaturesConfiguration{
			{
				Name:   awstypes.OrgFeature(plan.Name.ValueString()),
				Status: awstypes.FeatureStatus(plan.Status.ValueString()),
			},
		},
	}
	return in
}

func updateMemberDetectorFeature(ctx context.Context, conn *guardduty.Client, in *guardduty.UpdateMemberDetectorsInput) (*guardduty.UpdateMemberDetectorsOutput, error) {
	conns.GlobalMutexKV.Lock(*in.DetectorId)
	defer conns.GlobalMutexKV.Unlock(*in.DetectorId)

	out, err := conn.UpdateMemberDetectors(ctx, in)
	if err != nil {
		return out, err
	}

	if out == nil {
		return nil, errors.New("empty output")
	}

	// For example:
	// {"unprocessedAccounts":[{"result":"The request is rejected because the given account ID is not an associated member of account the current account.","accountId":"123456789012"}]}
	if len(out.UnprocessedAccounts) > 0 {
		return out, errors.New(*(out.UnprocessedAccounts[0].Result))
	}

	return out, err
}

// ==== FINDERS ====
func FindMemberDetectorFeatureByThreePartKey(ctx context.Context, client *guardduty.Client, detectorID, accountID, name string) (*awstypes.MemberFeaturesConfigurationResult, error) {
	output, err := findMemberConfigurationByDetectorAndAccountID(ctx, client, detectorID, accountID)

	if err != nil {
		return nil, err
	}

	for _, feature := range output.Features {
		if string(feature.Name) == name {
			return &feature, nil
		}
	}

	return nil, fmt.Errorf("no MemberFeaturesConfigurationResult found with name %s", name)

}

func findMemberConfigurationByDetectorAndAccountID(ctx context.Context, client *guardduty.Client, detectorID string, accountID string) (*awstypes.MemberDataSourceConfiguration, error) {
	input := &guardduty.GetMemberDetectorsInput{
		DetectorId: aws.String(detectorID),
		AccountIds: []string{accountID},
	}

	output, err := client.GetMemberDetectors(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if output.MemberDataSourceConfigurations == nil || len(output.MemberDataSourceConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.MemberDataSourceConfigurations[0], nil
}

// ==== MODEL ====
type resourceMemberDetectorFeatureModel struct {
	AccountID               types.String                                                  `tfsdk:"account_id"`
	AdditionalConfiguration fwtypes.ListNestedObjectValueOf[additionalConfigurationModel] `tfsdk:"additional_configuration"`
	DetectorID              types.String                                                  `tfsdk:"detector_id"`
	ID                      types.String                                                  `tfsdk:"id"`
	Name                    types.String                                                  `tfsdk:"name"`
	Status                  types.String                                                  `tfsdk:"status"`
}

type additionalConfigurationModel struct {
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

func (data *resourceMemberDetectorFeatureModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, memberDetectorFeatureResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.DetectorID = types.StringValue(parts[0])
	data.AccountID = types.StringValue(parts[1])
	data.Name = types.StringValue(parts[2])

	return nil
}

func (data *resourceMemberDetectorFeatureModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString()}, memberDetectorFeatureResourceIDPartCount, false)))
}
