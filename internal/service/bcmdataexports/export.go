// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bcmdataexports_export",name="Export")
// @Tags(identifierAttribute="id")
func newResourceExport(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceExport{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameExport = "Export"
)

type resourceExport struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceExport) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bcmdataexports_export"
}

func (r *resourceExport) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	dataQueryLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataQueryData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"query_statement": schema.StringAttribute{
					Required: true,
				},
				"table_configurations": schema.MapAttribute{
					// map[string]map[string]string
					CustomType: fwtypes.NewMapTypeOf[fwtypes.MapValueOf[types.String]](ctx),
					Optional:   true,
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.UseStateForUnknown(),
						mapplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}

	s3OutputConfigurationsLNB := schema.ListNestedBlock{
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

	s3DestinationLNB := schema.ListNestedBlock{
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
				"s3_output_configurations": s3OutputConfigurationsLNB,
			},
		},
	}

	destinationConfigurationsLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigurationsData](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"s3_destination": s3DestinationLNB,
			},
		},
	}

	refreshCadenceLNB := schema.ListNestedBlock{
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

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttribute(),
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
						"data_query":                 dataQueryLNB,
						"destination_configurations": destinationConfigurationsLNB,
						"refresh_cadence":            refreshCadenceLNB,
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

func (r *resourceExport) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var plan resourceExportData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bcmdataexports.CreateExportInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.ResourceTags = getTagsIn(ctx)

	out, err := conn.CreateExport(ctx, in)
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

	plan.ID = flex.StringToFramework(ctx, out.ExportArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaw, err := waitExportCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionWaitingForCreation, ResNameExport, plan.ID.String(), err),
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

func (r *resourceExport) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var state resourceExportData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findExportByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionSetting, ResNameExport, state.ID.String(), err),
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

func (r *resourceExport) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var plan, state resourceExportData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Export.Equal(state.Export) {
		in := &bcmdataexports.UpdateExportInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.ExportArn = aws.String(plan.ID.ValueString())

		out, err := conn.UpdateExport(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionUpdating, ResNameExport, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionUpdating, ResNameExport, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitExportUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionWaitingForUpdate, ResNameExport, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceExport) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BCMDataExportsClient(ctx)

	var state resourceExportData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bcmdataexports.DeleteExportInput{
		ExportArn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteExport(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDataExports, create.ErrActionDeleting, ResNameExport, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceExport) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func waitExportCreated(ctx context.Context, conn *bcmdataexports.Client, id string, timeout time.Duration) (*bcmdataexports.GetExportOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ExportStatusCodeHealthy),
		Refresh:                   statusExport(ctx, conn, id),
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
		Refresh:                   statusExport(ctx, conn, id),
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

func statusExport(ctx context.Context, conn *bcmdataexports.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findExportByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.ExportStatus.StatusCode), nil
	}
}

func findExportByID(ctx context.Context, conn *bcmdataexports.Client, exportArn string) (*bcmdataexports.GetExportOutput, error) {
	in := &bcmdataexports.GetExportInput{
		ExportArn: aws.String(exportArn),
	}

	out, err := conn.GetExport(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceExportData struct {
	Export   fwtypes.ListNestedObjectValueOf[exportData] `tfsdk:"export"`
	ID       types.String                                `tfsdk:"id"`
	Tags     types.Map                                   `tfsdk:"tags"`
	TagsAll  types.Map                                   `tfsdk:"tags_all"`
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
	QueryStatement      types.String                                         `tfsdk:"query_statement"`
	TableConfigurations fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]] `tfsdk:"table_configurations"`
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
