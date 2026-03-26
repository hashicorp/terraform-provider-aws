// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/uxc"
	awstypes "github.com/aws/aws-sdk-go-v2/service/uxc/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_uxc_account_customizations", name="Account Customizations")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(generator=false)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceAccountCustomizations(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAccountCustomizations{}, nil
}

const (
	ResNameAccountCustomizations = "Account Customizations"
)

type resourceAccountCustomizations struct {
	framework.ResourceWithModel[resourceAccountCustomizationsModel]
	framework.WithImportByIdentity
}

func (r *resourceAccountCustomizations) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_color": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccountColor](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.AccountColorNone)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
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

func (r *resourceAccountCustomizations) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAccountCustomizationsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	input := &uxc.UpdateAccountCustomizationsInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
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

	output, err := conn.UpdateAccountCustomizations(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating UXC %s", ResNameAccountCustomizations),
			err.Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	normalizeAccountCustomizationsModel(ctx, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAccountCustomizations) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceAccountCustomizationsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading UXC %s (%s)", ResNameAccountCustomizations, state.ID.ValueString()),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	normalizeAccountCustomizationsModel(ctx, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccountCustomizations) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceAccountCustomizationsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	input := &uxc.UpdateAccountCustomizationsInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if input.VisibleRegions == nil {
		input.VisibleRegions = []string{}
	}
	if input.VisibleServices == nil {
		input.VisibleServices = []string{}
	}

	output, err := conn.UpdateAccountCustomizations(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("updating UXC %s (%s)", ResNameAccountCustomizations, plan.ID.ValueString()),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	normalizeAccountCustomizationsModel(ctx, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAccountCustomizations) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceAccountCustomizationsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().UXCClient(ctx)

	// There is no Delete API. Reset all values to their defaults.
	_, err := conn.UpdateAccountCustomizations(ctx, &uxc.UpdateAccountCustomizationsInput{
		AccountColor:    awstypes.AccountColorNone,
		VisibleRegions:  []string{},
		VisibleServices: []string{},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting UXC %s (%s)", ResNameAccountCustomizations, state.ID.ValueString()),
			err.Error(),
		)
	}
}

// normalizeAccountCustomizationsModel maps empty API sets to null so that unconfigured
// optional attributes don't drift against a config that omits them.
func normalizeAccountCustomizationsModel(ctx context.Context, m *resourceAccountCustomizationsModel) {
	if !m.VisibleRegions.IsNull() && len(m.VisibleRegions.Elements()) == 0 {
		m.VisibleRegions = fwtypes.NewSetValueOfNull[types.String](ctx)
	}
	if !m.VisibleServices.IsNull() && len(m.VisibleServices.Elements()) == 0 {
		m.VisibleServices = fwtypes.NewSetValueOfNull[types.String](ctx)
	}
}

func findAccountCustomizations(ctx context.Context, conn *uxc.Client) (*uxc.GetAccountCustomizationsOutput, error) {
	output, err := conn.GetAccountCustomizations(ctx, &uxc.GetAccountCustomizationsInput{})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type resourceAccountCustomizationsModel struct {
	AccountColor    fwtypes.StringEnum[awstypes.AccountColor] `tfsdk:"account_color"`
	ID              types.String                              `tfsdk:"id"`
	VisibleRegions  fwtypes.SetOfString                       `tfsdk:"visible_regions"`
	VisibleServices fwtypes.SetOfString                       `tfsdk:"visible_services"`
}
