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

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_bedrockagentcore_consent_portal", name="Consent Portal")
// @Tags(identifierAttribute="consent_portal_arn")
// @IdentityAttribute("consent_portal_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol.GetConsentPortalOutput")
// @Testing(importStateIdAttribute="consent_portal_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckConsentPortals")
func newConsentPortalResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &consentPortalResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type consentPortalResource struct {
	framework.ResourceWithModel[consentPortalResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *consentPortalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"consent_portal_arn": framework.ARNAttributeComputedOnly(),
			"consent_portal_id":  framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 4096)},
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{Required: true, CustomType: fwtypes.ARNType},
			names.AttrName: schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"portal_url":      schema.StringAttribute{Computed: true},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"idp_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[consentPortalIDPConfigModel](ctx),
				Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"credential_provider_arn": schema.StringAttribute{Required: true, CustomType: fwtypes.ARNType},
						"scopes": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
							Validators: []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 255))},
						},
						"audience": schema.StringAttribute{Optional: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
					},
				},
			},
			"sources": schema.ListNestedBlock{
				CustomType:    fwtypes.NewListNestedObjectTypeOf[consentPortalSourceModel](ctx),
				Validators:    []validator.List{listvalidator.IsRequired(), listvalidator.SizeBetween(1, 1)},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIdentifier: schema.StringAttribute{Required: true},
						names.AttrType:       schema.StringAttribute{Required: true, CustomType: fwtypes.StringEnumType[awstypes.ConsentPortalSourceType]()},
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

func (r *consentPortalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan consentPortalResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateConsentPortalInput

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	// IAM roles can take time to become assumable after creation.
	out, err := tfresource.RetryWhenIsAErrorMessageContains[*bedrockagentcorecontrol.CreateConsentPortalOutput, *awstypes.ValidationException](ctx, propagationTimeout,
		func(ctx context.Context) (*bedrockagentcorecontrol.CreateConsentPortalOutput, error) {
			return conn.CreateConsentPortal(ctx, &input)
		}, "Execution role is not assumable")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	id := aws.ToString(out.ConsentPortalId)
	created, err := waitConsentPortalCreated(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("consent_portal_id"), id))
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *consentPortalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state consentPortalResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findConsentPortalByID(ctx, conn, state.ConsentPortalID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ConsentPortalID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *consentPortalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state consentPortalResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateConsentPortalInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ConsentPortalIdentifier = plan.ConsentPortalID.ValueStringPointer()
		// An omitted description leaves the remote value unchanged; send an empty string to clear it.
		if plan.Description.IsNull() {
			input.Description = aws.String("")
		}
		// Omitting audience preserves the existing value in AWS, including when the IdP config is supplied.
		if input.IdpConfig != nil && input.IdpConfig.Audience == nil {
			input.IdpConfig.Audience = aws.String("")
		}
		_, err := tfresource.RetryWhenIsAErrorMessageContains[*bedrockagentcorecontrol.UpdateConsentPortalOutput, *awstypes.ValidationException](ctx, propagationTimeout,
			func(ctx context.Context) (*bedrockagentcorecontrol.UpdateConsentPortalOutput, error) {
				return conn.UpdateConsentPortal(ctx, &input)
			}, "Execution role is not assumable")
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ConsentPortalID.String())
			return
		}
		if _, err := waitConsentPortalUpdated(ctx, conn, plan.ConsentPortalID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ConsentPortalID.String())
			return
		}
	}
	out, err := findConsentPortalByID(ctx, conn, plan.ConsentPortalID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ConsentPortalID.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *consentPortalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state consentPortalResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteConsentPortalInput{
		ConsentPortalIdentifier: state.ConsentPortalID.ValueStringPointer(),
	}

	_, err := conn.DeleteConsentPortal(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ConsentPortalID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitConsentPortalDeleted(ctx, conn, state.ConsentPortalID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ConsentPortalID.String())
		return
	}
}

func waitConsentPortalCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetConsentPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ConsentPortalStatusCreating),
		Target:                    enum.Slice(awstypes.ConsentPortalStatusActive),
		Refresh:                   statusConsentPortal(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetConsentPortalOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitConsentPortalUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetConsentPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ConsentPortalStatusUpdating),
		Target:                    enum.Slice(awstypes.ConsentPortalStatusActive),
		Refresh:                   statusConsentPortal(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetConsentPortalOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitConsentPortalDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetConsentPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConsentPortalStatusDeleting, awstypes.ConsentPortalStatusActive, awstypes.ConsentPortalStatusFailed, awstypes.ConsentPortalStatusUpdateFailed),
		Target:  []string{},
		Refresh: statusConsentPortal(conn, id),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetConsentPortalOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func statusConsentPortal(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findConsentPortalByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}
		return out, string(out.Status), nil
	}
}

func findConsentPortalByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetConsentPortalOutput, error) {
	// GetConsentPortal currently rejects ARNs despite documenting support for them.
	if arn.IsARN(id) {
		parsed, err := arn.Parse(id)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		portalID, ok := strings.CutPrefix(parsed.Resource, "consent-portal/")
		if parsed.Service != "bedrock-agentcore" || !ok || portalID == "" {
			return nil, smarterr.NewError(fmt.Errorf("invalid consent portal ARN: %s", id))
		}
		id = portalID
	}
	input := bedrockagentcorecontrol.GetConsentPortalInput{ConsentPortalIdentifier: aws.String(id)}
	out, err := conn.GetConsentPortal(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

type consentPortalResourceModel struct {
	framework.WithRegionModel
	ConsentPortalARN types.String                                                 `tfsdk:"consent_portal_arn"`
	ConsentPortalID  types.String                                                 `tfsdk:"consent_portal_id"`
	Description      types.String                                                 `tfsdk:"description" autoflex:",omitempty"`
	ExecutionRoleARN fwtypes.ARN                                                  `tfsdk:"execution_role_arn"`
	IDPConfig        fwtypes.ListNestedObjectValueOf[consentPortalIDPConfigModel] `tfsdk:"idp_config"`
	Name             types.String                                                 `tfsdk:"name"`
	PortalURL        types.String                                                 `tfsdk:"portal_url"`
	Sources          fwtypes.ListNestedObjectValueOf[consentPortalSourceModel]    `tfsdk:"sources"`
	Tags             tftags.Map                                                   `tfsdk:"tags"`
	TagsAll          tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts         timeouts.Value                                               `tfsdk:"timeouts"`
}

type consentPortalIDPConfigModel struct {
	Audience              types.String        `tfsdk:"audience" autoflex:",omitempty"`
	CredentialProviderARN fwtypes.ARN         `tfsdk:"credential_provider_arn"`
	Scopes                fwtypes.SetOfString `tfsdk:"scopes"`
}

type consentPortalSourceModel struct {
	Identifier types.String                                         `tfsdk:"identifier"`
	Type       fwtypes.StringEnum[awstypes.ConsentPortalSourceType] `tfsdk:"type"`
}
