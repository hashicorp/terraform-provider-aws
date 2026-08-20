// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_devopsagent_asset", name="Asset")
// @IdentityAttribute("agent_space_id")
// @IdentityAttribute("asset_id")
// @ImportIDHandler("assetImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsagent;devopsagent.GetAssetOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccAssetImportStateIDFunc)
// @Testing(importStateIdAttribute="asset_id")
// @Testing(importIgnore="content_body;content_path;metadata")
func newAssetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assetResource{}, nil
}

const (
	ResNameAsset = "Asset"
)

type assetResource struct {
	framework.ResourceWithModel[assetResourceModel]
	framework.WithImportByIdentity
}

func (r *assetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"asset_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_version": schema.Int32Attribute{
				Computed: true,
			},
			"content_body": schema.StringAttribute{
				Optional: true,
			},
			"content_path": schema.StringAttribute{
				Optional: true,
			},
			"filename": schema.StringAttribute{
				Optional: true,
			},
			"metadata": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
			},
		},
	}
}

func (r *assetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var plan assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := devopsagent.CreateAssetInput{
		AgentSpaceId: plan.AgentSpaceID.ValueStringPointer(),
		AssetType:    plan.AssetType.ValueStringPointer(),
	}

	// Set metadata if provided.
	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		doc, err := tfsmithy.DocumentFromJSONString(plan.Metadata.ValueString(), document.NewLazyDocument)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetType.ValueString())
			return
		}
		input.Metadata = doc
	}

	// Set content — either from inline content_body or from a local file via filename.
	// These are mutually exclusive.
	if !plan.Filename.IsNull() && !plan.Filename.IsUnknown() {
		content, err := buildContentFromFile(plan.Filename.ValueString(), plan.ContentPath)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetType.ValueString())
			return
		}
		input.Content = content
	} else if !plan.ContentBody.IsNull() && !plan.ContentBody.IsUnknown() {
		path := "SKILL.md"
		if !plan.ContentPath.IsNull() && !plan.ContentPath.IsUnknown() {
			path = plan.ContentPath.ValueString()
		}
		input.Content = &awstypes.AssetContentMemberFile{
			Value: awstypes.AssetFileContent{
				Path: aws.String(path),
				Body: &awstypes.AssetFileBodyMemberText{
					Value: plan.ContentBody.ValueString(),
				},
			},
		}
	}

	out, err := conn.CreateAsset(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetType.ValueString())
		return
	}
	if out == nil || out.Asset == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.AssetType.ValueString())
		return
	}

	plan.AssetID = types.StringPointerValue(out.Asset.AssetId)
	plan.AssetVersion = types.Int32PointerValue(out.Asset.Version)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *assetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var state assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAssetByID(ctx, conn, state.AgentSpaceID.ValueString(), state.AssetID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssetID.ValueString())
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *assetResource) flatten(_ context.Context, out *devopsagent.GetAssetOutput, state *assetResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if out.Asset == nil {
		diags.AddError("unexpected nil Asset in response", "")
		return diags
	}

	state.AssetID = types.StringPointerValue(out.Asset.AssetId)
	state.AssetType = types.StringPointerValue(out.Asset.AssetType)
	state.AssetVersion = types.Int32PointerValue(out.Asset.Version)

	// metadata is not read back from the API. The API adds server-managed defaults
	// (skill_type, status) that would cause spurious diffs since metadata is Optional-only.
	// The user's configured value is preserved in state.

	return diags
}

func (r *assetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var plan, state assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := devopsagent.UpdateAssetInput{
		AgentSpaceId: plan.AgentSpaceID.ValueStringPointer(),
		AssetId:      plan.AssetID.ValueStringPointer(),
	}

	needsUpdate := false

	// Copy version from state — will be overwritten if an update occurs.
	plan.AssetVersion = state.AssetVersion

	if !plan.Metadata.Equal(state.Metadata) {
		needsUpdate = true
		if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
			doc, err := tfsmithy.DocumentFromJSONString(plan.Metadata.ValueString(), document.NewLazyDocument)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetID.ValueString())
				return
			}
			input.Metadata = doc
		}
	}

	if !plan.ContentBody.Equal(state.ContentBody) || !plan.Filename.Equal(state.Filename) {
		needsUpdate = true
		if !plan.Filename.IsNull() && !plan.Filename.IsUnknown() {
			content, err := buildContentFromFile(plan.Filename.ValueString(), plan.ContentPath)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetID.ValueString())
				return
			}
			input.Content = content
		} else if !plan.ContentBody.IsNull() && !plan.ContentBody.IsUnknown() {
			path := "SKILL.md"
			if !plan.ContentPath.IsNull() && !plan.ContentPath.IsUnknown() {
				path = plan.ContentPath.ValueString()
			}
			input.Content = &awstypes.AssetContentMemberFile{
				Value: awstypes.AssetFileContent{
					Path: aws.String(path),
					Body: &awstypes.AssetFileBodyMemberText{
						Value: plan.ContentBody.ValueString(),
					},
				},
			}
		}
	}

	if needsUpdate {
		out, err := conn.UpdateAsset(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AssetID.ValueString())
			return
		}
		if out == nil || out.Asset == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.AssetID.ValueString())
			return
		}

		plan.AssetVersion = types.Int32PointerValue(out.Asset.Version)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *assetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var state assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := devopsagent.DeleteAssetInput{
		AgentSpaceId: state.AgentSpaceID.ValueStringPointer(),
		AssetId:      state.AssetID.ValueStringPointer(),
	}

	_, err := conn.DeleteAsset(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssetID.ValueString())
		return
	}
}

// buildContentFromFile reads a local file and returns the appropriate AssetContent union.
// For .zip files, it uses AssetContentMemberZip. For all other files, it uses
// AssetContentMemberFile with the content_path as the path within the asset (defaulting
// to the base filename if content_path is not set).
func buildContentFromFile(filename string, contentPath types.String) (awstypes.AssetContent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filename, err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".zip" {
		return &awstypes.AssetContentMemberZip{
			Value: awstypes.AssetZipContent{
				ZipFile: data,
			},
		}, nil
	}

	// Non-zip file: determine the path within the asset.
	path := filepath.Base(filename)
	if !contentPath.IsNull() && !contentPath.IsUnknown() {
		path = contentPath.ValueString()
	}

	return &awstypes.AssetContentMemberFile{
		Value: awstypes.AssetFileContent{
			Path: aws.String(path),
			Body: &awstypes.AssetFileBodyMemberBytes{
				Value: data,
			},
		},
	}, nil
}

func findAssetByID(ctx context.Context, conn *devopsagent.Client, agentSpaceID, assetID string) (*devopsagent.GetAssetOutput, error) {
	input := devopsagent.GetAssetInput{
		AgentSpaceId: aws.String(agentSpaceID),
		AssetId:      aws.String(assetID),
	}

	out, err := conn.GetAsset(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Asset == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type assetResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String         `tfsdk:"agent_space_id"`
	AssetID      types.String         `tfsdk:"asset_id"`
	AssetType    types.String         `tfsdk:"asset_type"`
	AssetVersion types.Int32          `tfsdk:"asset_version"`
	ContentBody  types.String         `tfsdk:"content_body"`
	ContentPath  types.String         `tfsdk:"content_path"`
	Filename     types.String         `tfsdk:"filename"`
	Metadata     jsontypes.Normalized `tfsdk:"metadata"`
}

var _ inttypes.ImportIDParser = assetImportID{}

type assetImportID struct{}

func (assetImportID) Parse(id string) (string, map[string]any, error) {
	agentSpaceID, assetID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <agent-space-id>"+intflex.ResourceIdSeparator+"<asset-id>", id)
	}

	result := map[string]any{
		"agent_space_id": agentSpaceID,
		"asset_id":       assetID,
	}

	return id, result, nil
}
