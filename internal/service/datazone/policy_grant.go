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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
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
// @Testing(importStateIdAttributes="domain_identifier;entity_type;entity_identifier;policy_type;grant_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(importStateIdAttribute="grant_id")
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
			names.AttrCreatedAt: schema.StringAttribute{
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
			"detail":            detailBlockSchema(ctx),
			names.AttrPrincipal: principalBlockSchema(ctx),
		},
	}
}

// Schema helpers.

func detailBlockSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[detailModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
			listvalidator.IsRequired(),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				tfobjectvalidator.ExactlyOneOfChildren(
					path.MatchRelative().AtName("add_to_project_member_pool"),
					path.MatchRelative().AtName("create_asset_type"),
					path.MatchRelative().AtName("create_domain_unit"),
					path.MatchRelative().AtName("create_environment"),
					path.MatchRelative().AtName("create_environment_from_blueprint"),
					path.MatchRelative().AtName("create_environment_profile"),
					path.MatchRelative().AtName("create_form_type"),
					path.MatchRelative().AtName("create_glossary"),
					path.MatchRelative().AtName("create_project"),
					path.MatchRelative().AtName("create_project_from_project_profile"),
					path.MatchRelative().AtName("delegate_create_environment_profile"),
					path.MatchRelative().AtName("override_domain_unit_owners"),
					path.MatchRelative().AtName("override_project_owners"),
					path.MatchRelative().AtName("use_asset_type"),
				),
			},
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
					CustomType:  fwtypes.ListOfStringType,
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
			listvalidator.IsRequired(),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				tfobjectvalidator.ExactlyOneOfChildren(
					path.MatchRelative().AtName("domain_unit"),
					path.MatchRelative().AtName("group"),
					path.MatchRelative().AtName("project"),
					path.MatchRelative().AtName("user"),
				),
			},
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

func (r *policyGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan policyGrantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input datazone.AddPolicyGrantInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.AddPolicyGrant(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DomainIdentifier.String())
		return
	}
	if out == nil || out.GrantId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.DomainIdentifier.String())
		return
	}

	plan.GrantID = fwflex.StringToFramework(ctx, out.GrantId)

	// Getting attributes not outputted by AddPolicyGrant API call.
	// i.e. CreatedAt, CreatedBy
	grant, err := findPolicyGrantByID(ctx, conn,
		plan.DomainIdentifier.ValueString(),
		plan.EntityType.ValueString(),
		plan.EntityIdentifier.ValueString(),
		plan.PolicyType.ValueString(),
		plan.GrantID.ValueString(),
	)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GrantID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, grant, &plan))
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state policyGrantResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// RemovePolicyGrant requires Principal (a union type), which AutoFlex cannot expand
	// automatically, so we fetch the current grant to obtain it.
	grant, err := findPolicyGrantByID(ctx, conn,
		state.DomainIdentifier.ValueString(),
		state.EntityType.ValueString(),
		state.EntityIdentifier.ValueString(),
		state.PolicyType.ValueString(),
		state.GrantID.ValueString(),
	)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GrantID.String())
		return
	}

	in := &datazone.RemovePolicyGrantInput{
		ClientToken:      aws.String(create.UniqueId(ctx)),
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		EntityIdentifier: state.EntityIdentifier.ValueStringPointer(),
		EntityType:       awstypes.TargetEntityType(state.EntityType.ValueString()),
		GrantIdentifier:  state.GrantID.ValueStringPointer(),
		PolicyType:       awstypes.ManagedPolicyType(state.PolicyType.ValueString()),
		Principal:        grant.Principal,
	}

	_, err = conn.RemovePolicyGrant(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if isResourceMissing(err) {
		return
	}
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "is not part of the existing Grant list") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GrantID.String())
		return
	}
}

func findPolicyGrantByID(ctx context.Context, conn *datazone.Client, domainID, entityType, entityID, policyType, grantID string) (*awstypes.PolicyGrantMember, error) {
	input := datazone.ListPolicyGrantsInput{
		DomainIdentifier: aws.String(domainID),
		EntityIdentifier: aws.String(entityID),
		EntityType:       awstypes.TargetEntityType(entityType),
		PolicyType:       awstypes.ManagedPolicyType(policyType),
	}

	pages := datazone.NewListPolicyGrantsPaginator(conn, &input)
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
	Detail           fwtypes.ListNestedObjectValueOf[detailModel]    `tfsdk:"detail"`
	DomainIdentifier types.String                                    `tfsdk:"domain_identifier"`
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
	IncludeChildDomainUnits types.Bool                        `tfsdk:"include_child_domain_units"`
	ProjectProfiles         fwtypes.ListValueOf[types.String] `tfsdk:"project_profiles"`
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

var (
	_ fwflex.Expander  = detailModel{}
	_ fwflex.Flattener = &detailModel{}
	_ fwflex.Expander  = principalModel{}
	_ fwflex.Flattener = &principalModel{}
)

// Expand implements fwflex.Expander for detailModel, converting the Terraform
// model to an AWS SDK PolicyGrantDetail union type.
func (m detailModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.AddToProjectMemberPool.IsNull():
		data, d := m.AddToProjectMemberPool.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberAddToProjectMemberPool{
			Value: awstypes.AddToProjectMemberPoolPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateAssetType.IsNull():
		data, d := m.CreateAssetType.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateAssetType{
			Value: awstypes.CreateAssetTypePolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateDomainUnit.IsNull():
		data, d := m.CreateDomainUnit.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateDomainUnit{
			Value: awstypes.CreateDomainUnitPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateEnvironment.IsNull():
		return &awstypes.PolicyGrantDetailMemberCreateEnvironment{
			Value: awstypes.Unit{},
		}, diags

	case !m.CreateEnvironmentFromBlueprint.IsNull():
		return &awstypes.PolicyGrantDetailMemberCreateEnvironmentFromBlueprint{
			Value: awstypes.Unit{},
		}, diags

	case !m.CreateEnvironmentProfile.IsNull():
		data, d := m.CreateEnvironmentProfile.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateEnvironmentProfile{
			Value: awstypes.CreateEnvironmentProfilePolicyGrantDetail{
				DomainUnitId: data.DomainUnitID.ValueStringPointer(),
			},
		}, diags

	case !m.CreateFormType.IsNull():
		data, d := m.CreateFormType.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateFormType{
			Value: awstypes.CreateFormTypePolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateGlossary.IsNull():
		data, d := m.CreateGlossary.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateGlossary{
			Value: awstypes.CreateGlossaryPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateProject.IsNull():
		data, d := m.CreateProject.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberCreateProject{
			Value: awstypes.CreateProjectPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.CreateProjectFromProjectProfile.IsNull():
		data, d := m.CreateProjectFromProjectProfile.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		val := awstypes.CreateProjectFromProjectProfilePolicyGrantDetail{
			IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
		}
		if !data.ProjectProfiles.IsNull() && !data.ProjectProfiles.IsUnknown() {
			smerr.AddEnrich(ctx, &diags, data.ProjectProfiles.ElementsAs(ctx, &val.ProjectProfiles, false))
			if diags.HasError() {
				return nil, diags
			}
		}
		return &awstypes.PolicyGrantDetailMemberCreateProjectFromProjectProfile{Value: val}, diags

	case !m.DelegateCreateEnvironmentProfile.IsNull():
		return &awstypes.PolicyGrantDetailMemberDelegateCreateEnvironmentProfile{
			Value: awstypes.Unit{},
		}, diags

	case !m.OverrideDomainUnitOwners.IsNull():
		data, d := m.OverrideDomainUnitOwners.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberOverrideDomainUnitOwners{
			Value: awstypes.OverrideDomainUnitOwnersPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.OverrideProjectOwners.IsNull():
		data, d := m.OverrideProjectOwners.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberOverrideProjectOwners{
			Value: awstypes.OverrideProjectOwnersPolicyGrantDetail{
				IncludeChildDomainUnits: data.IncludeChildDomainUnits.ValueBoolPointer(),
			},
		}, diags

	case !m.UseAssetType.IsNull():
		data, d := m.UseAssetType.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantDetailMemberUseAssetType{
			Value: awstypes.UseAssetTypePolicyGrantDetail{
				DomainUnitId: data.DomainUnitID.ValueStringPointer(),
			},
		}, diags
	}

	diags.AddError("Invalid detail", "Exactly one detail variant block must be specified.")
	return nil, diags
}

// Flatten implements fwflex.Flattener for detailModel, converting an AWS SDK
// PolicyGrantDetail union type to the Terraform model.
func (m *detailModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Initialize all variants to null; the matching case below will set the active one.
	m.AddToProjectMemberPool = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateAssetType = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateDomainUnit = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateEnvironment = fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx)
	m.CreateEnvironmentFromBlueprint = fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx)
	m.CreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfNull[domainUnitIDDetailModel](ctx)
	m.CreateFormType = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateGlossary = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateProject = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.CreateProjectFromProjectProfile = fwtypes.NewListNestedObjectValueOfNull[createProjectFromProjectProfileDetailModel](ctx)
	m.DelegateCreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx)
	m.OverrideDomainUnitOwners = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.OverrideProjectOwners = fwtypes.NewListNestedObjectValueOfNull[includeChildDomainUnitsDetailModel](ctx)
	m.UseAssetType = fwtypes.NewListNestedObjectValueOfNull[domainUnitIDDetailModel](ctx)

	switch t := v.(type) {
	case *awstypes.PolicyGrantDetailMemberAddToProjectMemberPool:
		m.AddToProjectMemberPool = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateAssetType:
		m.CreateAssetType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateDomainUnit:
		m.CreateDomainUnit = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateEnvironment:
		m.CreateEnvironment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})

	case *awstypes.PolicyGrantDetailMemberCreateEnvironmentFromBlueprint:
		m.CreateEnvironmentFromBlueprint = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})

	case *awstypes.PolicyGrantDetailMemberCreateEnvironmentProfile:
		m.CreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitIDDetailModel{
			DomainUnitID: fwflex.StringToFramework(ctx, t.Value.DomainUnitId),
		})

	case *awstypes.PolicyGrantDetailMemberCreateFormType:
		m.CreateFormType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateGlossary:
		m.CreateGlossary = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateProject:
		m.CreateProject = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberCreateProjectFromProjectProfile:
		ppModel := &createProjectFromProjectProfileDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
			ProjectProfiles:         fwtypes.NewListValueOfNull[types.String](ctx),
		}
		if len(t.Value.ProjectProfiles) > 0 {
			ppModel.ProjectProfiles = fwflex.FlattenFrameworkStringValueListOfString(ctx, t.Value.ProjectProfiles)
		}
		m.CreateProjectFromProjectProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, ppModel)

	case *awstypes.PolicyGrantDetailMemberDelegateCreateEnvironmentProfile:
		m.DelegateCreateEnvironmentProfile = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})

	case *awstypes.PolicyGrantDetailMemberOverrideDomainUnitOwners:
		m.OverrideDomainUnitOwners = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberOverrideProjectOwners:
		m.OverrideProjectOwners = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &includeChildDomainUnitsDetailModel{
			IncludeChildDomainUnits: types.BoolPointerValue(t.Value.IncludeChildDomainUnits),
		})

	case *awstypes.PolicyGrantDetailMemberUseAssetType:
		m.UseAssetType = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitIDDetailModel{
			DomainUnitID: fwflex.StringToFramework(ctx, t.Value.DomainUnitId),
		})

	default:
		diags.AddError("Unsupported PolicyGrantDetail type", fmt.Sprintf("detail flatten: unexpected type %T", v))
	}

	return diags
}

// Expand implements fwflex.Expander for principalModel, converting the Terraform
// model to an AWS SDK PolicyGrantPrincipal union type.
func (m principalModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.DomainUnit.IsNull():
		data, d := m.DomainUnit.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		result := awstypes.DomainUnitPolicyGrantPrincipal{
			DomainUnitDesignation: awstypes.DomainUnitDesignation(data.DomainUnitDesignation.ValueString()),
			DomainUnitIdentifier:  data.DomainUnitIdentifier.ValueStringPointer(),
		}
		if !data.AllDomainUnitsGrantFilter.IsNull() {
			result.DomainUnitGrantFilter = &awstypes.DomainUnitGrantFilterMemberAllDomainUnitsGrantFilter{
				Value: awstypes.AllDomainUnitsGrantFilter{},
			}
		}
		return &awstypes.PolicyGrantPrincipalMemberDomainUnit{Value: result}, diags

	case !m.Group.IsNull():
		data, d := m.Group.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.PolicyGrantPrincipalMemberGroup{
			Value: &awstypes.GroupPolicyGrantPrincipalMemberGroupIdentifier{
				Value: data.GroupIdentifier.ValueString(),
			},
		}, diags

	case !m.Project.IsNull():
		data, d := m.Project.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		result := awstypes.ProjectPolicyGrantPrincipal{
			ProjectDesignation: awstypes.ProjectDesignation(data.ProjectDesignation.ValueString()),
			ProjectIdentifier:  data.ProjectIdentifier.ValueStringPointer(),
		}
		if !data.DomainUnitFilter.IsNull() {
			filter, d := data.DomainUnitFilter.ToPtr(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return nil, diags
			}
			result.ProjectGrantFilter = &awstypes.ProjectGrantFilterMemberDomainUnitFilter{
				Value: awstypes.DomainUnitFilterForProject{
					DomainUnit:              filter.DomainUnit.ValueStringPointer(),
					IncludeChildDomainUnits: filter.IncludeChildDomainUnits.ValueBoolPointer(),
				},
			}
		}
		return &awstypes.PolicyGrantPrincipalMemberProject{Value: result}, diags

	case !m.User.IsNull():
		data, d := m.User.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		if !data.UserIdentifier.IsNull() && !data.UserIdentifier.IsUnknown() {
			return &awstypes.PolicyGrantPrincipalMemberUser{
				Value: &awstypes.UserPolicyGrantPrincipalMemberUserIdentifier{
					Value: data.UserIdentifier.ValueString(),
				},
			}, diags
		}
		if !data.AllUsersGrantFilter.IsNull() {
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

// Flatten implements fwflex.Flattener for principalModel, converting an AWS SDK
// PolicyGrantPrincipal union type to the Terraform model.
func (m *principalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Initialize all variants to null; the matching case below will set the active one.
	m.DomainUnit = fwtypes.NewListNestedObjectValueOfNull[domainUnitPrincipalModel](ctx)
	m.Group = fwtypes.NewListNestedObjectValueOfNull[groupPrincipalModel](ctx)
	m.Project = fwtypes.NewListNestedObjectValueOfNull[projectPrincipalModel](ctx)
	m.User = fwtypes.NewListNestedObjectValueOfNull[userPrincipalModel](ctx)

	switch t := v.(type) {
	case *awstypes.PolicyGrantPrincipalMemberDomainUnit:
		duModel := &domainUnitPrincipalModel{
			DomainUnitDesignation:     fwtypes.StringEnumValue(t.Value.DomainUnitDesignation),
			DomainUnitIdentifier:      fwflex.StringToFramework(ctx, t.Value.DomainUnitIdentifier),
			AllDomainUnitsGrantFilter: fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		}
		if _, ok := t.Value.DomainUnitGrantFilter.(*awstypes.DomainUnitGrantFilterMemberAllDomainUnitsGrantFilter); ok {
			duModel.AllDomainUnitsGrantFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
		}
		m.DomainUnit = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, duModel)

	case *awstypes.PolicyGrantPrincipalMemberGroup:
		groupID := ""
		if gm, ok := t.Value.(*awstypes.GroupPolicyGrantPrincipalMemberGroupIdentifier); ok {
			groupID = gm.Value
		}
		m.Group = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &groupPrincipalModel{
			GroupIdentifier: types.StringValue(groupID),
		})

	case *awstypes.PolicyGrantPrincipalMemberProject:
		projModel := &projectPrincipalModel{
			ProjectDesignation: fwtypes.StringEnumValue(t.Value.ProjectDesignation),
			ProjectIdentifier:  fwflex.StringToFramework(ctx, t.Value.ProjectIdentifier),
			DomainUnitFilter:   fwtypes.NewListNestedObjectValueOfNull[domainUnitFilterModel](ctx),
		}
		if pf, ok := t.Value.ProjectGrantFilter.(*awstypes.ProjectGrantFilterMemberDomainUnitFilter); ok {
			projModel.DomainUnitFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &domainUnitFilterModel{
				DomainUnit:              fwflex.StringToFramework(ctx, pf.Value.DomainUnit),
				IncludeChildDomainUnits: types.BoolPointerValue(pf.Value.IncludeChildDomainUnits),
			})
		}
		m.Project = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, projModel)

	case *awstypes.PolicyGrantPrincipalMemberUser:
		userModel := &userPrincipalModel{
			UserIdentifier:      types.StringNull(),
			AllUsersGrantFilter: fwtypes.NewListNestedObjectValueOfNull[unitModel](ctx),
		}
		switch u := t.Value.(type) {
		case *awstypes.UserPolicyGrantPrincipalMemberUserIdentifier:
			userModel.UserIdentifier = types.StringValue(u.Value)
		case *awstypes.UserPolicyGrantPrincipalMemberAllUsersGrantFilter:
			userModel.AllUsersGrantFilter = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &unitModel{})
		}
		m.User = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, userModel)

	default:
		diags.AddError("Unsupported PolicyGrantPrincipal type", fmt.Sprintf("principal flatten: unexpected type %T", v))
	}

	return diags
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
