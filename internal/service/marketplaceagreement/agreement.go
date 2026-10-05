// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceagreement"
	awstypes "github.com/aws/aws-sdk-go-v2/service/marketplaceagreement/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_marketplaceagreement_agreement", name="Agreement")
// @IdentityAttribute("agreement_id")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="agreement_id")
// @Testing(importIgnore="agreement_proposal_id", plannableImportAction="Update")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/marketplaceagreement;marketplaceagreement.DescribeAgreementOutput")
// @Testing(serialize=true)
// @Testing(generator=false)
func newAgreementResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &agreementResource{}

	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type agreementResource struct {
	framework.ResourceWithModel[agreementResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *agreementResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"acceptance_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agreement_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agreement_proposal_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					// The API doesn't return the proposal an agreement was created from, so an
					// imported agreement has none in state. Recording it must not replace the agreement.
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							response.RequiresReplace = !request.StateValue.IsNull()
						},
						"Changing the proposal on an agreement created by Terraform replaces the agreement.",
						"Changing the proposal on an agreement created by Terraform replaces the agreement.",
					),
				},
			},
			"agreement_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"offer_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"proposer_account_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStartTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"requested_term": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[requestedTermModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrConfiguration: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[requestedTermConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.ExactlyOneOfChildren(
										path.MatchRelative().AtName("configurable_upfront_pricing_term"),
										path.MatchRelative().AtName("renewal_term"),
										path.MatchRelative().AtName("variable_payment_term"),
									),
								},
								Blocks: map[string]schema.Block{
									"configurable_upfront_pricing_term": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[configurableUpfrontPricingTermConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"selector_value": schema.StringAttribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"dimension": schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[dimensionModel](ctx),
													Validators: []validator.Set{
														setvalidator.IsRequired(),
														setvalidator.SizeAtLeast(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"dimension_key": schema.StringAttribute{
																Required: true,
															},
															"dimension_value": schema.Int32Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"renewal_term": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[renewalTermConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"enable_auto_renew": schema.BoolAttribute{
													Required: true,
												},
											},
										},
									},
									"variable_payment_term": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[variablePaymentTermConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"expiration_duration": schema.StringAttribute{
													Optional: true,
												},
												"payment_request_approval_strategy": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.PaymentRequestApprovalStrategy](),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *agreementResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan agreementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MarketplaceAgreementClient(ctx)

	proposalID := fwflex.StringValueFromFramework(ctx, plan.AgreementProposalID)
	input := marketplaceagreement.CreateAgreementRequestInput{
		AgreementProposalIdentifier: aws.String(proposalID),
		ClientToken:                 aws.String(create.UniqueId(ctx)),
		Intent:                      awstypes.IntentNew,
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan.RequestedTerms, &input.RequestedTerms))
	if response.Diagnostics.HasError() {
		return
	}

	createOutput, err := conn.CreateAgreementRequest(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, proposalID)
		return
	}

	acceptInput := marketplaceagreement.AcceptAgreementRequestInput{
		AgreementRequestId: createOutput.AgreementRequestId,
	}
	acceptOutput, err := conn.AcceptAgreementRequest(ctx, &acceptInput)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, aws.ToString(createOutput.AgreementRequestId))
		return
	}

	agreementID := aws.ToString(acceptOutput.AgreementId)
	out, err := findAgreementByID(ctx, conn, agreementID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agreementID)
		return
	}

	// requested_term stays as planned; Read reconciles it with the accepted terms.
	smerr.AddEnrich(ctx, &response.Diagnostics, plan.setFromAgreement(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *agreementResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state agreementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MarketplaceAgreementClient(ctx)

	agreementID := fwflex.StringValueFromFramework(ctx, state.AgreementID)
	out, err := findAgreementByID(ctx, conn, agreementID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agreementID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, state.setFromAgreement(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	terms, err := findAcceptedTermsByAgreementID(ctx, conn, agreementID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agreementID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, terms, &state.RequestedTerms))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

// Update is only reached when agreement_proposal_id is set on an imported agreement;
// every other change replaces the agreement.
func (r *agreementResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan agreementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *agreementResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state agreementResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MarketplaceAgreementClient(ctx)

	agreementID := fwflex.StringValueFromFramework(ctx, state.AgreementID)
	input := marketplaceagreement.CancelAgreementInput{
		AgreementId: aws.String(agreementID),
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// A newly accepted agreement rejects cancellation with UNSUPPORTED_ACTION for
	// roughly the first minute and a half.
	_, err := tfresource.RetryWhen(ctx, deleteTimeout,
		func(ctx context.Context) (*marketplaceagreement.CancelAgreementOutput, error) {
			return conn.CancelAgreement(ctx, &input)
		},
		func(err error) (bool, error) {
			if isValidationExceptionReason(err, awstypes.ValidationExceptionReasonUnsupportedAction) {
				return true, err
			}
			return false, err
		},
	)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agreementID)
		return
	}

	if _, err := waitAgreementCancelled(ctx, conn, agreementID, deleteTimeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, agreementID)
		return
	}
}

func isValidationExceptionReason(err error, reason awstypes.ValidationExceptionReason) bool {
	var ve *awstypes.ValidationException
	return errors.As(err, &ve) && ve.Reason == reason
}

func findAgreementByID(ctx context.Context, conn *marketplaceagreement.Client, id string) (*marketplaceagreement.DescribeAgreementOutput, error) {
	input := marketplaceagreement.DescribeAgreementInput{
		AgreementId: aws.String(id),
	}
	output, err := findAgreement(ctx, conn, &input)
	if err != nil {
		return nil, err
	}

	// An agreement that isn't ACTIVE (cancelled, expired, renewed into a new agreement, ...)
	// no longer grants anything, so treat it as gone.
	if status := output.Status; status != awstypes.AgreementStatusActive {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(status),
		})
	}

	return output, nil
}

func findAgreement(ctx context.Context, conn *marketplaceagreement.Client, input *marketplaceagreement.DescribeAgreementInput) (*marketplaceagreement.DescribeAgreementOutput, error) {
	output, err := conn.DescribeAgreement(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func findAcceptedTermsByAgreementID(ctx context.Context, conn *marketplaceagreement.Client, id string) ([]awstypes.AcceptedTerm, error) {
	input := marketplaceagreement.GetAgreementTermsInput{
		AgreementId: aws.String(id),
	}
	var output []awstypes.AcceptedTerm

	pages := marketplaceagreement.NewGetAgreementTermsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.AcceptedTerms...)
	}

	return output, nil
}

func waitAgreementCancelled(ctx context.Context, conn *marketplaceagreement.Client, id string, timeout time.Duration) (*marketplaceagreement.DescribeAgreementOutput, error) {
	output, err := tfresource.RetryUntilNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findAgreementByID(ctx, conn, id)
	})
	if output, ok := output.(*marketplaceagreement.DescribeAgreementOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

var (
	_ fwflex.Flattener = &requestedTermModel{}
)

// setFromAgreement copies DescribeAgreement's output onto the computed attributes.
// AutoFlex matches the top-level fields by name; OfferId and the proposer's AccountId
// sit in nested structs, so those are mapped from the nested structs directly.
func (m *agreementResourceModel) setFromAgreement(ctx context.Context, out *marketplaceagreement.DescribeAgreementOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, m))
	if out.ProposalSummary != nil {
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out.ProposalSummary, m))
	}
	if out.Proposer != nil {
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out.Proposer, m, fwflex.WithFieldNamePrefix("Proposer")))
	}

	return diags
}

// Flatten maps an accepted term back to the requested term that produced it, so that
// requested_term reads back exactly as configured.
func (m *requestedTermModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	configuration := requestedTermConfigurationModel{
		ConfigurableUpfrontPricingTerm: fwtypes.NewListNestedObjectValueOfNull[configurableUpfrontPricingTermConfigurationModel](ctx),
		RenewalTerm:                    fwtypes.NewListNestedObjectValueOfNull[renewalTermConfigurationModel](ctx),
		VariablePaymentTerm:            fwtypes.NewListNestedObjectValueOfNull[variablePaymentTermConfigurationModel](ctx),
	}
	configured := false
	var id *string

	switch t := v.(type) {
	case awstypes.AcceptedTermMemberByolPricingTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberConfigurableUpfrontPricingTerm:
		id = t.Value.Id
		if c := t.Value.Configuration; c != nil {
			var model configurableUpfrontPricingTermConfigurationModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, c, &model))
			configuration.ConfigurableUpfrontPricingTerm = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			configured = true
		}
	case awstypes.AcceptedTermMemberFixedUpfrontPricingTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberFreeTrialPricingTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberLegalTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberNetPaymentTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberPaymentScheduleTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberRecurringPaymentTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberRenewalTerm:
		id = t.Value.Id
		if c := t.Value.Configuration; c != nil {
			var model renewalTermConfigurationModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, c, &model))
			configuration.RenewalTerm = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			configured = true
		}
	case awstypes.AcceptedTermMemberSupportTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberUsageBasedPricingTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberValidityTerm:
		id = t.Value.Id
	case awstypes.AcceptedTermMemberVariablePaymentTerm:
		id = t.Value.Id
		if c := t.Value.Configuration; c != nil {
			var model variablePaymentTermConfigurationModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, c, &model))
			configuration.VariablePaymentTerm = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			configured = true
		}
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("requestedTermModel.Flatten: %T", v))
	}
	if diags.HasError() {
		return diags
	}

	m.ID = fwflex.StringToFramework(ctx, id)
	m.Configuration = fwtypes.NewListNestedObjectValueOfNull[requestedTermConfigurationModel](ctx)
	if configured {
		m.Configuration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &configuration)
	}

	return diags
}

type agreementResourceModel struct {
	AcceptanceTime      timetypes.RFC3339                                  `tfsdk:"acceptance_time"`
	AgreementID         types.String                                       `tfsdk:"agreement_id"`
	AgreementProposalID types.String                                       `tfsdk:"agreement_proposal_id"`
	AgreementType       types.String                                       `tfsdk:"agreement_type"`
	EndTime             timetypes.RFC3339                                  `tfsdk:"end_time"`
	OfferID             types.String                                       `tfsdk:"offer_id"`
	ProposerAccountID   types.String                                       `tfsdk:"proposer_account_id"`
	RequestedTerms      fwtypes.SetNestedObjectValueOf[requestedTermModel] `tfsdk:"requested_term"`
	StartTime           timetypes.RFC3339                                  `tfsdk:"start_time"`
	Timeouts            timeouts.Value                                     `tfsdk:"timeouts"`
}

type requestedTermModel struct {
	Configuration fwtypes.ListNestedObjectValueOf[requestedTermConfigurationModel] `tfsdk:"configuration"`
	ID            types.String                                                     `tfsdk:"id"`
}

type requestedTermConfigurationModel struct {
	ConfigurableUpfrontPricingTerm fwtypes.ListNestedObjectValueOf[configurableUpfrontPricingTermConfigurationModel] `tfsdk:"configurable_upfront_pricing_term"`
	RenewalTerm                    fwtypes.ListNestedObjectValueOf[renewalTermConfigurationModel]                    `tfsdk:"renewal_term"`
	VariablePaymentTerm            fwtypes.ListNestedObjectValueOf[variablePaymentTermConfigurationModel]            `tfsdk:"variable_payment_term"`
}

var (
	_ fwflex.Expander = requestedTermConfigurationModel{}
)

func (m requestedTermConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.ConfigurableUpfrontPricingTerm.IsNull():
		model, d := m.ConfigurableUpfrontPricingTerm.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RequestedTermConfigurationMemberConfigurableUpfrontPricingTermConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.RenewalTerm.IsNull():
		model, d := m.RenewalTerm.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RequestedTermConfigurationMemberRenewalTermConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.VariablePaymentTerm.IsNull():
		model, d := m.VariablePaymentTerm.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.RequestedTermConfigurationMemberVariablePaymentTermConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type configurableUpfrontPricingTermConfigurationModel struct {
	Dimensions    fwtypes.SetNestedObjectValueOf[dimensionModel] `tfsdk:"dimension"`
	SelectorValue types.String                                   `tfsdk:"selector_value"`
}

type dimensionModel struct {
	DimensionKey   types.String `tfsdk:"dimension_key"`
	DimensionValue types.Int32  `tfsdk:"dimension_value"`
}

type renewalTermConfigurationModel struct {
	EnableAutoRenew types.Bool `tfsdk:"enable_auto_renew"`
}

type variablePaymentTermConfigurationModel struct {
	ExpirationDuration             types.String                                                `tfsdk:"expiration_duration"`
	PaymentRequestApprovalStrategy fwtypes.StringEnum[awstypes.PaymentRequestApprovalStrategy] `tfsdk:"payment_request_approval_strategy"`
}
