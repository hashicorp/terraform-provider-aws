// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
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
// @Testing(preCheck="testAccPreCheckScope")
func newScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &scopeResource{}, nil
}

// scopeDraftQualifier is the ARN qualifier of a scope's pending draft. A scope created
// unpublished, or a draft saved on top of a published scope, can only be read,
// updated, published or deleted through the DRAFT-qualified ARN.
const scopeDraftQualifier = ":DRAFT"

// scopeMaxExpressionOperands is the maximum number of operands of an and or or expression.
const scopeMaxExpressionOperands = 20

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
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
			},
			"is_published": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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
							PlanModifiers: []planmodifier.List{
								// The account filter cannot be added to or removed from an existing scope.
								listplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
									resp.RequiresReplace = req.PlanValue.Length(fwtypes.CollectionLengthUnhandledAsZero) != req.StateValue.Length(fwtypes.CollectionLengthUnhandledAsZero)
								}, "Adding or removing account_filter forces replacement.", "Adding or removing `account_filter` forces replacement."),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.ExactlyOneOfChildren(
										path.MatchRelative().AtName("exclude"),
										path.MatchRelative().AtName("include"),
										path.MatchRelative().AtName("include_all"),
									),
								},
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional: true,
										Validators: []validator.Bool{
											boolvalidator.Equals(true),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"exclude": accountSetBlock(ctx),
									"include": accountSetBlock(ctx),
								},
							},
						},
						"resource_scope": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceScopeModel](ctx),
							Validators: []validator.Set{
								setvalidator.IsRequired(),
								setvalidator.SizeAtLeast(1),
							},
							PlanModifiers: []planmodifier.Set{
								// A scope selects either global or regional resource types and cannot change between them.
								setplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.RequiresReplaceIfFuncResponse) {
									if req.PlanValue.IsUnknown() {
										return
									}

									var plan, state []resourceScopeModel
									resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &plan, false)...)
									resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &state, false)...)
									resp.RequiresReplace = hasGlobalResourceType(plan) != hasGlobalResourceType(state)
								}, "Changing between global and regional resource types forces replacement.", "Changing between global and regional resource types forces replacement."),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									exactlyOneOfValueAttributesValidator{names: []string{"exclude", "include", "include_all"}},
								},
								Attributes: map[string]schema.Attribute{
									"include_all": schema.BoolAttribute{
										Optional: true,
										Validators: []validator.Bool{
											boolvalidator.Equals(true),
										},
									},
									names.AttrResourceType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ScopeResourceType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"exclude": resourceSetBlock(ctx),
									"include": resourceSetBlock(ctx),
								},
							},
						},
					},
				},
			},
		},
	}
}

// ValidateConfig rejects resource_scope blocks that repeat a resource_type. The service keys
// resource scopes by resource type, so all but one of them would be silently dropped.
func (r *scopeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	scopeConfiguration, d := data.ScopeConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() || scopeConfiguration == nil {
		return
	}

	resourceScopes, d := scopeConfiguration.ResourceScope.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := make(map[awstypes.ScopeResourceType]bool, len(resourceScopes))
	for _, v := range resourceScopes {
		if v.ResourceType.IsNull() || v.ResourceType.IsUnknown() {
			continue
		}

		resourceType := v.ResourceType.ValueEnum()
		if seen[resourceType] {
			smerr.AddOne(ctx, &resp.Diagnostics, diag.NewAttributeErrorDiagnostic(
				path.Root("scope_configuration").AtListIndex(0).AtName("resource_scope"),
				"Duplicate resource_type",
				fmt.Sprintf("Each resource_scope must have a different resource_type, but %q is configured more than once.", resourceType),
			))
			return
		}
		seen[resourceType] = true
	}
}

func accountSetBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[accountSetModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"account_ids": schema.SetAttribute{
					CustomType: fwtypes.SetOfStringType,
					Optional:   true,
				},
				"organizational_units": schema.SetAttribute{
					CustomType: fwtypes.SetOfStringType,
					Optional:   true,
				},
			},
		},
	}
}

func resourceSetBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSetModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"explicit_arns": schema.SetAttribute{
					CustomType: fwtypes.SetOfStringType,
					Optional:   true,
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrExpression: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[expressionModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Validators: []validator.Object{
							exactlyOneOfValueAttributesValidator{names: []string{"and", "criteria", "not", "or"}},
						},
						Blocks: map[string]schema.Block{
							// The operands of and, not and or are criteria: the service rejects an operator nested under another operator.
							"and":      criteriaBlock(ctx, scopeMaxExpressionOperands),
							"criteria": criteriaBlock(ctx, 1),
							"not":      criteriaBlock(ctx, 1),
							"or":       criteriaBlock(ctx, scopeMaxExpressionOperands),
						},
					},
				},
			},
		},
	}
}

func criteriaBlock(ctx context.Context, maxItems int) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[criteriaModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(maxItems),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				exactlyOneOfValueAttributesValidator{names: []string{"alb_config", names.AttrTags}},
			},
			Attributes: map[string]schema.Attribute{
				names.AttrTags: schema.MapAttribute{
					CustomType: fwtypes.MapOfStringType,
					Optional:   true,
				},
			},
			Blocks: map[string]schema.Block{
				"alb_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[albConfigModel](ctx),
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

func (r *scopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreateScopeInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Scope")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateScope(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetScopeOutput)(out), &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *scopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findScopeByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state), smerr.ID, arn)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *scopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// Updates are guarded by the token returned from the most recent read or write.
		current, err := findScopeByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdateScopeInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, append(diff.IgnoredFieldNamesOpts(), fwflex.WithFieldNamePrefix("Scope"))...))
		if resp.Diagnostics.HasError() {
			return
		}

		// An omitted description leaves the current one in place, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.ScopeDescription = aws.String("")
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// IsPublished is required on every update, not only when it changes.
		input.IsPublished = fwflex.BoolFromFramework(ctx, plan.IsPublished)
		input.ScopeIdentifier = aws.String(scopeIdentifier(arn, current.Status))
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdateScope(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, (*networksecuritymanager.GetScopeOutput)(out), &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *scopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scopeResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	// A published scope with a pending draft cannot be deleted until the draft is,
	// and a scope that was never published exists only as a draft.
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

// findScopeByARN returns the version of the scope that an update acts on:
// the pending draft when there is one, otherwise the published scope.
func findScopeByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetScopeOutput, error) {
	out, err := findScopeByIdentifier(ctx, conn, arn+scopeDraftQualifier)
	if retry.NotFound(err) {
		return findScopeByIdentifier(ctx, conn, arn)
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func findScopeByIdentifier(ctx context.Context, conn *networksecuritymanager.Client, identifier string) (*networksecuritymanager.GetScopeOutput, error) {
	input := networksecuritymanager.GetScopeInput{
		ScopeIdentifier: aws.String(identifier),
	}

	out, err := conn.GetScope(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

// scopeBaseARN strips the DRAFT qualifier, so that the scope is always identified
// by the ARN it keeps across publication.
func scopeBaseARN(arn string) string {
	return strings.TrimSuffix(arn, scopeDraftQualifier)
}

func scopeIdentifier(arn string, status awstypes.EntityStatus) string {
	if status == awstypes.EntityStatusDraft {
		return arn + scopeDraftQualifier
	}

	return arn
}

func hasGlobalResourceType(resourceScopes []resourceScopeModel) bool {
	return slices.ContainsFunc(resourceScopes, func(v resourceScopeModel) bool {
		return v.ResourceType.ValueEnum() == awstypes.ScopeResourceTypeCfd
	})
}

// exactlyOneOfValueAttributesValidator reads the named attributes from the object's own value.
// Path-based validators such as tfobjectvalidator.ExactlyOneOfChildren find no siblings below
// resource_scope and always fail: the validator path identifies the set element by an
// fwtypes.ObjectValueOf value, which never equals the basetypes.ObjectValue that path matching
// builds from the config.
type exactlyOneOfValueAttributesValidator struct {
	names []string
}

func (v exactlyOneOfValueAttributesValidator) Description(context.Context) string {
	return fmt.Sprintf("exactly one of %s must be set", strings.Join(v.names, ", "))
}

func (v exactlyOneOfValueAttributesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v exactlyOneOfValueAttributesValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var count int
	for _, name := range v.names {
		value := req.ConfigValue.Attributes()[name]
		if value.IsUnknown() {
			return
		}
		if !value.IsNull() {
			count++
		}
	}

	if count != 1 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(req.Path, v.Description(ctx)))
	}
}

func (r *scopeResource) flatten(ctx context.Context, out *networksecuritymanager.GetScopeOutput, data *scopeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	description := data.Description
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("Scope")))
	if diags.HasError() {
		return diags
	}
	data.Description = normalizeDescription(description, data.Description)
	data.ARN = fwflex.StringValueToFramework(ctx, scopeBaseARN(aws.ToString(out.ScopeArn)))
	data.IsPublished = types.BoolValue(out.Status != awstypes.EntityStatusDraft)

	return diags
}

type scopeResourceModel struct {
	framework.WithRegionModel
	ARN                types.String                                             `tfsdk:"arn"`
	Description        types.String                                             `tfsdk:"description"`
	IsPublished        types.Bool                                               `tfsdk:"is_published"`
	Name               types.String                                             `tfsdk:"name"`
	ScopeConfiguration fwtypes.ListNestedObjectValueOf[scopeConfigurationModel] `tfsdk:"scope_configuration"`
	Tags               tftags.Map                                               `tfsdk:"tags"`
	TagsAll            tftags.Map                                               `tfsdk:"tags_all"`
}

type scopeConfigurationModel struct {
	AccountFilter fwtypes.ListNestedObjectValueOf[accountFilterModel] `tfsdk:"account_filter"`
	ResourceScope fwtypes.SetNestedObjectValueOf[resourceScopeModel]  `tfsdk:"resource_scope"`
}

var (
	_ fwflex.Expander  = scopeConfigurationModel{}
	_ fwflex.Flattener = &scopeConfigurationModel{}
)

// The service keys resource scopes by resource type, which is an attribute of each resource_scope block.
func (m scopeConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var apiObject awstypes.ScopeConfiguration

	diags.Append(fwflex.Expand(ctx, m.AccountFilter, &apiObject.AccountFilter)...)
	if diags.HasError() {
		return nil, diags
	}

	resourceScopes, d := m.ResourceScope.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	apiObject.ResourceScopes = make(map[string]awstypes.ResourceScope, len(resourceScopes))
	for _, v := range resourceScopes {
		var resourceScope awstypes.ResourceScope
		diags.Append(fwflex.Expand(ctx, v, &resourceScope)...)
		if diags.HasError() {
			return nil, diags
		}
		apiObject.ResourceScopes[v.ResourceType.ValueString()] = resourceScope
	}

	return &apiObject, diags
}

func (m *scopeConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	apiObject, ok := v.(awstypes.ScopeConfiguration)
	if !ok {
		return diags
	}

	if apiObject.AccountFilter != nil {
		diags.Append(fwflex.Flatten(ctx, apiObject.AccountFilter, &m.AccountFilter)...)
		if diags.HasError() {
			return diags
		}
	}

	resourceScopes := make([]*resourceScopeModel, 0, len(apiObject.ResourceScopes))
	for resourceType, resourceScope := range apiObject.ResourceScopes {
		var data resourceScopeModel
		diags.Append(fwflex.Flatten(ctx, resourceScope, &data)...)
		if diags.HasError() {
			return diags
		}
		data.ResourceType = fwtypes.StringEnumValue(awstypes.ScopeResourceType(resourceType))
		// Only a true include_all selects anything; keep false and unset equivalent.
		if !data.IncludeAll.ValueBool() {
			data.IncludeAll = types.BoolNull()
		}
		resourceScopes = append(resourceScopes, &data)
	}
	m.ResourceScope = fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, resourceScopes)

	return diags
}

type accountFilterModel struct {
	Exclude    fwtypes.ListNestedObjectValueOf[accountSetModel] `tfsdk:"exclude"`
	Include    fwtypes.ListNestedObjectValueOf[accountSetModel] `tfsdk:"include"`
	IncludeAll types.Bool                                       `tfsdk:"include_all"`
}

var (
	_ fwflex.Expander  = accountFilterModel{}
	_ fwflex.Flattener = &accountFilterModel{}
)

func (m accountFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case m.IncludeAll.ValueBool():
		return &awstypes.AccountFilterMemberIncludeAll{}, diags
	case m.Include.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.AccountFilterMemberInclude
		diags.Append(fwflex.Expand(ctx, m.Include, &apiObject.Value)...)
		return &apiObject, diags
	case m.Exclude.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.AccountFilterMemberExclude
		diags.Append(fwflex.Expand(ctx, m.Exclude, &apiObject.Value)...)
		return &apiObject, diags
	}

	return nil, diags
}

func (m *accountFilterModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.AccountFilterMemberExclude:
		diags.Append(fwflex.Flatten(ctx, &t.Value, &m.Exclude)...)
	case awstypes.AccountFilterMemberInclude:
		diags.Append(fwflex.Flatten(ctx, &t.Value, &m.Include)...)
	case awstypes.AccountFilterMemberIncludeAll:
		m.IncludeAll = types.BoolValue(true)
	}

	return diags
}

type accountSetModel struct {
	AccountIDs          fwtypes.SetOfString `tfsdk:"account_ids"`
	OrganizationalUnits fwtypes.SetOfString `tfsdk:"organizational_units"`
}

type resourceScopeModel struct {
	Exclude      fwtypes.ListNestedObjectValueOf[resourceSetModel] `tfsdk:"exclude"`
	Include      fwtypes.ListNestedObjectValueOf[resourceSetModel] `tfsdk:"include"`
	IncludeAll   types.Bool                                        `tfsdk:"include_all"`
	ResourceType fwtypes.StringEnum[awstypes.ScopeResourceType]    `tfsdk:"resource_type" autoflex:"-"`
}

var _ fwflex.Expander = resourceScopeModel{}

// Exactly one of include_all, include and exclude is sent, so unset blocks must stay nil.
func (m resourceScopeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var apiObject awstypes.ResourceScope

	if m.IncludeAll.ValueBool() {
		apiObject.IncludeAll = aws.Bool(true)
	}

	exclude, d := m.Exclude.ToPtr(ctx)
	diags.Append(d...)
	include, d := m.Include.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if exclude != nil {
		apiObject.Exclude = &awstypes.ResourceSet{}
		diags.Append(fwflex.Expand(ctx, exclude, apiObject.Exclude)...)
	}
	if include != nil {
		apiObject.Include = &awstypes.ResourceSet{}
		diags.Append(fwflex.Expand(ctx, include, apiObject.Include)...)
	}
	if diags.HasError() {
		return nil, diags
	}

	return &apiObject, diags
}

type resourceSetModel struct {
	ExplicitARNs fwtypes.SetOfString                              `tfsdk:"explicit_arns" autoflex:",omitempty"`
	Expression   fwtypes.ListNestedObjectValueOf[expressionModel] `tfsdk:"expression"`
}

// The operands of and, not and or are criteria: the service rejects an operator nested under another operator.
type expressionModel struct {
	And      fwtypes.ListNestedObjectValueOf[criteriaModel] `tfsdk:"and"`
	Criteria fwtypes.ListNestedObjectValueOf[criteriaModel] `tfsdk:"criteria"`
	Not      fwtypes.ListNestedObjectValueOf[criteriaModel] `tfsdk:"not"`
	Or       fwtypes.ListNestedObjectValueOf[criteriaModel] `tfsdk:"or"`
}

var (
	_ fwflex.Expander  = expressionModel{}
	_ fwflex.Flattener = &expressionModel{}
)

func (m expressionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case m.And.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.ResourceLogicalExpressionMemberAnd
		diags.Append(fwflex.Expand(ctx, m.And, &apiObject.Value)...)
		return &apiObject, diags
	case m.Criteria.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.ResourceLogicalExpressionMemberCriteria
		diags.Append(fwflex.Expand(ctx, m.Criteria, &apiObject.Value)...)
		return &apiObject, diags
	case m.Not.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.ResourceLogicalExpressionMemberNot
		diags.Append(fwflex.Expand(ctx, m.Not, &apiObject.Value)...)
		return &apiObject, diags
	case m.Or.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.ResourceLogicalExpressionMemberOr
		diags.Append(fwflex.Expand(ctx, m.Or, &apiObject.Value)...)
		return &apiObject, diags
	}

	return nil, diags
}

func (m *expressionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ResourceLogicalExpressionMemberAnd:
		diags.Append(fwflex.Flatten(ctx, t.Value, &m.And)...)
	case awstypes.ResourceLogicalExpressionMemberCriteria:
		diags.Append(fwflex.Flatten(ctx, t.Value, &m.Criteria)...)
	case awstypes.ResourceLogicalExpressionMemberNot:
		diags.Append(fwflex.Flatten(ctx, t.Value, &m.Not)...)
	case awstypes.ResourceLogicalExpressionMemberOr:
		diags.Append(fwflex.Flatten(ctx, t.Value, &m.Or)...)
	}

	return diags
}

type criteriaModel struct {
	ALBConfig fwtypes.ListNestedObjectValueOf[albConfigModel] `tfsdk:"alb_config"`
	Tags      fwtypes.MapOfString                             `tfsdk:"tags"`
}

var (
	_ fwflex.TypedExpander = criteriaModel{}
	_ fwflex.Flattener     = &criteriaModel{}
)

// A criteria is expanded on its own (expression.criteria) or as the operand of an operator (and, not, or).
func (m criteriaModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var criteria awstypes.ResourceCriteria
	switch {
	case m.ALBConfig.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0:
		var apiObject awstypes.ResourceCriteriaMemberAlbConfig
		diags.Append(fwflex.Expand(ctx, m.ALBConfig, &apiObject.Value)...)
		criteria = &apiObject
	case !m.Tags.IsNull():
		// An empty tag map is a valid criteria that matches every resource.
		criteria = &awstypes.ResourceCriteriaMemberTags{Value: fwflex.ExpandFrameworkStringValueMap(ctx, m.Tags)}
	default:
		return nil, diags
	}

	switch targetType {
	case reflect.TypeFor[awstypes.ResourceCriteria]():
		return criteria, diags
	case reflect.TypeFor[awstypes.ResourceLogicalExpression]():
		return &awstypes.ResourceLogicalExpressionMemberCriteria{Value: criteria}, diags
	}

	return nil, diags
}

func (m *criteriaModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ResourceCriteriaMemberAlbConfig:
		diags.Append(fwflex.Flatten(ctx, &t.Value, &m.ALBConfig)...)
	case awstypes.ResourceCriteriaMemberTags:
		diags.Append(fwflex.Flatten(ctx, t.Value, &m.Tags)...)
	case awstypes.ResourceLogicalExpressionMemberCriteria:
		// The operand of an operator.
		diags.Append(fwflex.Flatten(ctx, t.Value, m)...)
	}

	return diags
}

type albConfigModel struct {
	IPAddressType fwtypes.StringEnum[awstypes.IpAddressType] `tfsdk:"ip_address_type"`
	Scheme        fwtypes.StringEnum[awstypes.Scheme]        `tfsdk:"scheme"`
}
