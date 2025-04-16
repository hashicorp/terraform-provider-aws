// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"errors"
	"reflect"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	catalogLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[catalogOptIn](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrID: schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}

	dataCellsFilterLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataCellsFilterOptIn](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrDatabaseName: schema.StringAttribute{
					Optional: true,
				},
				names.AttrName: schema.StringAttribute{
					Optional: true,
				},
				"table_catalog_id": schema.StringAttribute{
					Optional: true,
				},
				names.AttrTableName: schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}

	dataLocationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataLocationOptIn](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrResourceARN: schema.StringAttribute{
					Required: true,
				},
				names.AttrCatalogID: schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}

	databaseLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[databaseOptIn](ctx),
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
	}

	lfTagLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[lfTagOptIn](ctx),
		Validators: []validator.List{
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
	}

	lftagExpressionLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[lfTagExpressionOptIn](ctx),
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
	}

	lfTagPolicyLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[lfTagPolicyOptIn](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrResourceType: schema.StringAttribute{
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
	}

	tableLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[tableOptIn](ctx),
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
	}

	tableWCLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[tableWithColumnsOptIn](ctx),
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
					CustomType: fwtypes.NewListNestedObjectTypeOf[columnWildcardDataOptIn](ctx),
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
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"last_updated_by": schema.StringAttribute{
				Computed: true,
			},
			"last_modified": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrCondition: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[conditionOptIn](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrExpression: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			names.AttrPrincipal: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakePrincipal](ctx),
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
			"resource_data": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceData](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"catalog":            catalogLNB,
						names.AttrDatabase:   databaseLNB,
						"data_cells_filter":  dataCellsFilterLNB,
						"data_location":      dataLocationLNB,
						"lf_tag":             lfTagLNB,
						"lf_tag_expression":  lftagExpressionLNB,
						"lf_tag_policy":      lfTagPolicyLNB,
						"table":              tableLNB,
						"table_with_columns": tableWCLNB,
					},
				},
			},
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

	principal, diags := plan.Principal.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var output *lakeformation.CreateLakeFormationOptInOutput
	err := retry.RetryContext(ctx, 2*IAMPropagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateLakeFormationOptIn(ctx, &in)
		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Insufficient Lake Formation permission(s) on Catalog") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateLakeFormationOptIn(ctx, &in)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameOptIn, principal.DataLakePrincipalIdentifier.ValueString(), err),
			err.Error(),
		)
		return
	}

	if output == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameOptIn, principal.DataLakePrincipalIdentifier.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	lstrsc, err := conn.ListLakeFormationOptIns(ctx, &lakeformation.ListLakeFormationOptInsInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameOptIn, principal.DataLakePrincipalIdentifier.ValueString(), err),
			err.Error(),
		)
		return
	}

	plan.LastModified = fwflex.TimeToFramework(ctx, lstrsc.LakeFormationOptInsInfoList[0].LastModified)
	plan.LastUpdatedBy = fwflex.StringValueToFramework(ctx, *lstrsc.LakeFormationOptInsInfoList[0].LastUpdatedBy)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	if out.LastModified != nil {
		state.LastModified = timetypes.NewRFC3339TimePointerValue(out.LastModified)
	}

	if out.LastUpdatedBy != nil {
		state.LastUpdatedBy = fwflex.StringToFramework(ctx, out.LastUpdatedBy)
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

	if optinResource == nil {
		resp.Diagnostics.AddWarning(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameOptIn, "unknown", errors.New("resource data is nil")),
			"resource data is nil",
		)
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

	in.Principal = &awstypes.DataLakePrincipal{
		DataLakePrincipalIdentifier: principalData.DataLakePrincipalIdentifier.ValueStringPointer(),
	}

	in.Resource = optin.expandOptInResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
}

func (r *resourceOptIn) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("resource_data").AtListIndex(0).AtName("catalog"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("data_cells_filter"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("data_location"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName(names.AttrDatabase),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("lf_tag"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("lf_tag_expression"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("lf_tag_policy"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("table"),
			path.MatchRoot("resource_data").AtListIndex(0).AtName("table_with_columns"),
		),
	}
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
	in := &lakeformation.ListLakeFormationOptInsInput{
		Principal: &awstypes.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(id),
		},
		Resource: resource,
	}

	return findOptIn(ctx, conn, in, func(v *awstypes.LakeFormationOptInsInfo) bool {
		return aws.ToString(v.Principal.DataLakePrincipalIdentifier) == id
	})
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
}

type catalogResource struct {
	data *resourceData
}

type dbResource struct {
	data *resourceData
}

type dcfResource struct {
	data *resourceData
}

type dlResource struct {
	data *resourceData
}

type lftagResource struct {
	data *resourceData
}

type lfteResource struct {
	data *resourceData
}

type lftpResource struct {
	data *resourceData
}

type tbResource struct {
	data *resourceData
}

type tbcResource struct {
	data *resourceData
}

func newOptInResourcer(data *resourceData, diags *diag.Diagnostics) optInResourcer {
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

func (d *tbcResource) expandOptInResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	tbcptr, err := d.data.TableWithColumns.ToPtr(ctx)
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

type resourceOptInData struct {
	Principal     fwtypes.ListNestedObjectValueOf[dataLakePrincipal] `tfsdk:"principal"`
	Resource      fwtypes.ListNestedObjectValueOf[resourceData]      `tfsdk:"resource_data"`
	Condition     fwtypes.ListNestedObjectValueOf[conditionOptIn]    `tfsdk:"condition"`
	LastUpdatedBy types.String                                       `tfsdk:"last_updated_by"`
	LastModified  timetypes.RFC3339                                  `tfsdk:"last_modified"`
}

type dataLakePrincipal struct {
	DataLakePrincipalIdentifier types.String `tfsdk:"data_lake_principal_identifier"`
}

type resourceData struct {
	Catalog          fwtypes.ListNestedObjectValueOf[catalogOptIn]          `tfsdk:"catalog"`
	DataCellsFilter  fwtypes.ListNestedObjectValueOf[dataCellsFilterOptIn]  `tfsdk:"data_cells_filter"`
	DataLocation     fwtypes.ListNestedObjectValueOf[dataLocationOptIn]     `tfsdk:"data_location"`
	Database         fwtypes.ListNestedObjectValueOf[databaseOptIn]         `tfsdk:"database"`
	LFTag            fwtypes.ListNestedObjectValueOf[lfTagOptIn]            `tfsdk:"lf_tag"`
	LFTagExpression  fwtypes.ListNestedObjectValueOf[lfTagExpressionOptIn]  `tfsdk:"lf_tag_expression"`
	LFTagPolicy      fwtypes.ListNestedObjectValueOf[lfTagPolicyOptIn]      `tfsdk:"lf_tag_policy"`
	Table            fwtypes.ListNestedObjectValueOf[tableOptIn]            `tfsdk:"table"`
	TableWithColumns fwtypes.ListNestedObjectValueOf[tableWithColumnsOptIn] `tfsdk:"table_with_columns"`
}

type catalogOptIn struct {
	ID types.String `tfsdk:"id"`
}

type conditionOptIn struct {
	Expression types.String `tfsdk:"expression"`
}

type dataCellsFilterOptIn struct {
	DatabaseName   types.String `tfsdk:"database_name"`
	Name           types.String `tfsdk:"name"`
	TableCatalogID types.String `tfsdk:"table_catalog_id"`
	TableName      types.String `tfsdk:"table_name"`
}

type databaseOptIn struct {
	CatalogID types.String `tfsdk:"catalog_id"`
	Name      types.String `tfsdk:"name"`
}

type dataLocationOptIn struct {
	ResourceArn types.String `tfsdk:"resource_arn"`
	CatalogID   types.String `tfsdk:"catalog_id"`
}

type lfTagOptIn struct {
	CatalogID types.String `tfsdk:"catalog_id"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
}

type lfTagExpressionOptIn struct {
	Name      types.String `tfsdk:"name"`
	CatalogID types.String `tfsdk:"catalog_id"`
}

type lfTagPolicyOptIn struct {
	ResourceType   fwtypes.StringEnum[awstypes.ResourceType] `tfsdk:"resource_type"`
	CatalogID      types.String                              `tfsdk:"catalog_id"`
	Expression     fwtypes.ListValueOf[types.String]         `tfsdk:"expression"`
	ExpressionName types.String                              `tfsdk:"expression_name"`
}

type tableOptIn struct {
	CatalogID    types.String `tfsdk:"catalog_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Name         types.String `tfsdk:"name"`
	Wildcard     types.Bool   `tfsdk:"wildcard"`
}

var (
	_ fwflex.Expander  = tableOptIn{}
	_ fwflex.Flattener = &tableOptIn{}
)

func (m tableOptIn) Expand(_ context.Context) (result any, diags diag.Diagnostics) {
	var r awstypes.TableResource

	r.CatalogId = m.CatalogID.ValueStringPointer()
	r.DatabaseName = m.DatabaseName.ValueStringPointer()
	r.Name = m.Name.ValueStringPointer()

	if m.Wildcard.ValueBool() {
		r.TableWildcard = &awstypes.TableWildcard{}
	}

	return &r, diags
}

func (m *tableOptIn) Flatten(ctx context.Context, input any) (diags diag.Diagnostics) {
	tbOpt, ok := input.(awstypes.TableResource)
	if !ok {
		diags.Append(fwflex.DiagFlatteningIncompatibleTypes(reflect.TypeOf(input), reflect.TypeFor[tableOptIn]()))
		return diags
	}

	m.CatalogID = fwflex.StringToFramework(ctx, tbOpt.CatalogId)
	m.DatabaseName = fwflex.StringToFramework(ctx, tbOpt.DatabaseName)
	m.Name = fwflex.StringToFramework(ctx, tbOpt.Name)
	if tbOpt.TableWildcard != nil {
		m.Wildcard = fwflex.BoolValueToFramework(ctx, true)
		m.Name = types.StringNull()
	}

	return diags
}

type tableWithColumnsOptIn struct {
	CatalogID      types.String                                             `tfsdk:"catalog_id"`
	ColumnNames    fwtypes.SetValueOf[types.String]                         `tfsdk:"column_names"`
	ColumnWildcard fwtypes.ListNestedObjectValueOf[columnWildcardDataOptIn] `tfsdk:"column_wildcard"`
	DatabaseName   types.String                                             `tfsdk:"database_name"`
	Name           types.String                                             `tfsdk:"name"`
}

type columnWildcardDataOptIn struct {
	ExcludedColumnNames fwtypes.SetValueOf[types.String] `tfsdk:"excluded_column_names"`
}
