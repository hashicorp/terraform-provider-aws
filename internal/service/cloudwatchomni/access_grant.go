// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"iter"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatchomni_access_grant", name="Access Grant")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("grant_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.AccessGrant")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="grant_id")
func newAccessGrantResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantResource{}

	return r, nil
}

// There is no UpdateAccessGrant operation, so every argument except tags forces replacement.
type accessGrantResource struct {
	framework.ResourceWithModel[accessGrantResourceModel]
	framework.WithImportByIdentity
}

func (r *accessGrantResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"domain_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^d-[0-9a-z]{1,25}$`), "must be a CloudWatch Omni domain ID"),
				},
			},
			"grant_id": framework.IDAttribute(),
			"grant_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessGrantType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]+$`), "use letters, numbers, underscores, and hyphens."),
				},
			},
			"permission": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessGrantPermission](),
				Required:   true,
				Validators: []validator.String{
					tfstringvalidator.AlsoRequiresWhenEquals(
						awstypes.AccessGrantPermissionCustom,
						path.MatchRelative().AtParent().AtName("scoped_action"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), "must be a CloudWatch Omni space ID"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrPrincipal: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[accessGrantPrincipalModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.AtLeastOneOfChildren(
							path.MatchRelative().AtName("principal_attribute"),
							path.MatchRelative().AtName("principal_id"),
						),
					},
					Attributes: map[string]schema.Attribute{
						"principal_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9:/_+=,.@-]*$`), "start with a letter or number and use letters, numbers, and the characters :/_+=,.@-"),
							},
						},
						"principal_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AccessGrantPrincipalType](),
							Required:   true,
							Validators: []validator.String{
								tfstringvalidator.ConflictsWithWhenNotEquals(
									awstypes.AccessGrantPrincipalTypeIdcUser,
									path.MatchRelative().AtParent().AtName("principal_attribute"),
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"principal_attribute": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[accessGrantPrincipalAttributeModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(10),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
								},
							},
						},
					},
				},
			},
			"scoped_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scopedActionsModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(20),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrActions: schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 50),
								listvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(3, 256),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9-]*:[A-Za-z][A-Za-z0-9]*$`), `use the form "vendor-code:Action", without wildcards.`),
								),
							},
						},
						// ContextConditions are not currently working server-side.
						// "context_conditions": schema.MapAttribute{
						// 	CustomType: fwtypes.MapOfListOfStringType,
						// 	Optional:   true,
						// 	Validators: []validator.Map{
						// 		mapvalidator.SizeAtMost(10),
						// 	},
						// },
					},
					Blocks: map[string]schema.Block{
						"resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceScopeModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(10),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.RequiredTogetherChildren(
										path.MatchRelative().AtParent().AtAnyListIndex().AtName("row_scope_group"),
										path.MatchRelative().AtParent().AtAnyListIndex().AtName("signal_types"),
									),
								},
								Attributes: map[string]schema.Attribute{
									"resource_arns": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.List{
											listvalidator.SizeAtMost(5),
										},
									},
									names.AttrResourceType: schema.StringAttribute{
										Required: true,
									},
									"signal_types": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringEnumType[awstypes.SignalType](),
										Optional:   true,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 3),
											listvalidator.UniqueValues(),
										},
									},
									names.AttrTags: schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.Map{
											mapvalidator.SizeAtMost(3),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"row_scope_group": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[rowScopeGroupModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(20),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"row_scope": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[rowScopeModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeBetween(1, 20),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrField: schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 256),
																},
															},
															"operator": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.RowScopeOperator](),
																Required:   true,
															},
															// The service sorts and deduplicates values, so they are a set.
															names.AttrValues: schema.SetAttribute{
																CustomType:  fwtypes.SetOfStringType,
																ElementType: types.StringType,
																Required:    true,
																Validators: []validator.Set{
																	setvalidator.SizeBetween(1, 100),
																	setvalidator.ValueStringsAre(
																		stringvalidator.LengthBetween(1, 256),
																	),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *accessGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudwatchomni.CreateAccessGrantInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	// CreateAccessGrant takes a tag map, but the tag operations live in the cloudwatch package and use a slice.
	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	// A newly created IAM principal can take a while to become visible to CloudWatch Omni.
	output, err := tfresource.RetryWhenIsAErrorMessageContains[*cloudwatchomni.CreateAccessGrantOutput, *awstypes.ValidationException](ctx, propagationTimeout,
		func(ctx context.Context) (*cloudwatchomni.CreateAccessGrantOutput, error) {
			return conn.CreateAccessGrant(ctx, &input)
		},
		"Principal does not exist",
	)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	// The CreateAccessGrant response omits the grant ARN, so read the grant back.
	grantID := aws.ToString(output.AccessGrant.GrantId)
	grant, err := findAccessGrantByID(ctx, conn, grantID)
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("grant_id"), grantID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, grantID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, grant, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *accessGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	grantID := fwflex.StringValueFromFramework(ctx, data.GrantID)
	output, err := findAccessGrantByID(ctx, conn, grantID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, grantID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *accessGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	grantID := fwflex.StringValueFromFramework(ctx, data.GrantID)
	input := cloudwatchomni.DeleteAccessGrantInput{
		GrantId: aws.String(grantID),
	}
	_, err := conn.DeleteAccessGrant(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, grantID)
		return
	}
}

func (r *accessGrantResource) flatten(ctx context.Context, grant *awstypes.AccessGrant, data *accessGrantResourceModel) diag.Diagnostics {
	// Unset collections are returned empty rather than absent; treat them as absent.
	if grant.Principal != nil && len(grant.Principal.PrincipalAttributes) == 0 {
		grant.Principal.PrincipalAttributes = nil
	}
	if len(grant.ScopedActions) == 0 {
		grant.ScopedActions = nil
	}

	return fwflex.Flatten(ctx, grant, data, fwflex.WithIgnoredFieldNamesAppend("Region"))
}

// findAccessGrantByID resolves an access grant scoped to the given region.
//
// Access grant IDs are account-wide and GetAccessGrant succeeds from any region,
// so the grant is only treated as present when its space is in this region.
func findAccessGrantByID(ctx context.Context, conn *cloudwatchomni.Client, id string) (*awstypes.AccessGrant, error) {
	input := cloudwatchomni.GetAccessGrantInput{
		GrantId: aws.String(id),
	}
	grant, err := findAccessGrant(ctx, conn, &input)
	if err != nil {
		return nil, err
	}

	spaceID := aws.ToString(grant.SpaceId)
	var inputLS cloudwatchomni.ListSpacesInput
	spaces, err := findSpaces(ctx, conn, &inputLS, tfslices.WithFilter(func(v awstypes.SpaceSummary) bool {
		return aws.ToString(v.SpaceId) == spaceID && aws.ToString(v.Region) == conn.Options().Region
	}))
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if len(spaces) == 0 {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: fmt.Sprintf("CloudWatch Omni Space (%s) not found in Region (%s)", spaceID, conn.Options().Region),
		})
	}

	return grant, nil
}

func findAccessGrant(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetAccessGrantInput) (*awstypes.AccessGrant, error) {
	output, err := conn.GetAccessGrant(ctx, input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case err != nil:
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.AccessGrant == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.AccessGrant, nil
}

func listAccessGrantPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListAccessGrantsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.AccessGrantSummary, error] {
	return func(yield func([]awstypes.AccessGrantSummary, error) bool) {
		pages := cloudwatchomni.NewListAccessGrantsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Access Grants: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}

type accessGrantResourceModel struct {
	framework.WithRegionModel
	DomainID      types.String                                               `tfsdk:"domain_id"`
	GrantARN      types.String                                               `tfsdk:"arn"`
	GrantID       types.String                                               `tfsdk:"grant_id"`
	GrantType     fwtypes.StringEnum[awstypes.AccessGrantType]               `tfsdk:"grant_type"`
	Name          types.String                                               `tfsdk:"name"`
	Permission    fwtypes.StringEnum[awstypes.AccessGrantPermission]         `tfsdk:"permission"`
	Principal     fwtypes.ListNestedObjectValueOf[accessGrantPrincipalModel] `tfsdk:"principal"`
	ScopedActions fwtypes.ListNestedObjectValueOf[scopedActionsModel]        `tfsdk:"scoped_action"`
	SpaceID       types.String                                               `tfsdk:"space_id"`
	Tags          tftags.Map                                                 `tfsdk:"tags"`
	TagsAll       tftags.Map                                                 `tfsdk:"tags_all"`
}

type accessGrantPrincipalModel struct {
	PrincipalAttributes fwtypes.ListNestedObjectValueOf[accessGrantPrincipalAttributeModel] `tfsdk:"principal_attribute"`
	PrincipalID         types.String                                                        `tfsdk:"principal_id"`
	PrincipalType       fwtypes.StringEnum[awstypes.AccessGrantPrincipalType]               `tfsdk:"principal_type"`
}

type accessGrantPrincipalAttributeModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type scopedActionsModel struct {
	Actions fwtypes.ListOfString `tfsdk:"actions"`
	// ContextConditions fwtypes.MapOfListOfString                           `tfsdk:"context_conditions"`
	Resources fwtypes.ListNestedObjectValueOf[resourceScopeModel] `tfsdk:"resource"`
}

type resourceScopeModel struct {
	ResourceARNs   fwtypes.ListOfString                                `tfsdk:"resource_arns"`
	ResourceType   types.String                                        `tfsdk:"resource_type"`
	RowScopeGroups fwtypes.ListNestedObjectValueOf[rowScopeGroupModel] `tfsdk:"row_scope_group" autoflex:"-"`
	SignalTypes    fwtypes.ListOfStringEnum[awstypes.SignalType]       `tfsdk:"signal_types"`
	Tags           fwtypes.MapOfString                                 `tfsdk:"tags"`
}

var (
	_ fwflex.Expander  = resourceScopeModel{}
	_ fwflex.Flattener = &resourceScopeModel{}
)

// Expand converts row_scope_group blocks, which AutoFlex cannot map onto the
// API's list-of-lists RowScopeGroups.
func (m resourceScopeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var r awstypes.ResourceScope
	// To prevent infinite recursion...
	type modelAlias resourceScopeModel
	alias := modelAlias(m)
	// No ignored field names to handle Tags.
	diags.Append(fwflex.Expand(ctx, alias, &r, fwflex.WithNoIgnoredFieldNames())...)
	if diags.HasError() {
		return nil, diags
	}

	groups, d := m.RowScopeGroups.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	for _, group := range groups {
		var rowScopes []awstypes.RowScope
		diags.Append(fwflex.Expand(ctx, group.RowScopes, &rowScopes)...)
		if diags.HasError() {
			return nil, diags
		}
		r.RowScopeGroups = append(r.RowScopeGroups, rowScopes)
	}

	return &r, diags
}

func (m *resourceScopeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// To prevent infinite recursion...
	type modelAlias *resourceScopeModel
	alias := modelAlias(m)

	apiObject := v.(awstypes.ResourceScope)
	// Unset collections are returned empty rather than absent; treat them as absent.
	if len(apiObject.ResourceArns) == 0 {
		apiObject.ResourceArns = nil
	}
	if len(apiObject.Tags) == 0 {
		apiObject.Tags = nil
	}

	// No ignored field names to handle Tags.
	diags.Append(fwflex.Flatten(ctx, apiObject, alias, fwflex.WithNoIgnoredFieldNames())...)
	if diags.HasError() {
		return diags
	}

	if len(apiObject.RowScopeGroups) == 0 {
		m.RowScopeGroups = fwtypes.NewListNestedObjectValueOfNull[rowScopeGroupModel](ctx)
		return diags
	}

	groups := make([]*rowScopeGroupModel, 0, len(apiObject.RowScopeGroups))
	for _, rowScopes := range apiObject.RowScopeGroups {
		var group rowScopeGroupModel
		diags.Append(fwflex.Flatten(ctx, rowScopes, &group.RowScopes)...)
		if diags.HasError() {
			return diags
		}
		groups = append(groups, &group)
	}

	rowScopeGroups, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, groups, nil)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.RowScopeGroups = rowScopeGroups

	return diags
}

type rowScopeGroupModel struct {
	RowScopes fwtypes.ListNestedObjectValueOf[rowScopeModel] `tfsdk:"row_scope"`
}

type rowScopeModel struct {
	Field    types.String                                  `tfsdk:"field"`
	Operator fwtypes.StringEnum[awstypes.RowScopeOperator] `tfsdk:"operator"`
	Values   fwtypes.SetOfString                           `tfsdk:"values"`
}
