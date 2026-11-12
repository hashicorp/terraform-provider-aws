// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Writes to files in the same asset use optimistic concurrency and can fail with
	// ConflictException when another client writes to the asset concurrently.
	assetFileConflictTimeout    = 2 * time.Minute
	assetFilePropagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_devopsagent_asset_file", name="Asset File")
// @IdentityAttribute("agent_space_id")
// @IdentityAttribute("asset_id")
// @IdentityAttribute("path")
// @ImportIDHandler("assetFileImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccAssetFileImportStateID")
// @Testing(importStateIdAttribute="path")
func newAssetFileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assetFileResource{}, nil
}

type assetFileResource struct {
	framework.ResourceWithModel[assetFileResourceModel]
	framework.WithImportByIdentity
}

func (r *assetFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_space_id": schema.StringAttribute{
				Description:   "Identifier of the Agent Space containing the asset. Changing this value requires replacement.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 2048)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"asset_id": schema.StringAttribute{
				Description:   "Identifier of the asset containing the file. Changing this value requires replacement.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Description:   "Creation timestamp in RFC 3339 format.",
				CustomType:    timetypes.RFC3339Type{},
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"metadata": schema.StringAttribute{
				Description: "JSON document describing the file. Updates replace supplied top-level fields and retain omitted fields in AWS; removing a field from this document does not delete it from AWS.",
				CustomType:  fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Optional:    true,
			},
			"metadata_actual": schema.StringAttribute{
				Description: "Full JSON metadata returned by AWS, including retained fields and service-provided defaults.",
				CustomType:  fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Computed:    true,
			},
			names.AttrPath: schema.StringAttribute{
				Description:   "Path of the file within the asset. Changing this value requires replacement.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp in RFC 3339 format.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			names.AttrVersion: schema.Int32Attribute{
				Description: "Asset version the file belongs to.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrContent: schema.ListNestedBlock{
				Description: "File content. Exactly one block is required.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[assetFileBodyModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bytes": schema.StringAttribute{
							Description: "Base64-encoded binary file content.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("text")),
							},
						},
						"text": schema.StringAttribute{
							Description: "Text file content.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *assetFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var plan assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input devopsagent.CreateAssetFileInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	key := assetFileMutexKey(plan.AgentSpaceID.ValueString(), plan.AssetID.ValueString())
	conns.GlobalMutexKV.Lock(key)
	defer conns.GlobalMutexKV.Unlock(key)

	_, err := tfresource.RetryWhenIsA[*devopsagent.CreateAssetFileOutput, *awstypes.ConflictException](ctx, assetFileConflictTimeout, func(ctx context.Context) (*devopsagent.CreateAssetFileOutput, error) {
		// The SDK sets ClientToken on input; reusing it replays the cached ConflictException.
		input.ClientToken = nil
		return conn.CreateAssetFile(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Path.ValueString())
		return
	}

	file, err := tfresource.RetryWhenNotFound(ctx, assetFilePropagationTimeout, func(ctx context.Context) (*awstypes.AssetFile, error) {
		return findAssetFileByThreePartKey(ctx, conn, plan.AgentSpaceID.ValueString(), plan.AssetID.ValueString(), plan.Path.ValueString())
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Path.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, file, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *assetFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var state assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findAssetFileByThreePartKey(ctx, conn, state.AgentSpaceID.ValueString(), state.AssetID.ValueString(), state.Path.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *assetFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var plan, state assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// UpdateAssetFile requires content on every update, including metadata-only changes.
		var input devopsagent.UpdateAssetFileInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.AgentSpaceId = state.AgentSpaceID.ValueStringPointer()
		input.AssetId = state.AssetID.ValueStringPointer()
		input.Path = state.Path.ValueStringPointer()
		key := assetFileMutexKey(state.AgentSpaceID.ValueString(), state.AssetID.ValueString())
		conns.GlobalMutexKV.Lock(key)
		defer conns.GlobalMutexKV.Unlock(key)

		_, err := tfresource.RetryWhenIsA[*devopsagent.UpdateAssetFileOutput, *awstypes.ConflictException](ctx, assetFileConflictTimeout, func(ctx context.Context) (*devopsagent.UpdateAssetFileOutput, error) {
			// The SDK sets ClientToken on input; reusing it replays the cached ConflictException.
			input.ClientToken = nil
			return conn.UpdateAssetFile(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
			return
		}

		file, err := findAssetFileByThreePartKey(ctx, conn, state.AgentSpaceID.ValueString(), state.AssetID.ValueString(), state.Path.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, file, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.MetadataActual = state.MetadataActual
		plan.UpdatedAt = state.UpdatedAt
		plan.Version = state.Version
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *assetFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var state assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	input := devopsagent.DeleteAssetFileInput{
		AgentSpaceId: state.AgentSpaceID.ValueStringPointer(),
		AssetId:      state.AssetID.ValueStringPointer(),
		Path:         state.Path.ValueStringPointer(),
	}
	key := assetFileMutexKey(state.AgentSpaceID.ValueString(), state.AssetID.ValueString())
	conns.GlobalMutexKV.Lock(key)
	defer conns.GlobalMutexKV.Unlock(key)

	_, err := tfresource.RetryWhenIsA[*devopsagent.DeleteAssetFileOutput, *awstypes.ConflictException](ctx, assetFileConflictTimeout, func(ctx context.Context) (*devopsagent.DeleteAssetFileOutput, error) {
		return conn.DeleteAssetFile(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	// DeleteAssetFile reports an already-deleted path as a ValidationException.
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "not found") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
	}
}

func findAssetFileByThreePartKey(ctx context.Context, conn *devopsagent.Client, agentSpaceID, assetID, filePath string) (*awstypes.AssetFile, error) {
	input := devopsagent.GetAssetFileInput{
		AgentSpaceId: aws.String(agentSpaceID),
		AssetId:      aws.String(assetID),
		Path:         aws.String(filePath),
	}
	out, err := conn.GetAssetFile(ctx, &input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	case errs.IsA[*awstypes.AccessDeniedException](err):
		// The service returns AccessDeniedException once the parent Agent Space is deleted.
		_, parentErr := findAssetByTwoPartKey(ctx, conn, agentSpaceID, assetID)
		if retry.NotFound(parentErr) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		if parentErr != nil {
			return nil, smarterr.NewError(parentErr)
		}
		return nil, smarterr.NewError(err)
	case err != nil:
		return nil, smarterr.NewError(err)
	}
	if out == nil || out.File == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out.File, nil
}

func (r *assetFileResource) flatten(ctx context.Context, file *awstypes.AssetFile, data *assetFileResourceModel) (diags diag.Diagnostics) {
	if file == nil {
		smerr.AddError(ctx, &diags, tfresource.NewEmptyResultError(), smerr.ID, data.Path.ValueString())
		return diags
	}
	diags.Append(flex.Flatten(ctx, file, data)...)
	if diags.HasError() {
		return diags
	}

	if file.Metadata == nil {
		data.MetadataActual = fwtypes.NewSmithyJSONNull[document.Interface]()
		return diags
	}
	metadata, err := tfsmithy.DocumentToJSONString(file.Metadata)
	if err != nil {
		smerr.AddError(ctx, &diags, err, smerr.ID, data.Path.ValueString())
		return diags
	}
	data.MetadataActual = fwtypes.NewSmithyJSONValue(metadata, document.NewLazyDocument)
	return diags
}

// assetFileMutexKey serializes this provider's writes to files in one asset.
// Concurrent writes to the same asset can succeed while silently dropping a file
// from the asset's latest version.
func assetFileMutexKey(agentSpaceID, assetID string) string {
	return "devopsagent-asset-" + agentSpaceID + "-" + assetID
}

var _ inttypes.ImportIDParser = assetFileImportID{}

type assetFileImportID struct{}

func (assetFileImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || slices.Contains(parts, "") {
		return "", nil, fmt.Errorf("unexpected format for import ID %q, expected AGENT_SPACE_ID,ASSET_ID,PATH", id)
	}
	return id, map[string]any{"agent_space_id": parts[0], "asset_id": parts[1], names.AttrPath: parts[2]}, nil
}

type assetFileResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID   types.String                                        `tfsdk:"agent_space_id"`
	AssetID        types.String                                        `tfsdk:"asset_id"`
	Content        fwtypes.ListNestedObjectValueOf[assetFileBodyModel] `tfsdk:"content"`
	CreatedAt      timetypes.RFC3339                                   `tfsdk:"created_at"`
	Metadata       fwtypes.SmithyJSON[document.Interface]              `tfsdk:"metadata" autoflex:",noflatten"`
	MetadataActual fwtypes.SmithyJSON[document.Interface]              `tfsdk:"metadata_actual" autoflex:"-"`
	Path           types.String                                        `tfsdk:"path"`
	UpdatedAt      timetypes.RFC3339                                   `tfsdk:"updated_at"`
	Version        types.Int32                                         `tfsdk:"version"`
}
