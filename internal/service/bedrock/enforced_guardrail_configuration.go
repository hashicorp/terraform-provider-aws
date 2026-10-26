// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_enforced_guardrail_configuration", name="Enforced Guardrail Configuration")
// @IdentityAttribute("config_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="config_id")
// @Testing(serialize=true)
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
			"config_id": framework.IDAttribute(),
			"guardrail_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"guardrail_id": schema.StringAttribute{
				Computed: true,
			},
			"guardrail_version": schema.StringAttribute{
				Required: true,
			},
			names.AttrOwner: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ConfigurationOwner](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
						},
						"included_models": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
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

	configID, diags := r.putEnforcedGuardrailConfiguration(ctx, conn, &data)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	// Read back the full configuration to populate computed attributes.
	id := aws.ToString(configID)
	output, err := findEnforcedGuardrailConfigurationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
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

	id := fwflex.StringValueFromFramework(ctx, data.ConfigID)
	output, err := findEnforcedGuardrailConfigurationByID(ctx, conn, id)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
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

	id := fwflex.StringValueFromFramework(ctx, data.ConfigID)
	_, diags := r.putEnforcedGuardrailConfiguration(ctx, conn, &data)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	// Read back the full configuration to populate computed attributes.
	output, err := findEnforcedGuardrailConfigurationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
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

	id := fwflex.StringValueFromFramework(ctx, data.ConfigID)
	input := bedrock.DeleteEnforcedGuardrailConfigurationInput{
		ConfigId: aws.String(id),
	}

	_, err := conn.DeleteEnforcedGuardrailConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *enforcedGuardrailConfigurationResource) flatten(ctx context.Context, cfg *awstypes.AccountEnforcedGuardrailOutputConfiguration, data *enforcedGuardrailConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, cfg, data)...)
	return diags
}

func (r *enforcedGuardrailConfigurationResource) putEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client, data *enforcedGuardrailConfigurationResourceModel) (*string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var inferenceConfig awstypes.AccountEnforcedGuardrailInferenceInputConfiguration
	diags.Append(fwflex.Expand(ctx, data, &inferenceConfig)...)
	if diags.HasError() {
		return nil, diags
	}
	inferenceConfig.GuardrailIdentifier = fwflex.StringFromFramework(ctx, data.GuardrailARN)

	input := bedrock.PutEnforcedGuardrailConfigurationInput{
		GuardrailInferenceConfig: &inferenceConfig,
	}
	// Pass configId on update.
	if configID := fwflex.StringValueFromFramework(ctx, data.ConfigID); configID != "" {
		input.ConfigId = aws.String(configID)
	}
	output, err := conn.PutEnforcedGuardrailConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &diags, err)
		return nil, diags
	}

	return output.ConfigId, diags
}

func findEnforcedGuardrailConfigurationByID(ctx context.Context, conn *bedrock.Client, id string) (*awstypes.AccountEnforcedGuardrailOutputConfiguration, error) {
	var input bedrock.ListEnforcedGuardrailsConfigurationInput
	return findEnforcedGuardrailConfiguration(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.AccountEnforcedGuardrailOutputConfiguration) bool {
		return aws.ToString(v.ConfigId) == id
	}))
}

func findEnforcedGuardrailConfiguration(ctx context.Context, conn *bedrock.Client, input *bedrock.ListEnforcedGuardrailsConfigurationInput, optFns ...tfslices.FinderOptionsFunc[awstypes.AccountEnforcedGuardrailOutputConfiguration]) (*awstypes.AccountEnforcedGuardrailOutputConfiguration, error) {
	output, err := findEnforcedGuardrailConfigurations(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEnforcedGuardrailConfigurations(ctx context.Context, conn *bedrock.Client, input *bedrock.ListEnforcedGuardrailsConfigurationInput, optFns ...tfslices.FinderOptionsFunc[awstypes.AccountEnforcedGuardrailOutputConfiguration]) ([]awstypes.AccountEnforcedGuardrailOutputConfiguration, error) {
	return tfslices.CollectAndConcatWithError(listEnforcedGuardrailConfigurationPages(ctx, conn, input), optFns...)
}

func listEnforcedGuardrailConfigurationPages(ctx context.Context, conn *bedrock.Client, input *bedrock.ListEnforcedGuardrailsConfigurationInput, optFns ...func(*bedrock.Options)) iter.Seq2[[]awstypes.AccountEnforcedGuardrailOutputConfiguration, error] {
	return func(yield func([]awstypes.AccountEnforcedGuardrailOutputConfiguration, error) bool) {
		pages := bedrock.NewListEnforcedGuardrailsConfigurationPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Bedrock Enforced Guardrail Configurations: %w", err))
				return
			}

			if !yield(page.GuardrailsConfig, nil) {
				return
			}
		}
	}
}

type enforcedGuardrailConfigurationResourceModel struct {
	framework.WithRegionModel
	ConfigID                 types.String                                                   `tfsdk:"config_id"`
	GuardrailARN             fwtypes.ARN                                                    `tfsdk:"guardrail_arn"`
	GuardrailID              types.String                                                   `tfsdk:"guardrail_id"`
	GuardrailVersion         types.String                                                   `tfsdk:"guardrail_version"`
	ModelEnforcement         fwtypes.ListNestedObjectValueOf[modelEnforcementModel]         `tfsdk:"model_enforcement"`
	Owner                    fwtypes.StringEnum[awstypes.ConfigurationOwner]                `tfsdk:"owner"`
	SelectiveContentGuarding fwtypes.ListNestedObjectValueOf[selectiveContentGuardingModel] `tfsdk:"selective_content_guarding"`
}

type modelEnforcementModel struct {
	ExcludedModels fwtypes.ListOfString `tfsdk:"excluded_models"`
	IncludedModels fwtypes.ListOfString `tfsdk:"included_models"`
}

type selectiveContentGuardingModel struct {
	Messages fwtypes.StringEnum[awstypes.SelectiveGuardingMode] `tfsdk:"messages"`
	System   fwtypes.StringEnum[awstypes.SelectiveGuardingMode] `tfsdk:"system"`
}
