// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_payment_credential_provider", name="Payment Credential Provider")
// @Tags(identifierAttribute="credential_provider_arn")
// @Testing(tagsTest=false)
func newPaymentCredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &paymentCredentialProviderResource{}, nil
}

type paymentCredentialProviderResource struct {
	framework.ResourceWithModel[paymentCredentialProviderResourceModel]
}

func (r *paymentCredentialProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"credential_provider_arn": framework.ARNAttributeComputedOnly(),
			"credential_provider_vendor": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PaymentCredentialProviderVendorType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[paymentProviderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"coinbase_cdp_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[coinbaseCdpConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("coinbase_cdp_configuration"),
									path.MatchRelative().AtParent().AtName("stripe_privy_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_key_id": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"api_key_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									"api_key_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
									},
									"wallet_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									"wallet_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"api_key_secret_config": paymentSecretReferenceBlock(ctx),
									"wallet_secret_config":  paymentSecretReferenceBlock(ctx),
								},
							},
						},
						"stripe_privy_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[stripePrivyConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("coinbase_cdp_configuration"),
									path.MatchRelative().AtParent().AtName("stripe_privy_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"app_id": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"app_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									"app_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
									},
									"authorization_id": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"authorization_private_key": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									"authorization_private_key_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"app_secret_config":                paymentSecretReferenceBlock(ctx),
									"authorization_private_key_config": paymentSecretReferenceBlock(ctx),
								},
							},
						},
					},
				},
			},
		},
	}
}

func paymentSecretReferenceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[paymentSecretReferenceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"json_key": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 128),
					},
				},
				"secret_id": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 2048),
					},
				},
			},
		},
	}
}

func (r *paymentCredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data paymentCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreatePaymentCredentialProviderInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNameSuffix("Input")))
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePaymentCredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	data.CredentialProviderARN = fwflex.StringToFramework(ctx, out.CredentialProviderArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentCredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data paymentCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findPaymentCredentialProviderByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// The provider configuration secrets are write-only; the API returns secret
	// ARNs rather than the submitted values, so preserve the configured block.
	providerConfiguration := data.ProviderConfiguration
	data.CredentialProviderARN = fwflex.StringToFramework(ctx, out.CredentialProviderArn)
	data.CredentialProviderVendor = fwtypes.StringEnumValue(out.CredentialProviderVendor)
	data.Name = fwflex.StringToFramework(ctx, out.Name)
	data.ProviderConfiguration = providerConfiguration

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentCredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old paymentCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, new.Name)
		var input bedrockagentcorecontrol.UpdatePaymentCredentialProviderInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, fwflex.WithFieldNameSuffix("Input")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdatePaymentCredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *paymentCredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data paymentCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeletePaymentCredentialProviderInput{
		Name: aws.String(name),
	}

	_, err := conn.DeletePaymentCredentialProvider(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *paymentCredentialProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func findPaymentCredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetPaymentCredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetPaymentCredentialProviderInput{
		Name: aws.String(name),
	}

	out, err := conn.GetPaymentCredentialProvider(ctx, &input)

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

type paymentCredentialProviderResourceModel struct {
	framework.WithRegionModel
	CredentialProviderARN    types.String                                                       `tfsdk:"credential_provider_arn"`
	CredentialProviderVendor fwtypes.StringEnum[awstypes.PaymentCredentialProviderVendorType]   `tfsdk:"credential_provider_vendor"`
	Name                     types.String                                                       `tfsdk:"name"`
	ProviderConfiguration    fwtypes.ListNestedObjectValueOf[paymentProviderConfigurationModel] `tfsdk:"provider_configuration"`
	Tags                     tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                         `tfsdk:"tags_all"`
}

type paymentProviderConfigurationModel struct {
	CoinbaseCDPConfiguration fwtypes.ListNestedObjectValueOf[coinbaseCdpConfigurationModel] `tfsdk:"coinbase_cdp_configuration"`
	StripePrivyConfiguration fwtypes.ListNestedObjectValueOf[stripePrivyConfigurationModel] `tfsdk:"stripe_privy_configuration"`
}

type coinbaseCdpConfigurationModel struct {
	APIKeyID           types.String                                                 `tfsdk:"api_key_id"`
	APIKeySecret       types.String                                                 `tfsdk:"api_key_secret"`
	APIKeySecretConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"api_key_secret_config"`
	APIKeySecretSource fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"api_key_secret_source"`
	WalletSecret       types.String                                                 `tfsdk:"wallet_secret"`
	WalletSecretConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"wallet_secret_config"`
	WalletSecretSource fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"wallet_secret_source"`
}

type stripePrivyConfigurationModel struct {
	AppID                         types.String                                                 `tfsdk:"app_id"`
	AppSecret                     types.String                                                 `tfsdk:"app_secret"`
	AppSecretConfig               fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"app_secret_config"`
	AppSecretSource               fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"app_secret_source"`
	AuthorizationID               types.String                                                 `tfsdk:"authorization_id"`
	AuthorizationPrivateKey       types.String                                                 `tfsdk:"authorization_private_key"`
	AuthorizationPrivateKeyConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"authorization_private_key_config"`
	AuthorizationPrivateKeySource fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"authorization_private_key_source"`
}

type paymentSecretReferenceModel struct {
	JSONKey  types.String `tfsdk:"json_key"`
	SecretID types.String `tfsdk:"secret_id"`
}

var (
	_ fwflex.Expander  = paymentProviderConfigurationModel{}
	_ fwflex.Flattener = &paymentProviderConfigurationModel{}
)

func (m paymentProviderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CoinbaseCDPConfiguration.IsNull():
		data, d := m.CoinbaseCDPConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PaymentProviderConfigurationInputMemberCoinbaseCdpConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags

	case !m.StripePrivyConfiguration.IsNull():
		data, d := m.StripePrivyConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PaymentProviderConfigurationInputMemberStripePrivyConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

func (m *paymentProviderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// The provider configuration is write-only on input; on read the API returns
	// secret ARNs instead of the submitted secret values. The configured block is
	// preserved by the resource's Read, so flattening here is intentionally a no-op
	// beyond recording which variant is present.
	switch v.(type) {
	case awstypes.PaymentProviderConfigurationOutputMemberCoinbaseCdpConfiguration,
		awstypes.PaymentProviderConfigurationOutputMemberStripePrivyConfiguration:
		// no-op: configured values are authoritative
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("payment provider configuration flatten: %T", v),
		)
	}
	return diags
}
