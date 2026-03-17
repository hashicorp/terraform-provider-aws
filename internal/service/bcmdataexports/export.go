// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bcmdataexports

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bcmdataexports_export",name="Export")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bcmdataexports;bcmdataexports.GetExportOutput")
// @Testing(skipEmptyTags=true, skipNullTags=true)
// @Testing(v60RefreshError=true)
func newExportResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &exportResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameExport = "Export"
)

type exportResource struct {
	framework.ResourceWithModel[exportResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *exportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrID:      framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"export": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[exportData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"export_arn": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"data_query":                 exportDataQuerySchema(ctx),
						"destination_configurations": exportDestinationConfigurationsSchema(ctx),
						"refresh_cadence":            exportRefreshCadenceSchema(ctx),
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func exportDataQuerySchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataQueryData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"query_statement": schema.StringAttribute{
					Required: true,
				},
				"table_configurations": schema.MapAttribute{
					CustomType: fwtypes.MapOfMapOfStringType,
					Optional:   true,
					Computed:   true,
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.UseStateForUnknown(),
						mapplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func exportS3OutputConfigurationsSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[s3OutputConfigurations](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compression": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.CompressionOption](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				names.AttrFormat: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.FormatOption](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"output_type": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.S3OutputType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"overwrite": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.OverwriteOption](),
				},
			},
		},
	}
}

func exportS3DestinationSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[s3Destination](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrS3Bucket: schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_prefix": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_region": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"s3_output_configurations": exportS3OutputConfigurationsSchema(ctx),
			},
		},
	}
}

func exportDestinationConfigurationsSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigurationsData](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"s3_destination": exportS3DestinationSchema(ctx),
			},
		},
	}
}

func exportRefreshCadenceSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[refreshCadenceData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"frequency": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.FrequencyOption](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func (r *exportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var plan exportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := bcmdataexports.CreateExportInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.ResourceTags = getTagsIn(ctx)

	out, err := conn.CreateExport(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionCreating, ResNameExport, "", err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionCreating, ResNameExport, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.ExportArn)
	plan.ID = flex.StringToFramework(ctx, out.ExportArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaw, err := waitExportCreated(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionWaitingForCreation, ResNameExport, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, outputRaw, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *exportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var state exportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findExportByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionReading, ResNameExport, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Export.ExportArn)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *exportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var plan, state exportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Export.Equal(state.Export) {
		in := bcmdataexports.UpdateExportInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.ExportArn = plan.ARN.ValueStringPointer()

		out, err := conn.UpdateExport(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionUpdating, ResNameExport, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionUpdating, ResNameExport, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitExportUpdated(ctx, conn, plan.ARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionWaitingForUpdate, ResNameExport, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *exportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var state exportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := bcmdataexports.DeleteExportInput{
		ExportArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteExport(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionDeleting, ResNameExport, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *exportResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := exportSchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeExportResourceStateFromV0,
		},
	}
}

func waitExportCreated(ctx context.Context, conn *bcmdataexports.Client, id string, timeout time.Duration) (*bcmdataexports.GetExportOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ExportStatusCodeHealthy),
		Refresh:                   statusExport(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bcmdataexports.GetExportOutput); ok {
		return out, err
	}

	return nil, err
}

func waitExportUpdated(ctx context.Context, conn *bcmdataexports.Client, id string, timeout time.Duration) (*bcmdataexports.GetExportOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ExportStatusCodeUnhealthy),
		Target:                    enum.Slice(awstypes.ExportStatusCodeHealthy),
		Refresh:                   statusExport(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bcmdataexports.GetExportOutput); ok {
		return out, err
	}

	return nil, err
}

func statusExport(conn *bcmdataexports.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findExportByARN(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.ExportStatus.StatusCode), nil
	}
}

func findExportByARN(ctx context.Context, conn *bcmdataexports.Client, exportArn string) (*bcmdataexports.GetExportOutput, error) {
	in := bcmdataexports.GetExportInput{
		ExportArn: aws.String(exportArn),
	}

	out, err := conn.GetExport(ctx, &in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type exportResourceModel struct {
	ARN      types.String                                `tfsdk:"arn"`
	Export   fwtypes.ListNestedObjectValueOf[exportData] `tfsdk:"export"`
	ID       types.String                                `tfsdk:"id"`
	Tags     tftags.Map                                  `tfsdk:"tags"`
	TagsAll  tftags.Map                                  `tfsdk:"tags_all"`
	Timeouts timeouts.Value                              `tfsdk:"timeouts"`
}

type exportData struct {
	Description               types.String                                                   `tfsdk:"description"`
	Name                      types.String                                                   `tfsdk:"name"`
	ExportArn                 types.String                                                   `tfsdk:"export_arn"`
	DataQuery                 fwtypes.ListNestedObjectValueOf[dataQueryData]                 `tfsdk:"data_query"`
	DestinationConfigurations fwtypes.ListNestedObjectValueOf[destinationConfigurationsData] `tfsdk:"destination_configurations"`
	RefreshCadence            fwtypes.ListNestedObjectValueOf[refreshCadenceData]            `tfsdk:"refresh_cadence"`
}

type dataQueryData struct {
	QueryStatement      types.String             `tfsdk:"query_statement"`
	TableConfigurations fwtypes.MapOfMapOfString `tfsdk:"table_configurations" autoflex:",omitempty"`
}

type s3OutputConfigurations struct {
	Compression fwtypes.StringEnum[awstypes.CompressionOption] `tfsdk:"compression"`
	Format      fwtypes.StringEnum[awstypes.FormatOption]      `tfsdk:"format"`
	OutputType  fwtypes.StringEnum[awstypes.S3OutputType]      `tfsdk:"output_type"`
	Overwrite   fwtypes.StringEnum[awstypes.OverwriteOption]   `tfsdk:"overwrite"`
}

type s3Destination struct {
	S3Bucket               types.String                                            `tfsdk:"s3_bucket"`
	S3Prefix               types.String                                            `tfsdk:"s3_prefix"`
	S3Region               types.String                                            `tfsdk:"s3_region"`
	S3OutputConfigurations fwtypes.ListNestedObjectValueOf[s3OutputConfigurations] `tfsdk:"s3_output_configurations"`
}

type destinationConfigurationsData struct {
	S3Destination fwtypes.ListNestedObjectValueOf[s3Destination] `tfsdk:"s3_destination"`
}

type refreshCadenceData struct {
	Frequency fwtypes.StringEnum[awstypes.FrequencyOption] `tfsdk:"frequency"`
}
