// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"errors"
	"fmt"
	"slices"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_guardduty_member_detector_feature", name="Member Detector Feature")
func newMemberDetectorFeatureResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &memberDetectorFeatureResource{}

	return r, nil
}

type memberDetectorFeatureResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *memberDetectorFeatureResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
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
			names.AttrName: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OrgFeature](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FeatureStatus](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"additional_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[memberAdditionalConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OrgFeatureAdditionalConfiguration](),
							Required:   true,
						},
						names.AttrStatus: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.FeatureStatus](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *memberDetectorFeatureResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data memberDetectorFeatureResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GuardDutyClient(ctx)

	input := guardduty.UpdateMemberDetectorsInput{
		AccountIds: fwflex.StringSliceValueFromFramework(ctx, data.AccountID),
		DetectorId: fwflex.StringFromFramework(ctx, data.DetectorID),
		Features: []awstypes.MemberFeaturesConfiguration{
			{
				Name:   data.Name.ValueEnum(),
				Status: data.Status.ValueEnum(),
			},
		},
	}

	if !data.AdditionalConfiguration.IsNull() {
		response.Diagnostics.Append(fwflex.Expand(ctx, &data.AdditionalConfiguration, &input.Features[0].AdditionalConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if err := updateMemberDetectors(ctx, conn, &input); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating GuardDuty Member Detector Feature (%s/%s/%s)", data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *memberDetectorFeatureResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data memberDetectorFeatureResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GuardDutyClient(ctx)

	output, err := findMemberDetectorFeatureByThreePartKey(ctx, conn, data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading GuardDuty Member Detector Feature (%s/%s/%s)", data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AdditionalConfiguration, &data.AdditionalConfiguration)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Status = fwtypes.StringEnumValue(output.Status)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *memberDetectorFeatureResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data memberDetectorFeatureResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GuardDutyClient(ctx)

	input := guardduty.UpdateMemberDetectorsInput{
		AccountIds: fwflex.StringSliceValueFromFramework(ctx, data.AccountID),
		DetectorId: fwflex.StringFromFramework(ctx, data.DetectorID),
		Features: []awstypes.MemberFeaturesConfiguration{
			{
				Name:   data.Name.ValueEnum(),
				Status: data.Status.ValueEnum(),
			},
		},
	}

	if !data.AdditionalConfiguration.IsNull() {
		response.Diagnostics.Append(fwflex.Expand(ctx, data.AdditionalConfiguration, &input.Features[0].AdditionalConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if err := updateMemberDetectors(ctx, conn, &input); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating GuardDuty Member Detector Feature (%s/%s/%s)", data.DetectorID.ValueString(), data.AccountID.ValueString(), data.Name.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func updateMemberDetectors(ctx context.Context, conn *guardduty.Client, input *guardduty.UpdateMemberDetectorsInput) error {
	dtectorID := aws.ToString(input.DetectorId)
	conns.GlobalMutexKV.Lock(dtectorID)
	defer conns.GlobalMutexKV.Unlock(dtectorID)

	output, err := conn.UpdateMemberDetectors(ctx, input)

	if err != nil {
		return err
	}

	if len(output.UnprocessedAccounts) > 0 {
		return errors.New(aws.ToString(output.UnprocessedAccounts[0].Result))
	}

	return nil
}

func findMemberDetectorFeatureByThreePartKey(ctx context.Context, client *guardduty.Client, detectorID, accountID, name string) (*awstypes.MemberFeaturesConfigurationResult, error) {
	detector, err := findMemberDetectorByTwoPartKey(ctx, client, detectorID, accountID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(detector.Features, func(v awstypes.MemberFeaturesConfigurationResult) bool {
		return string(v.Name) == name
	}))
}

func findMemberDetectorByTwoPartKey(ctx context.Context, client *guardduty.Client, detectorID, accountID string) (*awstypes.MemberDataSourceConfiguration, error) {
	input := guardduty.GetMemberDetectorsInput{
		DetectorId: aws.String(detectorID),
		AccountIds: []string{accountID},
	}

	detector, unprocessedAccounts, err := findMemberDetector(ctx, client, &input)

	if err != nil {
		return nil, err
	}

	if i := slices.IndexFunc(unprocessedAccounts, func(v awstypes.UnprocessedAccount) bool { return aws.ToString(v.AccountId) == accountID }); i >= 0 {
		return nil, errors.New(aws.ToString(unprocessedAccounts[i].Result))
	}

	return detector, nil
}

func findMemberDetector(ctx context.Context, client *guardduty.Client, input *guardduty.GetMemberDetectorsInput) (*awstypes.MemberDataSourceConfiguration, []awstypes.UnprocessedAccount, error) {
	detectors, unprocessedAccounts, err := findMemberDetectors(ctx, client, input)

	if err != nil {
		return nil, nil, err
	}

	detector, err := tfresource.AssertSingleValueResult(detectors)

	if err != nil {
		return nil, nil, err
	}

	return detector, unprocessedAccounts, nil
}

func findMemberDetectors(ctx context.Context, client *guardduty.Client, input *guardduty.GetMemberDetectorsInput) ([]awstypes.MemberDataSourceConfiguration, []awstypes.UnprocessedAccount, error) {
	output, err := client.GetMemberDetectors(ctx, input)

	if err != nil {
		return nil, nil, err
	}

	if output == nil {
		return nil, nil, tfresource.NewEmptyResultError(input)
	}

	return output.MemberDataSourceConfigurations, output.UnprocessedAccounts, nil
}

type memberDetectorFeatureResourceModel struct {
	AccountID               types.String                                                        `tfsdk:"account_id"`
	AdditionalConfiguration fwtypes.ListNestedObjectValueOf[memberAdditionalConfigurationModel] `tfsdk:"additional_configuration"`
	DetectorID              types.String                                                        `tfsdk:"detector_id"`
	Name                    fwtypes.StringEnum[awstypes.OrgFeature]                             `tfsdk:"name"`
	Status                  fwtypes.StringEnum[awstypes.FeatureStatus]                          `tfsdk:"status"`
}

type memberAdditionalConfigurationModel struct {
	Name   fwtypes.StringEnum[awstypes.OrgFeatureAdditionalConfiguration] `tfsdk:"name"`
	Status fwtypes.StringEnum[awstypes.FeatureStatus]                     `tfsdk:"status"`
}
