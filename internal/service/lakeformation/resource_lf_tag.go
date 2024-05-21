// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Resource LF Tag")
func newResourceResourceLFTag(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourceLFTag{}
	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameResourceLFTag = "Resource LF Tag"
)

type resourceResourceLFTag struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceResourceLFTag) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lakeformation_resource_lf_tag"
}

func (r *resourceResourceLFTag) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCatalogID: catalogIDSchemaOptional(),
			names.AttrID:        framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func catalogIDSchemaOptional() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func catalogIDSchemaOptionalComputed() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func (r *resourceResourceLFTag) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var plan ResourceResourceLFTagData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lakeformation.AddLFTagsToResourceInput{}

	if !plan.CatalogID.IsNull() {
		in.CatalogId = fwflex.StringFromFramework(ctx, plan.CatalogID)
	}

	lftagger := newLFTagTagger(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := lftagger.expandResource(ctx, &resp.Diagnostics)
	in.Resource = res
	if resp.Diagnostics.HasError() {
		return
	}

	lfTag, lfDiags := plan.LFTag.ToPtr(ctx)
	resp.Diagnostics.Append(lfDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.LFTags = []awstypes.LFTagPair{
		{
			TagKey: fwflex.StringFromFramework(ctx, lfTag.Key),
			TagValues: []string{
				lfTag.Value.ValueString(),
			},
		},
	}

	var output *lakeformation.AddLFTagsToResourceOutput
	err := retry.RetryContext(ctx, IAMPropagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.AddLFTagsToResource(ctx, in)
		if err != nil {
			if errs.IsA[*awstypes.ConcurrentModificationException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AddLFTagsToResource(ctx, in)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameResourceLFTag, prettify(in), err),
			err.Error(),
		)
		return
	}

	if output != nil && len(output.Failures) > 0 {
		var failureDiags diag.Diagnostics
		for _, v := range output.Failures {
			if v.LFTag == nil || v.Error == nil {
				continue
			}

			errSummary := fmt.Errorf("catalog id:%s, tag key:%s, value:%+v", aws.ToString(v.LFTag.CatalogId), aws.ToString(v.LFTag.TagKey), v.LFTag.TagValues)
			failureDiags.AddError(
				create.ProblemStandardMessage(names.LakeFormation, create.ErrActionCreating, ResNameResourceLFTag, prettify(in), errSummary),
				fmt.Sprintf("%s: %s", aws.ToString(v.Error.ErrorCode), aws.ToString(v.Error.ErrorMessage)))
		}
		resp.Diagnostics.Append(failureDiags...)
		return
	}

	state := plan

	id := fmt.Sprintf("%d", create.StringHashcode(prettify(in)))
	state.ID = fwflex.StringValueToFramework(ctx, id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaw, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (any, error) {
		return findResourceLFTagByID(ctx, conn, state.CatalogID.ValueString(), res)
	})

	out := outputRaw.(*lakeformation.GetResourceLFTagsOutput)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameResourceLFTag, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	outputTag := lftagger.findTag(ctx, out, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.LFTag = outputTag

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceResourceLFTag) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state ResourceResourceLFTagData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lftagger := newLFTagTagger(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	res := lftagger.expandResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceLFTagByID(ctx, conn, state.CatalogID.ValueString(), res)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionSetting, ResNameResourceLFTag, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	outputTag := lftagger.findTag(ctx, out, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.LFTag = outputTag

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourceLFTag) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceResourceLFTagData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResourceLFTag) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LakeFormationClient(ctx)

	var state ResourceResourceLFTagData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lakeformation.RemoveLFTagsFromResourceInput{}

	if !state.CatalogID.IsNull() {
		in.CatalogId = fwflex.StringFromFramework(ctx, state.CatalogID)
	}

	lftagger := newLFTagTagger(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	in.Resource = lftagger.expandResource(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lfTag, lfDiags := state.LFTag.ToPtr(ctx)
	resp.Diagnostics.Append(lfDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.LFTags = []awstypes.LFTagPair{
		{
			TagKey: fwflex.StringFromFramework(ctx, lfTag.Key),
			TagValues: []string{
				lfTag.Value.ValueString(),
			},
		},
	}

	if in.Resource == nil || reflect.DeepEqual(in.Resource, &awstypes.Resource{}) || len(in.LFTags) == 0 {
		resp.Diagnostics.AddWarning(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionDeleting, ResNameResourceLFTag, state.ID.String(), nil),
			"no LF-Tags to remove")
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		var err error
		_, err = conn.RemoveLFTagsFromResource(ctx, in)
		if err != nil {
			if errs.IsA[*awstypes.ConcurrentModificationException](err) {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(fmt.Errorf("removing Lake Formation LF-Tags: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.RemoveLFTagsFromResource(ctx, in)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LakeFormation, create.ErrActionWaitingForDeletion, ResNameResourceLFTag, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceLFTag) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrDatabase),
			path.MatchRoot("table"),
			path.MatchRoot("table_with_columns"),
		),
	}
}

func findResourceLFTagByID(ctx context.Context, conn *lakeformation.Client, catalogId string, resource *awstypes.Resource) (*lakeformation.GetResourceLFTagsOutput, error) {
	in := &lakeformation.GetResourceLFTagsInput{
		ShowAssignedLFTags: aws.Bool(true),
	}

	if catalogId != "" {
		in.CatalogId = aws.String(catalogId)
	}

	in.Resource = resource

	out, err := conn.GetResourceLFTags(ctx, in)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}

type lfTagTagger interface {
	expandResource(context.Context, *diag.Diagnostics) *awstypes.Resource
	findTag(context.Context, *lakeformation.GetResourceLFTagsOutput, *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[LFTag]
}

type dbTagger struct {
	data *ResourceResourceLFTagData
}

type tbTagger struct {
	data *ResourceResourceLFTagData
}

type tbcTagger struct {
	data *ResourceResourceLFTagData
}

func newLFTagTagger(r *ResourceResourceLFTagData, diags *diag.Diagnostics) lfTagTagger {
	switch {
	case !r.Database.IsNull():
		return &dbTagger{data: r}
	case !r.Table.IsNull():
		return &tbTagger{data: r}
	case !r.TableWithColumns.IsNull():
		return &tbcTagger{data: r}
	default:
		diags.AddError("unexpected resource type",
			"unexpected resource type")
		return nil
	}
}

func (d *dbTagger) expandResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
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

func (d *dbTagger) findTag(ctx context.Context, input *lakeformation.GetResourceLFTagsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[LFTag] {
	tag, err := d.data.LFTag.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
	}

	for _, v := range input.LFTagOnDatabase {
		if aws.ToString(v.TagKey) == tag.Key.ValueString() {
			t := slices.IndexFunc(v.TagValues, func(i string) bool {
				return i == tag.Value.ValueString()
			})

			if t != -1 {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTag{
					CatalogID: fwflex.StringToFramework(ctx, v.CatalogId),
					Key:       fwflex.StringToFramework(ctx, v.TagKey),
					Value:     fwflex.StringValueToFramework(ctx, v.TagValues[t]),
				})

				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
}

func (d *tbTagger) expandResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource

	tbptr, err := d.data.Table.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var tb awstypes.TableResource
	diags.Append(fwflex.Expand(ctx, tbptr, &tb)...)
	if diags.HasError() {
		return nil
	}

	r.Table = &tb

	return &r
}

func (d *tbTagger) findTag(ctx context.Context, input *lakeformation.GetResourceLFTagsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[LFTag] {
	tag, err := d.data.LFTag.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
	}

	for _, v := range input.LFTagsOnTable {
		if aws.ToString(v.TagKey) == tag.Key.ValueString() {
			t := slices.IndexFunc(v.TagValues, func(i string) bool {
				return i == tag.Value.ValueString()
			})

			if t != -1 {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTag{
					CatalogID: fwflex.StringToFramework(ctx, v.CatalogId),
					Key:       fwflex.StringToFramework(ctx, v.TagKey),
					Value:     fwflex.StringValueToFramework(ctx, v.TagValues[t]),
				})

				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
}

func (d *tbcTagger) expandResource(ctx context.Context, diags *diag.Diagnostics) *awstypes.Resource {
	var r awstypes.Resource
	tcbptr, err := d.data.TableWithColumns.ToPtr(ctx)
	diags.Append(err...)
	if diags.HasError() {
		return nil
	}

	var tcb awstypes.TableWithColumnsResource
	diags.Append(fwflex.Expand(ctx, tcbptr, &tcb)...)
	if diags.HasError() {
		return nil
	}

	r.TableWithColumns = &tcb

	return &r
}

func (d *tbcTagger) findTag(ctx context.Context, input *lakeformation.GetResourceLFTagsOutput, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[LFTag] {
	tag, err := d.data.LFTag.ToPtr(ctx)
	if err != nil {
		diags.Append(err...)
		return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
	}

	if len(input.LFTagsOnColumns) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
	}

	for _, v := range input.LFTagsOnColumns[0].LFTags {
		if aws.ToString(v.TagKey) == tag.Key.ValueString() {
			t := slices.IndexFunc(v.TagValues, func(i string) bool {
				return i == tag.Value.ValueString()
			})

			if t != -1 {
				out := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &LFTag{
					CatalogID: fwflex.StringToFramework(ctx, v.CatalogId),
					Key:       fwflex.StringToFramework(ctx, v.TagKey),
					Value:     fwflex.StringValueToFramework(ctx, v.TagValues[t]),
				})
				return out
			}
		}
	}

	return fwtypes.NewListNestedObjectValueOfNull[LFTag](ctx)
}

type ResourceResourceLFTagData struct {
	CatalogID        types.String                                      `tfsdk:"catalog_id"`
	Database         fwtypes.ListNestedObjectValueOf[Database]         `tfsdk:"database"`
	ID               types.String                                      `tfsdk:"id"`
	LFTag            fwtypes.ListNestedObjectValueOf[LFTag]            `tfsdk:"lf_tag"`
	Table            fwtypes.ListNestedObjectValueOf[table]            `tfsdk:"table"`
	TableWithColumns fwtypes.ListNestedObjectValueOf[tableWithColumns] `tfsdk:"table_with_columns"`
	Timeouts         timeouts.Value                                    `tfsdk:"timeouts"`
}

type Database struct {
	CatalogID types.String `tfsdk:"catalog_id"`
	Name      types.String `tfsdk:"name"`
}

type table struct {
	CatalogID    types.String `tfsdk:"catalog_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Name         types.String `tfsdk:"name"`
	Wildcard     types.Bool   `tfsdk:"wildcard"`
}

type tableWithColumns struct {
	CatalogID      types.String                                        `tfsdk:"catalog_id"`
	ColumnNames    fwtypes.SetValueOf[types.String]                    `tfsdk:"column_names"`
	ColumnWildcard fwtypes.ListNestedObjectValueOf[columnWildcardData] `tfsdk:"column_wildcard"`
	DatabaseName   types.String                                        `tfsdk:"database_name"`
	Name           types.String                                        `tfsdk:"name"`
}

type columnWildcardData struct {
	ExcludedColumnNames fwtypes.SetValueOf[types.String] `tfsdk:"excluded_column_names"`
}

type LFTag struct {
	CatalogID types.String `tfsdk:"catalog_id"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
}
