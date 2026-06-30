// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_devopsagent_asset_file", name="Asset File")
// @IdentityAttribute("agent_space_id")
// @IdentityAttribute("asset_id")
// @IdentityAttribute("path")
// @ImportIDHandler("assetFileImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsagent;devopsagent.GetAssetFileOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccAssetFileImportStateIDFunc)
// @Testing(importStateIdAttribute="path")
// @Testing(importIgnore="content_body")
// @Testing(plannableImportAction="NoOp")
func newAssetFileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assetFileResource{}, nil
}

const (
	ResNameAssetFile = "Asset File"
)

type assetFileResource struct {
	framework.ResourceWithModel[assetFileResourceModel]
	framework.WithImportByIdentity
}

func (r *assetFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_version": schema.Int32Attribute{
				Computed: true,
			},
			"content_body": schema.StringAttribute{
				Required: true,
			},
			names.AttrPath: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	input := devopsagent.CreateAssetFileInput{
		AgentSpaceId: plan.AgentSpaceID.ValueStringPointer(),
		AssetId:      plan.AssetID.ValueStringPointer(),
		Path:         plan.Path.ValueStringPointer(),
		Content: &awstypes.AssetFileBodyMemberText{
			Value: plan.ContentBody.ValueString(),
		},
	}

	out, err := conn.CreateAssetFile(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Path.ValueString())
		return
	}
	if out == nil || out.File == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Path.ValueString())
		return
	}

	plan.AssetVersion = types.Int32PointerValue(out.File.Version)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *assetFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var state assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAssetFileByPath(ctx, conn, state.AgentSpaceID.ValueString(), state.AssetID.ValueString(), state.Path.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *assetFileResource) flatten(_ context.Context, out *devopsagent.GetAssetFileOutput, state *assetFileResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if out.File == nil {
		diags.AddError("unexpected nil File in response", "")
		return diags
	}

	state.Path = types.StringPointerValue(out.File.Path)
	state.AssetVersion = types.Int32PointerValue(out.File.Version)

	// Unlike the Asset resource, content_body IS read back from the API here.
	// GetAssetFile returns the file content directly and the API does not add
	// server-managed fields to it, so config == API response with no spurious diffs.
	// The Asset resource treats content as write-only because GetAsset does not
	// return content (it requires a separate GetAssetContent call) and supports
	// zip bundles where reconstructing the original input would be complex.
	if out.File.Content != nil {
		switch v := out.File.Content.(type) {
		case *awstypes.AssetFileBodyMemberText:
			state.ContentBody = types.StringValue(v.Value)
		case *awstypes.AssetFileBodyMemberBytes:
			state.ContentBody = types.StringValue(string(v.Value))
		}
	}

	return diags
}

func (r *assetFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var plan, state assetFileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Copy version from state — will be overwritten if an update occurs.
	plan.AssetVersion = state.AssetVersion

	if !plan.ContentBody.Equal(state.ContentBody) {
		input := devopsagent.UpdateAssetFileInput{
			AgentSpaceId: plan.AgentSpaceID.ValueStringPointer(),
			AssetId:      plan.AssetID.ValueStringPointer(),
			Path:         plan.Path.ValueStringPointer(),
			Content: &awstypes.AssetFileBodyMemberText{
				Value: plan.ContentBody.ValueString(),
			},
		}

		out, err := conn.UpdateAssetFile(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Path.ValueString())
			return
		}
		if out == nil || out.File == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Path.ValueString())
			return
		}

		plan.AssetVersion = types.Int32PointerValue(out.File.Version)
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

	_, err := conn.DeleteAssetFile(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ValidationException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Path.ValueString())
		return
	}
}

func findAssetFileByPath(ctx context.Context, conn *devopsagent.Client, agentSpaceID, assetID, path string) (*devopsagent.GetAssetFileOutput, error) {
	input := devopsagent.GetAssetFileInput{
		AgentSpaceId: aws.String(agentSpaceID),
		AssetId:      aws.String(assetID),
		Path:         aws.String(path),
	}

	out, err := conn.GetAssetFile(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) || errs.IsA[*awstypes.ValidationException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.File == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type assetFileResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String `tfsdk:"agent_space_id"`
	AssetID      types.String `tfsdk:"asset_id"`
	AssetVersion types.Int32  `tfsdk:"asset_version"`
	ContentBody  types.String `tfsdk:"content_body"`
	Path         types.String `tfsdk:"path"`
}

var _ inttypes.ImportIDParser = assetFileImportID{}

type assetFileImportID struct{}

func (assetFileImportID) Parse(id string) (string, map[string]any, error) {
	// Format: <agent-space-id>,<asset-id>,<path>
	parts := strings.SplitN(id, intflex.ResourceIdSeparator, 3)
	if len(parts) != 3 {
		return "", nil, fmt.Errorf("id %q should be in the format <agent-space-id>"+intflex.ResourceIdSeparator+"<asset-id>"+intflex.ResourceIdSeparator+"<path>", id)
	}

	result := map[string]any{
		"agent_space_id": parts[0],
		"asset_id":       parts[1],
		names.AttrPath:   parts[2],
	}

	return id, result, nil
}
