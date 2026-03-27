// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_policy_engine", name="Policy Engine")
// @Tags(identifierAttribute="policy_engine_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol.GetPolicyEngineOutput")
// @Testing(hasNoPreExistingResource=true)
// @IdentityAttribute("id")
func newPolicyEngineResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyEngineResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNamePolicyEngine = "Policy Engine"
)

type policyEngineResource struct {
	framework.ResourceWithModel[policyEngineResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *policyEngineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`),
						"must start with a letter and contain only letters, numbers, and underscores",
					),
					stringvalidator.LengthBetween(1, 48),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_engine_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *policyEngineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan policyEngineResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreatePolicyEngineInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePolicyEngine(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	// CreatePolicyEngineOutput has inline fields (no nested .PolicyEngine struct).
	// Set the computed identifiers before waiting so the waiter can find the resource.
	plan.ID = types.StringPointerValue(out.PolicyEngineId)
	plan.PolicyEngineARN = types.StringPointerValue(out.PolicyEngineArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitPolicyEngineCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyEngineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state policyEngineResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPolicyEngineByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// Flatten the response. PolicyEngineId maps to ID which needs manual assignment
	// since the model field is named "id" not "policy_engine_id".
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringPointerValue(out.PolicyEngineId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyEngineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state policyEngineResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// UpdatePolicyEngine only supports updating description.
	// The API does not support clearing description once set (min length 1).
	// Tags are handled separately by the tag framework.
	if !plan.Description.Equal(state.Description) && !plan.Description.IsNull() {
		description := &awstypes.UpdatedDescription{
			OptionalValue: aws.String(plan.Description.ValueString()),
		}

		_, err := conn.UpdatePolicyEngine(ctx, &bedrockagentcorecontrol.UpdatePolicyEngineInput{
			PolicyEngineId: plan.ID.ValueStringPointer(),
			Description:    description,
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitPolicyEngineUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *policyEngineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state policyEngineResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeletePolicyEngine(ctx, &bedrockagentcorecontrol.DeletePolicyEngineInput{
		PolicyEngineId: state.ID.ValueStringPointer(),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPolicyEngineDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitPolicyEngineCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyEngineOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PolicyEngineStatusCreating),
		Target:                    enum.Slice(awstypes.PolicyEngineStatusActive),
		Refresh:                   statusPolicyEngine(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyEngineOutput); ok {
		if len(out.StatusReasons) > 0 {
			retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, ", ")))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPolicyEngineUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyEngineOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PolicyEngineStatusUpdating),
		Target:                    enum.Slice(awstypes.PolicyEngineStatusActive),
		Refresh:                   statusPolicyEngine(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyEngineOutput); ok {
		if len(out.StatusReasons) > 0 {
			retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, ", ")))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPolicyEngineDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPolicyEngineOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PolicyEngineStatusDeleting, awstypes.PolicyEngineStatusActive),
		Target:  []string{},
		Refresh: statusPolicyEngine(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPolicyEngineOutput); ok {
		if len(out.StatusReasons) > 0 {
			retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, ", ")))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPolicyEngine(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPolicyEngineByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findPolicyEngineByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetPolicyEngineOutput, error) {
	out, err := conn.GetPolicyEngine(ctx, &bedrockagentcorecontrol.GetPolicyEngineInput{
		PolicyEngineId: aws.String(id),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type policyEngineResourceModel struct {
	framework.WithRegionModel
	Description      types.String   `tfsdk:"description"`
	EncryptionKeyARN fwtypes.ARN    `tfsdk:"encryption_key_arn"`
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	PolicyEngineARN  types.String   `tfsdk:"policy_engine_arn"`
	Tags             tftags.Map     `tfsdk:"tags"`
	TagsAll          tftags.Map     `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
