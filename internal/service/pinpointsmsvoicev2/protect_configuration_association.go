// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_protect_configuration_association", name="Protect Configuration Association")
// @IdentityAttribute("configuration_set_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes;awstypes.ConfigurationSetInformation")
// @Testing(importStateIdAttribute="configuration_set_name")
// @Testing(preCheck="testAccPreCheckProtectConfiguration")
func newProtectConfigurationAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &protectConfigurationAssociationResource{}, nil
}

type protectConfigurationAssociationResource struct {
	framework.ResourceWithModel[protectConfigurationAssociationResourceModel]
	framework.WithImportByIdentity
}

func (r *protectConfigurationAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"configuration_set_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// A configuration set has at most one protect configuration, and associating another one
			// replaces it, so the protect configuration can be changed in place.
			"protect_configuration_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *protectConfigurationAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan protectConfigurationAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.ConfigurationSetName)
	var input pinpointsmsvoicev2.AssociateProtectConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := conn.AssociateProtectConfiguration(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *protectConfigurationAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state protectConfigurationAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.ConfigurationSetName)
	out, err := findProtectConfigurationAssociationByConfigurationSetName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state), smerr.ID, name)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *protectConfigurationAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state protectConfigurationAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only protect_configuration_id can change. Associating the protect configuration that is already
	// associated fails with a ConflictException, so the API is called only when it changes.
	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, plan.ConfigurationSetName)
		var input pinpointsmsvoicev2.AssociateProtectConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := conn.AssociateProtectConfiguration(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *protectConfigurationAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state protectConfigurationAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.ConfigurationSetName)
	var input pinpointsmsvoicev2.DisassociateProtectConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DisassociateProtectConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || isProtectConfigurationNotAssociatedError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

// isProtectConfigurationNotAssociatedError reports whether err is the conflict AWS returns when
// disassociating a protect configuration that isn't associated with the configuration set.
func isProtectConfigurationNotAssociatedError(err error) bool {
	if v, ok := errors.AsType[*awstypes.ConflictException](err); ok {
		return v.Reason == awstypes.ConflictExceptionReasonProtectConfigurationNotAssociatedWithConfigurationSet
	}

	return false
}

// findProtectConfigurationAssociationByConfigurationSetName returns the configuration set if it has
// a protect configuration associated with it.
func findProtectConfigurationAssociationByConfigurationSetName(ctx context.Context, conn *pinpointsmsvoicev2.Client, name string) (*awstypes.ConfigurationSetInformation, error) {
	out, err := findConfigurationSetByID(ctx, conn, name)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if aws.ToString(out.ProtectConfigurationId) == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type protectConfigurationAssociationResourceModel struct {
	framework.WithRegionModel
	ConfigurationSetName   types.String `tfsdk:"configuration_set_name"`
	ProtectConfigurationID types.String `tfsdk:"protect_configuration_id"`
}
