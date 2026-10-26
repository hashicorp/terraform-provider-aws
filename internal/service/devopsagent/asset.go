// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

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

// @FrameworkResource("aws_devopsagent_asset", name="Asset")
// @IdentityAttribute("agent_space_id")
// @IdentityAttribute("asset_id")
// @ImportIDHandler("assetImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccAssetImportStateID")
// @Testing(importStateIdAttribute="asset_id")
// @Testing(importIgnore="content;metadata")
func newAssetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assetResource{}, nil
}

type assetResource struct {
	framework.ResourceWithModel[assetResourceModel]
	framework.WithImportByIdentity
}

func (r *assetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_space_id": schema.StringAttribute{
				Description:   "Identifier of the Agent Space containing the asset. Changing this value requires replacement.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 2048)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"asset_id": schema.StringAttribute{
				Description:   "Identifier of the asset.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"asset_type": schema.StringAttribute{
				Description:   "Type of asset. Changing this value requires replacement. See https://docs.aws.amazon.com/devopsagent/latest/userguide/about-aws-devops-agent-managing-assets.html for supported types and their metadata requirements.",
				Required:      true,
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 64)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Description:   "Creation timestamp in RFC 3339 format.",
				CustomType:    timetypes.RFC3339Type{},
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"metadata": schema.StringAttribute{
				Description: "JSON document describing the asset. Required metadata fields depend on asset_type. Updates replace supplied top-level fields and retain omitted fields in AWS; removing a field from this document does not delete it from AWS. Changes to metadata alone do not re-upload content.",
				CustomType:  fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Optional:    true,
			},
			"metadata_all": schema.StringAttribute{
				Description: "Full JSON metadata returned by AWS, including retained fields and service-provided defaults.",
				CustomType:  fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp in RFC 3339 format.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			names.AttrVersion: schema.Int32Attribute{
				Description: "Current asset version.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrContent: schema.ListNestedBlock{
				Description: "Content to upload. Exactly one block is required.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[assetContentModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"file": schema.ListNestedBlock{
							Description: "Single file to add or replace.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[assetFileContentModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("zip"), path.MatchRelative().AtParent().AtName("source_url")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"metadata": schema.StringAttribute{
										Description: "JSON metadata for the file.",
										CustomType:  fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
										Optional:    true,
									},
									names.AttrPath: schema.StringAttribute{
										Description: "Path of the file within the asset.",
										Required:    true,
									},
								},
								Blocks: map[string]schema.Block{
									"body": schema.ListNestedBlock{
										Description: "File content.",
										CustomType:  fwtypes.NewListNestedObjectTypeOf[assetFileBodyModel](ctx),
										Validators:  []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtMost(1)},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"bytes": schema.StringAttribute{
													Optional:    true,
													Description: "Base64-encoded binary file content.",
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
							},
						},
						"source_url": schema.ListNestedBlock{
							Description: "Source URL to import.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[assetSourceURLModel](ctx),
							Validators:  []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrURL: schema.StringAttribute{
										Description: "Source URL from which to import asset content.",
										Required:    true,
									},
								},
							},
						},
						"zip": schema.ListNestedBlock{
							Description: "ZIP archive containing asset files.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[assetZipContentModel](ctx),
							Validators:  []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"zip_file": schema.StringAttribute{
										Required:    true,
										Description: "Base64-encoded ZIP archive. Use filebase64() to load a local ZIP file.",
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

func (r *assetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var plan assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input devopsagent.CreateAssetInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateAsset(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.AgentSpaceID.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Asset, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *assetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var state assetResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findAssetByTwoPartKey(ctx, conn, state.AgentSpaceID.ValueString(), state.AssetID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssetID.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *assetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)
	var plan, state assetResourceModel
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
		var input devopsagent.UpdateAssetInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...))
		if resp.Diagnostics.HasError() {
			return
		}
		input.AgentSpaceId = state.AgentSpaceID.ValueStringPointer()
		input.AssetId = state.AssetID.ValueStringPointer()
		out, err := conn.UpdateAsset(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssetID.ValueString())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Asset, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.MetadataAll = state.MetadataAll
		plan.UpdatedAt = state.UpdatedAt
		plan.Version = state.Version
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
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssetID.ValueString())
	}
}

func findAssetByTwoPartKey(ctx context.Context, conn *devopsagent.Client, agentSpaceID, assetID string) (*awstypes.Asset, error) {
	input := devopsagent.GetAssetInput{
		AgentSpaceId: aws.String(agentSpaceID),
		AssetId:      aws.String(assetID),
	}
	out, err := conn.GetAsset(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	// GetAsset returns AccessDeniedException once the parent Agent Space is deleted.
	if errs.IsA[*awstypes.AccessDeniedException](err) {
		if _, parentErr := findAgentSpaceByID(ctx, conn, agentSpaceID); retry.NotFound(parentErr) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil || out.Asset == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out.Asset, nil
}

func (r *assetResource) flatten(ctx context.Context, asset *awstypes.Asset, data *assetResourceModel) (diags diag.Diagnostics) {
	if asset == nil {
		smerr.AddError(ctx, &diags, tfresource.NewEmptyResultError(), smerr.ID, data.AssetID.ValueString())
		return diags
	}
	// Content and metadata are inputs; AWS returns merged metadata, not the original patch.
	diags.Append(flex.Flatten(ctx, asset, data)...)
	if diags.HasError() {
		return diags
	}
	if asset.Metadata == nil {
		data.MetadataAll = fwtypes.NewSmithyJSONNull[document.Interface]()
		return diags
	}
	metadata, err := tfsmithy.DocumentToJSONString(asset.Metadata)
	if err != nil {
		smerr.AddError(ctx, &diags, err, smerr.ID, data.AssetID.ValueString())
		return diags
	}
	data.MetadataAll = fwtypes.NewSmithyJSONValue(metadata, document.NewLazyDocument)
	return diags
}

var _ inttypes.ImportIDParser = assetImportID{}

type assetImportID struct{}

func (assetImportID) Parse(id string) (string, map[string]any, error) {
	agentSpaceID, assetID, found := strings.Cut(id, ",")
	if !found || agentSpaceID == "" || assetID == "" || strings.Contains(assetID, ",") {
		return "", nil, fmt.Errorf("unexpected format for import ID %q, expected AGENT_SPACE_ID,ASSET_ID", id)
	}
	return id, map[string]any{"agent_space_id": agentSpaceID, "asset_id": assetID}, nil
}

type assetResourceModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String                                       `tfsdk:"agent_space_id"`
	AssetID      types.String                                       `tfsdk:"asset_id"`
	AssetType    types.String                                       `tfsdk:"asset_type"`
	Content      fwtypes.ListNestedObjectValueOf[assetContentModel] `tfsdk:"content" autoflex:",noflatten"`
	CreatedAt    timetypes.RFC3339                                  `tfsdk:"created_at"`
	Metadata     fwtypes.SmithyJSON[document.Interface]             `tfsdk:"metadata" autoflex:",noflatten"`
	MetadataAll  fwtypes.SmithyJSON[document.Interface]             `tfsdk:"metadata_all" autoflex:"-"`
	UpdatedAt    timetypes.RFC3339                                  `tfsdk:"updated_at"`
	Version      types.Int32                                        `tfsdk:"version"`
}

type assetContentModel struct {
	File      fwtypes.ListNestedObjectValueOf[assetFileContentModel] `tfsdk:"file"`
	SourceURL fwtypes.ListNestedObjectValueOf[assetSourceURLModel]   `tfsdk:"source_url"`
	Zip       fwtypes.ListNestedObjectValueOf[assetZipContentModel]  `tfsdk:"zip"`
}

type assetFileContentModel struct {
	Body     fwtypes.ListNestedObjectValueOf[assetFileBodyModel] `tfsdk:"body"`
	Metadata fwtypes.SmithyJSON[document.Interface]              `tfsdk:"metadata"`
	Path     types.String                                        `tfsdk:"path"`
}

type assetFileBodyModel struct {
	Bytes types.String `tfsdk:"bytes"`
	Text  types.String `tfsdk:"text"`
}

type assetSourceURLModel struct {
	URL types.String `tfsdk:"url"`
}

type assetZipContentModel struct {
	ZipFile types.String `tfsdk:"zip_file"`
}

var (
	_ flex.Expander = assetContentModel{}
	_ flex.Expander = assetFileBodyModel{}
	_ flex.Expander = assetZipContentModel{}
)

func (m assetContentModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.File.IsNull():
		var content awstypes.AssetFileContent
		diags.Append(flex.Expand(ctx, m.File, &content)...)
		return &awstypes.AssetContentMemberFile{Value: content}, diags
	case !m.SourceURL.IsNull():
		var content awstypes.AssetSourceUrlContent
		diags.Append(flex.Expand(ctx, m.SourceURL, &content)...)
		return &awstypes.AssetContentMemberSourceUrl{Value: content}, diags
	case !m.Zip.IsNull():
		var content awstypes.AssetZipContent
		diags.Append(flex.Expand(ctx, m.Zip, &content)...)
		return &awstypes.AssetContentMemberZip{Value: content}, diags
	default:
		smerr.AddError(ctx, &diags, fmt.Errorf("exactly one of file, source_url, or zip must be configured"))
		return nil, diags
	}
}

func (m assetFileBodyModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !m.Text.IsNull() {
		return &awstypes.AssetFileBodyMemberText{Value: m.Text.ValueString()}, diags
	}
	if !m.Bytes.IsNull() {
		value, err := base64.StdEncoding.DecodeString(m.Bytes.ValueString())
		if err != nil {
			smerr.AddError(ctx, &diags, fmt.Errorf("decoding file body bytes: %w", err))
			return nil, diags
		}
		return &awstypes.AssetFileBodyMemberBytes{Value: value}, diags
	}
	smerr.AddError(ctx, &diags, fmt.Errorf("exactly one of bytes or text must be configured"))
	return nil, diags
}

func (m assetZipContentModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, err := base64.StdEncoding.DecodeString(m.ZipFile.ValueString())
	if err != nil {
		smerr.AddError(ctx, &diags, fmt.Errorf("decoding zip_file: %w", err))
		return nil, diags
	}
	return &awstypes.AssetZipContent{ZipFile: value}, diags
}
