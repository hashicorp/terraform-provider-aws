// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package taxsettings

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/taxsettings"
	awstypes "github.com/aws/aws-sdk-go-v2/service/taxsettings/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_taxsettings_tax_registration", name="Tax Registration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/taxsettings/types;awstypes.TaxRegistration")
// @Testing(preCheck="testAccPreCheck")
func newTaxRegistrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &taxRegistrationResource{}, nil
}

const ResNameTaxRegistration = "Tax Registration"

type taxRegistrationResource struct {
	framework.ResourceWithModel[taxRegistrationResourceModel]
	framework.WithImportByIdentity
}

func (r *taxRegistrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certified_email_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"legal_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"registration_id": schema.StringAttribute{
				Required: true,
			},
			"registration_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TaxRegistrationType](),
				Required:   true,
			},
			"sector": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Sector](),
				Optional:   true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"legal_address": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[addressModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address_line1": schema.StringAttribute{
							Required: true,
						},
						"address_line2": schema.StringAttribute{
							Optional: true,
						},
						"address_line3": schema.StringAttribute{
							Optional: true,
						},
						"city": schema.StringAttribute{
							Required: true,
						},
						"country_code": schema.StringAttribute{
							Required: true,
						},
						"district_or_county": schema.StringAttribute{
							Optional: true,
						},
						"postal_code": schema.StringAttribute{
							Required: true,
						},
						"state_or_region": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *taxRegistrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TaxSettingsClient(ctx)

	var plan taxRegistrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &taxsettings.PutTaxRegistrationInput{
		TaxRegistrationEntry: expandTaxRegistrationEntry(ctx, plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.AccountID.IsNull() && !plan.AccountID.IsUnknown() {
		input.AccountId = plan.AccountID.ValueStringPointer()
	}

	out, err := conn.PutTaxRegistration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("creating Tax Registration", err.Error())
		return
	}

	accountID := plan.AccountID.ValueString()
	if accountID == "" {
		accountID = r.Meta().AccountID(ctx)
	}
	plan.AccountID = types.StringValue(accountID)
	plan.ID = types.StringValue(accountID)
	plan.Status = types.StringValue(string(out.Status))

	// Read back to populate computed fields (legal_name is returned in GetTaxRegistration).
	reg, err := findTaxRegistrationByAccountID(ctx, conn, accountID)
	if err != nil {
		resp.Diagnostics.AddError("reading Tax Registration after create", err.Error())
		return
	}
	flattenTaxRegistration(ctx, reg, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *taxRegistrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TaxSettingsClient(ctx)

	var state taxRegistrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTaxRegistrationByAccountID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading Tax Registration", err.Error())
		return
	}

	flattenTaxRegistration(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *taxRegistrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().TaxSettingsClient(ctx)

	var plan taxRegistrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &taxsettings.PutTaxRegistrationInput{
		AccountId:            plan.AccountID.ValueStringPointer(),
		TaxRegistrationEntry: expandTaxRegistrationEntry(ctx, plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutTaxRegistration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("updating Tax Registration", err.Error())
		return
	}

	plan.Status = types.StringValue(string(out.Status))

	reg, err := findTaxRegistrationByAccountID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("reading Tax Registration after update", err.Error())
		return
	}
	flattenTaxRegistration(ctx, reg, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *taxRegistrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TaxSettingsClient(ctx)

	var state taxRegistrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &taxsettings.DeleteTaxRegistrationInput{
		AccountId: state.AccountID.ValueStringPointer(),
	}

	_, err := conn.DeleteTaxRegistration(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError("deleting Tax Registration", err.Error())
		return
	}
}

func findTaxRegistrationByAccountID(ctx context.Context, conn *taxsettings.Client, accountID string) (*awstypes.TaxRegistration, error) {
	input := &taxsettings.GetTaxRegistrationInput{
		AccountId: aws.String(accountID),
	}

	out, err := conn.GetTaxRegistration(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}
		return nil, err
	}

	if out == nil || out.TaxRegistration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.TaxRegistration, nil
}

func expandTaxRegistrationEntry(ctx context.Context, plan taxRegistrationResourceModel, diags *diag.Diagnostics) *awstypes.TaxRegistrationEntry {
	entry := &awstypes.TaxRegistrationEntry{
		RegistrationId:   plan.RegistrationID.ValueStringPointer(),
		RegistrationType: awstypes.TaxRegistrationType(plan.RegistrationType.ValueString()),
	}

	if !plan.LegalName.IsNull() && !plan.LegalName.IsUnknown() {
		entry.LegalName = plan.LegalName.ValueStringPointer()
	}
	if !plan.CertifiedEmailID.IsNull() && !plan.CertifiedEmailID.IsUnknown() {
		entry.CertifiedEmailId = plan.CertifiedEmailID.ValueStringPointer()
	}
	if !plan.Sector.IsNull() && !plan.Sector.IsUnknown() {
		s := awstypes.Sector(plan.Sector.ValueString())
		entry.Sector = s
	}

	if !plan.LegalAddress.IsNull() && !plan.LegalAddress.IsUnknown() {
		addrs, d := plan.LegalAddress.ToSlice(ctx)
		diags.Append(d...)
		if len(addrs) > 0 {
			a := addrs[0]
			entry.LegalAddress = &awstypes.Address{
				AddressLine1:     a.AddressLine1.ValueStringPointer(),
				AddressLine2:     a.AddressLine2.ValueStringPointer(),
				AddressLine3:     a.AddressLine3.ValueStringPointer(),
				City:             a.City.ValueStringPointer(),
				CountryCode:      a.CountryCode.ValueStringPointer(),
				DistrictOrCounty: a.DistrictOrCounty.ValueStringPointer(),
				PostalCode:       a.PostalCode.ValueStringPointer(),
				StateOrRegion:    a.StateOrRegion.ValueStringPointer(),
			}
		}
	}

	return entry
}

func flattenTaxRegistration(ctx context.Context, reg *awstypes.TaxRegistration, state *taxRegistrationResourceModel, diags *diag.Diagnostics) {
	state.LegalName = types.StringPointerValue(reg.LegalName)
	state.RegistrationID = types.StringPointerValue(reg.RegistrationId)
	state.RegistrationType = fwtypes.StringEnumValue(reg.RegistrationType)
	state.Status = types.StringValue(string(reg.Status))

	if reg.Sector != "" {
		state.Sector = fwtypes.StringEnumValue(reg.Sector)
	} else {
		state.Sector = fwtypes.StringEnumNull[awstypes.Sector]()
	}
	state.CertifiedEmailID = types.StringPointerValue(reg.CertifiedEmailId)

	if reg.LegalAddress != nil {
		addr := addressModel{
			AddressLine1:     types.StringPointerValue(reg.LegalAddress.AddressLine1),
			AddressLine2:     types.StringPointerValue(reg.LegalAddress.AddressLine2),
			AddressLine3:     types.StringPointerValue(reg.LegalAddress.AddressLine3),
			City:             types.StringPointerValue(reg.LegalAddress.City),
			CountryCode:      types.StringPointerValue(reg.LegalAddress.CountryCode),
			DistrictOrCounty: types.StringPointerValue(reg.LegalAddress.DistrictOrCounty),
			PostalCode:       types.StringPointerValue(reg.LegalAddress.PostalCode),
			StateOrRegion:    types.StringPointerValue(reg.LegalAddress.StateOrRegion),
		}
		legalAddr, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &addr)
		diags.Append(d...)
		state.LegalAddress = legalAddr
	}
}

type taxRegistrationResourceModel struct {
	AccountID        types.String                                     `tfsdk:"account_id"`
	CertifiedEmailID types.String                                     `tfsdk:"certified_email_id"`
	ID               types.String                                     `tfsdk:"id"`
	LegalAddress     fwtypes.ListNestedObjectValueOf[addressModel]    `tfsdk:"legal_address"`
	LegalName        types.String                                     `tfsdk:"legal_name"`
	RegistrationID   types.String                                     `tfsdk:"registration_id"`
	RegistrationType fwtypes.StringEnum[awstypes.TaxRegistrationType] `tfsdk:"registration_type"`
	Sector           fwtypes.StringEnum[awstypes.Sector]              `tfsdk:"sector"`
	Status           types.String                                     `tfsdk:"status"`
}

type addressModel struct {
	AddressLine1     types.String `tfsdk:"address_line1"`
	AddressLine2     types.String `tfsdk:"address_line2"`
	AddressLine3     types.String `tfsdk:"address_line3"`
	City             types.String `tfsdk:"city"`
	CountryCode      types.String `tfsdk:"country_code"`
	DistrictOrCounty types.String `tfsdk:"district_or_county"`
	PostalCode       types.String `tfsdk:"postal_code"`
	StateOrRegion    types.String `tfsdk:"state_or_region"`
}
