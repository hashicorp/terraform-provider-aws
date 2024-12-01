package dataexchange

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameJob = "Job"
)

func ResourceJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceJob{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceJob struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_dataexchange_job"
}

func (r *resourceJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"state": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"details": schema.ObjectAttribute{
				Required: true,
				AttributeTypes: map[string]attr.Type{
					"import_assets_from_s3": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"data_set_id": types.StringType,
							"revision_id": types.StringType,
							"asset_sources": types.SetType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"bucket": types.StringType,
										"key":    types.StringType,
									},
								},
							},
						},
					},
					"export_assets_to_s3": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"data_set_id": types.StringType,
							"revision_id": types.StringType,
							"asset_destinations": types.SetType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"asset_id": types.StringType,
										"bucket":   types.StringType,
										"key":      types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

type resourceJobModel struct {
	ARN       types.String   `tfsdk:"arn"`
	CreatedAt types.String   `tfsdk:"created_at"`
	Details   types.Object   `tfsdk:"details"`
	ID        types.String   `tfsdk:"id"`
	State     types.String   `tfsdk:"state"`
	Type      types.String   `tfsdk:"type"`
	UpdatedAt types.String   `tfsdk:"updated_at"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

type detailsModel struct {
	ImportAssetsFromS3 *importAssetsFromS3Details `tfsdk:"import_assets_from_s3"`
	ExportAssetsToS3   *exportAssetsToS3Details   `tfsdk:"export_assets_to_s3"`
}

type importAssetsFromS3Details struct {
	DataSetId    types.String       `tfsdk:"data_set_id"`
	RevisionId   types.String       `tfsdk:"revision_id"`
	AssetSources []assetSourceEntry `tfsdk:"asset_sources"`
}

type exportAssetsToS3Details struct {
	DataSetId         types.String            `tfsdk:"data_set_id"`
	RevisionId        types.String            `tfsdk:"revision_id"`
	AssetDestinations []assetDestinationEntry `tfsdk:"asset_destinations"`
}

type assetSourceEntry struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

type assetDestinationEntry struct {
	AssetId types.String `tfsdk:"asset_id"`
	Bucket  types.String `tfsdk:"bucket"`
	Key     types.String `tfsdk:"key"`
}

func getImportAssetsFromS3Details(ctx context.Context, obj types.Object) (*importAssetsFromS3Details, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var detailsValue importAssetsFromS3Details

	// Get the import_assets_from_s3 attribute
	importAttr, ok := obj.Attributes()["import_assets_from_s3"]
	if !ok || importAttr.IsNull() {
		return nil, nil
	}

	importObj, ok := importAttr.(types.Object)
	if !ok {
		return nil, fmt.Errorf("unexpected type for import_assets_from_s3")
	}

	importAttrs := importObj.Attributes()

	// Extract data_set_id and revision_id
	if v, ok := importAttrs["data_set_id"].(types.String); ok {
		detailsValue.DataSetId = v
	}
	if v, ok := importAttrs["revision_id"].(types.String); ok {
		detailsValue.RevisionId = v
	}

	// Handle asset_sources which is a Set
	if assetSourcesAttr, ok := importAttrs["asset_sources"].(types.Set); ok {
		var assetSources []assetSourceEntry
		diags := assetSourcesAttr.ElementsAs(ctx, &assetSources, false)
		if diags.HasError() {
			return nil, fmt.Errorf("error converting asset_sources: %s", diags.Errors())
		}
		detailsValue.AssetSources = assetSources
	}

	return &detailsValue, nil
}

func getExportAssetsToS3Details(ctx context.Context, obj types.Object) (*exportAssetsToS3Details, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	var detailsValue exportAssetsToS3Details

	// Get the export_assets_to_s3 attribute
	exportAttr, ok := obj.Attributes()["export_assets_to_s3"]
	if !ok || exportAttr.IsNull() {
		return nil, nil
	}

	exportObj, ok := exportAttr.(types.Object)
	if !ok {
		return nil, fmt.Errorf("unexpected type for export_assets_to_s3")
	}

	exportAttrs := exportObj.Attributes()

	// Extract data_set_id and revision_id
	if v, ok := exportAttrs["data_set_id"].(types.String); ok {
		detailsValue.DataSetId = v
	}
	if v, ok := exportAttrs["revision_id"].(types.String); ok {
		detailsValue.RevisionId = v
	}

	// Handle asset_destinations which is a Set
	if assetDestinationsAttr, ok := exportAttrs["asset_destinations"].(types.Set); ok {
		var assetDestinations []assetDestinationEntry
		diags := assetDestinationsAttr.ElementsAs(ctx, &assetDestinations, false)
		if diags.HasError() {
			return nil, fmt.Errorf("error converting asset_destinations: %s", diags.Errors())
		}
		detailsValue.AssetDestinations = assetDestinations
	}

	return &detailsValue, nil
}

func (r *resourceJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceJobModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	importDetails, err := getImportAssetsFromS3Details(ctx, plan.Details)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing import details",
			err.Error(),
		)
		return
	}

	exportDetails, err := getExportAssetsToS3Details(ctx, plan.Details)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing export details",
			err.Error(),
		)
		return
	}

	// Validate job type and required blocks
	switch plan.Type.ValueString() {
	case string(awstypes.TypeImportAssetsFromS3):
		if importDetails == nil {
			resp.Diagnostics.AddError(
				"Missing Required Configuration",
				"When type is IMPORT_ASSETS_FROM_S3, the details.import_assets_from_s3 block is required",
			)
			return
		}
		if exportDetails != nil {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When type is IMPORT_ASSETS_FROM_S3, the details.export_assets_to_s3 block must not be present",
			)
			return
		}
		// Validate asset sources
		if len(importDetails.AssetSources) == 0 {
			resp.Diagnostics.AddError(
				"Missing Required Configuration",
				"At least one asset_sources block is required in details.import_assets_from_s3",
			)
			return
		}

	case string(awstypes.TypeExportAssetsToS3):
		if exportDetails == nil {
			resp.Diagnostics.AddError(
				"Missing Required Configuration",
				"When type is EXPORT_ASSETS_TO_S3, the details.export_assets_to_s3 block is required",
			)
			return
		}
		if importDetails != nil {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When type is EXPORT_ASSETS_TO_S3, the details.import_assets_from_s3 block must not be present",
			)
			return
		}
		// Validate asset destinations
		if len(exportDetails.AssetDestinations) == 0 {
			resp.Diagnostics.AddError(
				"Missing Required Configuration",
				"At least one asset_destinations block is required in details.export_assets_to_s3",
			)
			return
		}

	default:
		resp.Diagnostics.AddError(
			"Invalid Type Configuration",
			fmt.Sprintf("Type must be either IMPORT_ASSETS_FROM_S3 or EXPORT_ASSETS_TO_S3, got: %s", plan.Type.ValueString()),
		)
		return
	}

	conn := r.Meta().DataExchangeClient(ctx)

	input := &dataexchange.CreateJobInput{
		Type:    awstypes.Type(plan.Type.ValueString()),
		Details: &awstypes.RequestDetails{},
	}

	switch input.Type {
	case awstypes.TypeImportAssetsFromS3:
		var assetSources []awstypes.AssetSourceEntry
		for _, source := range importDetails.AssetSources {
			assetSources = append(assetSources, awstypes.AssetSourceEntry{
				Bucket: aws.String(source.Bucket.ValueString()),
				Key:    aws.String(source.Key.ValueString()),
			})
		}

		// For API calls, we need just the revision ID
		revisionID := importDetails.RevisionId.ValueString()
		if parts := strings.Split(revisionID, ":"); len(parts) > 1 {
			revisionID = parts[1]
		}

		input.Details.ImportAssetsFromS3 = &awstypes.ImportAssetsFromS3RequestDetails{
			AssetSources: assetSources,
			DataSetId:    aws.String(importDetails.DataSetId.ValueString()),
			RevisionId:   aws.String(revisionID),
		}

	case awstypes.TypeExportAssetsToS3:
		var assetDestinations []awstypes.AssetDestinationEntry
		for _, dest := range exportDetails.AssetDestinations {
			assetDestinations = append(assetDestinations, awstypes.AssetDestinationEntry{
				AssetId: aws.String(dest.AssetId.ValueString()),
				Bucket:  aws.String(dest.Bucket.ValueString()),
				Key:     aws.String(dest.Key.ValueString()),
			})
		}
		input.Details.ExportAssetsToS3 = &awstypes.ExportAssetsToS3RequestDetails{
			AssetDestinations: assetDestinations,
			DataSetId:         aws.String(exportDetails.DataSetId.ValueString()),
			RevisionId:        aws.String(exportDetails.RevisionId.ValueString()),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Creating DataExchange Job with input: %#v", input))

	out, err := conn.CreateJob(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameJob, plan.Type.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameJob, plan.Type.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = types.StringValue(aws.ToString(out.Id))
	plan.ARN = types.StringValue(aws.ToString(out.Arn))
	plan.State = types.StringValue(string(out.State))
	plan.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(out.UpdatedAt.Format(time.RFC3339))

	// Start the job
	startInput := &dataexchange.StartJobInput{
		JobId: aws.String(plan.ID.ValueString()),
	}

	_, err = conn.StartJob(ctx, startInput)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, "starting", ResNameJob, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitJobCompleted(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionWaitingForCreation, ResNameJob, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func waitJobCompleted(ctx context.Context, conn *dataexchange.Client, id string, timeout time.Duration) (*dataexchange.GetJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.StateInProgress),
			string(awstypes.StateWaiting),
		},
		Target:  []string{string(awstypes.StateCompleted)},
		Refresh: statusJob(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dataexchange.GetJobOutput); ok {
		return out, err
	}

	return nil, err
}

func (r *resourceJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataExchangeClient(ctx)

	out, err := FindJobById(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionReading, ResNameJob, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = types.StringValue(aws.ToString(out.Arn))
	state.CreatedAt = types.StringValue(out.CreatedAt.Format(time.RFC3339))
	state.State = types.StringValue(string(out.State))
	state.Type = types.StringValue(string(out.Type))
	state.UpdatedAt = types.StringValue(out.UpdatedAt.Format(time.RFC3339))

	// Handle details based on job type
	detailsAttrs := make(map[string]attr.Value)
	if out.Type == awstypes.TypeImportAssetsFromS3 && out.Details.ImportAssetsFromS3 != nil {
		importDetails := out.Details.ImportAssetsFromS3

		// Try to preserve the revision ID format from state
		var revisionID string
		var oldState resourceJobModel
		req.State.Get(ctx, &oldState)
		if !oldState.Details.IsNull() {
			importDetailsFromState, err := getImportAssetsFromS3Details(ctx, oldState.Details)
			if err == nil && importDetailsFromState != nil {
				revisionID = importDetailsFromState.RevisionId.ValueString()
			}
		}
		// If we couldn't get it from state, construct the compound ID
		if revisionID == "" {
			revisionID = fmt.Sprintf("%s:%s",
				aws.ToString(importDetails.DataSetId),
				aws.ToString(importDetails.RevisionId))
		}

		assetSources := make([]attr.Value, 0)
		for _, source := range importDetails.AssetSources {
			sourceAttrs := map[string]attr.Value{
				"bucket": types.StringValue(aws.ToString(source.Bucket)),
				"key":    types.StringValue(aws.ToString(source.Key)),
			}
			sourceObj, diags := types.ObjectValue(
				map[string]attr.Type{
					"bucket": types.StringType,
					"key":    types.StringType,
				},
				sourceAttrs,
			)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}
			assetSources = append(assetSources, sourceObj)
		}

		assetSourcesSet, diags := types.SetValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bucket": types.StringType,
					"key":    types.StringType,
				},
			},
			assetSources,
		)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		importAttrs := map[string]attr.Value{
			"data_set_id":   types.StringValue(aws.ToString(importDetails.DataSetId)),
			"revision_id":   types.StringValue(revisionID),
			"asset_sources": assetSourcesSet,
		}

		importObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"data_set_id": types.StringType,
				"revision_id": types.StringType,
				"asset_sources": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bucket": types.StringType,
							"key":    types.StringType,
						},
					},
				},
			},
			importAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		detailsAttrs["import_assets_from_s3"] = importObj
		detailsAttrs["export_assets_to_s3"] = types.ObjectNull(
			map[string]attr.Type{
				"data_set_id": types.StringType,
				"revision_id": types.StringType,
				"asset_destinations": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"asset_id": types.StringType,
							"bucket":   types.StringType,
							"key":      types.StringType,
						},
					},
				},
			},
		)
	} else if out.Type == awstypes.TypeExportAssetsToS3 && out.Details.ExportAssetsToS3 != nil {
		exportDetails := out.Details.ExportAssetsToS3

		assetDestinations := make([]attr.Value, 0)
		for _, dest := range exportDetails.AssetDestinations {
			destAttrs := map[string]attr.Value{
				"asset_id": types.StringValue(aws.ToString(dest.AssetId)),
				"bucket":   types.StringValue(aws.ToString(dest.Bucket)),
				"key":      types.StringValue(aws.ToString(dest.Key)),
			}
			destObj, diags := types.ObjectValue(
				map[string]attr.Type{
					"asset_id": types.StringType,
					"bucket":   types.StringType,
					"key":      types.StringType,
				},
				destAttrs,
			)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				return
			}
			assetDestinations = append(assetDestinations, destObj)
		}

		assetDestinationsSet, diags := types.SetValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"asset_id": types.StringType,
					"bucket":   types.StringType,
					"key":      types.StringType,
				},
			},
			assetDestinations,
		)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		exportAttrs := map[string]attr.Value{
			"data_set_id":        types.StringValue(aws.ToString(exportDetails.DataSetId)),
			"revision_id":        types.StringValue(aws.ToString(exportDetails.RevisionId)),
			"asset_destinations": assetDestinationsSet,
		}

		exportObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"data_set_id": types.StringType,
				"revision_id": types.StringType,
				"asset_destinations": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"asset_id": types.StringType,
							"bucket":   types.StringType,
							"key":      types.StringType,
						},
					},
				},
			},
			exportAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		detailsAttrs["export_assets_to_s3"] = exportObj
		detailsAttrs["import_assets_from_s3"] = types.ObjectNull(
			map[string]attr.Type{
				"data_set_id": types.StringType,
				"revision_id": types.StringType,
				"asset_sources": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"bucket": types.StringType,
							"key":    types.StringType,
						},
					},
				},
			},
		)
	}

	detailsObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"import_assets_from_s3": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"data_set_id": types.StringType,
					"revision_id": types.StringType,
					"asset_sources": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"bucket": types.StringType,
								"key":    types.StringType,
							},
						},
					},
				},
			},
			"export_assets_to_s3": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"data_set_id": types.StringType,
					"revision_id": types.StringType,
					"asset_destinations": types.SetType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"asset_id": types.StringType,
								"bucket":   types.StringType,
								"key":      types.StringType,
							},
						},
					},
				},
			},
		},
		detailsAttrs,
	)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	state.Details = detailsObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"AWS DataExchange Jobs cannot be updated after creation. All changes require replacement.",
	)
}

func (r *resourceJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceJobModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataExchangeClient(ctx)

	// Only try to cancel if job is in WAITING state
	out, err := FindJobById(ctx, conn, state.ID.ValueString())
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameJob, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if out.State == awstypes.StateWaiting {
		input := &dataexchange.CancelJobInput{
			JobId: aws.String(state.ID.ValueString()),
		}

		_, err = conn.CancelJob(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return
			}
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}

		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		_, err = waitJobDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionWaitingForDeletion, ResNameJob, state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}
	// For completed or other state jobs, just remove from state
}

func (r *resourceJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitJobDeleted(ctx context.Context, conn *dataexchange.Client, id string, timeout time.Duration) (*dataexchange.GetJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.StateInProgress),
			string(awstypes.StateWaiting),
		},
		Target: []string{
			string(awstypes.StateCancelled),
			string(awstypes.StateCompleted),
			string(awstypes.StateError),
		},
		Refresh: statusJob(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*dataexchange.GetJobOutput); ok {
		return out, err
	}

	return nil, err
}

func statusJob(ctx context.Context, conn *dataexchange.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindJobById(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindJobById(ctx context.Context, conn *dataexchange.Client, id string) (*dataexchange.GetJobOutput, error) {
	input := &dataexchange.GetJobInput{
		JobId: aws.String(id),
	}

	output, err := conn.GetJob(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func JobParseRevisionID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("unexpected format for revision ID (%s), expected DATA-SET_ID:REVISION-ID", id)
}
