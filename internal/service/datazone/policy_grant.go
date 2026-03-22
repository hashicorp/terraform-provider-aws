// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_datazone_policy_grant", name="Policy Grant")
// @IdentityAttribute("domain_identifier")
// @IdentityAttribute("entity_type")
// @IdentityAttribute("entity_identifier")
// @IdentityAttribute("policy_type")
// @IdentityAttribute("grant_id")
// @ImportIDHandler("policyGrantImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/datazone/types;types.PolicyGrantMember")
// @Testing(preCheck="testAccPreCheck")
func newPolicyGrantResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &policyGrantResource{}, nil
}

const (
	ResNamePolicyGrant = "Policy Grant"
)

type policyGrantResource struct {
	framework.ResourceWithModel[policyGrantResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *policyGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TargetEntityType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grant_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ManagedPolicyType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"detail":    detailBlockSchema(ctx),
			"principal": principalBlockSchema(ctx),
		},
	}
}

func (r *policyGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan policyGrantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	detailData, d := plan.Detail.ToSlice(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	principalData, d := plan.Principal.ToSlice(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(detailData) == 0 {
		resp.Diagnostics.AddError("Missing detail", "The detail block is required.")
		return
	}
	if len(principalData) == 0 {
		resp.Diagnostics.AddError("Missing principal", "The principal block is required.")
		return
	}

	detail, d := expandPolicyGrantDetail(ctx, detailData[0])
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	principal, d := expandPolicyGrantPrincipal(ctx, principalData[0])
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &datazone.AddPolicyGrantInput{
		DomainIdentifier: plan.DomainIdentifier.ValueStringPointer(),
		EntityIdentifier: plan.EntityIdentifier.ValueStringPointer(),
		EntityType:       awstypes.TargetEntityType(plan.EntityType.ValueString()),
		PolicyType:       awstypes.ManagedPolicyType(plan.PolicyType.ValueString()),
		Detail:           detail,
		Principal:        principal,
	}

	out, err := conn.AddPolicyGrant(ctx, input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DomainIdentifier.String())
		return
	}
	if out == nil || out.GrantId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.DomainIdentifier.String())
		return
	}

	plan.GrantID = flex.StringToFramework(ctx, out.GrantId)

	grant, err := findPolicyGrantByID(ctx, conn,
		plan.DomainIdentifier.ValueString(),
		plan.EntityType.ValueString(),
		plan.EntityIdentifier.ValueString(),
		plan.PolicyType.ValueString(),
		plan.GrantID.ValueString(),
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GrantID.String())
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, grant, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state policyGrantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPolicyGrantByID(ctx, conn,
		state.DomainIdentifier.ValueString(),
		state.EntityType.ValueString(),
		state.EntityIdentifier.ValueString(),
		state.PolicyType.ValueString(),
		state.GrantID.ValueString(),
	)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GrantID.String())
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyGrantResource) flatten(ctx context.Context, grant *awstypes.PolicyGrantMember, data *policyGrantResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.GrantID = flex.StringToFramework(ctx, grant.GrantId)
	data.CreatedAt = timetypes.NewRFC3339TimePointerValue(grant.CreatedAt)
	data.CreatedBy = flex.StringToFramework(ctx, grant.CreatedBy)

	detailModel, d := flattenPolicyGrantDetail(ctx, grant.Detail)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.Detail = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, detailModel)

	principalModel, d := flattenPolicyGrantPrincipal(ctx, grant.Principal)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.Principal = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, principalModel)

	return diags
}

func (r *policyGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state policyGrantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	principalData, d := state.Principal.ToSlice(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(principalData) == 0 {
		resp.Diagnostics.AddError("Missing principal", "The principal block is required for deletion.")
		return
	}

	principal, d := expandPolicyGrantPrincipal(ctx, principalData[0])
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &datazone.RemovePolicyGrantInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		EntityIdentifier: state.EntityIdentifier.ValueStringPointer(),
		EntityType:       awstypes.TargetEntityType(state.EntityType.ValueString()),
		PolicyType:       awstypes.ManagedPolicyType(state.PolicyType.ValueString()),
		Principal:        principal,
		GrantIdentifier:  state.GrantID.ValueStringPointer(),
	}

	_, err := conn.RemovePolicyGrant(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		if isResourceMissing(err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GrantID.String())
		return
	}
}

func findPolicyGrantByID(ctx context.Context, conn *datazone.Client, domainID, entityType, entityID, policyType, grantID string) (*awstypes.PolicyGrantMember, error) {
	input := &datazone.ListPolicyGrantsInput{
		DomainIdentifier: aws.String(domainID),
		EntityIdentifier: aws.String(entityID),
		EntityType:       awstypes.TargetEntityType(entityType),
		PolicyType:       awstypes.ManagedPolicyType(policyType),
	}

	pages := datazone.NewListPolicyGrantsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			if isResourceMissing(err) {
				return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
			}
			return nil, smarterr.NewError(err)
		}

		for _, grant := range page.GrantList {
			if aws.ToString(grant.GrantId) == grantID {
				return &grant, nil
			}
		}
	}

	return nil, smarterr.NewError(&retry.NotFoundError{
		LastError: fmt.Errorf("DataZone Policy Grant %s not found", grantID),
	})
}

// Model types.

type policyGrantResourceModel struct {
	framework.WithRegionModel
	CreatedAt        timetypes.RFC3339                               `tfsdk:"created_at"`
	CreatedBy        types.String                                    `tfsdk:"created_by"`
	DomainIdentifier types.String                                    `tfsdk:"domain_identifier"`
	Detail           fwtypes.ListNestedObjectValueOf[detailModel]    `tfsdk:"detail"`
	EntityIdentifier types.String                                    `tfsdk:"entity_identifier"`
	EntityType       fwtypes.StringEnum[awstypes.TargetEntityType]   `tfsdk:"entity_type"`
	GrantID          types.String                                    `tfsdk:"grant_id"`
	PolicyType       fwtypes.StringEnum[awstypes.ManagedPolicyType]  `tfsdk:"policy_type"`
	Principal        fwtypes.ListNestedObjectValueOf[principalModel] `tfsdk:"principal"`
}

type detailModel struct {
	AddToProjectMemberPool           fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"add_to_project_member_pool"`
	CreateAssetType                  fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"create_asset_type"`
	CreateDomainUnit                 fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"create_domain_unit"`
	CreateEnvironment                fwtypes.ListNestedObjectValueOf[unitModel]                                  `tfsdk:"create_environment"`
	CreateEnvironmentFromBlueprint   fwtypes.ListNestedObjectValueOf[unitModel]                                  `tfsdk:"create_environment_from_blueprint"`
	CreateEnvironmentProfile         fwtypes.ListNestedObjectValueOf[domainUnitIDDetailModel]                    `tfsdk:"create_environment_profile"`
	CreateFormType                   fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"create_form_type"`
	CreateGlossary                   fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"create_glossary"`
	CreateProject                    fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"create_project"`
	CreateProjectFromProjectProfile  fwtypes.ListNestedObjectValueOf[createProjectFromProjectProfileDetailModel] `tfsdk:"create_project_from_project_profile"`
	DelegateCreateEnvironmentProfile fwtypes.ListNestedObjectValueOf[unitModel]                                  `tfsdk:"delegate_create_environment_profile"`
	OverrideDomainUnitOwners         fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"override_domain_unit_owners"`
	OverrideProjectOwners            fwtypes.ListNestedObjectValueOf[includeChildDomainUnitsDetailModel]         `tfsdk:"override_project_owners"`
	UseAssetType                     fwtypes.ListNestedObjectValueOf[domainUnitIDDetailModel]                    `tfsdk:"use_asset_type"`
}

type unitModel struct{}

type includeChildDomainUnitsDetailModel struct {
	IncludeChildDomainUnits types.Bool `tfsdk:"include_child_domain_units"`
}

type domainUnitIDDetailModel struct {
	DomainUnitID types.String `tfsdk:"domain_unit_id"`
}

type createProjectFromProjectProfileDetailModel struct {
	IncludeChildDomainUnits types.Bool `tfsdk:"include_child_domain_units"`
	ProjectProfiles         types.List `tfsdk:"project_profiles"`
}

type principalModel struct {
	DomainUnit fwtypes.ListNestedObjectValueOf[domainUnitPrincipalModel] `tfsdk:"domain_unit"`
	Group      fwtypes.ListNestedObjectValueOf[groupPrincipalModel]      `tfsdk:"group"`
	Project    fwtypes.ListNestedObjectValueOf[projectPrincipalModel]    `tfsdk:"project"`
	User       fwtypes.ListNestedObjectValueOf[userPrincipalModel]       `tfsdk:"user"`
}

type domainUnitPrincipalModel struct {
	DomainUnitDesignation     fwtypes.StringEnum[awstypes.DomainUnitDesignation] `tfsdk:"domain_unit_designation"`
	DomainUnitIdentifier      types.String                                       `tfsdk:"domain_unit_identifier"`
	AllDomainUnitsGrantFilter fwtypes.ListNestedObjectValueOf[unitModel]         `tfsdk:"all_domain_units_grant_filter"`
}

type projectPrincipalModel struct {
	ProjectDesignation fwtypes.StringEnum[awstypes.ProjectDesignation]        `tfsdk:"project_designation"`
	ProjectIdentifier  types.String                                           `tfsdk:"project_identifier"`
	DomainUnitFilter   fwtypes.ListNestedObjectValueOf[domainUnitFilterModel] `tfsdk:"domain_unit_filter"`
}

type groupPrincipalModel struct {
	GroupIdentifier types.String `tfsdk:"group_identifier"`
}

type userPrincipalModel struct {
	UserIdentifier      types.String                               `tfsdk:"user_identifier"`
	AllUsersGrantFilter fwtypes.ListNestedObjectValueOf[unitModel] `tfsdk:"all_users_grant_filter"`
}

type domainUnitFilterModel struct {
	DomainUnit              types.String `tfsdk:"domain_unit"`
	IncludeChildDomainUnits types.Bool   `tfsdk:"include_child_domain_units"`
}

// Schema helpers.

func detailBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[detailModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"add_to_project_member_pool":          includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_asset_type":                   includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_domain_unit":                  includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_environment":                  unitBlockSchema(ctx),
				"create_environment_from_blueprint":   unitBlockSchema(ctx),
				"create_environment_profile":          domainUnitIDDetailBlockSchema(ctx),
				"create_form_type":                    includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_glossary":                     includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_project":                      includeChildDomainUnitsDetailBlockSchema(ctx),
				"create_project_from_project_profile": createProjectFromProjectProfileDetailBlockSchema(ctx),
				"delegate_create_environment_profile": unitBlockSchema(ctx),
				"override_domain_unit_owners":         includeChildDomainUnitsDetailBlockSchema(ctx),
				"override_project_owners":             includeChildDomainUnitsDetailBlockSchema(ctx),
				"use_asset_type":                      domainUnitIDDetailBlockSchema(ctx),
			},
		},
	}
}

func unitBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[unitModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{},
	}
}

func includeChildDomainUnitsDetailBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[includeChildDomainUnitsDetailModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"include_child_domain_units": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func domainUnitIDDetailBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[domainUnitIDDetailModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"domain_unit_id": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func createProjectFromProjectProfileDetailBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[createProjectFromProjectProfileDetailModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"include_child_domain_units": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
				"project_profiles": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func principalBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[principalModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"domain_unit": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[domainUnitPrincipalModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"domain_unit_designation": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.DomainUnitDesignation](),
								Required:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"domain_unit_identifier": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"all_domain_units_grant_filter": unitBlockSchema(ctx),
						},
					},
				},
				"group": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[groupPrincipalModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"group_identifier": schema.StringAttribute{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				"project": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[projectPrincipalModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"project_designation": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.ProjectDesignation](),
								Required:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"project_identifier": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"domain_unit_filter": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[domainUnitFilterModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"domain_unit": schema.StringAttribute{
											Required: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"include_child_domain_units": schema.BoolAttribute{
											Optional: true,
											PlanModifiers: []planmodifier.Bool{
												boolplanmodifier.RequiresReplace(),
											},
										},
									},
								},
							},
						},
					},
				},
				"user": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[userPrincipalModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"user_identifier": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"all_users_grant_filter": unitBlockSchema(ctx),
						},
					},
				},
			},
		},
	}
}

// Expand helpers convert Terraform model types to AWS SDK types.

func expandPolicyGrantDetail(ctx context.Context, detail *detailModel) (awstypes.PolicyGrantDetail, diag.Diagnostics) {
	var diags diag.Diagnostics

	if elements, d := detail.AddToProjectMemberPool.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberAddToProjectMemberPool{
			Value: awstypes.AddToProjectMemberPoolPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateAssetType.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateAssetType{
			Value: awstypes.CreateAssetTypePolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateDomainUnit.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateDomainUnit{
			Value: awstypes.CreateDomainUnitPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateEnvironment.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateEnvironment{
			Value: awstypes.Unit{},
		}, diags
	}

	if elements, d := detail.CreateEnvironmentFromBlueprint.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateEnvironmentFromBlueprint{
			Value: awstypes.Unit{},
		}, diags
	}

	if elements, d := detail.CreateEnvironmentProfile.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateEnvironmentProfile{
			Value: awstypes.CreateEnvironmentProfilePolicyGrantDetail{
				DomainUnitId: elements[0].DomainUnitID.ValueStringPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateFormType.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateFormType{
			Value: awstypes.CreateFormTypePolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateGlossary.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateGlossary{
			Value: awstypes.CreateGlossaryPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateProject.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberCreateProject{
			Value: awstypes.CreateProjectPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.CreateProjectFromProjectProfile.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		val := awstypes.CreateProjectFromProjectProfilePolicyGrantDetail{
			IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
		}
		if !elements[0].ProjectProfiles.IsNull() && !elements[0].ProjectProfiles.IsUnknown() {
			var profiles []string
			d := elements[0].ProjectProfiles.ElementsAs(ctx, &profiles, false)
			diags.Append(d...)
			val.ProjectProfiles = profiles
		}
		return &awstypes.PolicyGrantDetailMemberCreateProjectFromProjectProfile{Value: val}, diags
	}

	if elements, d := detail.DelegateCreateEnvironmentProfile.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberDelegateCreateEnvironmentProfile{
			Value: awstypes.Unit{},
		}, diags
	}

	if elements, d := detail.OverrideDomainUnitOwners.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberOverrideDomainUnitOwners{
			Value: awstypes.OverrideDomainUnitOwnersPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.OverrideProjectOwners.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberOverrideProjectOwners{
			Value: awstypes.OverrideProjectOwnersPolicyGrantDetail{
				IncludeChildDomainUnits: elements[0].IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags
	}

	if elements, d := detail.UseAssetType.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantDetailMemberUseAssetType{
			Value: awstypes.UseAssetTypePolicyGrantDetail{
				DomainUnitId: elements[0].DomainUnitID.ValueStringPointer(),
			},
		}, diags
	}

	diags.AddError("Invalid detail", "Exactly one detail variant block must be specified.")
	return nil, diags
}

func expandPolicyGrantPrincipal(ctx context.Context, principal *principalModel) (awstypes.PolicyGrantPrincipal, diag.Diagnostics) {
	var diags diag.Diagnostics

	if elements, d := principal.DomainUnit.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		du := elements[0]
		result := awstypes.DomainUnitPolicyGrantPrincipal{
			DomainUnitDesignation: awstypes.DomainUnitDesignation(du.DomainUnitDesignation.ValueString()),
			DomainUnitIdentifier:  du.DomainUnitIdentifier.ValueStringPointer(),
		}
		if filterElements, fd := du.AllDomainUnitsGrantFilter.ToSlice(ctx); fd.HasError() {
			diags.Append(fd...)
			return nil, diags
		} else if len(filterElements) > 0 {
			result.DomainUnitGrantFilter = &awstypes.DomainUnitGrantFilterMemberAllDomainUnitsGrantFilter{
				Value: awstypes.AllDomainUnitsGrantFilter{},
			}
		}
		return &awstypes.PolicyGrantPrincipalMemberDomainUnit{Value: result}, diags
	}

	if elements, d := principal.Group.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		return &awstypes.PolicyGrantPrincipalMemberGroup{
			Value: &awstypes.GroupPolicyGrantPrincipalMemberGroupIdentifier{
				Value: elements[0].GroupIdentifier.ValueString(),
			},
		}, diags
	}

	if elements, d := principal.Project.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		proj := elements[0]
		result := awstypes.ProjectPolicyGrantPrincipal{
			ProjectDesignation: awstypes.ProjectDesignation(proj.ProjectDesignation.ValueString()),
			ProjectIdentifier:  proj.ProjectIdentifier.ValueStringPointer(),
		}
		if filterElements, fd := proj.DomainUnitFilter.ToSlice(ctx); fd.HasError() {
			diags.Append(fd...)
			return nil, diags
		} else if len(filterElements) > 0 {
			result.ProjectGrantFilter = &awstypes.ProjectGrantFilterMemberDomainUnitFilter{
				Value: awstypes.DomainUnitFilterForProject{
					DomainUnit:              filterElements[0].DomainUnit.ValueStringPointer(),
					IncludeChildDomainUnits: filterElements[0].IncludeChildDomainUnits.ValueBoolPointer(),
				},
			}
		}
		return &awstypes.PolicyGrantPrincipalMemberProject{Value: result}, diags
	}

	if elements, d := principal.User.ToSlice(ctx); d.HasError() {
		diags.Append(d...)
		return nil, diags
	} else if len(elements) > 0 {
		u := elements[0]
		if !u.UserIdentifier.IsNull() && !u.UserIdentifier.IsUnknown() {
			return &awstypes.PolicyGrantPrincipalMemberUser{
				Value: &awstypes.UserPolicyGrantPrincipalMemberUserIdentifier{
					Value: u.UserIdentifier.ValueString(),
				},
			}, diags
		}
		if filterElements, fd := u.AllUsersGrantFilter.ToSlice(ctx); fd.HasError() {
			diags.Append(fd...)
			return nil, diags
		} else if len(filterElements) > 0 {
			return &awstypes.PolicyGrantPrincipalMemberUser{
				Value: &awstypes.UserPolicyGrantPrincipalMemberAllUsersGrantFilter{
					Value: awstypes.AllUsersGrantFilter{},
				},
			}, diags
		}
	}

	diags.AddError("Invalid principal", "Exactly one principal variant block must be specified.")
	return nil, diags
}

// Flatten helpers convert AWS SDK types to Terraform model types.

func flattenPolicyGrantDetail(ctx context.Context, detail awstypes.PolicyGrantDetail) (*detailModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := &detailModel{
		AddToProjectMemberPool:           fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateAssetType:                  fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateDomainUnit:                 fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateEnvironment:                fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		CreateEnvironmentFromBlueprint:   fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		CreateEnvironmentProfile:         fwtypes.NewListNestedObjectValueOfNull[domainUnitIDDetailModel](ctx),
		CreateFormType:                   fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateGlossary:                   fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateProject:                    fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		CreateProjectFromProjectProfile:  fwtypes.NewListNestedObjectValueOfNull[createProjectFromProjectProfileDetailModel](ctx),
		DelegateCreateEnvironmentProfile: fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		OverrideDomainUnitOwners:         fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		OverrideProjectOwners:            fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx),
		UseAssetType:                     fwtypes.NewListNestedObjectValueOfNull[domainUnitIDDetailModel](ctx),
	}

	switch v := detail.(type) {
	case *awstypes.PolicyGrantDetailMemberAddToProjectMemberPool:
		m.AddToProjectMemberPool = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateAssetType:
		m.CreateAssetType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateDomainUnit:
		m.CreateDomainUnit = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateEnvironment:
		m.CreateEnvironment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
	case *awstypes.PolicyGrantDetailMemberCreateEnvironmentFromBlueprint:
		m.CreateEnvironmentFromBlueprint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
	case *awstypes.PolicyGrantDetailMemberCreateEnvironmentProfile:
		m.CreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitIDDetailModel{
			DomainUnitID: flex.StringToFramework(ctx, v.Value.DomainUnitId),
		})
	case *awstypes.PolicyGrantDetailMemberCreateFormType:
		m.CreateFormType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateGlossary:
		m.CreateGlossary = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateProject:
		m.CreateProject = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberCreateProjectFromProjectProfile:
		ppModel := &createProjectFromProjectProfileDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
			ProjectProfiles:         types.ListNull(types.StringType),
		}
		if len(v.Value.ProjectProfiles) > 0 {
			elements := make([]attr.Value, len(v.Value.ProjectProfiles))
			for i, p := range v.Value.ProjectProfiles {
				elements[i] = types.StringValue(p)
			}
			listVal, d := types.ListValue(types.StringType, elements)
			diags.Append(d...)
			ppModel.ProjectProfiles = listVal
		}
		m.CreateProjectFromProjectProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, ppModel)
	case *awstypes.PolicyGrantDetailMemberDelegateCreateEnvironmentProfile:
		m.DelegateCreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
	case *awstypes.PolicyGrantDetailMemberOverrideDomainUnitOwners:
		m.OverrideDomainUnitOwners = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberOverrideProjectOwners:
		m.OverrideProjectOwners = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(v.Value.IncludeChildDomainUnits),
		})
	case *awstypes.PolicyGrantDetailMemberUseAssetType:
		m.UseAssetType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitIDDetailModel{
			DomainUnitID: flex.StringToFramework(ctx, v.Value.DomainUnitId),
		})
	default:
		diags.AddError("Unknown detail type", fmt.Sprintf("Unexpected policy grant detail type: %T", detail))
	}

	return m, diags
}

func flattenPolicyGrantPrincipal(ctx context.Context, principal awstypes.PolicyGrantPrincipal) (*principalModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := &principalModel{
		DomainUnit: fwtypes.NewListNestedObjectValueOfNull[domainUnitPrincipalModel](ctx),
		Group:      fwtypes.NewListNestedObjectValueOfNull[groupPrincipalModel](ctx),
		Project:    fwtypes.NewListNestedObjectValueOfNull[projectPrincipalModel](ctx),
		User:       fwtypes.NewListNestedObjectValueOfNull[userPrincipalModel](ctx),
	}

	switch v := principal.(type) {
	case *awstypes.PolicyGrantPrincipalMemberDomainUnit:
		duModel := &domainUnitPrincipalModel{
			DomainUnitDesignation:     fwtypes.StringEnumValue(v.Value.DomainUnitDesignation),
			DomainUnitIdentifier:      flex.StringToFramework(ctx, v.Value.DomainUnitIdentifier),
			AllDomainUnitsGrantFilter: fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		}
		if v.Value.DomainUnitGrantFilter != nil {
			if _, ok := v.Value.DomainUnitGrantFilter.(*awstypes.DomainUnitGrantFilterMemberAllDomainUnitsGrantFilter); ok {
				duModel.AllDomainUnitsGrantFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
			}
		}
		m.DomainUnit = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, duModel)

	case *awstypes.PolicyGrantPrincipalMemberGroup:
		groupID := ""
		if gm, ok := v.Value.(*awstypes.GroupPolicyGrantPrincipalMemberGroupIdentifier); ok {
			groupID = gm.Value
		}
		m.Group = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &groupPrincipalModel{
			GroupIdentifier: types.StringValue(groupID),
		})

	case *awstypes.PolicyGrantPrincipalMemberProject:
		projModel := &projectPrincipalModel{
			ProjectDesignation: fwtypes.StringEnumValue(v.Value.ProjectDesignation),
			ProjectIdentifier:  flex.StringToFramework(ctx, v.Value.ProjectIdentifier),
			DomainUnitFilter:   fwtypes.NewListNestedObjectValueOfNull[domainUnitFilterModel](ctx),
		}
		if v.Value.ProjectGrantFilter != nil {
			if pf, ok := v.Value.ProjectGrantFilter.(*awstypes.ProjectGrantFilterMemberDomainUnitFilter); ok {
				projModel.DomainUnitFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitFilterModel{
					DomainUnit:              flex.StringToFramework(ctx, pf.Value.DomainUnit),
					IncludeChildDomainUnits: types.BoolPointerValue(pf.Value.IncludeChildDomainUnits),
				})
			}
		}
		m.Project = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, projModel)

	case *awstypes.PolicyGrantPrincipalMemberUser:
		userModel := &userPrincipalModel{
			UserIdentifier:      types.StringNull(),
			AllUsersGrantFilter: fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		}
		switch u := v.Value.(type) {
		case *awstypes.UserPolicyGrantPrincipalMemberUserIdentifier:
			userModel.UserIdentifier = types.StringValue(u.Value)
		case *awstypes.UserPolicyGrantPrincipalMemberAllUsersGrantFilter:
			userModel.AllUsersGrantFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
		}
		m.User = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, userModel)

	default:
		diags.AddError("Unknown principal type", fmt.Sprintf("Unexpected policy grant principal type: %T", principal))
	}

	return m, diags
}

// Import ID Handler.

var (
	_ inttypes.ImportIDParser = policyGrantImportID{}
)

type policyGrantImportID struct{}

func (policyGrantImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, intflex.ResourceIdSeparator)
	if len(parts) != 5 {
		return "", nil, fmt.Errorf("id %q should be in the format <domain-identifier>%s<entity-type>%s<entity-identifier>%s<policy-type>%s<grant-id>",
			id,
			intflex.ResourceIdSeparator,
			intflex.ResourceIdSeparator,
			intflex.ResourceIdSeparator,
			intflex.ResourceIdSeparator,
		)
	}

	result := map[string]any{
		"domain_identifier": parts[0],
		"entity_type":       parts[1],
		"entity_identifier": parts[2],
		"policy_type":       parts[3],
		"grant_id":          parts[4],
	}

	return id, result, nil
}
