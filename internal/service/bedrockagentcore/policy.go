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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_policy", name="Policy")
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithTimeouts
}

func (r *policyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
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
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status_reasons": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"validation_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyValidationMode](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cedar": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cedarPolicyModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"statement": schema.StringAttribute{
										Required: true,
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

	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	policyEngineID, policyID := aws.ToString(out.PolicyEngineId), aws.ToString(out.PolicyId)

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	waited, err := waitPolicyCreated(ctx, conn, policyEngineID, policyID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root("policy_engine_id"), policyEngineID)
		response.State.SetAttribute(ctx, path.Root("policy_id"), policyID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, waited, &data))
	if response.Diagnostics.HasError() {
		return
	}

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

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
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
		input := bedrockagentcorecontrol.UpdatePolicyInput{
			PolicyEngineId: aws.String(policyEngineID),
			PolicyId:       aws.String(policyID),
		}

		// UpdatePolicy requires at least one of Definition / Description, and
		// always re-applies ValidationMode against the new (or current) Definition.
		// To keep server-side validation aligned with the configured validation_mode
		// across single-attribute changes, always send the planned Definition and
		// ValidationMode.
		definition, dd := plan.Definition.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, dd)
		if response.Diagnostics.HasError() {
			return
		}
		expanded, dd := definition.Expand(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, dd)
		if response.Diagnostics.HasError() {
			return
		}
		input.Definition = expanded.(awstypes.PolicyDefinition)
		input.ValidationMode = plan.ValidationMode.ValueEnum()

		if !plan.Description.Equal(state.Description) {
			input.Description = &awstypes.UpdatedDescription{
				OptionalValue: plan.Description.ValueStringPointer(),
			}
		}

		_, err := conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
			return
		}

		waited, err := waitPolicyUpdated(ctx, conn, policyEngineID, policyID, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, waited, &plan))
		if response.Diagnostics.HasError() {
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

func (r *policyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")

	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`Unexpected format for import ID (%s), use: "policy_engine_id,policy_id"`, request.ID))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("policy_engine_id"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("policy_id"), parts[1]))
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

type policyResourceModel struct {
	framework.WithRegionModel
	CreatedAt      timetypes.RFC3339                                      `tfsdk:"created_at"`
	Definition     fwtypes.ListNestedObjectValueOf[policyDefinitionModel] `tfsdk:"definition"`
	Description    types.String                                           `tfsdk:"description"`
	Name           types.String                                           `tfsdk:"name"`
	PolicyARN      types.String                                           `tfsdk:"policy_arn"`
	PolicyEngineID types.String                                           `tfsdk:"policy_engine_id"`
	PolicyID       types.String                                           `tfsdk:"policy_id"`
	Status         fwtypes.StringEnum[awstypes.PolicyStatus]              `tfsdk:"status"`
	StatusReasons  fwtypes.ListOfString                                   `tfsdk:"status_reasons"`
	UpdatedAt      timetypes.RFC3339                                      `tfsdk:"updated_at"`
	ValidationMode fwtypes.StringEnum[awstypes.PolicyValidationMode]      `tfsdk:"validation_mode"`
	Timeouts       timeouts.Value                                         `tfsdk:"timeouts"`
}

type policyDefinitionModel struct {
	Cedar fwtypes.ListNestedObjectValueOf[cedarPolicyModel] `tfsdk:"cedar"`
}

var (
	_ fwflex.Expander  = policyDefinitionModel{}
	_ fwflex.Flattener = &policyDefinitionModel{}
)

func (m policyDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	cedar, d := m.Cedar.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}

	var r awstypes.PolicyDefinitionMemberCedar
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, cedar, &r.Value))
	if diags.HasError() {
		return nil, diags
	}
	return &r, diags
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
