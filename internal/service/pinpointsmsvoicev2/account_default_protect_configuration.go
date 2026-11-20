// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_account_default_protect_configuration", name="Account Default Protect Configuration")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheckProtectConfiguration")
func newAccountDefaultProtectConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountDefaultProtectConfigurationResource{}, nil
}

type accountDefaultProtectConfigurationResource struct {
	framework.ResourceWithModel[accountDefaultProtectConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *accountDefaultProtectConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			// Setting another account default replaces the current one, so it can be changed in place.
			"protect_configuration_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *accountDefaultProtectConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountDefaultProtectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var input pinpointsmsvoicev2.SetAccountDefaultProtectConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := conn.SetAccountDefaultProtectConfiguration(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ProtectConfigurationID.ValueString())
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *accountDefaultProtectConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountDefaultProtectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findAccountDefaultProtectConfiguration(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *accountDefaultProtectConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state accountDefaultProtectConfigurationResourceModel
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

	if diff.HasChanges() {
		var input pinpointsmsvoicev2.SetAccountDefaultProtectConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := conn.SetAccountDefaultProtectConfiguration(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ProtectConfigurationID.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accountDefaultProtectConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	// DeleteAccountDefaultProtectConfiguration succeeds when no account default is set.
	var input pinpointsmsvoicev2.DeleteAccountDefaultProtectConfigurationInput
	if _, err := conn.DeleteAccountDefaultProtectConfiguration(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
}

// findAccountDefaultProtectConfiguration returns the protect configuration that is the account default.
func findAccountDefaultProtectConfiguration(ctx context.Context, conn *pinpointsmsvoicev2.Client) (*awstypes.ProtectConfigurationInformation, error) {
	input := pinpointsmsvoicev2.DescribeProtectConfigurationsInput{
		Filters: []awstypes.ProtectConfigurationFilter{
			{
				Name:   awstypes.ProtectConfigurationFilterNameAccountDefault,
				Values: []string{"true"},
			},
		},
	}

	return smarterr.Assert(findProtectConfiguration(ctx, conn, &input))
}

type accountDefaultProtectConfigurationResourceModel struct {
	framework.WithRegionModel
	ID                     types.String `tfsdk:"id" autoflex:"-"`
	ProtectConfigurationID types.String `tfsdk:"protect_configuration_id"`
}
