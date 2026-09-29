// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_user_journey", name="User Journey")
// @IdentityAttribute("system_arn")
// @IdentityAttribute("user_journey_id")
// @ImportIDHandler("userJourneyImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.UserJourney")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccCheckUserJourneyImportStateIDFunc")
// @Testing(importStateIdAttribute="user_journey_id")
func newUserJourneyResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &userJourneyResource{}, nil
}

type userJourneyResource struct {
	framework.ResourceWithModel[userJourneyResourceModel]
	framework.WithImportByIdentity
}

func (r *userJourneyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 500),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
			},
			"policy_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"system_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_journey_id": framework.IDAttribute(),
		},
	}
}

func (r *userJourneyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userJourneyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input resiliencehubv2.CreateUserJourneyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := conn.CreateUserJourney(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	plan.UserJourneyID = fwflex.StringToFramework(ctx, output.UserJourney.UserJourneyId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userJourneyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userJourneyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	systemARN, userJourneyID := fwflex.StringValueFromFramework(ctx, state.SystemARN), fwflex.StringValueFromFramework(ctx, state.UserJourneyID)
	uj, err := findUserJourneyByTwoPartKey(ctx, conn, systemARN, userJourneyID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userJourneyID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, uj, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *userJourneyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userJourneyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	userJourneyID := fwflex.StringValueFromFramework(ctx, plan.UserJourneyID)
	var input resiliencehubv2.UpdateUserJourneyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateUserJourney(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userJourneyID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userJourneyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userJourneyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	systemARN, userJourneyID := fwflex.StringValueFromFramework(ctx, state.SystemARN), fwflex.StringValueFromFramework(ctx, state.UserJourneyID)
	input := resiliencehubv2.DeleteUserJourneyInput{
		SystemArn:     aws.String(systemARN),
		UserJourneyId: aws.String(userJourneyID),
	}
	_, err := conn.DeleteUserJourney(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userJourneyID)
	}
}

func (r *userJourneyResource) flatten(ctx context.Context, uj *awstypes.UserJourney, data *userJourneyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, uj, data)...)
	if diags.HasError() {
		return diags
	}

	return diags
}

type userJourneyImportID struct{}

func (userJourneyImportID) Parse(id string) (string, map[string]any, error) {
	const (
		userJourneyImportIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, userJourneyImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"system_arn":      parts[0],
		"user_journey_id": parts[1],
	}

	return id, result, nil
}

func findUserJourneyByTwoPartKey(ctx context.Context, conn *resiliencehubv2.Client, systemArn, userJourneyId string) (*awstypes.UserJourney, error) {
	input := resiliencehubv2.GetUserJourneyInput{
		SystemArn:     aws.String(systemArn),
		UserJourneyId: aws.String(userJourneyId),
	}

	return findUserJourney(ctx, conn, &input)
}

func findUserJourney(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.GetUserJourneyInput) (*awstypes.UserJourney, error) {
	output, err := conn.GetUserJourney(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.UserJourney == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.UserJourney, nil
}

type userJourneyResourceModel struct {
	framework.WithRegionModel
	Description   types.String `tfsdk:"description"`
	Name          types.String `tfsdk:"name"`
	PolicyARN     fwtypes.ARN  `tfsdk:"policy_arn"`
	SystemARN     fwtypes.ARN  `tfsdk:"system_arn"`
	UserJourneyID types.String `tfsdk:"user_journey_id"`
}
