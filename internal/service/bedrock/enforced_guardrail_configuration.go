// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_enforced_guardrail_configuration", name="Enforced Guardrail Configuration")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
func newEnforcedGuardrailConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &enforcedGuardrailConfigurationResource{}, nil
}

type enforcedGuardrailConfigurationResource struct {
	framework.ResourceWithModel[enforcedGuardrailConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *enforcedGuardrailConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			"config_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"guardrail_identifier": schema.StringAttribute{
				Required: true,
			},
			"guardrail_version": schema.StringAttribute{
				Required: true,
			},
			"guardrail_arn": schema.StringAttribute{
				Computed: true,
			},
			"guardrail_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwner: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ConfigurationOwner](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"model_enforcement": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelEnforcementModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"excluded_models": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							Required:    true,
							ElementType: types.StringType,
						},
						"included_models": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"selective_content_guarding": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[selectiveContentGuardingModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"messages": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SelectiveGuardingMode](),
							Optional:   true,
						},
						"system": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SelectiveGuardingMode](),
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *enforcedGuardrailConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().Region(ctx))

	r.putEnforcedGuardrailConfiguration(ctx, conn, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Read back the full configuration to populate computed attributes.
	output, err := findEnforcedGuardrailConfiguration(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	r.flattenOutput(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *enforcedGuardrailConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findEnforcedGuardrailConfiguration(ctx, conn)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	r.flattenOutput(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *enforcedGuardrailConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	r.putEnforcedGuardrailConfiguration(ctx, conn, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// Read back the full configuration to populate computed attributes.
	output, err := findEnforcedGuardrailConfiguration(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	r.flattenOutput(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *enforcedGuardrailConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	input := bedrock.DeleteEnforcedGuardrailConfigurationInput{
		ConfigId: data.ConfigID.ValueStringPointer(),
	}

	_, err := conn.DeleteEnforcedGuardrailConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}
}

func (r *enforcedGuardrailConfigurationResource) putEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client, data *enforcedGuardrailConfigurationResourceModel, diags *diag.Diagnostics) {
	var inferenceConfig awstypes.AccountEnforcedGuardrailInferenceInputConfiguration
	smerr.AddEnrich(ctx, diags, fwflex.Expand(ctx, data, &inferenceConfig))
	if diags.HasError() {
		return
	}

	input := bedrock.PutEnforcedGuardrailConfigurationInput{
		GuardrailInferenceConfig: &inferenceConfig,
	}

	// Pass configId on update.
	if !data.ConfigID.IsNull() && !data.ConfigID.IsUnknown() {
		input.ConfigId = data.ConfigID.ValueStringPointer()
	}

	_, err := conn.PutEnforcedGuardrailConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, diags, err, smerr.ID, data.ID.String())
	}
}

func (r *enforcedGuardrailConfigurationResource) flattenOutput(ctx context.Context, output *awstypes.AccountEnforcedGuardrailOutputConfiguration, data *enforcedGuardrailConfigurationResourceModel, diags *diag.Diagnostics) {
	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, output, data))
	if diags.HasError() {
		return
	}

	// Populate guardrail_identifier from guardrail_arn on read so that import works correctly.
	if data.GuardrailIdentifier.IsNull() || data.GuardrailIdentifier.ValueString() == "" {
		data.GuardrailIdentifier = types.StringPointerValue(output.GuardrailArn)
	}
}

func findEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client) (*awstypes.AccountEnforcedGuardrailOutputConfiguration, error) {
	input := bedrock.ListEnforcedGuardrailsConfigurationInput{}

	output, err := conn.ListEnforcedGuardrailsConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || len(output.GuardrailsConfig) == 0 {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return &output.GuardrailsConfig[0], nil
}

type enforcedGuardrailConfigurationResourceModel struct {
	framework.WithRegionModel
	ID                       types.String                                                   `tfsdk:"id"`
	ConfigID                 types.String                                                   `tfsdk:"config_id"`
	CreatedAt                timetypes.RFC3339                                              `tfsdk:"created_at"`
	CreatedBy                types.String                                                   `tfsdk:"created_by"`
	GuardrailArn             types.String                                                   `tfsdk:"guardrail_arn"`
	GuardrailId              types.String                                                   `tfsdk:"guardrail_id"`
	GuardrailIdentifier      types.String                                                   `tfsdk:"guardrail_identifier"`
	GuardrailVersion         types.String                                                   `tfsdk:"guardrail_version"`
	ModelEnforcement         fwtypes.ListNestedObjectValueOf[modelEnforcementModel]         `tfsdk:"model_enforcement"`
	Owner                    fwtypes.StringEnum[awstypes.ConfigurationOwner]                `tfsdk:"owner"`
	SelectiveContentGuarding fwtypes.ListNestedObjectValueOf[selectiveContentGuardingModel] `tfsdk:"selective_content_guarding"`
	UpdatedAt                timetypes.RFC3339                                              `tfsdk:"updated_at"`
	UpdatedBy                types.String                                                   `tfsdk:"updated_by"`
}

type modelEnforcementModel struct {
	ExcludedModels fwtypes.ListOfString `tfsdk:"excluded_models"`
	IncludedModels fwtypes.ListOfString `tfsdk:"included_models"`
}

type selectiveContentGuardingModel struct {
	Messages fwtypes.StringEnum[awstypes.SelectiveGuardingMode] `tfsdk:"messages"`
	System   fwtypes.StringEnum[awstypes.SelectiveGuardingMode] `tfsdk:"system"`
}
