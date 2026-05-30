// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_catalog", name="Catalog")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="name")
// @Testing(serialize=true)
func newCatalogResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &catalogResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type catalogResource struct {
	framework.ResourceWithModel[catalogResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *catalogResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("catalog_properties").AtListIndex(0).AtName("data_lake_access_properties"),
			path.MatchRoot("federated_catalog"),
			path.MatchRoot("target_redshift_catalog"),
		),
	}
}

func (r *catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_full_table_external_data_access": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.AllowFullTableExternalDataAccessEnum](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCatalogID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreateTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"overwrite_child_resource_permissions_with_default": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.OverwriteChildResourcePermissionsWithDefaultEnum](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrParameters: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"update_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"catalog_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[catalogPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"custom_properties": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Map{
								mapplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"data_lake_access_properties": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeAccessPropertiesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"catalog_type": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"data_lake_access": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"data_transfer_role": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrKMSKey: schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"managed_workgroup_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"managed_workgroup_status": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"redshift_database_name": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrStatusMessage: schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"iceberg_optimization_properties": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[icebergOptimizationPropertiesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"compaction": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									"orphan_file_deletion": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									"retention": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
			"create_database_default_permissions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[principalPermissionsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPermissions: schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrPrincipal: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakePrincipalModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_lake_principal_identifier": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"create_table_default_permissions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[principalPermissionsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPermissions: schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrPrincipal: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakePrincipalModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_lake_principal_identifier": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"federated_catalog": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[federatedCatalogModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"connection_name": schema.StringAttribute{
							Optional: true,
						},
						"connection_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrIdentifier: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"target_redshift_catalog": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetRedshiftCatalogModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"catalog_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GlueClient(ctx)

	var plan catalogResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	catalogPropertiesWasNull := plan.CatalogProperties.IsNull()

	var catalogInput awstypes.CatalogInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &catalogInput))
	if resp.Diagnostics.HasError() {
		return
	}
	// API contract requires explicit empty slices rather than nil for these.
	if catalogInput.CreateDatabaseDefaultPermissions == nil {
		catalogInput.CreateDatabaseDefaultPermissions = []awstypes.PrincipalPermissions{}
	}
	if catalogInput.CreateTableDefaultPermissions == nil {
		catalogInput.CreateTableDefaultPermissions = []awstypes.PrincipalPermissions{}
	}
	catalogInput.OverwriteChildResourcePermissionsWithDefault = plan.OverwriteChildResourcePermissionsWithDefault.ValueEnum()

	input := &glue.CreateCatalogInput{
		Name:         plan.Name.ValueStringPointer(),
		CatalogInput: &catalogInput,
		Tags:         getTagsIn(ctx),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.FederationSourceException](
		ctx, iamPropagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateCatalog(ctx, input)
		},
		"Invalid role provided",
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	out, err := waitCatalogReady(ctx, conn, plan.Name.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	readCatalogIntoModel(ctx, out, &plan, catalogPropertiesWasNull, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().GlueClient(ctx)

	var state catalogResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	catalogPropertiesWasNull := state.CatalogProperties.IsNull()

	out, err := findCatalogByName(ctx, conn, state.Name.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
		return
	}

	readCatalogIntoModel(ctx, out, &state, catalogPropertiesWasNull, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, r.Meta().GlueClient(ctx), state.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *catalogResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Not applicable on create or destroy.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state catalogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Federated catalogs do not support UpdateCatalog — force replacement
	// when any catalog attribute changes.
	if !plan.FederatedCatalog.IsNull() || !state.FederatedCatalog.IsNull() {
		if !plan.Description.Equal(state.Description) ||
			!plan.Parameters.Equal(state.Parameters) ||
			!plan.CatalogProperties.Equal(state.CatalogProperties) ||
			!plan.FederatedCatalog.Equal(state.FederatedCatalog) ||
			!plan.AllowFullTableExternalDataAccess.Equal(state.AllowFullTableExternalDataAccess) ||
			!plan.CreateDatabaseDefaultPermissions.Equal(state.CreateDatabaseDefaultPermissions) ||
			!plan.CreateTableDefaultPermissions.Equal(state.CreateTableDefaultPermissions) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("federated_catalog"))
		}
	}
}

func (r *catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().GlueClient(ctx)

	var plan, state catalogResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	catalogPropertiesWasNull := plan.CatalogProperties.IsNull()

	if !plan.Description.Equal(state.Description) ||
		!plan.Parameters.Equal(state.Parameters) ||
		!plan.CatalogProperties.Equal(state.CatalogProperties) ||
		!plan.FederatedCatalog.Equal(state.FederatedCatalog) ||
		!plan.TargetRedshiftCatalog.Equal(state.TargetRedshiftCatalog) ||
		!plan.AllowFullTableExternalDataAccess.Equal(state.AllowFullTableExternalDataAccess) ||
		!plan.CreateDatabaseDefaultPermissions.Equal(state.CreateDatabaseDefaultPermissions) ||
		!plan.CreateTableDefaultPermissions.Equal(state.CreateTableDefaultPermissions) {
		var catalogInput awstypes.CatalogInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &catalogInput))
		if resp.Diagnostics.HasError() {
			return
		}
		if catalogInput.CreateDatabaseDefaultPermissions == nil {
			catalogInput.CreateDatabaseDefaultPermissions = []awstypes.PrincipalPermissions{}
		}
		if catalogInput.CreateTableDefaultPermissions == nil {
			catalogInput.CreateTableDefaultPermissions = []awstypes.PrincipalPermissions{}
		}
		catalogInput.OverwriteChildResourcePermissionsWithDefault = plan.OverwriteChildResourcePermissionsWithDefault.ValueEnum()

		input := &glue.UpdateCatalogInput{
			CatalogId:    plan.Name.ValueStringPointer(),
			CatalogInput: &catalogInput,
		}

		_, err := conn.UpdateCatalog(ctx, input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
			return
		}
	}

	out, err := findCatalogByName(ctx, conn, plan.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	readCatalogIntoModel(ctx, out, &plan, catalogPropertiesWasNull, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle tags update
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, r.Meta().GlueClient(ctx), state.ARN.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().GlueClient(ctx)

	var state catalogResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Retry on ConcurrentModificationException: Redshift-managed catalogs
	// (catalog_type = "aws:redshift") delegate to the Redshift workgroup,
	// which rejects deletes while another operation is still running.
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConcurrentModificationException](
		ctx,
		r.DeleteTimeout(ctx, state.Timeouts),
		func(ctx context.Context) (any, error) {
			return conn.DeleteCatalog(ctx, &glue.DeleteCatalogInput{
				CatalogId: state.Name.ValueStringPointer(),
			})
		},
	)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
	}
}

func findCatalogByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.Catalog, error) {
	input := &glue.GetCatalogInput{
		CatalogId: aws.String(name),
	}

	out, err := conn.GetCatalog(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		// Lake Formation returns AccessDeniedException when catalog doesn't exist
		// and caller lacks Lake Formation permissions
		if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Lake Formation permission") {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Catalog == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Catalog, nil
}

// Managed workgroup status observed for catalogs created with
// DataLakeAccessProperties (RMS / "aws:redshift"). Redshift provisions the
// backing workgroup asynchronously and the status transitions through
// undocumented values (observed: CREATING, MODIFYING) before reaching
// AVAILABLE. Empty string means no managed workgroup (catalogs without
// data_lake_access_properties), which is also terminal.
const (
	iamPropagationTimeout           = 2 * time.Minute
	managedWorkgroupStatusAvailable = "AVAILABLE"
)

// waitCatalogReady polls GetCatalog until the catalog's managed workgroup is
// AVAILABLE. For catalogs without data_lake_access_properties, no managed
// workgroup exists and the first read returns immediately. Any non-AVAILABLE
// non-empty status is treated as pending — AWS does not publish the enum, so
// we can't enumerate all transitional states and must accept anything that
// isn't the terminal value.
func waitCatalogReady(ctx context.Context, conn *glue.Client, id string, timeout time.Duration) (*awstypes.Catalog, error) {
	var catalog *awstypes.Catalog
	err := tfresource.Retry(ctx, timeout, func(ctx context.Context) *tfresource.RetryError {
		c, err := findCatalogByName(ctx, conn, id)
		if err != nil {
			return tfresource.NonRetryableError(err)
		}
		catalog = c

		if c.CatalogProperties == nil || c.CatalogProperties.DataLakeAccessProperties == nil {
			return nil
		}
		status := aws.ToString(c.CatalogProperties.DataLakeAccessProperties.ManagedWorkgroupStatus)
		if status == "" || status == managedWorkgroupStatusAvailable {
			return nil
		}
		return tfresource.RetryableError(smarterr.NewError(&retry.NotFoundError{
			Message: "managed workgroup status " + status,
		}))
	}, tfresource.WithPollInterval(10*time.Second))

	return catalog, err
}

// readCatalogIntoModel copies AWS fields onto the resource model.
// catalog_properties is only populated when the user declared it (model is non-null)
// or when the API returned real user-configured data (data_lake_access_properties),
// preventing a permanent diff when AWS auto-populates custom_properties on
// LF-managed catalogs. ARN is set manually because flex does not rename
// ResourceArn -> ARN, and ID mirrors CatalogId.
func readCatalogIntoModel(ctx context.Context, catalog *awstypes.Catalog, model *catalogResourceModel, _ bool, diags *diag.Diagnostics) {
	// Flatten everything except catalog_properties first.
	savedCatalogProperties := model.CatalogProperties
	diags.Append(fwflex.Flatten(ctx, catalog, model)...)
	if diags.HasError() {
		return
	}
	model.ARN = types.StringPointerValue(catalog.ResourceArn)
	// Only populate catalog_properties when the user declared it or the API
	// returned real user-configured data; suppress AWS-auto-populated values
	// (e.g. custom_properties={"aws:PermissionsModel":"LAKEFORMATION"}).
	if savedCatalogProperties.IsNull() && catalog.CatalogProperties != nil && catalog.CatalogProperties.DataLakeAccessProperties == nil {
		model.CatalogProperties = fwtypes.NewListNestedObjectValueOfNull[catalogPropertiesModel](ctx)
	}
}

type catalogResourceModel struct {
	framework.WithRegionModel
	AllowFullTableExternalDataAccess             fwtypes.StringEnum[awstypes.AllowFullTableExternalDataAccessEnum]             `tfsdk:"allow_full_table_external_data_access"`
	ARN                                          types.String                                                                  `tfsdk:"arn" autoflex:"-"`
	CatalogID                                    types.String                                                                  `tfsdk:"catalog_id"`
	CatalogProperties                            fwtypes.ListNestedObjectValueOf[catalogPropertiesModel]                       `tfsdk:"catalog_properties"`
	CreateDatabaseDefaultPermissions             fwtypes.ListNestedObjectValueOf[principalPermissionsModel]                    `tfsdk:"create_database_default_permissions"`
	CreateTableDefaultPermissions                fwtypes.ListNestedObjectValueOf[principalPermissionsModel]                    `tfsdk:"create_table_default_permissions"`
	CreateTime                                   timetypes.RFC3339                                                             `tfsdk:"create_time"`
	Description                                  types.String                                                                  `tfsdk:"description"`
	FederatedCatalog                             fwtypes.ListNestedObjectValueOf[federatedCatalogModel]                        `tfsdk:"federated_catalog"`
	Name                                         types.String                                                                  `tfsdk:"name"`
	OverwriteChildResourcePermissionsWithDefault fwtypes.StringEnum[awstypes.OverwriteChildResourcePermissionsWithDefaultEnum] `tfsdk:"overwrite_child_resource_permissions_with_default" autoflex:"-"`
	Parameters                                   fwtypes.MapOfString                                                           `tfsdk:"parameters"`
	Tags                                         tftags.Map                                                                    `tfsdk:"tags"`
	TagsAll                                      tftags.Map                                                                    `tfsdk:"tags_all"`
	TargetRedshiftCatalog                        fwtypes.ListNestedObjectValueOf[targetRedshiftCatalogModel]                   `tfsdk:"target_redshift_catalog"`
	Timeouts                                     timeouts.Value                                                                `tfsdk:"timeouts"`
	UpdateTime                                   timetypes.RFC3339                                                             `tfsdk:"update_time"`
}

type catalogPropertiesModel struct {
	CustomProperties              fwtypes.MapOfString                                                 `tfsdk:"custom_properties"`
	DataLakeAccessProperties      fwtypes.ListNestedObjectValueOf[dataLakeAccessPropertiesModel]      `tfsdk:"data_lake_access_properties"`
	IcebergOptimizationProperties fwtypes.ListNestedObjectValueOf[icebergOptimizationPropertiesModel] `tfsdk:"iceberg_optimization_properties"`
}

type dataLakeAccessPropertiesModel struct {
	CatalogType            types.String `tfsdk:"catalog_type"`
	DataLakeAccess         types.Bool   `tfsdk:"data_lake_access"`
	DataTransferRole       fwtypes.ARN  `tfsdk:"data_transfer_role"`
	KmsKey                 types.String `tfsdk:"kms_key"`
	ManagedWorkgroupName   types.String `tfsdk:"managed_workgroup_name"`
	ManagedWorkgroupStatus types.String `tfsdk:"managed_workgroup_status"`
	RedshiftDatabaseName   types.String `tfsdk:"redshift_database_name"`
	StatusMessage          types.String `tfsdk:"status_message"`
}

type icebergOptimizationPropertiesModel struct {
	Compaction         fwtypes.MapValueOf[types.String] `tfsdk:"compaction" autoflex:",omitempty"`
	OrphanFileDeletion fwtypes.MapValueOf[types.String] `tfsdk:"orphan_file_deletion" autoflex:",omitempty"`
	Retention          fwtypes.MapValueOf[types.String] `tfsdk:"retention" autoflex:",omitempty"`
	RoleArn            fwtypes.ARN                      `tfsdk:"role_arn" autoflex:",omitempty"`
}

type federatedCatalogModel struct {
	ConnectionName types.String `tfsdk:"connection_name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Identifier     types.String `tfsdk:"identifier"`
}

type targetRedshiftCatalogModel struct {
	CatalogArn fwtypes.ARN `tfsdk:"catalog_arn"`
}

type principalPermissionsModel struct {
	Permissions fwtypes.ListOfString                                    `tfsdk:"permissions"`
	Principal   fwtypes.ListNestedObjectValueOf[dataLakePrincipalModel] `tfsdk:"principal"`
}

type dataLakePrincipalModel struct {
	DataLakePrincipalIdentifier types.String `tfsdk:"data_lake_principal_identifier"`
}
