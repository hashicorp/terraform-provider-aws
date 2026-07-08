// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_configuration_bundle", name="Configuration Bundle")
// @Tags(identifierAttribute="bundle_arn")
// @IdentityAttribute("bundle_id", identityDuplicateAttributes="bundle_id")
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="bundle_id")
func newConfigurationBundleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &configurationBundleResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type configurationBundleResource struct {
	framework.ResourceWithModel[configurationBundleResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *configurationBundleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"branch_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_/-]{0,127}$`), ""),
				},
			},
			"bundle_arn": framework.ARNAttributeComputedOnly(),
			"bundle_id":  framework.IDAttribute(),
			"bundle_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,99}$`), ""),
				},
			},
			"commit_message": schema.StringAttribute{
				Optional: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws(|-cn|-us-gov):kms:[a-zA-Z0-9-]*:[0-9]{12}:key/[a-zA-Z0-9-]{36}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			// version_id and updated_at advance on every update; NO UseStateForUnknown
			// (they must stay unknown-in-plan and be re-read after apply).
			"version_id": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"lineage_metadata": framework.ResourceComputedListOfObjectsAttribute[versionLineageMetadataModel](ctx),
		},
		Blocks: map[string]schema.Block{
			"component": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[componentConfigurationModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"component_identifier": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								// Length bounds the identifier to 2048; the SDK's {0,2047}
								// quantifier exceeds Go regexp's 1000 repeat limit, so the
								// upper bound is enforced by LengthBetween, not the regex.
								stringvalidator.LengthBetween(1, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_:/.\-]*$`), ""),
							},
						},
						names.AttrConfiguration: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
						},
					},
				},
			},
			"created_by": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[versionCreatedBySourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *configurationBundleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data configurationBundleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateConfigurationBundleInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Components = data.expandComponents(ctx, &response.Diagnostics)
	input.Tags = getTagsIn(ctx)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateConfigurationBundle(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.BundleName.ValueString())
		return
	}

	data.BundleID = fwflex.StringToFramework(ctx, out.BundleId)

	bundle, err := findConfigurationBundleByID(ctx, conn, data.BundleID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.BundleID.ValueString())
		return
	}

	// Re-hydrate computed/read-back fields WITHOUT clobbering the write-only inputs
	// (branch_name, commit_message, created_by are not returned at the top level of
	// Get — they surface only inside lineage_metadata).
	smerr.AddEnrich(ctx, &response.Diagnostics, data.flattenReadBack(ctx, bundle))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *configurationBundleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data configurationBundleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	out, err := findConfigurationBundleByID(ctx, conn, data.BundleID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.BundleID.ValueString())
		return
	}

	// branch_name/commit_message/created_by are write-only top-level inputs (not in
	// Get output); flattenReadBack leaves the prior state values untouched.
	smerr.AddEnrich(ctx, &response.Diagnostics, data.flattenReadBack(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *configurationBundleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state configurationBundleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateConfigurationBundleInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}
		input.BundleId = fwflex.StringFromFramework(ctx, plan.BundleID)
		input.Components = plan.expandComponents(ctx, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
		// Versioning is git-like: an update creates a new version whose parent is
		// the current version. parentVersionIds must not be null, and the service
		// requires a commitMessage whenever components are (re)sent on update.
		input.ParentVersionIds = []string{state.VersionID.ValueString()}
		if input.CommitMessage == nil {
			input.CommitMessage = aws.String("Updated via Terraform")
		}

		if _, err := conn.UpdateConfigurationBundle(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.BundleID.ValueString())
			return
		}
	}

	// Always re-read so the computed values that advance on update (version_id,
	// updated_at, lineage_metadata) are known after apply — including tags-only
	// updates where diff.HasChanges() is false but the plan still marks them unknown.
	bundle, err := findConfigurationBundleByID(ctx, conn, plan.BundleID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.BundleID.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flattenReadBack(ctx, bundle))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *configurationBundleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data configurationBundleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.DeleteConfigurationBundleInput{
		BundleId: fwflex.StringFromFramework(ctx, data.BundleID),
	}
	if _, err := conn.DeleteConfigurationBundle(ctx, &input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.BundleID.ValueString())
		return
	}
}

// expandComponents builds the SDK component map from the `component` set blocks.
// AutoFlex can't build a map keyed by a nested attribute, so it's done by hand.
func (m configurationBundleResourceModel) expandComponents(ctx context.Context, diags *diag.Diagnostics) map[string]awstypes.ComponentConfiguration {
	if m.Components.IsNull() || m.Components.IsUnknown() {
		return nil
	}
	elems, d := m.Components.ToSlice(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() {
		return nil
	}
	out := make(map[string]awstypes.ComponentConfiguration, len(elems))
	for _, e := range elems {
		id := e.ComponentIdentifier.ValueString()
		var cc awstypes.ComponentConfiguration
		if !e.Configuration.IsNull() && !e.Configuration.IsUnknown() {
			doc, err := tfsmithy.DocumentFromJSONString(e.Configuration.ValueString(), document.NewLazyDocument)
			if err != nil {
				diags.AddError("creating Smithy document", err.Error())
				return nil
			}
			cc.Configuration = doc
		}
		out[id] = cc
	}
	return out
}

// flattenReadBack copies the round-tripping fields (components, computed metadata,
// tags) from a Get output into the model, WITHOUT touching the write-only
// top-level inputs branch_name/commit_message/created_by (absent from Get).
func (m *configurationBundleResourceModel) flattenReadBack(ctx context.Context, out *bedrockagentcorecontrol.GetConfigurationBundleOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	m.BundleARN = fwflex.StringToFramework(ctx, out.BundleArn)
	m.BundleID = fwflex.StringToFramework(ctx, out.BundleId)
	m.BundleName = fwflex.StringToFramework(ctx, out.BundleName)
	m.Description = fwflex.StringToFramework(ctx, out.Description)
	if out.KmsKeyArn != nil {
		m.KMSKeyARN = fwtypes.ARNValue(aws.ToString(out.KmsKeyArn))
	} else {
		m.KMSKeyARN = fwtypes.ARNNull()
	}
	m.VersionID = fwflex.StringToFramework(ctx, out.VersionId)
	m.CreatedAt = timetypes.NewRFC3339TimePointerValue(out.CreatedAt)
	m.UpdatedAt = timetypes.NewRFC3339TimePointerValue(out.UpdatedAt)

	components := make([]*componentConfigurationModel, 0, len(out.Components))
	for id, cc := range out.Components {
		cm := &componentConfigurationModel{
			ComponentIdentifier: fwflex.StringValueToFramework(ctx, id),
		}
		if cc.Configuration != nil {
			s, err := tfsmithy.DocumentToJSONString(cc.Configuration)
			if err != nil {
				diags.AddError("reading Smithy document", err.Error())
				return diags
			}
			cm.Configuration = fwtypes.NewSmithyJSONValue(s, document.NewLazyDocument)
		} else {
			cm.Configuration = fwtypes.NewSmithyJSONNull[document.Interface]()
		}
		components = append(components, cm)
	}
	m.Components = fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, components)

	if out.LineageMetadata != nil {
		var lm versionLineageMetadataModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out.LineageMetadata, &lm))
		if diags.HasError() {
			return diags
		}
		m.LineageMetadata = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &lm)
	} else {
		m.LineageMetadata = fwtypes.NewListNestedObjectValueOfNull[versionLineageMetadataModel](ctx)
	}

	return diags
}

func findConfigurationBundleByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, bundleID string) (*bedrockagentcorecontrol.GetConfigurationBundleOutput, error) {
	input := bedrockagentcorecontrol.GetConfigurationBundleInput{
		BundleId: aws.String(bundleID),
	}

	out, err := conn.GetConfigurationBundle(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

type configurationBundleResourceModel struct {
	framework.WithRegionModel
	BranchName      types.String                                                 `tfsdk:"branch_name"`
	BundleARN       types.String                                                 `tfsdk:"bundle_arn"`
	BundleID        types.String                                                 `tfsdk:"bundle_id"`
	BundleName      types.String                                                 `tfsdk:"bundle_name"`
	CommitMessage   types.String                                                 `tfsdk:"commit_message"`
	Components      fwtypes.SetNestedObjectValueOf[componentConfigurationModel]  `tfsdk:"component" autoflex:"-"`
	CreatedAt       timetypes.RFC3339                                            `tfsdk:"created_at"`
	CreatedBy       fwtypes.ListNestedObjectValueOf[versionCreatedBySourceModel] `tfsdk:"created_by"`
	Description     types.String                                                 `tfsdk:"description"`
	KMSKeyARN       fwtypes.ARN                                                  `tfsdk:"kms_key_arn"`
	LineageMetadata fwtypes.ListNestedObjectValueOf[versionLineageMetadataModel] `tfsdk:"lineage_metadata" autoflex:"-"`
	Tags            tftags.Map                                                   `tfsdk:"tags"`
	TagsAll         tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                               `tfsdk:"timeouts"`
	UpdatedAt       timetypes.RFC3339                                            `tfsdk:"updated_at"`
	VersionID       types.String                                                 `tfsdk:"version_id"`
}

type componentConfigurationModel struct {
	ComponentIdentifier types.String                           `tfsdk:"component_identifier"`
	Configuration       fwtypes.SmithyJSON[document.Interface] `tfsdk:"configuration"`
}

type versionCreatedBySourceModel struct {
	ARN  types.String `tfsdk:"arn"`
	Name types.String `tfsdk:"name"`
}

type versionLineageMetadataModel struct {
	BranchName       types.String                                                 `tfsdk:"branch_name"`
	CommitMessage    types.String                                                 `tfsdk:"commit_message"`
	CreatedBy        fwtypes.ListNestedObjectValueOf[versionCreatedBySourceModel] `tfsdk:"created_by"`
	ParentVersionIDs fwtypes.ListOfString                                         `tfsdk:"parent_version_ids"`
}
