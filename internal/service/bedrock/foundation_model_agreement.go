// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_foundation_model_agreement", name="Foundation Model Agreement")
// @IdentityAttribute("model_id")
// @Testing(importIgnore="offer_token")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="model_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrock;bedrock.GetFoundationModelAvailabilityOutput")
// @Testing(requireEnvVarValue="AWS_BEDROCK_FOUNDATION_MODEL_ID")
// @Testing(serialize=true)
// @Testing(plannableImportAction="NoOp")
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheckFoundationModelAgreement")
// @Testing(preCheck="testAccPreCheckFoundationModelUseCase")
func newFoundationModelAgreementResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &foundationModelAgreementResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameFoundationModelAgreement = "Foundation Model Agreement"
)

type foundationModelAgreementResource struct {
	framework.ResourceWithModel[foundationModelAgreementResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *foundationModelAgreementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"model_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"offer_token": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *foundationModelAgreementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan foundationModelAgreementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrock.CreateFoundationModelAgreementInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateFoundationModelAgreement(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ModelID.String())
		return
	}
	if out == nil || out.ModelId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ModelID.String())
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitFoundationModelAgreementCreated(ctx, conn, plan.ModelID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ModelID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *foundationModelAgreementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state foundationModelAgreementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFoundationModelAgreementByID(ctx, conn, state.ModelID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ModelID.String())
		return
	}
	// If agreement is found but not available, treat as not found to allow for proper recreation and avoid confusion.
	if out != nil && out.AgreementAvailability != nil && out.AgreementAvailability.Status == awstypes.AgreementStatusNotAvailable {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(errors.New("agreement not available")))
		resp.State.RemoveResource(ctx)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *foundationModelAgreementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state foundationModelAgreementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.DeleteFoundationModelAgreementInput{
		ModelId: state.ModelID.ValueStringPointer(),
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err := tfresource.RetryWhen(ctx, deleteTimeout, func(ctx context.Context) (*bedrock.DeleteFoundationModelAgreementOutput, error) {
		return conn.DeleteFoundationModelAgreement(ctx, &input)
	}, func(err error) (bool, error) {
		// Sometimes the delete can not be done even though the resource is is an active state.
		if errs.IsA[*awstypes.ConflictException](err) {
			return true, err
		}
		return false, err
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		// Functionally equivalent to a not found error
		if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "No agreement exists to cancel") {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ModelID.String())
		return
	}

	_, err = waitFoundationModelAgreementDeleted(ctx, conn, state.ModelID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ModelID.String())
		return
	}
}

func waitFoundationModelAgreementCreated(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetFoundationModelAvailabilityOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgreementStatusPending),
		Target:                    enum.Slice(awstypes.AgreementStatusAvailable),
		Refresh:                   statusFoundationModelAgreement(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetFoundationModelAvailabilityOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFoundationModelAgreementDeleted(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetFoundationModelAvailabilityOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgreementStatusPending),
		Target:  enum.Slice(awstypes.AgreementStatusNotAvailable),
		Refresh: statusFoundationModelAgreement(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetFoundationModelAvailabilityOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusFoundationModelAgreement(conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findFoundationModelAgreementByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.AgreementAvailability.Status), nil
	}
}

func findFoundationModelAgreementByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetFoundationModelAvailabilityOutput, error) {
	input := bedrock.GetFoundationModelAvailabilityInput{
		ModelId: aws.String(id),
	}

	out, err := conn.GetFoundationModelAvailability(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AgreementAvailability == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type foundationModelAgreementResourceModel struct {
	framework.WithRegionModel
	ModelID    types.String   `tfsdk:"model_id"`
	OfferToken types.String   `tfsdk:"offer_token"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
