package quicksight

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceDataSet(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceDataSet{}, nil
}

const (
	ResNameDataSet = "DataSet"
)

type resourceDataSet struct {
	framework.ResourceWithConfigure
}

func (r *resourceDataSet) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_data_set"
}

func (r *resourceDataSet) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aws_account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"import_mode": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(quicksight.DataSetImportMode_Values()...),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"physical_table_map": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 32),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"physical_table_map_id": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"s3_source": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_source_arn": schema.StringAttribute{
										Required: true,
										// TODO: verify valid ARN
									},
									"upload_settings": schema.ListAttribute{
										Optional:    true,
										ElementType: types.ObjectType{AttrTypes: uploadSettingsAttrTypes},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											newUploadSettingsPlanModifier(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"input_columns": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 2048),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 128),
													},
												},
												"type": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.OneOf(quicksight.InputColumnDataType_Values()...),
													},
												},
											},
										},
									},
									// NOTE: The custom upload settings plan modifier does not work as expected
									// with a ListNestedBlock attribute. Changing to a ListAttribute (see above)
									// results in the intended behavior. Keeping this note and the attempted block
									// design around while research continues.
									//
									// "upload_settings": schema.ListNestedBlock{
									// 	Validators: []validator.List{
									// 		listvalidator.SizeAtMost(1),
									// 	},
									// 	PlanModifiers: []planmodifier.List{
									// 		newUploadSettingsPlanModifier(),
									// 	},
									// 	NestedObject: schema.NestedBlockObject{
									// 		Attributes: map[string]schema.Attribute{
									// 			"contains_header": schema.BoolAttribute{
									// 				Computed: true,
									// 				Optional: true,
									// 				PlanModifiers: []planmodifier.Bool{
									// 					boolplanmodifier.UseStateForUnknown(),
									// 				},
									// 			},
									// 			"delimiter": schema.StringAttribute{
									// 				Computed: true,
									// 				Optional: true,
									// 				Validators: []validator.String{
									// 					stringvalidator.LengthBetween(1, 1),
									// 				},
									// 				PlanModifiers: []planmodifier.String{
									// 					stringplanmodifier.UseStateForUnknown(),
									// 				},
									// 			},
									// 			"format": schema.StringAttribute{
									// 				Computed: true,
									// 				Optional: true,
									// 				Validators: []validator.String{
									// 					stringvalidator.OneOf(quicksight.FileFormat_Values()...),
									// 				},
									// 				PlanModifiers: []planmodifier.String{
									// 					stringplanmodifier.UseStateForUnknown(),
									// 				},
									// 			},
									// 			"start_from_row": schema.Int64Attribute{
									// 				Computed: true,
									// 				Optional: true,
									// 				Validators: []validator.Int64{
									// 					int64validator.AtLeast(1),
									// 				},
									// 				PlanModifiers: []planmodifier.Int64{
									// 					int64planmodifier.UseStateForUnknown(),
									// 				},
									// 			},
									// 			"text_qualifier": schema.StringAttribute{
									// 				Computed: true,
									// 				Optional: true,
									// 				Validators: []validator.String{
									// 					stringvalidator.OneOf(quicksight.TextQualifier_Values()...),
									// 				},
									// 				PlanModifiers: []planmodifier.String{
									// 					stringplanmodifier.UseStateForUnknown(),
									// 				},
									// 			},
									// 		},
									// 	},
									// },
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceDataSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn()

	var plan resourceDataSetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}

	// TODO: move into helper function
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.AWSAccountID.ValueString(), plan.DataSetID.ValueString()))

	var physicalTableMap []physicalTableMapData
	resp.Diagnostics.Append(plan.PhysicalTableMap.ElementsAs(ctx, &physicalTableMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	physicalTableMapObj, d := expandPhysicalTableMap(ctx, physicalTableMap)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := quicksight.CreateDataSetInput{
		AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
		DataSetId:        aws.String(plan.DataSetID.ValueString()),
		ImportMode:       aws.String(plan.ImportMode.ValueString()),
		PhysicalTableMap: physicalTableMapObj,
		Name:             aws.String(plan.Name.ValueString()),
	}

	out, err := conn.CreateDataSetWithContext(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameDataSet, plan.DataSetID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameDataSet, plan.DataSetID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// Create output does not return a complete data set object, so a read is required
	// before refreshing state
	outFind, err := FindDataSetByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameDataSet, plan.DataSetID.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), outFind)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceDataSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn()

	var state resourceDataSetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindDataSetByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, ResNameDataSet, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightConn()

	var plan, state resourceDataSetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ImportMode.Equal(state.ImportMode) ||
		!plan.PhysicalTableMap.Equal(state.PhysicalTableMap) ||
		!plan.Name.Equal(state.Name) {
		var physicalTableMap []physicalTableMapData
		resp.Diagnostics.Append(plan.PhysicalTableMap.ElementsAs(ctx, &physicalTableMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		physicalTableMapObj, d := expandPhysicalTableMap(ctx, physicalTableMap)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := quicksight.UpdateDataSetInput{
			AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
			DataSetId:        aws.String(plan.DataSetID.ValueString()),
			ImportMode:       aws.String(plan.ImportMode.ValueString()),
			PhysicalTableMap: physicalTableMapObj,
			Name:             aws.String(plan.Name.ValueString()),
		}

		out, err := conn.UpdateDataSetWithContext(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameDataSet, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameDataSet, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// Update output does not return a complete data set object, so a read is required
		// before refreshing state
		outFind, err := FindDataSetByID(ctx, conn, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameDataSet, plan.DataSetID.String(), nil),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), outFind)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceDataSet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn()

	var state resourceDataSetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteDataSetWithContext(ctx, &quicksight.DeleteDataSetInput{
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		DataSetId:    aws.String(state.DataSetID.ValueString()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameDataSet, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceDataSet) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceDataSet) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindDataSetByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.DataSet, error) {
	awsAccountId, dataSetId, err := ParseDataSetID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeDataSetInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSetId:    aws.String(dataSetId),
	}

	out, err := conn.DescribeDataSetWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil, &sdkv2resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DataSet == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DataSet, nil
}

func ParseDataSetID(id string) (string, string, error) {
	// TODO: switch delim to ","
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], nil
}

// uploadSettingsPlanModifier is a plan modifier that handles setting the
// planned value to unknown (ie. making the attribute computed) when not
// configured. This avoids "inconsistent final plan" errors when the AWS
// default values are returned in the API response.
type uploadSettingsPlanModifier struct{}

func (m uploadSettingsPlanModifier) Description(ctx context.Context) string {
	return "If value is not configured, plan as unknown"
}

func (m uploadSettingsPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "If value is not configured, plan as unknown"
}

func (m uploadSettingsPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If plan and state values are null (ie. create with upload_settings omitted),
	// set the planned value to unknown.
	if req.PlanValue.IsNull() && req.StateValue.IsNull() {
		// resp.Diagnostics.AddAttributeWarning(req.Path, "Create modification", "Null plan set to unknown")
		log.Print("[DEBUG] null upload_settings plan set to unknown")
		resp.PlanValue = types.ListUnknown(types.ObjectType{AttrTypes: uploadSettingsAttrTypes})
	}

	// If plan is null but state is not, "UseStateForNull" and set the planned
	// value to the stored AWS defaults, preventing a persistent diff.
	//
	// TODO: this will also suppress changes when a configured block is removed.
	// Need a way to distinguish the "keep default" case from removing explicitly
	// configured values.
	if req.PlanValue.IsNull() && !req.StateValue.IsNull() {
		// resp.Diagnostics.AddAttributeWarning(req.Path, "Update modification", fmt.Sprintf("Null plan set to state value: %s", req.StateValue.String()))
		log.Printf("[DEBUG] null upload_settings plan set to state value: %s", req.StateValue.String())
		resp.PlanValue = req.StateValue
	}
}

func newUploadSettingsPlanModifier() planmodifier.List {
	return uploadSettingsPlanModifier{}
}

var (
	physicalTableMapAttrTypes = map[string]attr.Type{
		"physical_table_map_id": types.StringType,
		"s3_source":             types.ListType{ElemType: types.ObjectType{AttrTypes: s3SourceAttrTypes}},
	}

	s3SourceAttrTypes = map[string]attr.Type{
		"data_source_arn": types.StringType,
		"input_columns":   types.ListType{ElemType: types.ObjectType{AttrTypes: inputColumnsAttrTypes}},
		"upload_settings": types.ListType{ElemType: types.ObjectType{AttrTypes: uploadSettingsAttrTypes}},
	}

	inputColumnsAttrTypes = map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
	}

	uploadSettingsAttrTypes = map[string]attr.Type{
		"contains_header": types.BoolType,
		"delimiter":       types.StringType,
		"format":          types.StringType,
		"start_from_row":  types.Int64Type,
		"text_qualifier":  types.StringType,
	}
)

type resourceDataSetData struct {
	ARN              types.String `tfsdk:"arn"`
	AWSAccountID     types.String `tfsdk:"aws_account_id"`
	DataSetID        types.String `tfsdk:"data_set_id"`
	ImportMode       types.String `tfsdk:"import_mode"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	PhysicalTableMap types.Set    `tfsdk:"physical_table_map"`
	Tags             types.Map    `tfsdk:"tags"`
	TagsAll          types.Map    `tfsdk:"tags_all"`
}

type physicalTableMapData struct {
	PhysicalTableMapID types.String `tfsdk:"physical_table_map_id"`
	S3Source           types.List   `tfsdk:"s3_source"`
}

type s3SourceData struct {
	DataSourceARN  types.String `tfsdk:"data_source_arn"`
	InputColumns   types.List   `tfsdk:"input_columns"`
	UploadSettings types.List   `tfsdk:"upload_settings"`
}

type inputColumnsData struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type uploadSettingsData struct {
	ContainsHeader types.Bool   `tfsdk:"contains_header"`
	Delimiter      types.String `tfsdk:"delimiter"`
	Format         types.String `tfsdk:"format"`
	StartFromRow   types.Int64  `tfsdk:"start_from_row"`
	TextQualifier  types.String `tfsdk:"text_qualifier"`
}

func (rd *resourceDataSetData) refreshFromOutput(ctx context.Context, meta *conns.AWSClient, out *quicksight.DataSet) diag.Diagnostics {
	var diags diag.Diagnostics
	if out == nil {
		return diags
	}

	rd.ARN = flex.StringToFramework(ctx, out.Arn)
	rd.DataSetID = flex.StringToFramework(ctx, out.DataSetId)
	rd.ImportMode = flex.StringToFramework(ctx, out.ImportMode)
	rd.Name = flex.StringToFramework(ctx, out.Name)

	physicalTableMap, d := flattenPhysicalTableMap(ctx, out.PhysicalTableMap)
	diags.Append(d...)
	rd.PhysicalTableMap = physicalTableMap

	tagsOut, err := ListTags(ctx, meta.QuickSightConn(), aws.StringValue(out.Arn))
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameDataSet, rd.ID.String(), nil),
			err.Error(),
		)
		return diags
	}

	defaultTagsConfig := meta.DefaultTagsConfig
	ignoreTagsConfig := meta.IgnoreTagsConfig
	tags := tagsOut.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// AWS APIs often return empty lists of tags when none have been configured.
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 {
		rd.Tags = tftags.Null
	} else {
		rd.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags)
	}
	rd.TagsAll = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	return diags
}

func expandPhysicalTableMap(ctx context.Context, tfSet []physicalTableMapData) (map[string]*quicksight.PhysicalTable, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfSet) == 0 {
		return nil, diags
	}

	obj := make(map[string]*quicksight.PhysicalTable)
	for _, item := range tfSet {
		var s3Source []s3SourceData
		diags.Append(item.S3Source.ElementsAs(ctx, &s3Source, false)...)
		s3SourceObj, d := expandS3Source(ctx, s3Source)
		diags.Append(d...)

		obj[item.PhysicalTableMapID.ValueString()] = &quicksight.PhysicalTable{
			S3Source: s3SourceObj,
		}
	}
	return obj, diags
}

func expandS3Source(ctx context.Context, tfList []s3SourceData) (*quicksight.S3Source, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil, diags
	}

	item := tfList[0]
	var inputColumns []inputColumnsData
	diags.Append(item.InputColumns.ElementsAs(ctx, &inputColumns, false)...)

	var uploadSettings []uploadSettingsData
	if !item.UploadSettings.IsUnknown() {
		diags.Append(item.UploadSettings.ElementsAs(ctx, &uploadSettings, false)...)
	}

	obj := &quicksight.S3Source{
		DataSourceArn:  aws.String(item.DataSourceARN.ValueString()),
		InputColumns:   expandInputColumns(inputColumns),
		UploadSettings: expandUploadSettings(uploadSettings),
	}
	return obj, diags
}

func expandInputColumns(tfList []inputColumnsData) []*quicksight.InputColumn {
	if len(tfList) == 0 {
		return nil
	}

	var obj []*quicksight.InputColumn
	for _, item := range tfList {
		obj = append(obj, &quicksight.InputColumn{
			Name: aws.String(item.Name.ValueString()),
			Type: aws.String(item.Type.ValueString()),
		})
	}
	return obj
}

func expandUploadSettings(tfList []uploadSettingsData) *quicksight.UploadSettings {
	if len(tfList) == 0 {
		return nil
	}
	item := tfList[0]
	return &quicksight.UploadSettings{
		ContainsHeader: aws.Bool(item.ContainsHeader.ValueBool()),
		Delimiter:      aws.String(item.Delimiter.ValueString()),
		Format:         aws.String(item.Format.ValueString()),
		StartFromRow:   aws.Int64(item.StartFromRow.ValueInt64()),
		TextQualifier:  aws.String(item.TextQualifier.ValueString()),
	}
}

func flattenPhysicalTableMap(ctx context.Context, apiObject map[string]*quicksight.PhysicalTable) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: physicalTableMapAttrTypes}

	if apiObject == nil {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for k, v := range apiObject {
		s3Source, d := flattenS3Source(ctx, v.S3Source)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"physical_table_map_id": flex.StringValueToFramework(ctx, k),
			"s3_source":             s3Source,
		}
		objVal, d := types.ObjectValue(physicalTableMapAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenS3Source(ctx context.Context, apiObject *quicksight.S3Source) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: s3SourceAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	inputColumns, d := flattenInputColumns(ctx, apiObject.InputColumns)
	diags.Append(d...)
	uploadSettings, d := flattenUploadSettings(ctx, apiObject.UploadSettings)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"data_source_arn": flex.StringToFramework(ctx, apiObject.DataSourceArn),
		"input_columns":   inputColumns,
		"upload_settings": uploadSettings,
	}
	objVal, d := types.ObjectValue(s3SourceAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenInputColumns(ctx context.Context, apiObject []*quicksight.InputColumn) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: inputColumnsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, ic := range apiObject {
		obj := map[string]attr.Value{
			"name": flex.StringToFramework(ctx, ic.Name),
			"type": flex.StringToFramework(ctx, ic.Type),
		}
		objVal, d := types.ObjectValue(inputColumnsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenUploadSettings(ctx context.Context, apiObject *quicksight.UploadSettings) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: uploadSettingsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"contains_header": flex.BoolToFramework(ctx, apiObject.ContainsHeader),
		"delimiter":       flex.StringToFramework(ctx, apiObject.Delimiter),
		"format":          flex.StringToFramework(ctx, apiObject.Format),
		"start_from_row":  flex.Int64ToFramework(ctx, apiObject.StartFromRow),
		"text_qualifier":  flex.StringToFramework(ctx, apiObject.TextQualifier),
	}
	objVal, d := types.ObjectValue(uploadSettingsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}
