// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_api_key_credential_provider", name="Api Key Credential Provider")
// @Tags(identifierAttribute="credential_provider_arn")
// @Testing(tagsTest=false)
func newAPIKeyCredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &apiKeyCredentialProviderResource{}
	return r, nil
}

type apiKeyCredentialProviderResource struct {
	framework.ResourceWithModel[apiKeyCredentialProviderResourceModel]
}

func (r *apiKeyCredentialProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("api_key"),
						path.MatchRoot("api_key_wo"),
						path.MatchRoot("api_key_secret_config"),
					),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("api_key_wo"),
					}...),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("api_key_wo")),
				},
			},
			"api_key_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("api_key"),
						path.MatchRoot("api_key_wo"),
						path.MatchRoot("api_key_secret_config"),
					),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("api_key_wo_version"),
					}...),
				},
			},
			"api_key_wo_version": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRoot("api_key_wo"),
					}...),
				},
			},
			"api_key_secret_arn": framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
			"api_key_secret_source": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credential_provider_arn": framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"api_key_secret_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[secretReferenceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.AlsoRequires(path.MatchRoot("api_key_secret_source")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"secret_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
						"json_key": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
						},
					},
				},
			},
		},
	}
}

func (r *apiKeyCredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Prefer write-only value. It's only in Config, not Plan.
	if !config.APIKeyWO.IsNull() {
		input.ApiKey = fwflex.StringFromFramework(ctx, config.APIKeyWO)
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateApiKeyCredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *apiKeyCredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findAPIKeyCredentialProviderByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *apiKeyCredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state, config apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
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
		name := fwflex.StringValueFromFramework(ctx, plan.Name)
		var input bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Prefer write-only value. It's only in Config, not Plan.
		if !config.APIKeyWO.IsNull() {
			input.ApiKey = fwflex.StringFromFramework(ctx, config.APIKeyWO)
		}

		_, err := conn.UpdateApiKeyCredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *apiKeyCredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeleteApiKeyCredentialProviderInput{
		Name: aws.String(name),
	}
	_, err := conn.DeleteApiKeyCredentialProvider(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *apiKeyCredentialProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func (r *apiKeyCredentialProviderResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var config, state apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if config.APIKeySecretSource.IsUnknown() || state.APIKeySecretSource.IsNull() || state.APIKeySecretSource.IsUnknown() {
		return
	}

	// Derive the effective secret source from configuration: api_key/api_key_wo
	// imply MANAGED, api_key_secret_config implies EXTERNAL. This cannot rely on
	// the planned value alone: api_key_secret_source is Optional+Computed, so when
	// it's absent from configuration, UseStateForUnknown fills it from state.
	var effective awstypes.SecretSourceType
	switch {
	case !config.APIKeySecretSource.IsNull():
		effective = config.APIKeySecretSource.ValueEnum()
	case !config.APIKeySecretConfig.IsNull():
		effective = awstypes.SecretSourceTypeExternal
	case !config.APIKey.IsNull() || !config.APIKeyWO.IsNull():
		effective = awstypes.SecretSourceTypeManaged
	default:
		return
	}

	// The API rejects switching the secret source between MANAGED and EXTERNAL in place.
	if effective != state.APIKeySecretSource.ValueEnum() {
		// Overwrite the UseStateForUnknown-filled plan value; without a value diff,
		// Terraform Core ignores RequiresReplace.
		smerr.AddEnrich(ctx, &response.Diagnostics, response.Plan.SetAttribute(ctx, path.Root("api_key_secret_source"), fwtypes.StringEnumValue(effective)))
		if response.Diagnostics.HasError() {
			return
		}
		response.RequiresReplace = append(response.RequiresReplace, path.Root("api_key_secret_source"))
	}
}

func (r *apiKeyCredentialProviderResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data apiKeyCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if data.APIKeySecretSource.IsUnknown() || data.APIKeySecretSource.IsNull() {
		return
	}

	switch data.APIKeySecretSource.ValueEnum() {
	case awstypes.SecretSourceTypeExternal:
		if !data.APIKey.IsNull() {
			response.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(path.Root("api_key"), path.Root("api_key_secret_source"), string(awstypes.SecretSourceTypeExternal)))
		}
		if !data.APIKeyWO.IsNull() {
			response.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(path.Root("api_key_wo"), path.Root("api_key_secret_source"), string(awstypes.SecretSourceTypeExternal)))
		}
		if data.APIKeySecretConfig.IsNull() {
			response.Diagnostics.Append(fwdiag.NewAttributeRequiredWhenError(path.Root("api_key_secret_config"), path.Root("api_key_secret_source"), string(awstypes.SecretSourceTypeExternal)))
		}
	case awstypes.SecretSourceTypeManaged:
		if !data.APIKeySecretConfig.IsNull() {
			response.Diagnostics.Append(fwdiag.NewAttributeConflictsWhenError(path.Root("api_key_secret_config"), path.Root("api_key_secret_source"), string(awstypes.SecretSourceTypeManaged)))
		}
	}
}

func findAPIKeyCredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetApiKeyCredentialProviderInput{
		Name: aws.String(name),
	}

	return findAPIKeyCredentialProvider(ctx, conn, &input)
}

func findAPIKeyCredentialProvider(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetApiKeyCredentialProviderInput) (*bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput, error) {
	out, err := conn.GetApiKeyCredentialProvider(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type apiKeyCredentialProviderResourceModel struct {
	framework.WithRegionModel
	APIKey                types.String                                          `tfsdk:"api_key"`
	APIKeySecretARN       fwtypes.ListNestedObjectValueOf[secretModel]          `tfsdk:"api_key_secret_arn"`
	APIKeySecretConfig    fwtypes.ListNestedObjectValueOf[secretReferenceModel] `tfsdk:"api_key_secret_config"`
	APIKeySecretSource    fwtypes.StringEnum[awstypes.SecretSourceType]         `tfsdk:"api_key_secret_source"`
	APIKeyWO              types.String                                          `tfsdk:"api_key_wo"`
	APIKeyWOVersion       types.Int64                                           `tfsdk:"api_key_wo_version"`
	CredentialProviderARN types.String                                          `tfsdk:"credential_provider_arn"`
	Name                  types.String                                          `tfsdk:"name"`
	Tags                  tftags.Map                                            `tfsdk:"tags"`
	TagsAll               tftags.Map                                            `tfsdk:"tags_all"`
}

type secretModel struct {
	SecretARN fwtypes.ARN `tfsdk:"secret_arn"`
}

type secretReferenceModel struct {
	SecretID types.String `tfsdk:"secret_id"`
	JSONKey  types.String `tfsdk:"json_key"`
}
