// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/uxc"
	awstypes "github.com/aws/aws-sdk-go-v2/service/uxc/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
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
// @Testing(importStateIdAttribute="account_color")
func newAccountCustomizationsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountCustomizationsResource{}, nil
}

type accountCustomizationsResource struct {
	framework.ResourceWithModel[accountCustomizationsResourceModel]
	framework.WithImportByIdentity
}

func (r *accountCustomizationsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Computed:    true,
				Default:     setdefault.StaticValue(fwtypes.NewSetValueOfEmpty[types.String](ctx).SetValue),
			},
			"visible_services": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(fwtypes.NewSetValueOfEmpty[types.String](ctx).SetValue),
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

	output, err := conn.UpdateAccountCustomizations(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))

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

	output, err := conn.UpdateAccountCustomizations(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))

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
	// ImportState needs to set some value
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("account_color"), awstypes.AccountColorNone))
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
	AccountColor fwtypes.StringEnum[awstypes.AccountColor] `tfsdk:"account_color"`
	// TODO: `legacy` mode is used here for the behavior. It is not a legacy resource.
	// Needs a new mode that flattens to an empty collection instead of null.
	VisibleRegions  fwtypes.SetOfString `tfsdk:"visible_regions" autoflex:",legacy"`
	VisibleServices fwtypes.SetOfString `tfsdk:"visible_services" autoflex:",legacy"`
}
