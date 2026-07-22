// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_policy", name="Policy")
// @IdentityAttribute("policy_engine_id")
// @IdentityAttribute("policy_id")
// @ImportIDHandler(policyImportID)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccPolicyImportStateIDFunc")
// @Testing(importStateIdAttribute="policy_id")
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(importIgnore="validation_mode")
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *policyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"enforcement_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnforcementMode](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 48),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_arn": framework.ARNAttributeComputedOnly(),
			"policy_engine_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(12, 59),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z][A-Za-z0-9_]*-[a-z0-9_]{10}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": framework.IDAttribute(),
			"validation_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyValidationMode](),
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
							// UpdatePolicy rejects changing the definition type in place
							// ("Changing policy type is not permitted"), so switching the
							// union variant (cedar <-> policy) forces replacement. Editing
							// the statement within a variant remains an in-place update.
							var prev, plan policyDefinitionModel
							smerr.AddEnrich(ctx, &response.Diagnostics, request.State.GetAttribute(ctx, path.Root("definition").AtListIndex(0), &prev))
							smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, path.Root("definition").AtListIndex(0), &plan))
							if response.Diagnostics.HasError() {
								return
							}
							if (!prev.Cedar.IsNull() && !plan.Policy.IsNull()) ||
								(!prev.Policy.IsNull() && !plan.Cedar.IsNull()) {
								response.RequiresReplace = true
							}
						},
						"Changing the policy definition type (cedar <-> policy) requires replacement",
						"",
					),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cedar": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cedarPolicyModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									// If another member is added to the union, this will need to be updated.
									path.MatchRelative().AtParent().AtName("cedar"),
									path.MatchRelative().AtParent().AtName(names.AttrPolicy),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"statement": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(35, 10000),
										},
									},
								},
							},
						},
						names.AttrPolicy: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[policyStatementModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									// If another member is added to the union, this will need to be updated.
									path.MatchRelative().AtParent().AtName("cedar"),
									path.MatchRelative().AtParent().AtName(names.AttrPolicy),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"statement": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(35, 10000),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *policyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data policyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreatePolicyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.ValueString())
		return
	}

	policyEngineID, policyID := aws.ToString(out.PolicyEngineId), aws.ToString(out.PolicyId)

	policy, err := waitPolicyCreated(ctx, conn, policyEngineID, policyID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("policy_engine_id"), policyEngineID)
		response.State.SetAttribute(ctx, path.Root("policy_id"), policyID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
		return
	}

	// Set values for unknowns.
	data.PolicyARN = fwflex.StringToFramework(ctx, out.PolicyArn)
	data.PolicyID = fwflex.StringValueToFramework(ctx, policyID)
	// enforcement_mode is Optional+Computed; populate the server-resolved value.
	data.EnforcementMode = fwtypes.StringEnumValue(policy.EnforcementMode)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *policyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data policyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	policyEngineID, policyID := fwflex.StringValueFromFramework(ctx, data.PolicyEngineID), fwflex.StringValueFromFramework(ctx, data.PolicyID)
	out, err := findPolicyByTwoPartKey(ctx, conn, policyEngineID, policyID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *policyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state policyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		policyEngineID, policyID := fwflex.StringValueFromFramework(ctx, plan.PolicyEngineID), fwflex.StringValueFromFramework(ctx, plan.PolicyID)
		var input bedrockagentcorecontrol.UpdatePolicyInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithIgnoredFieldNamesAppend("Description")))
		if response.Diagnostics.HasError() {
			return
		}

		if !plan.Description.Equal(state.Description) {
			input.Description = &awstypes.UpdatedDescription{}
			if !plan.Description.IsNull() {
				input.Description.OptionalValue = plan.Description.ValueStringPointer()
			}
		}

		_, err := conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
			return
		}

		if _, err := waitPolicyUpdated(ctx, conn, policyEngineID, policyID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *policyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data policyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	policyEngineID, policyID := fwflex.StringValueFromFramework(ctx, data.PolicyEngineID), fwflex.StringValueFromFramework(ctx, data.PolicyID)
	input := bedrockagentcorecontrol.DeletePolicyInput{
		PolicyEngineId: aws.String(policyEngineID),
		PolicyId:       aws.String(policyID),
	}

	_, err := conn.DeletePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
		return
	}

	if _, err := waitPolicyDeleted(ctx, conn, policyEngineID, policyID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
		return
	}
}

func (r *policyResource) flatten(ctx context.Context, policy any, data *policyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, policy, data)...)
	return diags
}

func waitPolicyCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, policyEngineID, policyID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PolicyStatusCreating),
		Target:                    enum.Slice(awstypes.PolicyStatusActive),
		Refresh:                   statusPolicy(conn, policyEngineID, policyID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPolicyUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, policyEngineID, policyID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PolicyStatusUpdating),
		Target:                    enum.Slice(awstypes.PolicyStatusActive),
		Refresh:                   statusPolicy(conn, policyEngineID, policyID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPolicyDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, policyEngineID, policyID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PolicyStatusDeleting, awstypes.PolicyStatusActive),
		Target:  []string{},
		Refresh: statusPolicy(conn, policyEngineID, policyID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPolicy(conn *bedrockagentcorecontrol.Client, policyEngineID, policyID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPolicyByTwoPartKey(ctx, conn, policyEngineID, policyID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findPolicyByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, policyEngineID, policyID string) (*bedrockagentcorecontrol.GetPolicyOutput, error) {
	input := bedrockagentcorecontrol.GetPolicyInput{
		PolicyEngineId: aws.String(policyEngineID),
		PolicyId:       aws.String(policyID),
	}

	return findPolicy(ctx, conn, &input)
}

func findPolicy(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetPolicyInput) (*bedrockagentcorecontrol.GetPolicyOutput, error) {
	out, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

var (
	_ inttypes.ImportIDParser = policyImportID{}
)

type policyImportID struct{}

func (policyImportID) Parse(id string) (string, map[string]any, error) {
	const (
		policyIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, policyIDParts, true)

	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"policy_engine_id": parts[0],
		"policy_id":        parts[1],
	}

	return id, result, nil
}

type policyResourceModel struct {
	framework.WithRegionModel
	Definition      fwtypes.ListNestedObjectValueOf[policyDefinitionModel] `tfsdk:"definition"`
	Description     types.String                                           `tfsdk:"description"`
	EnforcementMode fwtypes.StringEnum[awstypes.EnforcementMode]           `tfsdk:"enforcement_mode"`
	Name            types.String                                           `tfsdk:"name"`
	PolicyARN       types.String                                           `tfsdk:"policy_arn"`
	PolicyEngineID  types.String                                           `tfsdk:"policy_engine_id"`
	PolicyID        types.String                                           `tfsdk:"policy_id"`
	ValidationMode  fwtypes.StringEnum[awstypes.PolicyValidationMode]      `tfsdk:"validation_mode"`
	Timeouts        timeouts.Value                                         `tfsdk:"timeouts"`
}

type policyDefinitionModel struct {
	Cedar  fwtypes.ListNestedObjectValueOf[cedarPolicyModel]     `tfsdk:"cedar"`
	Policy fwtypes.ListNestedObjectValueOf[policyStatementModel] `tfsdk:"policy"`
}

var (
	_ fwflex.Expander  = policyDefinitionModel{}
	_ fwflex.Flattener = &policyDefinitionModel{}
)

func (m policyDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Cedar.IsNull():
		data, d := m.Cedar.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PolicyDefinitionMemberCedar
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags

	case !m.Policy.IsNull():
		data, d := m.Policy.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PolicyDefinitionMemberPolicy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

func (m *policyDefinitionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PolicyDefinitionMemberCedar:
		var model cedarPolicyModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Cedar = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.PolicyDefinitionMemberPolicy:
		var model policyStatementModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Policy = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("policy definition flatten: %T", v),
		)
	}
	return diags
}

type cedarPolicyModel struct {
	Statement types.String `tfsdk:"statement"`
}

type policyStatementModel struct {
	Statement types.String `tfsdk:"statement"`
}
