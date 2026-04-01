// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/uxc"
	awstypes "github.com/aws/aws-sdk-go-v2/service/uxc/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_uxc_account_customizations", name="Account Customizations")
// @SingletonIdentity
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
func newAccountCustomizationsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountCustomizationsResource{}, nil
}

type accountCustomizationsResource struct {
	framework.ResourceWithModel[accountCustomizationsResourceModel]
	framework.WithImportByIdentity
}

func (r *accountCustomizationsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	accountColorType := fwtypes.StringEnumType[awstypes.AccountColor]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_color": schema.StringAttribute{
				CustomType: accountColorType,
				Optional:   true,
				Computed:   true,
				Default:    accountColorType.AttributeDefault(awstypes.AccountColorNone),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visible_regions": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"visible_services": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *accountCustomizationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountCustomizationsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	input := uxc.UpdateAccountCustomizationsInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	// nil means "no change" in the API; use empty slice to explicitly clear any pre-existing restrictions.
	if input.VisibleRegions == nil {
		input.VisibleRegions = []string{}
	}
	if input.VisibleServices == nil {
		input.VisibleServices = []string{}
	}

	output, err := conn.UpdateAccountCustomizations(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	normalizeAccountCustomizationsModel(ctx, &plan)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accountCustomizationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountCustomizationsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	output, err := findAccountCustomizations(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &state))
	normalizeAccountCustomizationsModel(ctx, &state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *accountCustomizationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountCustomizationsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	input := uxc.UpdateAccountCustomizationsInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	if input.VisibleRegions == nil {
		input.VisibleRegions = []string{}
	}
	if input.VisibleServices == nil {
		input.VisibleServices = []string{}
	}

	output, err := conn.UpdateAccountCustomizations(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	normalizeAccountCustomizationsModel(ctx, &plan)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accountCustomizationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountCustomizationsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	// There is no Delete API. Reset all values to their defaults.
	input := uxc.UpdateAccountCustomizationsInput{
		AccountColor:    awstypes.AccountColorNone,
		VisibleRegions:  []string{},
		VisibleServices: []string{},
	}
	_, err := conn.UpdateAccountCustomizations(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
	}
}

func (r *accountCustomizationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	conn := r.Meta().UXCClient(ctx)

	output, err := findAccountCustomizations(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	var state accountCustomizationsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeAccountCustomizationsModel(ctx, &state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

// normalizeAccountCustomizationsModel maps empty API sets to null so that unconfigured
// optional attributes don't drift against a config that omits them.
func normalizeAccountCustomizationsModel(ctx context.Context, m *accountCustomizationsResourceModel) {
	if !m.VisibleRegions.IsNull() && len(m.VisibleRegions.Elements()) == 0 {
		m.VisibleRegions = fwtypes.NewSetValueOfNull[types.String](ctx)
	}
	if !m.VisibleServices.IsNull() && len(m.VisibleServices.Elements()) == 0 {
		m.VisibleServices = fwtypes.NewSetValueOfNull[types.String](ctx)
	}
}

func findAccountCustomizations(ctx context.Context, conn *uxc.Client) (*uxc.GetAccountCustomizationsOutput, error) {
	input := uxc.GetAccountCustomizationsInput{}
	output, err := conn.GetAccountCustomizations(ctx, &input)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type accountCustomizationsResourceModel struct {
	AccountColor    fwtypes.StringEnum[awstypes.AccountColor] `tfsdk:"account_color"`
	VisibleRegions  fwtypes.SetOfString                       `tfsdk:"visible_regions"`
	VisibleServices fwtypes.SetOfString                       `tfsdk:"visible_services"`
}
