// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Data Cells Filter")
func newResourceDataCellsFilter(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDataCellsFilter{}

	return r, nil
}

const (
	ResNameDataCellsFilter = "Data Cells Filter"
)

type resourceDataCellsFilter struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceDataCellsFilter) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lakeformation_data_cells_filter"
}

func (r *resourceDataCellsFilter) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"table_data": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tableData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"column_names": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
							//Validators: []validator.Set{
							//	setvalidator.ConflictsWith(
							//		path.MatchRelative().AtParent().AtName("column_wildcard"),
							//	),
							//},
						},
						"database_name": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"table_catalog_id": schema.StringAttribute{
							Required: true,
						},
						"table_name": schema.StringAttribute{
							Required: true,
						},
						"version_id": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"column_wildcard": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[columnWildcard](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								//listvalidator.ConflictsWith(
								//	path.MatchRelative().AtParent().AtName("column_names"),
								//),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"excluded_column_names": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Optional:   true,
									},
								},
							},
						},
						"row_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[rowFilter](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"filter_expression": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"all_rows_wildcard": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[allRowsWildcard](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
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

func (r *resourceDataCellsFilter) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan resourceDataCellsFilterData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lakeformation.CreateDataCellsFilterInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	td, diags := plan.TableData.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateDataCellsFilter(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameDataCellsFilter, td.Name.String(), err),
			err.Error(),
		)
		return
	}

	idParts := []string{
		td.DatabaseName.ValueString(),
		td.Name.ValueString(),
		td.TableCatalogID.ValueString(),
		td.TableName.ValueString(),
	}
	id, err := intflex.FlattenResourceId(idParts, len(idParts), false)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionFlatteningResourceId, ResNameDataCellsFilter, td.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataCellsFilter) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceDataCellsFilterData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDataCellsFilterByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameDataCellsFilter, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataCellsFilter) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan, state resourceDataCellsFilterData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TableData.Equal(state.TableData) {
		in := &lakeformation.UpdateDataCellsFilterInput{}

		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDataCellsFilter(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LakeFormation, create.ErrActionUpdating, ResNameDataCellsFilter, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataCellsFilter) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state resourceDataCellsFilterData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	td, diags := state.TableData.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts, err := intflex.ExpandResourceId(state.ID.ValueString(), 4, false)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionExpandingResourceId, ResNameDataCellsFilter, td.Name.String(), err),
			err.Error(),
		)
		return
	}

	in := &lakeformation.DeleteDataCellsFilterInput{
		DatabaseName:   aws.String(idParts[0]),
		Name:           aws.String(idParts[1]),
		TableCatalogId: aws.String(idParts[2]),
		TableName:      aws.String(idParts[3]),
	}

	_, err = conn.DeleteDataCellsFilter(ctx, in)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameDataCellsFilter, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDataCellsFilter) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("table_data").AtListIndex(0).AtName("column_names"),
			path.MatchRoot("table_data").AtListIndex(0).AtName("column_wildcard"),
		),
	}
}

type identifier string

func (i *identifier) String() string {
	return string(*i)
}

func (i *identifier) Len() int {
	return len(strings.Split(string(*i), intflex.ResourceIdSeparator))
}

func findDataCellsFilterByID(ctx context.Context, conn *lakeformation.Client, id string) (*awstypes.DataCellsFilter, error) {
	idd := identifier(id)
	idParts, err := intflex.ExpandResourceId(idd.String(), idd.Len(), false)

	if err != nil {
		return nil, err
	}

	in := &lakeformation.GetDataCellsFilterInput{
		DatabaseName:   aws.String(idParts[0]),
		Name:           aws.String(idParts[1]),
		TableCatalogId: aws.String(idParts[2]),
		TableName:      aws.String(idParts[3]),
	}

	out, err := conn.GetDataCellsFilter(ctx, in)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.DataCellsFilter == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DataCellsFilter, nil
}

type resourceDataCellsFilterData struct {
	ID        types.String                               `tfsdk:"id"`
	TableData fwtypes.ListNestedObjectValueOf[tableData] `tfsdk:"table_data"`
}

type tableData struct {
	DatabaseName   types.String                                    `tfsdk:"database_name"`
	Name           types.String                                    `tfsdk:"name"`
	TableCatalogID types.String                                    `tfsdk:"table_catalog_id"`
	TableName      types.String                                    `tfsdk:"table_name"`
	ColumnNames    fwtypes.SetValueOf[types.String]                `tfsdk:"column_names"`
	ColumnWildcard fwtypes.ListNestedObjectValueOf[columnWildcard] `tfsdk:"column_wildcard"`
	RowFilter      fwtypes.ListNestedObjectValueOf[rowFilter]      `tfsdk:"row_filter"`
	VersionID      types.String                                    `tfsdk:"version_id"`
}

type columnWildcard struct {
	ExcludedColumnNames fwtypes.ListValueOf[types.String] `tfsdk:"excluded_column_names"`
}

type rowFilter struct {
	AllRowsWildcard  fwtypes.ListNestedObjectValueOf[allRowsWildcard] `tfsdk:"all_rows_wildcard"`
	FilterExpression types.String                                     `tfsdk:"filter_expression"`
}

type allRowsWildcard struct{}
