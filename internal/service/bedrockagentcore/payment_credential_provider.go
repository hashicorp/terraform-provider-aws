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

// @FrameworkResource("aws_bedrockagentcore_payment_credential_provider", name="Payment Credential Provider")
// @Tags(identifierAttribute="credential_provider_arn")
// @Testing(tagsTest=false)
func newPaymentCredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &paymentCredentialProviderResource{}, nil
}

type paymentCredentialProviderResource struct {
	framework.ResourceWithModel[paymentCredentialProviderResourceModel]
}

var _ resource.ResourceWithValidateConfig = &paymentCredentialProviderResource{}

func (r *paymentCredentialProviderResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data paymentCredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}
	if data.ProviderConfiguration.IsNull() || data.ProviderConfiguration.IsUnknown() {
		return
	}
	config, d := data.ProviderConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() || config == nil {
		return
	}

	providerPath := path.Root("provider_configuration").AtListIndex(0)
	if !config.CoinbaseCDPConfiguration.IsNull() && !config.CoinbaseCDPConfiguration.IsUnknown() {
		if cb, d := config.CoinbaseCDPConfiguration.ToPtr(ctx); !d.HasError() && cb != nil {
			p := providerPath.AtName("coinbase_cdp_configuration").AtListIndex(0)
			validateExternalSecretRequiresConfig(&response.Diagnostics, cb.APIKeySecretSource, cb.APIKeySecretConfig, "api_key_secret_source", p.AtName("api_key_secret_config"))
			validateExternalSecretRequiresConfig(&response.Diagnostics, cb.WalletSecretSource, cb.WalletSecretConfig, "wallet_secret_source", p.AtName("wallet_secret_config"))
		}
	}
	if !config.StripePrivyConfiguration.IsNull() && !config.StripePrivyConfiguration.IsUnknown() {
		if sp, d := config.StripePrivyConfiguration.ToPtr(ctx); !d.HasError() && sp != nil {
			p := providerPath.AtName("stripe_privy_configuration").AtListIndex(0)
			validateExternalSecretRequiresConfig(&response.Diagnostics, sp.AppSecretSource, sp.AppSecretConfig, "app_secret_source", p.AtName("app_secret_config"))
			validateExternalSecretRequiresConfig(&response.Diagnostics, sp.AuthorizationPrivateKeySource, sp.AuthorizationPrivateKeyConfig, "authorization_private_key_source", p.AtName("authorization_private_key_config"))
		}
	}
}

func validateExternalSecretRequiresConfig(diags *diag.Diagnostics, source fwtypes.StringEnum[awstypes.SecretSourceType], secretConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel], sourceName string, configPath path.Path) {
	if source.IsNull() || source.IsUnknown() {
		return
	}
	if source.ValueEnum() != awstypes.SecretSourceTypeExternal {
		return
	}
	if !secretConfig.IsNull() && !secretConfig.IsUnknown() {
		return
	}
	diags.AddAttributeError(
		configPath,
		"Missing Required Configuration",
		fmt.Sprintf("%q is set to %q, so %s must be configured.", sourceName, awstypes.SecretSourceTypeExternal, configPath.String()),
	)
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
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"api_key_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9+/=\-_\s]*$`), "must contain only base64 characters and whitespace"),
										},
									},
									"api_key_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"wallet_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9+/=\-_\s]*$`), "must contain only base64 characters and whitespace"),
										},
									},
									"wallet_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"api_key_secret_arn": framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
									"wallet_secret_arn":  framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
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
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"app_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9+/=\-_\s]*$`), "must contain only base64 characters and whitespace"),
										},
									},
									"app_secret_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"authorization_id": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]+$`), "must contain only letters, numbers, hyphens, and underscores"),
										},
									},
									"authorization_private_key": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^(wallet-auth:)?[a-zA-Z0-9+/=\-_\s]*$`), "must contain only base64 characters and whitespace, optionally prefixed with wallet-auth:"),
										},
									},
									"authorization_private_key_source": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"app_secret_arn":                framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
									"authorization_private_key_arn": framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
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
	smerr.AddEnrich(ctx, &response.Diagnostics, flattenPaymentProviderConfigurationSecretARNs(ctx, out.ProviderConfigurationOutput, &data))
	if response.Diagnostics.HasError() {
		return
	}

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
	smerr.AddEnrich(ctx, &response.Diagnostics, flattenPaymentProviderConfigurationSecretARNs(ctx, out.ProviderConfigurationOutput, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// GetPaymentCredentialProvider does not populate Tags in its response, so
	// tags are read back by the transparent-tagging interceptor via ListTags.
	// Calling setTagsOut with the (empty) response tags would clobber that,
	// producing a perpetual tags diff.

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

		out, err := conn.UpdatePaymentCredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, flattenPaymentProviderConfigurationSecretARNs(ctx, out.ProviderConfigurationOutput, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

// flattenPaymentProviderConfigurationSecretARNs overlays the server-returned
// secret ARNs onto the (write-only) provider configuration block, which is
// otherwise preserved from prior state. The ARN attributes are Computed, so they
// are populated from the Create/Read/Update response rather than from config.
func flattenPaymentProviderConfigurationSecretARNs(ctx context.Context, apiObject awstypes.PaymentProviderConfigurationOutput, data *paymentCredentialProviderResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiObject == nil {
		return diags
	}

	config, d := data.ProviderConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	// On import the write-only provider_configuration block is absent from state.
	// Reconstruct a minimal block so the computed managed secret ARNs the API
	// returns become known in state; otherwise they plan as null on the reconcile
	// apply and it fails with "inconsistent result after apply". The write-only
	// secret values and other inputs are re-supplied from configuration on the
	// next apply.
	if config == nil {
		config = &paymentProviderConfigurationModel{
			CoinbaseCDPConfiguration: fwtypes.NewListNestedObjectValueOfNull[coinbaseCdpConfigurationModel](ctx),
			StripePrivyConfiguration: fwtypes.NewListNestedObjectValueOfNull[stripePrivyConfigurationModel](ctx),
		}
	}

	switch v := apiObject.(type) {
	case *awstypes.PaymentProviderConfigurationOutputMemberCoinbaseCdpConfiguration:
		coinbase, d := config.CoinbaseCDPConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		if coinbase == nil {
			// On import, hydrate every field the API returns (identifier, secret
			// sources, and for EXTERNAL secrets the secret reference), leaving only
			// the write-only secret values null so the post-import plan is limited
			// to re-supplying them (and is empty when both secrets are EXTERNAL).
			coinbase = &coinbaseCdpConfigurationModel{
				APIKeyID:           fwflex.StringToFramework(ctx, v.Value.ApiKeyId),
				APIKeySecret:       types.StringNull(),
				APIKeySecretARN:    fwtypes.NewListNestedObjectValueOfNull[secretModel](ctx),
				APIKeySecretConfig: fwtypes.NewListNestedObjectValueOfNull[paymentSecretReferenceModel](ctx),
				APIKeySecretSource: fwtypes.StringEnumValue(v.Value.ApiKeySecretSource),
				WalletSecret:       types.StringNull(),
				WalletSecretARN:    fwtypes.NewListNestedObjectValueOfNull[secretModel](ctx),
				WalletSecretConfig: fwtypes.NewListNestedObjectValueOfNull[paymentSecretReferenceModel](ctx),
				WalletSecretSource: fwtypes.StringEnumValue(v.Value.WalletSecretSource),
			}
			if v.Value.ApiKeySecretSource == awstypes.SecretSourceTypeExternal {
				coinbase.APIKeySecretConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &paymentSecretReferenceModel{
					JSONKey:  fwflex.StringToFramework(ctx, v.Value.ApiKeySecretJsonKey),
					SecretID: fwflex.StringToFramework(ctx, v.Value.ApiKeySecretArn.SecretArn),
				})
			}
			if v.Value.WalletSecretSource == awstypes.SecretSourceTypeExternal {
				coinbase.WalletSecretConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &paymentSecretReferenceModel{
					JSONKey:  fwflex.StringToFramework(ctx, v.Value.WalletSecretJsonKey),
					SecretID: fwflex.StringToFramework(ctx, v.Value.WalletSecretArn.SecretArn),
				})
			}
		}
		// The secret ARNs and sources are computed; the API is authoritative for
		// them on create, update, and import, so always set them from the response.
		coinbase.APIKeySecretSource = fwtypes.StringEnumValue(v.Value.ApiKeySecretSource)
		coinbase.WalletSecretSource = fwtypes.StringEnumValue(v.Value.WalletSecretSource)
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v.Value.ApiKeySecretArn, &coinbase.APIKeySecretARN))
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v.Value.WalletSecretArn, &coinbase.WalletSecretARN))
		if diags.HasError() {
			return diags
		}
		config.CoinbaseCDPConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, coinbase)

	case *awstypes.PaymentProviderConfigurationOutputMemberStripePrivyConfiguration:
		stripe, d := config.StripePrivyConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		if stripe == nil {
			// On import, hydrate every field the API returns (identifiers, secret
			// sources, and for EXTERNAL secrets the secret reference), leaving only
			// the write-only secret values null so the post-import plan is limited
			// to re-supplying them (and is empty when both secrets are EXTERNAL).
			stripe = &stripePrivyConfigurationModel{
				AppID:                         fwflex.StringToFramework(ctx, v.Value.AppId),
				AppSecret:                     types.StringNull(),
				AppSecretARN:                  fwtypes.NewListNestedObjectValueOfNull[secretModel](ctx),
				AppSecretConfig:               fwtypes.NewListNestedObjectValueOfNull[paymentSecretReferenceModel](ctx),
				AppSecretSource:               fwtypes.StringEnumValue(v.Value.AppSecretSource),
				AuthorizationID:               fwflex.StringToFramework(ctx, v.Value.AuthorizationId),
				AuthorizationPrivateKey:       types.StringNull(),
				AuthorizationPrivateKeyARN:    fwtypes.NewListNestedObjectValueOfNull[secretModel](ctx),
				AuthorizationPrivateKeyConfig: fwtypes.NewListNestedObjectValueOfNull[paymentSecretReferenceModel](ctx),
				AuthorizationPrivateKeySource: fwtypes.StringEnumValue(v.Value.AuthorizationPrivateKeySource),
			}
			if v.Value.AppSecretSource == awstypes.SecretSourceTypeExternal {
				stripe.AppSecretConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &paymentSecretReferenceModel{
					JSONKey:  fwflex.StringToFramework(ctx, v.Value.AppSecretJsonKey),
					SecretID: fwflex.StringToFramework(ctx, v.Value.AppSecretArn.SecretArn),
				})
			}
			if v.Value.AuthorizationPrivateKeySource == awstypes.SecretSourceTypeExternal {
				stripe.AuthorizationPrivateKeyConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &paymentSecretReferenceModel{
					JSONKey:  fwflex.StringToFramework(ctx, v.Value.AuthorizationPrivateKeyJsonKey),
					SecretID: fwflex.StringToFramework(ctx, v.Value.AuthorizationPrivateKeyArn.SecretArn),
				})
			}
		}
		// The secret ARNs and sources are computed; the API is authoritative for
		// them on create, update, and import, so always set them from the response.
		stripe.AppSecretSource = fwtypes.StringEnumValue(v.Value.AppSecretSource)
		stripe.AuthorizationPrivateKeySource = fwtypes.StringEnumValue(v.Value.AuthorizationPrivateKeySource)
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v.Value.AppSecretArn, &stripe.AppSecretARN))
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v.Value.AuthorizationPrivateKeyArn, &stripe.AuthorizationPrivateKeyARN))
		if diags.HasError() {
			return diags
		}
		config.StripePrivyConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, stripe)
	}

	data.ProviderConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, config)
	return diags
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
	APIKeySecretARN    fwtypes.ListNestedObjectValueOf[secretModel]                 `tfsdk:"api_key_secret_arn"`
	APIKeySecretConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"api_key_secret_config"`
	APIKeySecretSource fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"api_key_secret_source"`
	WalletSecret       types.String                                                 `tfsdk:"wallet_secret"`
	WalletSecretARN    fwtypes.ListNestedObjectValueOf[secretModel]                 `tfsdk:"wallet_secret_arn"`
	WalletSecretConfig fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"wallet_secret_config"`
	WalletSecretSource fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"wallet_secret_source"`
}

type stripePrivyConfigurationModel struct {
	AppID                         types.String                                                 `tfsdk:"app_id"`
	AppSecret                     types.String                                                 `tfsdk:"app_secret"`
	AppSecretARN                  fwtypes.ListNestedObjectValueOf[secretModel]                 `tfsdk:"app_secret_arn"`
	AppSecretConfig               fwtypes.ListNestedObjectValueOf[paymentSecretReferenceModel] `tfsdk:"app_secret_config"`
	AppSecretSource               fwtypes.StringEnum[awstypes.SecretSourceType]                `tfsdk:"app_secret_source"`
	AuthorizationID               types.String                                                 `tfsdk:"authorization_id"`
	AuthorizationPrivateKey       types.String                                                 `tfsdk:"authorization_private_key"`
	AuthorizationPrivateKeyARN    fwtypes.ListNestedObjectValueOf[secretModel]                 `tfsdk:"authorization_private_key_arn"`
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
