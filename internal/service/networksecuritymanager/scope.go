// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_scope", name="Scope")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager;networksecuritymanager.GetScopeOutput")
// @Testing(identityRegionOverrideTest=false)
func newScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &scopeResource{}

	return r, nil
}

const (
	// scopeDraftQualifier is the ARN qualifier of a scope's pending draft. A
	// draft saved on top of a published scope, or a scope created unpublished,
	// can only be read, updated, published or deleted through the
	// DRAFT-qualified ARN.
	scopeDraftQualifier = ":DRAFT"

	// scopeMaxExpressionOperands is the API limit on the number of criteria
	// under a single "and" or "or" operator.
	scopeMaxExpressionOperands = 20
)

type scopeResource struct {
	framework.ResourceWithModel[scopeResourceModel]
	framework.WithImportByIdentity
}

func (r *scopeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9 _.:/=+\-@]*$`), "must contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"has_published_version": schema.BoolAttribute{
				Computed: true,
			},
			"is_published": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.:/=+\-@]*$`), "must start with an alphanumeric character and contain only alphanumeric characters, spaces, and _.:/=+-@"),
				},
			},
			"scope_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EntityStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"scope_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scopeConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"account_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[accountFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"exclude": scopeAccountSetBlock(ctx),
									"include": scopeAccountSetBlock(ctx),
								},
							},
						},
						"resource_scope": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceScopeModel](ctx),
							Validators: []validator.Set{
								setvalidator.IsRequired(),
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional: true,
									},
									names.AttrResourceType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ScopeResourceType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"exclude": scopeResourceSetBlock(ctx),
									"include": scopeResourceSetBlock(ctx),
								},
							},
						},
					},
				},
			},
		},
	}
}

func scopeAccountSetBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[accountSetModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"account_ids": schema.SetAttribute{
					CustomType:  fwtypes.SetOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtMost(1000),
						setvalidator.ValueStringsAre(fwvalidators.AWSAccountID()),
					},
				},
				"organizational_units": schema.SetAttribute{
					CustomType:  fwtypes.SetOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtMost(1000),
						setvalidator.ValueStringsAre(
							stringvalidator.RegexMatches(regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[0-9a-z]{8,32}$`), "must be an organizational unit id"),
						),
					},
				},
			},
		},
	}
}

func scopeResourceSetBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSetModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"explicit_arns": schema.SetAttribute{
					CustomType:  fwtypes.SetOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Set{
						setvalidator.SizeAtMost(100),
						setvalidator.ValueStringsAre(fwvalidators.ARN()),
					},
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrExpression: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[resourceExpressionModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"and": scopeExpressionOperandsBlock(ctx, scopeMaxExpressionOperands),
							"criteria": scopeCriteriaBlock(ctx, []validator.List{
								listvalidator.SizeAtMost(1),
							}),
							"not": scopeExpressionOperandsBlock(ctx, 1),
							"or":  scopeExpressionOperandsBlock(ctx, scopeMaxExpressionOperands),
						},
					},
				},
			},
		},
	}
}

// scopeExpressionOperandsBlock is an "and", "or" or "not" operator. The API
// allows criteria directly under an operator but no operator under another
// one, so an operator's operands are all criteria.
//
// The exactly-one-of rules of the expression and criteria blocks are checked
// in ValidateConfig: path expressions do not resolve through the enclosing
// resource_scope set.
func scopeExpressionOperandsBlock(ctx context.Context, maxOperands int) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceExpressionOperandsModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"criteria": scopeCriteriaBlock(ctx, []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, maxOperands),
				}),
			},
		},
	}
}

func scopeCriteriaBlock(ctx context.Context, validators []validator.List) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceCriteriaModel](ctx),
		Validators: validators,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrTags: schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeAtMost(100),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"alb_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[albConfigurationModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrIPAddressType: schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.IpAddressType](),
								Optional:   true,
							},
							"scheme": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.Scheme](),
								Optional:   true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *scopeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ScopeConfiguration.IsUnknown() {
		return
	}

	scopeConfiguration, d := config.ScopeConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() || scopeConfiguration == nil {
		return
	}

	configurationPath := path.Root("scope_configuration").AtListIndex(0)

	if !scopeConfiguration.AccountFilter.IsUnknown() {
		accountFilter, d := scopeConfiguration.AccountFilter.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if d.HasError() {
			return
		}

		if accountFilter != nil {
			resp.Diagnostics.Append(validateScopeSelection(
				configurationPath.AtName("account_filter").AtListIndex(0),
				accountFilter.IncludeAll,
				accountFilter.Include.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0,
				accountFilter.Exclude.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0,
			)...)
		}
	}

	if scopeConfiguration.ResourceScope.IsUnknown() {
		return
	}

	// Only a failure to read the configuration stops the checks below: an
	// invalid account filter must not hide an invalid resource scope.
	resourceScopes, d := scopeConfiguration.ResourceScope.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if d.HasError() {
		return
	}

	resourceScopePath := configurationPath.AtName("resource_scope")
	seen := make(map[awstypes.ScopeResourceType]struct{}, len(resourceScopes))
	var global, regional []string

	for _, resourceScope := range resourceScopes {
		if resourceScope == nil {
			continue
		}

		resp.Diagnostics.Append(validateScopeSelection(
			resourceScopePath,
			resourceScope.IncludeAll,
			resourceScope.Include.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0,
			resourceScope.Exclude.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0,
		)...)

		resp.Diagnostics.Append(validateScopeResourceSet(ctx, resourceScopePath.AtName("exclude"), resourceScope.Exclude)...)
		resp.Diagnostics.Append(validateScopeResourceSet(ctx, resourceScopePath.AtName("include"), resourceScope.Include)...)

		if resourceScope.ResourceType.IsUnknown() || resourceScope.ResourceType.IsNull() {
			continue
		}

		resourceType := resourceScope.ResourceType.ValueEnum()
		if _, ok := seen[resourceType]; ok {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				resourceScopePath,
				"must not contain the same resource_type more than once",
				string(resourceType),
			))
		}
		seen[resourceType] = struct{}{}

		if scopeResourceTypeIsGlobal(resourceType) {
			global = append(global, string(resourceType))
		} else {
			regional = append(regional, string(resourceType))
		}
	}

	// A scope selects either global or regional resource types.
	if len(global) > 0 && len(regional) > 0 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			resourceScopePath,
			"must not combine global and regional resource types",
			fmt.Sprintf("%s (global) with %s (regional)", strings.Join(global, ", "), strings.Join(regional, ", ")),
		))
	}
}

// validateScopeResourceSet checks the exactly-one-of rules of an include or
// exclude resource set's expression and criteria blocks.
func validateScopeResourceSet(ctx context.Context, p path.Path, tfList fwtypes.ListNestedObjectValueOf[resourceSetModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	if tfList.IsUnknown() {
		return diags
	}

	resourceSet, d := tfList.ToPtr(ctx)
	diags.Append(d...)
	if d.HasError() || resourceSet == nil || resourceSet.Expression.IsUnknown() {
		return diags
	}

	expression, d := resourceSet.Expression.ToPtr(ctx)
	diags.Append(d...)
	if d.HasError() || expression == nil {
		return diags
	}

	expressionPath := p.AtListIndex(0).AtName(names.AttrExpression).AtListIndex(0)

	// Iterated in a fixed order: reporting a subset of the operands' errors
	// depending on Go's map ordering makes the diagnostics irreproducible.
	operators := []struct {
		name     string
		operands fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel]
	}{
		{"and", expression.And},
		{"not", expression.Not},
		{"or", expression.Or},
	}

	n := 0
	if expression.Criteria.IsUnknown() || expression.Criteria.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
		n++
		diags.Append(validateScopeCriteria(ctx, expressionPath.AtName("criteria"), expression.Criteria)...)
	}
	for _, operator := range operators {
		if operator.operands.IsUnknown() {
			n++
			continue
		}
		if operator.operands.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
			continue
		}
		n++

		operands, d := operator.operands.ToPtr(ctx)
		diags.Append(d...)
		if d.HasError() || operands == nil {
			continue
		}
		diags.Append(validateScopeCriteria(ctx, expressionPath.AtName(operator.name).AtListIndex(0).AtName("criteria"), operands.Criteria)...)
	}
	if n != 1 {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			expressionPath,
			`exactly one of "and", "criteria", "not" or "or" must be set`,
		))
	}

	return diags
}

// validateScopeCriteria checks that every criteria sets exactly one of tags
// and alb_config.
func validateScopeCriteria(ctx context.Context, p path.Path, tfList fwtypes.ListNestedObjectValueOf[resourceCriteriaModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	if tfList.IsUnknown() {
		return diags
	}

	criteria, d := tfList.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	for i, v := range criteria {
		if v == nil || v.Tags.IsUnknown() || v.ALBConfig.IsUnknown() {
			continue
		}

		n := 0
		if !v.Tags.IsNull() {
			n++
		}
		if v.ALBConfig.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
			n++
		}
		if n != 1 {
			diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
				p.AtListIndex(i),
				`exactly one of "alb_config" or "tags" must be set`,
			))
		}
	}

	return diags
}

// validateScopeSelection checks that exactly one of include_all (set to true),
// include or exclude is configured on an account filter or a resource scope.
func validateScopeSelection(p path.Path, includeAll types.Bool, include, exclude bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if includeAll.IsUnknown() {
		return diags
	}

	if !includeAll.IsNull() && !includeAll.ValueBool() {
		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			p.AtName("include_all"),
			"must be true when set",
			"false",
		))
	}

	n := 0
	for _, set := range []bool{includeAll.ValueBool(), include, exclude} {
		if set {
			n++
		}
	}
	if n != 1 {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			p,
			`exactly one of "include_all", "include" or "exclude" must be set`,
		))
	}

	return diags
}

// scopeResourceTypeIsGlobal reports whether a resource type is global. A scope
// cannot combine global and regional resource types, and a scope's kind cannot
// change once it is created.
func scopeResourceTypeIsGlobal(resourceType awstypes.ScopeResourceType) bool {
	return resourceType == awstypes.ScopeResourceTypeCfd
}

func (r *scopeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only an update can change these; both the account filter's presence and
	// the resource types' kind are fixed when the scope is created.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ScopeConfiguration.IsUnknown() {
		return
	}

	planned, d := plan.ScopeConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	current, d := state.ScopeConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() || planned == nil || current == nil {
		return
	}

	configurationPath := path.Root("scope_configuration").AtListIndex(0)

	if !planned.AccountFilter.IsUnknown() {
		plannedFilter := planned.AccountFilter.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
		currentFilter := current.AccountFilter.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
		if plannedFilter != currentFilter {
			resp.RequiresReplace = append(resp.RequiresReplace, configurationPath.AtName("account_filter"))
		}
	}

	if !planned.ResourceScope.IsUnknown() {
		plannedGlobal, d := scopeConfigurationIsGlobal(ctx, planned)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		currentGlobal, d := scopeConfigurationIsGlobal(ctx, current)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		if plannedGlobal != currentGlobal {
			resp.RequiresReplace = append(resp.RequiresReplace, configurationPath.AtName("resource_scope"))
		}
	}
}

func scopeConfigurationIsGlobal(ctx context.Context, data *scopeConfigurationModel) (bool, diag.Diagnostics) {
	resourceScopes, diags := data.ResourceScope.ToSlice(ctx)
	if diags.HasError() {
		return false, diags
	}

	for _, resourceScope := range resourceScopes {
		if resourceScope == nil || resourceScope.ResourceType.IsUnknown() || resourceScope.ResourceType.IsNull() {
			continue
		}
		if scopeResourceTypeIsGlobal(resourceScope.ResourceType.ValueEnum()) {
			return true, diags
		}
	}

	return false, diags
}

func (r *scopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input networksecuritymanager.CreateScopeInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Scope")))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	scopeConfiguration, d := expandScopeConfiguration(ctx, plan.ScopeConfiguration)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ScopeConfiguration = scopeConfiguration
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateScope(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenScope(ctx, (*networksecuritymanager.GetScopeOutput)(output), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *scopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := scopeBaseARN(state.ARN.ValueString())
	output, err := findScopeByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenScope(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *scopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var plan, state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := scopeBaseARN(state.ARN.ValueString())

	// Every write returns a new update token and the token is required on the
	// next update, so the current one is read back rather than kept in state:
	// a change made outside Terraform would otherwise fail the update with a
	// ConflictException.
	current, err := findScopeByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if !plan.ScopeConfiguration.Equal(state.ScopeConfiguration) || !plan.Description.Equal(state.Description) || !plan.IsPublished.Equal(state.IsPublished) {
		var input networksecuritymanager.UpdateScopeInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Scope")))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		scopeConfiguration, d := expandScopeConfiguration(ctx, plan.ScopeConfiguration)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ScopeConfiguration = scopeConfiguration
		// A pending draft can only be updated (or published) through its
		// DRAFT-qualified ARN; the base ARN is rejected while a draft exists.
		input.ScopeIdentifier = aws.String(scopeUpdateIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken
		if plan.Description.IsNull() {
			// An omitted description leaves the current one in place; an
			// empty string clears it.
			input.ScopeDescription = aws.String("")
		}

		output, err := conn.UpdateScope(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		current = (*networksecuritymanager.GetScopeOutput)(output)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenScope(ctx, current, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *scopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	var state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := scopeBaseARN(state.ARN.ValueString())

	// A published scope with a pending draft cannot be deleted until the draft
	// is; a scope that was never published exists only as a draft, and
	// deleting its base ARN succeeds without removing anything.
	for _, identifier := range []string{arn + scopeDraftQualifier, arn} {
		input := networksecuritymanager.DeleteScopeInput{
			ScopeIdentifier: aws.String(identifier),
		}

		_, err := conn.DeleteScope(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}
}

// findScopeByARN returns the version of the scope that an update would act on:
// the pending draft when there is one, otherwise the published scope.
func findScopeByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetScopeOutput, error) {
	output, err := findScopeByIdentifier(ctx, conn, arn+scopeDraftQualifier)
	if retry.NotFound(err) {
		return findScopeByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findScopeByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetScopeOutput, error) {
	input := networksecuritymanager.GetScopeInput{
		ScopeIdentifier: aws.String(identifier),
	}

	output, err := conn.GetScope(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

// scopeBaseARN strips the DRAFT qualifier, so that the scope is always
// identified by the ARN it keeps across publication.
func scopeBaseARN(arn string) string {
	return strings.TrimSuffix(arn, scopeDraftQualifier)
}

func scopeUpdateIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + scopeDraftQualifier
	}

	return arn
}

func expandScopeConfiguration(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[scopeConfigurationModel]) (*awstypes.ScopeConfiguration, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	tfObject, d := tfList.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || tfObject == nil {
		return nil, diags
	}

	apiObject := &awstypes.ScopeConfiguration{
		ResourceScopes: map[string]awstypes.ResourceScope{},
	}

	accountFilter, d := tfObject.AccountFilter.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	if accountFilter != nil {
		apiObject.AccountFilter, d = expandAccountFilter(ctx, accountFilter)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}

	resourceScopes, d := tfObject.ResourceScope.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	for _, resourceScope := range resourceScopes {
		if resourceScope == nil {
			continue
		}

		var v awstypes.ResourceScope

		if resourceScope.IncludeAll.ValueBool() {
			v.IncludeAll = aws.Bool(true)
		}

		v.Include, d = expandResourceSet(ctx, resourceScope.Include)
		smerr.AddEnrich(ctx, &diags, d)
		v.Exclude, d = expandResourceSet(ctx, resourceScope.Exclude)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		apiObject.ResourceScopes[resourceScope.ResourceType.ValueString()] = v
	}

	return apiObject, diags
}

func expandAccountFilter(ctx context.Context, tfObject *accountFilterModel) (awstypes.AccountFilter, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	switch {
	case tfObject.IncludeAll.ValueBool():
		return &awstypes.AccountFilterMemberIncludeAll{Value: awstypes.Unit{}}, diags

	case tfObject.Include.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		accountSet, d := expandAccountSet(ctx, tfObject.Include)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() || accountSet == nil {
			return nil, diags
		}
		return &awstypes.AccountFilterMemberInclude{Value: *accountSet}, diags

	case tfObject.Exclude.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		accountSet, d := expandAccountSet(ctx, tfObject.Exclude)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() || accountSet == nil {
			return nil, diags
		}
		return &awstypes.AccountFilterMemberExclude{Value: *accountSet}, diags
	}

	return nil, diags
}

func expandAccountSet(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[accountSetModel]) (*awstypes.AccountSet, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	tfObject, d := tfList.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || tfObject == nil {
		return nil, diags
	}

	return &awstypes.AccountSet{
		AccountIds:          fwflex.ExpandFrameworkStringValueSet(ctx, tfObject.AccountIDs),
		OrganizationalUnits: fwflex.ExpandFrameworkStringValueSet(ctx, tfObject.OrganizationalUnits),
	}, diags
}

func expandResourceSet(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[resourceSetModel]) (*awstypes.ResourceSet, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	tfObject, d := tfList.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || tfObject == nil {
		return nil, diags
	}

	apiObject := &awstypes.ResourceSet{
		ExplicitArns: fwflex.ExpandFrameworkStringValueSet(ctx, tfObject.ExplicitARNs),
	}

	expression, d := tfObject.Expression.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	if expression != nil {
		apiObject.Expression, d = expandResourceLogicalExpression(ctx, expression)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}

	return apiObject, diags
}

func expandResourceLogicalExpression(ctx context.Context, tfObject *resourceExpressionModel) (awstypes.ResourceLogicalExpression, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	switch {
	case tfObject.Criteria.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		criteria, d := expandResourceCriteriaList(ctx, tfObject.Criteria)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() || len(criteria) == 0 {
			return nil, diags
		}
		return criteria[0], diags

	case tfObject.And.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		operands, d := expandResourceExpressionOperands(ctx, tfObject.And)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.ResourceLogicalExpressionMemberAnd{Value: operands}, diags

	case tfObject.Or.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		operands, d := expandResourceExpressionOperands(ctx, tfObject.Or)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.ResourceLogicalExpressionMemberOr{Value: operands}, diags

	case tfObject.Not.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		operands, d := expandResourceExpressionOperands(ctx, tfObject.Not)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() || len(operands) == 0 {
			return nil, diags
		}
		return &awstypes.ResourceLogicalExpressionMemberNot{Value: operands[0]}, diags
	}

	return nil, diags
}

func expandResourceExpressionOperands(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel]) ([]awstypes.ResourceLogicalExpression, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	tfObject, d := tfList.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || tfObject == nil {
		return nil, diags
	}

	return expandResourceCriteriaList(ctx, tfObject.Criteria)
}

func expandResourceCriteriaList(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[resourceCriteriaModel]) ([]awstypes.ResourceLogicalExpression, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	tfObjects, d := tfList.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}

	apiObjects := make([]awstypes.ResourceLogicalExpression, 0, len(tfObjects))

	for _, tfObject := range tfObjects {
		if tfObject == nil {
			continue
		}

		criteria, d := expandResourceCriteria(ctx, tfObject)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		apiObjects = append(apiObjects, &awstypes.ResourceLogicalExpressionMemberCriteria{Value: criteria})
	}

	return apiObjects, diags
}

func expandResourceCriteria(ctx context.Context, tfObject *resourceCriteriaModel) (awstypes.ResourceCriteria, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	switch {
	case !tfObject.Tags.IsNull():
		tags := fwflex.ExpandFrameworkStringValueMap(ctx, tfObject.Tags)
		if tags == nil {
			tags = map[string]string{}
		}
		return &awstypes.ResourceCriteriaMemberTags{Value: tags}, diags

	case tfObject.ALBConfig.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		albConfig, d := tfObject.ALBConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() || albConfig == nil {
			return nil, diags
		}
		return &awstypes.ResourceCriteriaMemberAlbConfig{Value: awstypes.AlbConfiguration{
			IpAddressType: albConfig.IPAddressType.ValueEnum(),
			Scheme:        albConfig.Scheme.ValueEnum(),
		}}, diags
	}

	return nil, diags
}

func flattenScope(ctx context.Context, output *networksecuritymanager.GetScopeOutput, data *scopeResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// The API returns an empty description when none was set.
	description := data.Description

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("Scope")))
	if diags.HasError() {
		return diags
	}

	// A draft is returned with the DRAFT-qualified ARN.
	data.ARN = types.StringValue(scopeBaseARN(aws.ToString(output.ScopeArn)))
	data.HasPublishedVersion = types.BoolValue(aws.ToBool(output.HasPublishedVersion))
	data.IsPublished = types.BoolValue(output.Status != awstypes.EntityStatusDraft)
	if aws.ToString(output.ScopeDescription) == "" && (description.IsNull() || description.IsUnknown()) {
		data.Description = types.StringNull()
	}

	scopeConfiguration, d := flattenScopeConfiguration(ctx, output.ScopeConfiguration)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	data.ScopeConfiguration = scopeConfiguration

	return diags
}

func flattenScopeConfiguration(ctx context.Context, apiObject *awstypes.ScopeConfiguration) (fwtypes.ListNestedObjectValueOf[scopeConfigurationModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[scopeConfigurationModel](ctx), diags
	}

	tfObject := &scopeConfigurationModel{
		AccountFilter: fwtypes.NewListNestedObjectValueOfNull[accountFilterModel](ctx),
		ResourceScope: fwtypes.NewSetNestedObjectValueOfNull[resourceScopeModel](ctx),
	}

	if apiObject.AccountFilter != nil {
		accountFilter, d := flattenAccountFilter(ctx, apiObject.AccountFilter)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[scopeConfigurationModel](ctx), diags
		}
		tfObject.AccountFilter, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, accountFilter)
		smerr.AddEnrich(ctx, &diags, d)
	}

	// The API keys resource scopes by resource type in no particular order.
	resourceTypes := make([]string, 0, len(apiObject.ResourceScopes))
	for resourceType := range apiObject.ResourceScopes {
		resourceTypes = append(resourceTypes, resourceType)
	}
	slices.Sort(resourceTypes)

	resourceScopes := make([]*resourceScopeModel, 0, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		v := apiObject.ResourceScopes[resourceType]

		resourceScope := &resourceScopeModel{
			IncludeAll:   types.BoolNull(),
			ResourceType: fwtypes.StringEnumValue(awstypes.ScopeResourceType(resourceType)),
		}
		if aws.ToBool(v.IncludeAll) {
			resourceScope.IncludeAll = types.BoolValue(true)
		}

		var d diag.Diagnostics
		resourceScope.Include, d = flattenResourceSet(ctx, v.Include)
		smerr.AddEnrich(ctx, &diags, d)
		resourceScope.Exclude, d = flattenResourceSet(ctx, v.Exclude)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[scopeConfigurationModel](ctx), diags
		}

		resourceScopes = append(resourceScopes, resourceScope)
	}

	var d diag.Diagnostics
	tfObject.ResourceScope, d = fwtypes.NewSetNestedObjectValueOfSlice(ctx, resourceScopes, nil)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[scopeConfigurationModel](ctx), diags
	}

	tfList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, tfObject)
	smerr.AddEnrich(ctx, &diags, d)

	return tfList, diags
}

func flattenAccountFilter(ctx context.Context, apiObject awstypes.AccountFilter) (*accountFilterModel, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	tfObject := &accountFilterModel{
		Exclude:    fwtypes.NewListNestedObjectValueOfNull[accountSetModel](ctx),
		Include:    fwtypes.NewListNestedObjectValueOfNull[accountSetModel](ctx),
		IncludeAll: types.BoolNull(),
	}

	switch v := apiObject.(type) {
	case *awstypes.AccountFilterMemberIncludeAll:
		tfObject.IncludeAll = types.BoolValue(true)
	case *awstypes.AccountFilterMemberInclude:
		var d diag.Diagnostics
		tfObject.Include, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenAccountSet(ctx, &v.Value))
		smerr.AddEnrich(ctx, &diags, d)
	case *awstypes.AccountFilterMemberExclude:
		var d diag.Diagnostics
		tfObject.Exclude, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenAccountSet(ctx, &v.Value))
		smerr.AddEnrich(ctx, &diags, d)
	case *awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, v.Tag, &diags)
	default:
		diags.AddError(
			"Unsupported Account Filter",
			fmt.Sprintf("flattenAccountFilter: %T", apiObject),
		)
	}

	return tfObject, diags
}

func flattenAccountSet(ctx context.Context, apiObject *awstypes.AccountSet) *accountSetModel { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	return &accountSetModel{
		AccountIDs:          fwflex.FlattenFrameworkStringValueSetOfString(ctx, apiObject.AccountIds),
		OrganizationalUnits: fwflex.FlattenFrameworkStringValueSetOfString(ctx, apiObject.OrganizationalUnits),
	}
}

func flattenResourceSet(ctx context.Context, apiObject *awstypes.ResourceSet) (fwtypes.ListNestedObjectValueOf[resourceSetModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[resourceSetModel](ctx), diags
	}

	tfObject := &resourceSetModel{
		ExplicitARNs: fwflex.FlattenFrameworkStringValueSetOfString(ctx, apiObject.ExplicitArns),
		Expression:   fwtypes.NewListNestedObjectValueOfNull[resourceExpressionModel](ctx),
	}

	if apiObject.Expression != nil {
		expression, d := flattenResourceLogicalExpression(ctx, apiObject.Expression)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[resourceSetModel](ctx), diags
		}
		tfObject.Expression, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, expression)
		smerr.AddEnrich(ctx, &diags, d)
	}

	tfList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, tfObject)
	smerr.AddEnrich(ctx, &diags, d)

	return tfList, diags
}

func flattenResourceLogicalExpression(ctx context.Context, apiObject awstypes.ResourceLogicalExpression) (*resourceExpressionModel, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	tfObject := &resourceExpressionModel{
		And:      fwtypes.NewListNestedObjectValueOfNull[resourceExpressionOperandsModel](ctx),
		Criteria: fwtypes.NewListNestedObjectValueOfNull[resourceCriteriaModel](ctx),
		Not:      fwtypes.NewListNestedObjectValueOfNull[resourceExpressionOperandsModel](ctx),
		Or:       fwtypes.NewListNestedObjectValueOfNull[resourceExpressionOperandsModel](ctx),
	}

	switch v := apiObject.(type) {
	case *awstypes.ResourceLogicalExpressionMemberCriteria:
		criteria, d := flattenResourceCriteriaList(ctx, []awstypes.ResourceLogicalExpression{v})
		smerr.AddEnrich(ctx, &diags, d)
		tfObject.Criteria = criteria
	case *awstypes.ResourceLogicalExpressionMemberAnd:
		operands, d := flattenResourceExpressionOperands(ctx, v.Value)
		smerr.AddEnrich(ctx, &diags, d)
		tfObject.And = operands
	case *awstypes.ResourceLogicalExpressionMemberOr:
		operands, d := flattenResourceExpressionOperands(ctx, v.Value)
		smerr.AddEnrich(ctx, &diags, d)
		tfObject.Or = operands
	case *awstypes.ResourceLogicalExpressionMemberNot:
		operands, d := flattenResourceExpressionOperands(ctx, []awstypes.ResourceLogicalExpression{v.Value})
		smerr.AddEnrich(ctx, &diags, d)
		tfObject.Not = operands
	case *awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, v.Tag, &diags)
	default:
		diags.AddError(
			"Unsupported Resource Logical Expression",
			fmt.Sprintf("flattenResourceLogicalExpression: %T", apiObject),
		)
	}

	return tfObject, diags
}

func flattenResourceExpressionOperands(ctx context.Context, apiObjects []awstypes.ResourceLogicalExpression) (fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	criteria, d := flattenResourceCriteriaList(ctx, apiObjects)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[resourceExpressionOperandsModel](ctx), diags
	}

	tfList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &resourceExpressionOperandsModel{
		Criteria: criteria,
	})
	smerr.AddEnrich(ctx, &diags, d)

	return tfList, diags
}

// flattenResourceCriteriaList flattens the operands of an operator. Every
// operand is a criteria: the API rejects an operator nested under another.
func flattenResourceCriteriaList(ctx context.Context, apiObjects []awstypes.ResourceLogicalExpression) (fwtypes.ListNestedObjectValueOf[resourceCriteriaModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	tfObjects := make([]*resourceCriteriaModel, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		v, ok := apiObject.(*awstypes.ResourceLogicalExpressionMemberCriteria)
		if !ok {
			diags.AddError(
				"Unsupported Resource Logical Expression",
				fmt.Sprintf("flattenResourceCriteriaList: nested %T", apiObject),
			)
			return fwtypes.NewListNestedObjectValueOfNull[resourceCriteriaModel](ctx), diags
		}

		tfObject, d := flattenResourceCriteria(ctx, v.Value)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[resourceCriteriaModel](ctx), diags
		}

		tfObjects = append(tfObjects, tfObject)
	}

	tfList, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, tfObjects, nil)
	smerr.AddEnrich(ctx, &diags, d)

	return tfList, diags
}

func flattenResourceCriteria(ctx context.Context, apiObject awstypes.ResourceCriteria) (*resourceCriteriaModel, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	tfObject := &resourceCriteriaModel{
		ALBConfig: fwtypes.NewListNestedObjectValueOfNull[albConfigurationModel](ctx),
		Tags:      fwtypes.NewMapValueOfNull[types.String](ctx),
	}

	switch v := apiObject.(type) {
	case *awstypes.ResourceCriteriaMemberTags:
		// An empty tag map is a valid (match-everything) criteria and must
		// stay distinct from no tags criteria at all.
		elements := make(map[string]attr.Value, len(v.Value))
		for k, val := range v.Value {
			elements[k] = types.StringValue(val)
		}
		var d diag.Diagnostics
		tfObject.Tags, d = fwtypes.NewMapValueOf[types.String](ctx, elements)
		smerr.AddEnrich(ctx, &diags, d)
	case *awstypes.ResourceCriteriaMemberAlbConfig:
		albConfig := &albConfigurationModel{
			IPAddressType: fwtypes.StringEnumNull[awstypes.IpAddressType](),
			Scheme:        fwtypes.StringEnumNull[awstypes.Scheme](),
		}
		if v.Value.IpAddressType != "" {
			albConfig.IPAddressType = fwtypes.StringEnumValue(v.Value.IpAddressType)
		}
		if v.Value.Scheme != "" {
			albConfig.Scheme = fwtypes.StringEnumValue(v.Value.Scheme)
		}
		var d diag.Diagnostics
		tfObject.ALBConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, albConfig)
		smerr.AddEnrich(ctx, &diags, d)
	case *awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, v.Tag, &diags)
	default:
		diags.AddError(
			"Unsupported Resource Criteria",
			fmt.Sprintf("flattenResourceCriteria: %T", apiObject),
		)
	}

	return tfObject, diags
}

type scopeResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String                                             `tfsdk:"arn"`
	Description         types.String                                             `tfsdk:"description"`
	HasPublishedVersion types.Bool                                               `tfsdk:"has_published_version"`
	IsPublished         types.Bool                                               `tfsdk:"is_published"`
	Name                types.String                                             `tfsdk:"name"`
	ScopeConfiguration  fwtypes.ListNestedObjectValueOf[scopeConfigurationModel] `tfsdk:"scope_configuration" autoflex:"-"`
	ScopeID             types.String                                             `tfsdk:"scope_id"`
	Status              fwtypes.StringEnum[awstypes.EntityStatus]                `tfsdk:"status"`
	Tags                tftags.Map                                               `tfsdk:"tags"`
	TagsAll             tftags.Map                                               `tfsdk:"tags_all"`
	UpdatedAt           timetypes.RFC3339                                        `tfsdk:"updated_at"`
	Version             types.String                                             `tfsdk:"version"`
}

type scopeConfigurationModel struct {
	AccountFilter fwtypes.ListNestedObjectValueOf[accountFilterModel] `tfsdk:"account_filter"`
	ResourceScope fwtypes.SetNestedObjectValueOf[resourceScopeModel]  `tfsdk:"resource_scope"`
}

type accountFilterModel struct {
	Exclude    fwtypes.ListNestedObjectValueOf[accountSetModel] `tfsdk:"exclude"`
	Include    fwtypes.ListNestedObjectValueOf[accountSetModel] `tfsdk:"include"`
	IncludeAll types.Bool                                       `tfsdk:"include_all"`
}

type accountSetModel struct {
	AccountIDs          fwtypes.SetOfString `tfsdk:"account_ids"`
	OrganizationalUnits fwtypes.SetOfString `tfsdk:"organizational_units"`
}

type resourceScopeModel struct {
	Exclude      fwtypes.ListNestedObjectValueOf[resourceSetModel] `tfsdk:"exclude"`
	Include      fwtypes.ListNestedObjectValueOf[resourceSetModel] `tfsdk:"include"`
	IncludeAll   types.Bool                                        `tfsdk:"include_all"`
	ResourceType fwtypes.StringEnum[awstypes.ScopeResourceType]    `tfsdk:"resource_type"`
}

type resourceSetModel struct {
	ExplicitARNs fwtypes.SetOfString                                      `tfsdk:"explicit_arns"`
	Expression   fwtypes.ListNestedObjectValueOf[resourceExpressionModel] `tfsdk:"expression"`
}

type resourceExpressionModel struct {
	And      fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel] `tfsdk:"and"`
	Criteria fwtypes.ListNestedObjectValueOf[resourceCriteriaModel]           `tfsdk:"criteria"`
	Not      fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel] `tfsdk:"not"`
	Or       fwtypes.ListNestedObjectValueOf[resourceExpressionOperandsModel] `tfsdk:"or"`
}

type resourceExpressionOperandsModel struct {
	Criteria fwtypes.ListNestedObjectValueOf[resourceCriteriaModel] `tfsdk:"criteria"`
}

type resourceCriteriaModel struct {
	ALBConfig fwtypes.ListNestedObjectValueOf[albConfigurationModel] `tfsdk:"alb_config"`
	Tags      fwtypes.MapOfString                                    `tfsdk:"tags"`
}

type albConfigurationModel struct {
	IPAddressType fwtypes.StringEnum[awstypes.IpAddressType] `tfsdk:"ip_address_type"`
	Scheme        fwtypes.StringEnum[awstypes.Scheme]        `tfsdk:"scheme"`
}
