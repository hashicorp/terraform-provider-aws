// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

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
			"owner": schema.StringAttribute{
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
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	response.Diagnostics.Append(r.putEnforcedGuardrailConfiguration(ctx, conn, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().Region(ctx))

	// Read back the full configuration to populate computed attributes.
	output, err := findEnforcedGuardrailConfiguration(ctx, conn)
	if err != nil {
		response.Diagnostics.AddError("reading Bedrock Enforced Guardrail Configuration after create", err.Error())
		return
	}

	response.Diagnostics.Append(r.flattenOutput(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *enforcedGuardrailConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findEnforcedGuardrailConfiguration(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Enforced Guardrail Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(r.flattenOutput(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *enforcedGuardrailConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	response.Diagnostics.Append(r.putEnforcedGuardrailConfiguration(ctx, conn, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read back the full configuration to populate computed attributes.
	output, err := findEnforcedGuardrailConfiguration(ctx, conn)
	if err != nil {
		response.Diagnostics.AddError("reading Bedrock Enforced Guardrail Configuration after update", err.Error())
		return
	}

	response.Diagnostics.Append(r.flattenOutput(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *enforcedGuardrailConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data enforcedGuardrailConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteEnforcedGuardrailConfiguration(ctx, &bedrock.DeleteEnforcedGuardrailConfigurationInput{
		ConfigId: data.ConfigID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Enforced Guardrail Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *enforcedGuardrailConfigurationResource) putEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client, data *enforcedGuardrailConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	inferenceConfig := &awstypes.AccountEnforcedGuardrailInferenceInputConfiguration{
		GuardrailIdentifier: data.GuardrailIdentifier.ValueStringPointer(),
		GuardrailVersion:    data.GuardrailVersion.ValueStringPointer(),
	}

	// Expand ModelEnforcement if set.
	if !data.ModelEnforcement.IsNull() && !data.ModelEnforcement.IsUnknown() {
		modelEnforcementData, d := data.ModelEnforcement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if modelEnforcementData != nil {
			var me awstypes.ModelEnforcement
			diags.Append(fwflex.Expand(ctx, modelEnforcementData, &me)...)
			if diags.HasError() {
				return diags
			}
			inferenceConfig.ModelEnforcement = &me
		}
	}

	// Expand SelectiveContentGuarding if set.
	if !data.SelectiveContentGuarding.IsNull() && !data.SelectiveContentGuarding.IsUnknown() {
		scgData, d := data.SelectiveContentGuarding.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if scgData != nil {
			var scg awstypes.SelectiveContentGuarding
			diags.Append(fwflex.Expand(ctx, scgData, &scg)...)
			if diags.HasError() {
				return diags
			}
			inferenceConfig.SelectiveContentGuarding = &scg
		}
	}

	input := &bedrock.PutEnforcedGuardrailConfigurationInput{
		GuardrailInferenceConfig: inferenceConfig,
	}

	// Pass configId on update.
	if !data.ConfigID.IsNull() && !data.ConfigID.IsUnknown() {
		input.ConfigId = data.ConfigID.ValueStringPointer()
	}

	_, err := conn.PutEnforcedGuardrailConfiguration(ctx, input)
	if err != nil {
		diags.AddError("putting Bedrock Enforced Guardrail Configuration", err.Error())
		return diags
	}

	return diags
}

func (r *enforcedGuardrailConfigurationResource) flattenOutput(ctx context.Context, output *awstypes.AccountEnforcedGuardrailOutputConfiguration, data *enforcedGuardrailConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ConfigID = types.StringPointerValue(output.ConfigId)
	data.GuardrailArn = types.StringPointerValue(output.GuardrailArn)
	data.GuardrailId = types.StringPointerValue(output.GuardrailId)
	data.GuardrailVersion = types.StringPointerValue(output.GuardrailVersion)

	// Populate guardrail_identifier from guardrail_arn on read so that import works correctly.
	if data.GuardrailIdentifier.IsNull() || data.GuardrailIdentifier.ValueString() == "" {
		data.GuardrailIdentifier = types.StringPointerValue(output.GuardrailArn)
	}
	data.CreatedAt = fwflex.TimeToFramework(ctx, output.CreatedAt)
	data.CreatedBy = types.StringPointerValue(output.CreatedBy)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, output.UpdatedAt)
	data.UpdatedBy = types.StringPointerValue(output.UpdatedBy)
	data.Owner = fwtypes.StringEnumValue(output.Owner)

	if output.ModelEnforcement != nil {
		var me modelEnforcementModel
		diags.Append(fwflex.Flatten(ctx, output.ModelEnforcement, &me)...)
		if diags.HasError() {
			return diags
		}
		data.ModelEnforcement = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &me)
	} else {
		data.ModelEnforcement = fwtypes.NewListNestedObjectValueOfNull[modelEnforcementModel](ctx)
	}

	if output.SelectiveContentGuarding != nil {
		var scg selectiveContentGuardingModel
		diags.Append(fwflex.Flatten(ctx, output.SelectiveContentGuarding, &scg)...)
		if diags.HasError() {
			return diags
		}
		data.SelectiveContentGuarding = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &scg)
	} else {
		data.SelectiveContentGuarding = fwtypes.NewListNestedObjectValueOfNull[selectiveContentGuardingModel](ctx)
	}

	return diags
}

func findEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client) (*awstypes.AccountEnforcedGuardrailOutputConfiguration, error) {
	input := &bedrock.ListEnforcedGuardrailsConfigurationInput{}

	output, err := conn.ListEnforcedGuardrailsConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.GuardrailsConfig) == 0 {
		return nil, tfresource.NewEmptyResultError()
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
