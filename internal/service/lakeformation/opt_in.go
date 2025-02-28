// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lakeformation_opt_in", name="Opt In")
func newResourceOptIn(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOptIn{}

	// r.SetDefaultCreateTimeout(30 * time.Minute)
	// r.SetDefaultUpdateTimeout(30 * time.Minute)
	// r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOptIn = "Opt In"
)

type resourceOptIn struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpUpdate[resourceOptIn]
}

func (r *resourceOptIn) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated_by": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_modified": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[Condition](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"expression": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"principal": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[DataLakePrincipal](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_lake_principal_identifier": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"catalog": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[Catalog](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"catalog_id": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"data_cells_filter": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[DataCellsFilter](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							Optional: true,
						},
						"name": schema.StringAttribute{
							Optional: true,
						},
						"table_catalog_id": schema.StringAttribute{
							Optional: true,
						},
						"table_name": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"data_location": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[DataLocation](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrResourceARN: schema.StringAttribute{
							Required: true,
						},
						"catalog_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			names.AttrDatabase: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[Database](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional(),
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"lf_tag": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[LFTag](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptionalComputed(),
						names.AttrKey: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
								stringvalidator.RegexMatches(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:\*\/=+\-@%]*)$`), ""),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"lf_tag_expression": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional(),
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"lf_tag_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[LFTagPolicy](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"resource_type": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ResourceType](),
						},
						names.AttrCatalogID: catalogIDSchemaOptionalComputed(),
						names.AttrExpression: schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"expression_name": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"table": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[table](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional(),
						names.AttrDatabaseName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrName),
									path.MatchRelative().AtParent().AtName("wildcard"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"wildcard": schema.BoolAttribute{
							Optional: true,
							Validators: []validator.Bool{
								boolvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrName),
									path.MatchRelative().AtParent().AtName("wildcard"),
								),
							},
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"table_with_columns": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tableWithColumns](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCatalogID: catalogIDSchemaOptional(),
						"column_names": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
							Validators: []validator.Set{
								setvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName("column_names"),
									path.MatchRelative().AtParent().AtName("column_wildcard"),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						names.AttrDatabaseName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"column_wildcard": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[columnWildcardData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName("column_names"),
									path.MatchRelative().AtParent().AtName("column_wildcard"),
								),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"excluded_column_names": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},

			// "timeouts": timeouts.Block(ctx, timeouts.Opts{
			// 	Create: true,
			// 	Update: true,
			// 	Delete: true,
			// }),
		},
	}
}

func (r *resourceOptIn) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan resourceOptInData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := lakeformation.CreateLakeFormationOptInInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateLakeFormationOptIn(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameOptIn, plan.Principal.String(), err),
			err.Error(),
		)
		return
	}

	if output == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameOptIn, plan.Principal.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceOptIn) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceOptInData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	optinResource, diags := state.Resource.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	optin := newOptInResourcer(optinResource, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	opinr := optin.expandOptInResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	principalData, diags := state.Principal.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOptInByID(ctx, conn, principalData.DataLakePrincipalIdentifier.ValueString(), opinr)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameOptIn, principalData.DataLakePrincipalIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceOptIn) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceOptInData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lakeformation.DeleteLakeFormationOptInInput{}
	optinResource, diags := state.Resource.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	optin := newOptInResourcer(optinResource, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	principalData, diags := state.Principal.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}


	in.Resource = optin.expandOptInResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// out, err := findOptInByID(ctx, conn, principalData.DataLakePrincipalIdentifier.ValueString(), in.Resource)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameOptIn, principalData.DataLakePrincipalIdentifier.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	if _, err := conn.DeleteLakeFormationOptIn(ctx, in); err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameOptIn, principalData.DataLakePrincipalIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitOptInDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.LakeFormation, create.ErrActionWaitingForDeletion, ResNameOptIn, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
}

func (r *resourceOptIn) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findOptIns(ctx context.Context, conn *lakeformation.Client, input *lakeformation.ListLakeFormationOptInsInput, filter tfslices.Predicate[*awstypes.LakeFormationOptInsInfo]) ([]awstypes.LakeFormationOptInsInfo, error) {
	var output []awstypes.LakeFormationOptInsInfo

	pages := lakeformation.NewListLakeFormationOptInsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.LakeFormationOptInsInfoList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findOptInByID(ctx context.Context, conn *lakeformation.Client, id string, resource *awstypes.Resource) (*awstypes.LakeFormationOptInsInfo, error) {

	in := &lakeformation.ListLakeFormationOptInsInput{}

	in.Resource = resource

	return findOptIn(ctx, conn, in, tfslices.Predicate[*awstypes.LakeFormationOptInsInfo](func(v *awstypes.LakeFormationOptInsInfo) bool {
		return aws.ToString(v.Principal.DataLakePrincipalIdentifier) == id
	}))
}

func findOptIn(ctx context.Context, conn *lakeformation.Client, input *lakeformation.ListLakeFormationOptInsInput, filter tfslices.Predicate[*awstypes.LakeFormationOptInsInfo]) (*awstypes.LakeFormationOptInsInfo, error) {
	output, err := findOptIns(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

type optInResourcer interface {
	expandOptInResource(context.Context, *diag.Diagnostics) *awstypes.Resource
	findOptIn(context.Context, *lakeformation.ListLakeFormationOptInsOutput, *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData]
	// findOptInByAttr(context.Context, *lakeformation.Client, string, string) (*awstypes.LakeFormationOptInsInfo, error)
}

type catalogResource struct {
	data *ResourceData
}

type dbResource struct {
	data *ResourceData
}

type dcfResource struct {
	data *ResourceData
}

type dlResource struct {
	data *ResourceData
}

type lftagResource struct {
	data *ResourceData
}

type lfteResource struct {
	data *ResourceData
}

type lftpResource struct {
	data *ResourceData
}

type tbResource struct {
	data *ResourceData
}

type tbcResource struct {
	data *ResourceData
}

func newOptInResourcer(data *ResourceData, diags *diag.Diagnostics) optInResourcer {
	switch {
	case !data.Catalog.IsNull():
		return &catalogResource{data: data}
	case !data.Database.IsNull():
		return &dbResource{data: data}
	case !data.DataCellsFilter.IsNull():
		return &dcfResource{data: data}
	case !data.DataLocation.IsNull():
		return &dlResource{data: data}
	case !data.LFTag.IsNull():
		return &lftagResource{data: data}
	case !data.LFTagExpression.IsNull():
		return &lfteResource{data: data}
	case !data.LFTagPolicy.IsNull():
		return &lftpResource{data: data}
	case !data.Table.IsNull():
		return &tbResource{data: data}
	case !data.TableWithColumns.IsNull():
		return &tbcResource{data: data}
	default:
		diags.AddError("unexpected resource type",
			"unexpected resource type")
		return nil
	}
}

// //////////////////////// CATALOG //////////////////////////
func (d *catalogResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	catalogptr, err := d.data.Catalog.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var catalog awstypes.CatalogResource
	diags.Append(fwflex.Expand(ctx, catalogptr, &catalog)...)
	if diags.HasError() {
		return nil
	}

	r.Catalog = &catalog
	return &r
}

func (d *catalogResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	catalog, err := d.data.Catalog.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.Catalog != nil {
			if aws.ToString(v.Resource.Catalog.Id) == catalog.ID.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					Catalog: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &Catalog{
						ID: fwflex.StringToFramework(ctx, v.Resource.Catalog.Id),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

////////////////////////////////////////////////////////////

////////////////////////// DATABASE //////////////////////////

func (d *dbResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	dbptr, err := d.data.Database.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var db awstypes.DatabaseResource
	diags.Append(fwflex.Expand(ctx, dbptr, &db)...)
	if diags.HasError() {
		return nil
	}

	r.Database = &db
	return &r
}

func (d *dbResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	db, err := d.data.Database.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.Database != nil {
			if aws.ToString(v.Resource.Database.Name) == db.Name.ValueString() &&
				aws.ToString(v.Resource.Database.CatalogId) == db.CatalogID.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					Database: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &Database{
						Name:      fwflex.StringToFramework(ctx, v.Resource.Database.Name),
						CatalogID: fwflex.StringToFramework(ctx, v.Resource.Database.CatalogId),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

//////////////////////////////////////////////////////////////

// //////////////////DATA_CELLS_FILTER//////////////////////////
func (d *dcfResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	dcfptr, err := d.data.DataCellsFilter.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var dcf awstypes.DataCellsFilterResource
	diags.Append(fwflex.Expand(ctx, dcfptr, &dcf)...)
	if diags.HasError() {
		return nil
	}

	r.DataCellsFilter = &dcf
	return &r
}

func (d *dcfResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	dcf, err := d.data.DataCellsFilter.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.DataCellsFilter != nil {
			if aws.ToString(v.Resource.DataCellsFilter.Name) == dcf.Name.ValueString() &&
				aws.ToString(v.Resource.DataCellsFilter.DatabaseName) == dcf.DatabaseName.ValueString() &&
				aws.ToString(v.Resource.DataCellsFilter.TableCatalogId) == dcf.TableCatalogID.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					DataCellsFilter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &DataCellsFilter{
						Name:           fwflex.StringToFramework(ctx, v.Resource.DataCellsFilter.Name),
						DatabaseName:   fwflex.StringToFramework(ctx, v.Resource.DataCellsFilter.DatabaseName),
						TableCatalogID: fwflex.StringToFramework(ctx, v.Resource.DataCellsFilter.TableCatalogId),
						TableName:      fwflex.StringToFramework(ctx, v.Resource.DataCellsFilter.TableName),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

/////////////////////////////////////////////////////////////////////////////

// /////////////////////DATA_LOCATION////////////////////////////
func (d *dlResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	dlptr, err := d.data.DataLocation.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var dl awstypes.DataLocationResource
	diags.Append(fwflex.Expand(ctx, dlptr, &dl)...)
	if diags.HasError() {
		return nil
	}

	r.DataLocation = &dl
	return &r
}

func (d *dlResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	dl, err := d.data.DataLocation.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.DataLocation != nil {
			if aws.ToString(v.Resource.DataLocation.ResourceArn) == dl.ResourceArn.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					DataLocation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &DataLocation{
						ResourceArn: fwflex.StringToFramework(ctx, v.Resource.DataLocation.ResourceArn),
						CatalogID:   fwflex.StringToFramework(ctx, v.Resource.DataLocation.CatalogId),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

/////////////////////////////////////////////////////////////////////////////////////////////

// //////////////////////// LFTAG ////////////////////////////////////////////////////
func (d *lftagResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	lftagptr, err := d.data.LFTag.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var lftag awstypes.LFTagKeyResource
	diags.Append(fwflex.Expand(ctx, lftagptr, &lftag)...)
	if diags.HasError() {
		return nil
	}

	r.LFTag = &lftag
	return &r
}

func (d *lftagResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	lftag, err := d.data.LFTag.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.LFTag != nil {
			if aws.ToString(v.Resource.LFTag.TagKey) == lftag.Key.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					LFTag: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTag{
						Key: fwflex.StringToFramework(ctx, v.Resource.LFTag.TagKey),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

// ///////////////////////LFTAG EXPRESSION//////////////////////
func (d *lfteResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	lfteptr, err := d.data.LFTagExpression.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var lfte awstypes.LFTagExpressionResource
	diags.Append(fwflex.Expand(ctx, lfteptr, &lfte)...)
	if diags.HasError() {
		return nil
	}

	r.LFTagExpression = &lfte
	return &r
}

func (d *lfteResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	lfte, err := d.data.LFTagExpression.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.LFTagExpression != nil {
			if aws.ToString(v.Resource.LFTagExpression.Name) == lfte.Name.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					LFTagExpression: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTagExpression{
						Name:      fwflex.StringToFramework(ctx, v.Resource.LFTagExpression.Name),
						CatalogID: fwflex.StringToFramework(ctx, v.Resource.LFTagExpression.CatalogId),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

// /////////////////////LFTAG POLICY ////////////////////////////////////
func (d *lftpResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	lftptr, err := d.data.LFTagPolicy.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var lft awstypes.LFTagPolicyResource
	diags.Append(fwflex.Expand(ctx, lftptr, &lft)...)
	if diags.HasError() {
		return nil
	}

	r.LFTagPolicy = &lft
	return &r
}

func (d *lftpResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	lftp, err := d.data.LFTagPolicy.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.LFTagPolicy != nil {
			if aws.ToString((*string)(&v.Resource.LFTagPolicy.ResourceType)) == lftp.ResourceType.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					LFTagPolicy: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTagPolicy{
						ResourceType:   fwtypes.StringEnumValue(v.Resource.LFTagPolicy.ResourceType),
						CatalogID:      fwflex.StringToFramework(ctx, v.Resource.LFTagPolicy.CatalogId),
						ExpressionName: fwflex.StringToFramework(ctx, v.Resource.LFTagPolicy.ExpressionName),
					}),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

// ///////////////////////TABLE////////////////////////////////
func (d *tbResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	tableptr, err := d.data.Table.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var table awstypes.TableResource
	diags.Append(fwflex.Expand(ctx, tableptr, &table)...)
	if diags.HasError() {
		return nil
	}

	r.Table = &table
	return &r
}

func (d *tbResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	tb, err := d.data.Table.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.Table != nil {
			if aws.ToString(v.Resource.Table.Name) == tb.Name.ValueString() &&
				aws.ToString(v.Resource.Table.DatabaseName) == tb.DatabaseName.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					Table: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &table{
						Name: fwflex.StringToFramework(ctx, v.Resource.Table.Name),
					}),
				})
				return out
			}
		}
	}
	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

// //////////////////////////////////////////////////////////////////////////////////////////////
func (d *tbcResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	tbcptr, err := d.data.Table.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var tbc awstypes.TableWithColumnsResource
	diags.Append(fwflex.Expand(ctx, tbcptr, &tbc)...)
	if diags.HasError() {
		return nil
	}

	r.TableWithColumns = &tbc
	return &r
}

func (d *tbcResource) findOptIn(ctx context.Context, input *lakeformation.ListLakeFormationOptInsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[ResourceData] {
	tbc, err := d.data.TableWithColumns.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
	}

	for _, v := range input.LakeFormationOptInsInfoList {
		if v.Resource != nil && v.Resource.TableWithColumns != nil {
			if aws.ToString(v.Resource.TableWithColumns.Name) == tbc.Name.ValueString() &&
				aws.ToString(v.Resource.TableWithColumns.DatabaseName) == tbc.DatabaseName.ValueString() {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ResourceData{
					TableWithColumns: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tableWithColumns{
						Name: fwflex.StringToFramework(ctx, v.Resource.TableWithColumns.Name),
					}),
				})
				return out
			}
		}
	}
	return fwtypes.NewListNestedObjectValueOfNull[ResourceData](ctx)
}

////////////////////////////////////////////////////////////////////////////////////////////////

type resourceOptInData struct {
	Principal     fwtypes.ListNestedObjectValueOf[DataLakePrincipal] `tfsdk:"principal"`
	Resource      fwtypes.ListNestedObjectValueOf[ResourceData]      `tfsdk:"resource"`
	Condition     fwtypes.ListNestedObjectValueOf[Condition]         `tfsdk:"condition"`
	LastUpdatedBy timetypes.RFC3339                                  `tfsdk:"last_updated_by"`
	LastModified  types.String                                       `tfsdk:"last_modified"`
}

type DataLakePrincipal struct {
	DataLakePrincipalIdentifier types.String `tfsdk:"data_lake_principal_identifier"`
}

type ResourceData struct {
	Catalog          fwtypes.ListNestedObjectValueOf[Catalog]          `tfsdk:"catalog"`
	DataCellsFilter  fwtypes.ListNestedObjectValueOf[DataCellsFilter]  `tfsdk:"data_cells_filter"`
	DataLocation     fwtypes.ListNestedObjectValueOf[DataLocation]     `tfsdk:"data_location"`
	Database         fwtypes.ListNestedObjectValueOf[Database]         `tfsdk:"database"`
	LFTag            fwtypes.ListNestedObjectValueOf[LFTag]            `tfsdk:"lf_tag"`
	LFTagExpression  fwtypes.ListNestedObjectValueOf[LFTagExpression]  `tfsdk:"lf_tag_expression"`
	LFTagPolicy      fwtypes.ListNestedObjectValueOf[LFTagPolicy]      `tfsdk:"lf_tag_policy"`
	Table            fwtypes.ListNestedObjectValueOf[table]            `tfsdk:"table"`
	TableWithColumns fwtypes.ListNestedObjectValueOf[tableWithColumns] `tfsdk:"table_with_columns"`
}

type Catalog struct {
	ID types.String `tfsdk:"id"`
}

type Condition struct {
	Expression types.String `tfsdk:"expression"`
}

type DataCellsFilter struct {
	DatabaseName   types.String `tfsdk:"database_name"`
	Name           types.String `tfsdk:"name"`
	TableCatalogID types.String `tfsdk:"table_catalog_id"`
	TableName      types.String `tfsdk:"table_name"`
}

type DataLocation struct {
	ResourceArn types.String `tfsdk:"resource_arn"`
	CatalogID   types.String `tfsdk:"catalog_id"`
}

type LFTagExpression struct {
	Name      types.String `tfsdk:"name"`
	CatalogID types.String `tfsdk:"catalog_id"`
}

type LFTagPolicy struct {
	ResourceType   fwtypes.StringEnum[awstypes.ResourceType] `tfsdk:"resource_type"`
	CatalogID      types.String                              `tfsdk:"catalog_id"`
	Expression     fwtypes.ListValueOf[types.String]         `tfsdk:"expression"`
	ExpressionName types.String                              `tfsdk:"expression_name"`
}
