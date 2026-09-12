// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	oauth2ClientCredentialsCtxKey = inttypes.NewContextKey[oauth2CredentialProviderWriteOnlyCredentialsModel]()
)

// @FrameworkResource("aws_bedrockagentcore_oauth2_credential_provider", name="OAuth2 Credential Provider")
// @IdentityAttribute("name")
// @Tags(identifierAttribute="credential_provider_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol;bedrockagentcorecontrol.GetOauth2CredentialProviderOutput")
// Generated tests use only github_oauth2_provider_config. Testing other providers requires ignoring their client credential attributes on import.
// @Testing(importIgnore="oauth2_provider_config.0.github_oauth2_provider_config.0.client_credentials_wo_version;oauth2_provider_config.0.github_oauth2_provider_config.0.client_id;oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret;oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret_config;oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret_source")
// @Testing(importStateIdAttribute="name")
// @Testing(preCheck="testAccPreCheckOAuth2CredentialProviders")
// @Testing(preIdentityVersion="v6.63.0")
func newOAuth2CredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &oauth2CredentialProviderResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type oauth2CredentialProviderResource struct {
	framework.ResourceWithModel[oauth2CredentialProviderResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

// Client credentials schema shared by all OAuth2 providers.
// See oauth2ProviderClientCredentialsModel.
func oauth2ProviderClientCredentialsAttributes(context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"client_credentials_wo_version": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				),
			},
		},
		names.AttrClientID: schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				),
				//stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("client_id_wo")),
			},
		},
		"client_id_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
				),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("client_credentials_wo_version"),
				),
			},
		},
		names.AttrClientSecret: schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 2048),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("client_secret_wo"),
				),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
				),
				//stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("client_secret_wo")),
			},
		},
		"client_secret_source": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.SecretSourceType](),
			Optional:   true,
			Validators: []validator.String{
				// Secret-source-conditional client-secret rules. The service rejects an inline
				// client secret when the secret source is EXTERNAL (the secret is referenced via
				// client_secret_config); outside EXTERNAL a client_id must be paired with an
				// inline client secret (the pairing the static AlsoRequires used to enforce).
				tfstringvalidator.AlsoRequiresWhenEquals(
					awstypes.SecretSourceTypeExternal,
					path.MatchRelative().AtParent().AtName("client_secret_config"),
				),
				tfstringvalidator.ConflictsWithWhenEquals(
					awstypes.SecretSourceTypeExternal,
					path.MatchRelative().AtParent().AtName(names.AttrClientSecret),
					path.MatchRelative().AtParent().AtName("client_secret_wo"),
				),
			},
		},
		"client_secret_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 2048),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName(names.AttrClientSecret),
				),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("client_credentials_wo_version"),
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				),
			},
		},
	}
}

// Configuration schema shared by all non-custom OAuth2 providers.
// See basicOAuth2ProviderConfigModel.
func basicOAuth2ProviderConfigBlock[T any](ctx context.Context) schema.Block {
	attrs := oauth2ProviderClientCredentialsAttributes(ctx)
	maps.Copy(attrs, map[string]schema.Attribute{
		"oauth_discovery": framework.ResourceComputedListOfObjectsAttribute[oauth2DiscoveryModel](ctx),
	})

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: attrs,
			Blocks: map[string]schema.Block{
				"client_secret_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[secretReferenceModel](ctx),
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
				},
			},
		},
	}
}

func customOAuth2ProviderConfigBlock(ctx context.Context) schema.Block {
	block := basicOAuth2ProviderConfigBlock[customOAuth2ProviderConfigModel](ctx).(schema.ListNestedBlock)
	// Replace the Computed oauth_discovery attribute with a configurable block.
	delete(block.NestedObject.Attributes, "oauth_discovery")
	maps.Copy(block.NestedObject.Attributes, map[string]schema.Attribute{
		"client_authentication_method": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.ClientAuthenticationMethodType](),
			Optional:   true,
			Validators: []validator.String{
				tfstringvalidator.ExactlyOneOfWhenEquals(
					awstypes.ClientAuthenticationMethodTypeClientSecretBasic,
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				),
				tfstringvalidator.ExactlyOneOfWhenEquals(
					awstypes.ClientAuthenticationMethodTypeClientSecretPost,
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				),
			},
		},
	})
	maps.Copy(block.NestedObject.Blocks, map[string]schema.Block{
		"oauth_discovery": schema.ListNestedBlock{
			CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2DiscoveryModel](ctx),
			Validators: []validator.List{
				listvalidator.IsRequired(),
				listvalidator.SizeAtLeast(1),
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Validators: []validator.Object{
					tfobjectvalidator.ExactlyOneOfChildren(
						path.MatchRelative().AtName("authorization_server_metadata"),
						path.MatchRelative().AtName("discovery_url"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"discovery_url": schema.StringAttribute{
						Optional: true,
					},
				},
				Blocks: map[string]schema.Block{
					"authorization_server_metadata": schema.ListNestedBlock{
						CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2AuthorizationServerMetadataModel](ctx),
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"authorization_endpoint": schema.StringAttribute{
									Required: true,
								},
								names.AttrIssuer: schema.StringAttribute{
									Required: true,
								},
								"response_types": schema.SetAttribute{
									CustomType: fwtypes.SetOfStringType,
									Optional:   true,
								},
								"token_endpoint": schema.StringAttribute{
									Required: true,
								},
								"token_endpoint_auth_methods": schema.ListAttribute{
									CustomType: fwtypes.ListOfStringType,
									Optional:   true,
									Validators: []validator.List{
										listvalidator.SizeBetween(1, 2),
										listvalidator.ValueStringsAre(
											stringvalidator.RegexMatches(regexache.MustCompile(`^(client_secret_post|client_secret_basic)$`), ""),
										),
									},
								},
							},
						},
					},
				},
			},
		},
		"on_behalf_of_token_exchange_config": schema.ListNestedBlock{
			CustomType: fwtypes.NewListNestedObjectTypeOf[onBehalfOfTokenExchangeConfigTypeModel](ctx),
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"grant_type": schema.StringAttribute{
						CustomType: fwtypes.StringEnumType[awstypes.OnBehalfOfTokenExchangeGrantTypeType](),
						Required:   true,
						Validators: []validator.String{
							tfstringvalidator.AlsoRequiresWhenEquals(
								awstypes.OnBehalfOfTokenExchangeGrantTypeTypeTokenExchange,
								path.MatchRelative().AtParent().AtName("token_exchange_grant_type_config"),
							),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"token_exchange_grant_type_config": schema.ListNestedBlock{
						CustomType: fwtypes.NewListNestedObjectTypeOf[tokenExchangeGrantTypeConfigTypeModel](ctx),
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"actor_token_content": schema.StringAttribute{
									CustomType: fwtypes.StringEnumType[awstypes.ActorTokenContentType](),
									Required:   true,
									Validators: []validator.String{
										tfstringvalidator.ConflictsWithWhenNotEquals(
											awstypes.ActorTokenContentTypeM2m,
											path.MatchRelative().AtParent().AtName("actor_token_scopes"),
										),
									},
								},
								"actor_token_scopes": schema.SetAttribute{
									CustomType: fwtypes.SetOfStringType,
									Optional:   true,
								},
							},
						},
					},
				},
			},
		},
		"private_endpoint": privateEndpointBlock(ctx),
		"private_endpoint_override": privateEndpointOverrideBlock(ctx, listvalidator.AlsoRequires(
			path.MatchRelative().AtParent().AtName("private_endpoint"),
		)),
		"private_key_jwt_config": schema.ListNestedBlock{
			CustomType: fwtypes.NewListNestedObjectTypeOf[privateKeyJWTConfigModel](ctx),
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"additional_header_claims": schema.MapAttribute{
						CustomType: fwtypes.MapOfStringType,
						Optional:   true,
					},
					"additional_payload_claims": schema.MapAttribute{
						CustomType: fwtypes.MapOfStringType,
						Optional:   true,
					},
					"signing_algorithm": schema.StringAttribute{
						CustomType: fwtypes.StringEnumType[awstypes.SigningAlgorithm](),
						Optional:   true,
					},
				},
				Blocks: map[string]schema.Block{
					"private_key_source": schema.ListNestedBlock{
						CustomType: fwtypes.NewListNestedObjectTypeOf[privateKeySourceModel](ctx),
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Validators: []validator.Object{
								tfobjectvalidator.ExactlyOneOfChildren(
									path.MatchRelative().AtName("kms_key_source"),
								),
							},
							Blocks: map[string]schema.Block{
								"kms_key_source": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[kmsKeySourceTypeModel](ctx),
									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											names.AttrKMSKeyARN: schema.StringAttribute{
												CustomType: fwtypes.ARNType,
												Required:   true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return block
}

func includedOAuth2ProviderConfigBlock(ctx context.Context) schema.Block {
	block := basicOAuth2ProviderConfigBlock[includedOAuth2ProviderConfigModel](ctx).(schema.ListNestedBlock)
	maps.Copy(block.NestedObject.Attributes, map[string]schema.Attribute{
		"authorization_endpoint": schema.StringAttribute{
			Optional: true,
		},
		names.AttrIssuer: schema.StringAttribute{
			Optional: true,
		},
		"token_endpoint": schema.StringAttribute{
			Optional: true,
		},
	})

	return block
}

func microsoftOAuth2ProviderConfigBlock(ctx context.Context) schema.Block {
	block := basicOAuth2ProviderConfigBlock[microsoftOAuth2ProviderConfigModel](ctx).(schema.ListNestedBlock)
	maps.Copy(block.NestedObject.Attributes, map[string]schema.Attribute{
		"tenant_id": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("tenant_id_wo"),
				),
			},
		},
		"tenant_id_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("tenant_id"),
				),
				stringvalidator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("tenant_id_wo_version"),
				),
			},
		},
		"tenant_id_wo_version": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(
					path.MatchRelative().AtParent().AtName("tenant_id_wo"),
				),
			},
		},
	})

	return block
}

func (r *oauth2CredentialProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"callback_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret_arn":       framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
			"credential_provider_arn": framework.ARNAttributeComputedOnly(),
			"credential_provider_vendor": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CredentialProviderVendorType](),
				Required:   true,
				Validators: []validator.String{
					tfstringvalidator.DiscriminatorRequires(map[awstypes.CredentialProviderVendorType]path.Expression{
						awstypes.CredentialProviderVendorTypeAtlassianOauth2:  r.providerConfigPathExpression("atlassian_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeCustomOauth2:     r.providerConfigPathExpression("custom_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeGithubOauth2:     r.providerConfigPathExpression("github_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeGoogleOauth2:     r.providerConfigPathExpression("google_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeLinkedinOauth2:   r.providerConfigPathExpression("linkedin_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeMicrosoftOauth2:  r.providerConfigPathExpression("microsoft_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeSalesforceOauth2: r.providerConfigPathExpression("salesforce_oauth2_provider_config"),
						awstypes.CredentialProviderVendorTypeSlackOauth2:      r.providerConfigPathExpression("slack_oauth2_provider_config"),
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]{1,128}$`), "Valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 50 characters."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"oauth2_provider_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2ProviderConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							r.providerConfigPathExpression("atlassian_oauth2_provider_config"),
							r.providerConfigPathExpression("custom_oauth2_provider_config"),
							r.providerConfigPathExpression("github_oauth2_provider_config"),
							r.providerConfigPathExpression("google_oauth2_provider_config"),
							r.providerConfigPathExpression("included_oauth2_provider_config"),
							r.providerConfigPathExpression("linkedin_oauth2_provider_config"),
							r.providerConfigPathExpression("microsoft_oauth2_provider_config"),
							r.providerConfigPathExpression("salesforce_oauth2_provider_config"),
							r.providerConfigPathExpression("slack_oauth2_provider_config"),
						),
					},
					Blocks: map[string]schema.Block{
						"atlassian_oauth2_provider_config":  basicOAuth2ProviderConfigBlock[atlassianOAuth2ProviderConfigModel](ctx),
						"custom_oauth2_provider_config":     customOAuth2ProviderConfigBlock(ctx),
						"github_oauth2_provider_config":     basicOAuth2ProviderConfigBlock[githubOAuth2ProviderConfigModel](ctx),
						"google_oauth2_provider_config":     basicOAuth2ProviderConfigBlock[googleOAuth2ProviderConfigModel](ctx),
						"included_oauth2_provider_config":   includedOAuth2ProviderConfigBlock(ctx),
						"linkedin_oauth2_provider_config":   basicOAuth2ProviderConfigBlock[linkedinOAuth2ProviderConfigModel](ctx),
						"microsoft_oauth2_provider_config":  microsoftOAuth2ProviderConfigBlock(ctx),
						"salesforce_oauth2_provider_config": basicOAuth2ProviderConfigBlock[salesforceOAuth2ProviderConfigModel](ctx),
						"slack_oauth2_provider_config":      basicOAuth2ProviderConfigBlock[slackOAuth2ProviderConfigModel](ctx),
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

func (r *oauth2CredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	// Get the effective client credentials.
	clientCredentials, d := plan.writeOnlyCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	// Terraform WO attributes are only in Config.
	fromConfig, d := config.writeOnlyCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}
	clientCredentials.ClientIDWO = fromConfig.ClientIDWO
	clientCredentials.ClientSecretWO = fromConfig.ClientSecretWO

	// Stuff the client credentials into Context for AutoFlEx.
	ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateOauth2CredentialProviderInput
	smerr.AddEnrich(ctx, &response.Diagnostics,
		fwflex.Expand(ctx, plan, &input,
			fwflex.WithFieldNameSuffix("Input"),
		))
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateOauth2CredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	provider, err := waitOAuth2CredentialProviderCreated(ctx, conn, name, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root(names.AttrName), name)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, provider, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Get the client credentials from State.
	clientCredentials, d := data.writeOnlyCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findOAuth2CredentialProviderByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	// Stuff the client credentials into Context for AutoFlEx.
	ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *oauth2CredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state, config oauth2CredentialProviderResourceModel
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
		// Get the effective client credentials.
		clientCredentials, d := plan.writeOnlyCredentials(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		// Terraform WO attributes are only in Config.
		fromConfig, d := config.writeOnlyCredentials(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}
		clientCredentials.ClientIDWO = fromConfig.ClientIDWO
		clientCredentials.ClientSecretWO = fromConfig.ClientSecretWO

		// Stuff the client credentials into Context for AutoFlEx.
		ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

		name := fwflex.StringValueFromFramework(ctx, plan.Name)
		var input bedrockagentcorecontrol.UpdateOauth2CredentialProviderInput
		smerr.AddEnrich(ctx, &response.Diagnostics,
			fwflex.Expand(ctx, plan, &input,
				fwflex.WithFieldNameSuffix("Input")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateOauth2CredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
			return
		}

		provider, err := waitOAuth2CredentialProviderUpdated(ctx, conn, name, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, provider, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	const (
		conflictTimeout = 2 * time.Minute
	)
	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeleteOauth2CredentialProviderInput{
		Name: aws.String(name),
	}
	// "ConflictException: UPDATE in progress for resource with id '...'. Please retry after update completes".
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ConflictException](ctx, conflictTimeout, func(ctx context.Context) (any, error) {
		return conn.DeleteOauth2CredentialProvider(ctx, &input)
	}, "Please retry")
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	if _, err := waitOAuth2CredentialProviderDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}
}

func (r *oauth2CredentialProviderResource) flatten(ctx context.Context, out *bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, data *oauth2CredentialProviderResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, out, data, fwflex.WithFieldNameSuffix("Output"))...)
	return diags
}

func (*oauth2CredentialProviderResource) providerConfigPath(name string) path.Path {
	return path.Root("oauth2_provider_config").AtListIndex(0).AtName(name)
}

func (r *oauth2CredentialProviderResource) providerConfigPathExpression(name string) path.Expression {
	return r.providerConfigPath(name).Expression()
}

func findOAuth2CredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetOauth2CredentialProviderInput{
		Name: aws.String(name),
	}
	return findOAuth2CredentialProvider(ctx, conn, &input)
}

func findOAuth2CredentialProvider(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetOauth2CredentialProviderInput) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	out, err := conn.GetOauth2CredentialProvider(ctx, input)

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

func statusOAuth2CredentialProvider(conn *bedrockagentcorecontrol.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findOAuth2CredentialProviderByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitOAuth2CredentialProviderCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating),
		Target:                    enum.Slice(awstypes.StatusReady),
		Refresh:                   statusOAuth2CredentialProvider(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOAuth2CredentialProviderUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusUpdating),
		Target:                    enum.Slice(awstypes.StatusReady),
		Refresh:                   statusOAuth2CredentialProvider(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOAuth2CredentialProviderDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeleting, awstypes.StatusReady),
		Target:  []string{},
		Refresh: statusOAuth2CredentialProvider(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type oauth2CredentialProviderResourceModel struct {
	framework.WithRegionModel
	CallbackURL              types.String                                               `tfsdk:"callback_url"`
	ClientSecretARN          fwtypes.ListNestedObjectValueOf[secretModel]               `tfsdk:"client_secret_arn"`
	CredentialProviderARN    types.String                                               `tfsdk:"credential_provider_arn"`
	CredentialProviderVendor fwtypes.StringEnum[awstypes.CredentialProviderVendorType]  `tfsdk:"credential_provider_vendor"`
	Name                     types.String                                               `tfsdk:"name"`
	OAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"oauth2_provider_config"`
	Tags                     tftags.Map                                                 `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                             `tfsdk:"timeouts"`
}

func (m *oauth2CredentialProviderResourceModel) writeOnlyCredentials(ctx context.Context) (oauth2CredentialProviderWriteOnlyCredentialsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model, d := m.OAuth2ProviderConfig.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || model == nil {
		v, d := fwtypes.Nullified[oauth2CredentialProviderWriteOnlyCredentialsModel](ctx)
		diags.Append(d...)
		return v, diags
	}

	v, d := model.writeOnlyCredentials(ctx)
	diags.Append(d...)

	return v, diags
}

type oauth2ProviderConfigModel struct {
	AtlassianOAuth2ProviderConfig  fwtypes.ListNestedObjectValueOf[atlassianOAuth2ProviderConfigModel]  `tfsdk:"atlassian_oauth2_provider_config"`
	CustomOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[customOAuth2ProviderConfigModel]     `tfsdk:"custom_oauth2_provider_config"`
	GithubOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[githubOAuth2ProviderConfigModel]     `tfsdk:"github_oauth2_provider_config"`
	GoogleOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[googleOAuth2ProviderConfigModel]     `tfsdk:"google_oauth2_provider_config"`
	IncludedOAuth2ProviderConfig   fwtypes.ListNestedObjectValueOf[includedOAuth2ProviderConfigModel]   `tfsdk:"included_oauth2_provider_config"`
	LinkedinOAuth2ProviderConfig   fwtypes.ListNestedObjectValueOf[linkedinOAuth2ProviderConfigModel]   `tfsdk:"linkedin_oauth2_provider_config"`
	MicrosoftOAuth2ProviderConfig  fwtypes.ListNestedObjectValueOf[microsoftOAuth2ProviderConfigModel]  `tfsdk:"microsoft_oauth2_provider_config"`
	SalesforceOAuth2ProviderConfig fwtypes.ListNestedObjectValueOf[salesforceOAuth2ProviderConfigModel] `tfsdk:"salesforce_oauth2_provider_config"`
	SlackOAuth2ProviderConfig      fwtypes.ListNestedObjectValueOf[slackOAuth2ProviderConfigModel]      `tfsdk:"slack_oauth2_provider_config"`
}

func (m *oauth2ProviderConfigModel) writeOnlyCredentials(ctx context.Context) (oauth2CredentialProviderWriteOnlyCredentialsModel, diag.Diagnostics) {
	r, diags := fwtypes.Nullified[oauth2CredentialProviderWriteOnlyCredentialsModel](ctx)
	if diags.HasError() {
		return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
	}

	switch {
	case !m.AtlassianOAuth2ProviderConfig.IsNull():
		model, d := m.AtlassianOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.CustomOAuth2ProviderConfig.IsNull():
		model, d := m.CustomOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.GithubOAuth2ProviderConfig.IsNull():
		model, d := m.GithubOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.GoogleOAuth2ProviderConfig.IsNull():
		model, d := m.GoogleOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.IncludedOAuth2ProviderConfig.IsNull():
		model, d := m.IncludedOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.LinkedinOAuth2ProviderConfig.IsNull():
		model, d := m.LinkedinOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.MicrosoftOAuth2ProviderConfig.IsNull():
		model, d := m.MicrosoftOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.microsoftOAuth2ProviderTenantIDModel = model.microsoftOAuth2ProviderTenantIDModel
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.SalesforceOAuth2ProviderConfig.IsNull():
		model, d := m.SalesforceOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel

	case !m.SlackOAuth2ProviderConfig.IsNull():
		model, d := m.SlackOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2CredentialProviderWriteOnlyCredentialsModel](), diags
		}
		r.oauth2ProviderClientCredentialsModel = model.oauth2ProviderClientCredentialsModel
	}

	return r, diags
}

var (
	_ fwflex.Expander  = oauth2ProviderConfigModel{}
	_ fwflex.Flattener = &oauth2ProviderConfigModel{}
)

func (m *oauth2ProviderConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	// Propagate client credentials from State.
	clientCredentials := oauth2ClientCredentialsCtxKey.FromContext(ctx)

	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.Oauth2ProviderConfigOutputMemberAtlassianOauth2ProviderConfig:
		var model atlassianOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.AtlassianOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberCustomOauth2ProviderConfig:
		var model customOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.CustomOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberGithubOauth2ProviderConfig:
		var model githubOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.GithubOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberGoogleOauth2ProviderConfig:
		var model googleOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.GoogleOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberIncludedOauth2ProviderConfig:
		v := t.Value
		var model includedOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		// Copy over additional attributes from oauth_discovery.authorization_server_metadata.
		switch t := v.OauthDiscovery.(type) {
		case *awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata:
			v := t.Value
			model.AuthorizationEndpoint = fwflex.StringToFramework(ctx, v.AuthorizationEndpoint)
			model.Issuer = fwflex.StringToFramework(ctx, v.Issuer)
			model.TokenEndpoint = fwflex.StringToFramework(ctx, v.TokenEndpoint)
		}
		var d diag.Diagnostics
		m.IncludedOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberLinkedinOauth2ProviderConfig:
		var model linkedinOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.LinkedinOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberMicrosoftOauth2ProviderConfig:
		var model microsoftOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.microsoftOAuth2ProviderTenantIDModel = clientCredentials.microsoftOAuth2ProviderTenantIDModel
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.MicrosoftOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberSalesforceOauth2ProviderConfig:
		var model salesforceOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.SalesforceOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	case awstypes.Oauth2ProviderConfigOutputMemberSlackOauth2ProviderConfig:
		var model slackOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var d diag.Diagnostics
		m.SlackOAuth2ProviderConfig, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("oauth2ProviderConfigModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m oauth2ProviderConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	// Set API fields for client credentials.
	clientCredentials := oauth2ClientCredentialsCtxKey.FromContext(ctx)
	if !clientCredentials.ClientIDWO.IsNull() && !clientCredentials.ClientSecretWO.IsNull() {
		clientCredentials.ClientID = clientCredentials.ClientIDWO
		clientCredentials.ClientSecret = clientCredentials.ClientSecretWO
	}

	var diags diag.Diagnostics
	switch {
	case !m.AtlassianOAuth2ProviderConfig.IsNull():
		model, d := m.AtlassianOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberAtlassianOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.CustomOAuth2ProviderConfig.IsNull():
		model, d := m.CustomOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberCustomOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.GithubOAuth2ProviderConfig.IsNull():
		model, d := m.GithubOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberGithubOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.GoogleOAuth2ProviderConfig.IsNull():
		model, d := m.GoogleOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberGoogleOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.IncludedOAuth2ProviderConfig.IsNull():
		model, d := m.IncludedOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberIncludedOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.LinkedinOAuth2ProviderConfig.IsNull():
		model, d := m.LinkedinOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberLinkedinOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MicrosoftOAuth2ProviderConfig.IsNull():
		model, d := m.MicrosoftOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.microsoftOAuth2ProviderTenantIDModel = clientCredentials.microsoftOAuth2ProviderTenantIDModel
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberMicrosoftOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SalesforceOAuth2ProviderConfig.IsNull():
		model, d := m.SalesforceOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberSalesforceOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SlackOAuth2ProviderConfig.IsNull():
		model, d := m.SlackOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		model.oauth2ProviderClientCredentialsModel = clientCredentials.oauth2ProviderClientCredentialsModel
		var r awstypes.Oauth2ProviderConfigInputMemberSlackOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

// Client credentials attributes shared by all OAuth2 providers.
type oauth2ProviderClientCredentialsModel struct {
	ClientCredentialsWOVersion types.Int64                                           `tfsdk:"client_credentials_wo_version"`
	ClientID                   types.String                                          `tfsdk:"client_id"`
	ClientIDWO                 types.String                                          `tfsdk:"client_id_wo"`
	ClientSecret               types.String                                          `tfsdk:"client_secret"`
	ClientSecretConfig         fwtypes.ListNestedObjectValueOf[secretReferenceModel] `tfsdk:"client_secret_config"`
	ClientSecretSource         fwtypes.StringEnum[awstypes.SecretSourceType]         `tfsdk:"client_secret_source"`
	ClientSecretWO             types.String                                          `tfsdk:"client_secret_wo"`
}

type secretReferenceModel struct {
	JSONKey  types.String `tfsdk:"json_key"`
	SecretID types.String `tfsdk:"secret_id"`
}

// Configuration attributes shared by all OAuth2 providers.
type basicOAuth2ProviderConfigModel struct {
	oauth2ProviderClientCredentialsModel
	OAuthDiscovery fwtypes.ListNestedObjectValueOf[oauth2DiscoveryModel] `tfsdk:"oauth_discovery"`
}

type customOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
	ClientAuthenticationMethod    fwtypes.StringEnum[awstypes.ClientAuthenticationMethodType]             `tfsdk:"client_authentication_method"`
	OnBehalfOfTokenExchangeConfig fwtypes.ListNestedObjectValueOf[onBehalfOfTokenExchangeConfigTypeModel] `tfsdk:"on_behalf_of_token_exchange_config"`
	PrivateEndpoint               fwtypes.ListNestedObjectValueOf[privateEndpointModel]                   `tfsdk:"private_endpoint"`
	PrivateEndpointOverrides      fwtypes.ListNestedObjectValueOf[privateEndpointOverrideModel]           `tfsdk:"private_endpoint_override"`
	PrivateKeyJWTConfig           fwtypes.ListNestedObjectValueOf[privateKeyJWTConfigModel]               `tfsdk:"private_key_jwt_config"`
}

type onBehalfOfTokenExchangeConfigTypeModel struct {
	GrantType                    fwtypes.StringEnum[awstypes.OnBehalfOfTokenExchangeGrantTypeType]      `tfsdk:"grant_type"`
	TokenExchangeGrantTypeConfig fwtypes.ListNestedObjectValueOf[tokenExchangeGrantTypeConfigTypeModel] `tfsdk:"token_exchange_grant_type_config"`
}

type tokenExchangeGrantTypeConfigTypeModel struct {
	ActorTokenContent fwtypes.StringEnum[awstypes.ActorTokenContentType] `tfsdk:"actor_token_content"`
	ActorTokenScopes  fwtypes.SetOfString                                `tfsdk:"actor_token_scopes"`
}

type privateKeyJWTConfigModel struct {
	AdditionalHeaderClaims  fwtypes.MapOfString                                    `tfsdk:"additional_header_claims"`
	AdditionalPayloadClaims fwtypes.MapOfString                                    `tfsdk:"additional_payload_claims"`
	PrivateKeySource        fwtypes.ListNestedObjectValueOf[privateKeySourceModel] `tfsdk:"private_key_source"`
	SigningAlgorithm        fwtypes.StringEnum[awstypes.SigningAlgorithm]          `tfsdk:"signing_algorithm"`
}

type privateKeySourceModel struct {
	KMSKeySource fwtypes.ListNestedObjectValueOf[kmsKeySourceTypeModel] `tfsdk:"kms_key_source"`
}

var (
	_ fwflex.Expander  = privateKeySourceModel{}
	_ fwflex.Flattener = &privateKeySourceModel{}
)

func (m *privateKeySourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PrivateKeySourceMemberKmsKeySource:
		var model kmsKeySourceTypeModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.KMSKeySource, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("privateKeySourceModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m privateKeySourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.KMSKeySource.IsNull():
		model, d := m.KMSKeySource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.PrivateKeySourceMemberKmsKeySource
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type kmsKeySourceTypeModel struct {
	KMSKeyARN fwtypes.ARN `tfsdk:"kms_key_arn"`
}

// These OAuth2 providers have only the common attributes.
type (
	atlassianOAuth2ProviderConfigModel  basicOAuth2ProviderConfigModel
	githubOAuth2ProviderConfigModel     basicOAuth2ProviderConfigModel
	googleOAuth2ProviderConfigModel     basicOAuth2ProviderConfigModel
	linkedinOAuth2ProviderConfigModel   basicOAuth2ProviderConfigModel
	salesforceOAuth2ProviderConfigModel basicOAuth2ProviderConfigModel
	slackOAuth2ProviderConfigModel      basicOAuth2ProviderConfigModel
)

type includedOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	Issuer                types.String `tfsdk:"issuer"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
}

type microsoftOAuth2ProviderTenantIDModel struct {
	TenantID          types.String `tfsdk:"tenant_id"`
	TenantIDWO        types.String `tfsdk:"tenant_id_wo"`
	TenantIDWOVersion types.Int64  `tfsdk:"tenant_id_wo_version"`
}

type microsoftOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
	microsoftOAuth2ProviderTenantIDModel
}

type oauth2DiscoveryModel struct {
	AuthorizationServerMetadata fwtypes.ListNestedObjectValueOf[oauth2AuthorizationServerMetadataModel] `tfsdk:"authorization_server_metadata"`
	DiscoveryURL                types.String                                                            `tfsdk:"discovery_url"`
}

var (
	_ fwflex.Expander  = oauth2DiscoveryModel{}
	_ fwflex.Flattener = &oauth2DiscoveryModel{}
)

func (m *oauth2DiscoveryModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.Oauth2DiscoveryMemberDiscoveryUrl:
		m.DiscoveryURL = fwflex.StringValueToFramework(ctx, t.Value)

	case awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata:
		var model oauth2AuthorizationServerMetadataModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.AuthorizationServerMetadata, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("oauth2DiscoveryModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m oauth2DiscoveryModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.DiscoveryURL.IsNull():
		var r awstypes.Oauth2DiscoveryMemberDiscoveryUrl
		r.Value = fwflex.StringValueFromFramework(ctx, m.DiscoveryURL)
		return &r, diags

	case !m.AuthorizationServerMetadata.IsNull():
		model, d := m.AuthorizationServerMetadata.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type oauth2AuthorizationServerMetadataModel struct {
	AuthorizationEndpoint    types.String         `tfsdk:"authorization_endpoint"`
	Issuer                   types.String         `tfsdk:"issuer"`
	ResponseTypes            fwtypes.SetOfString  `tfsdk:"response_types"`
	TokenEndpoint            types.String         `tfsdk:"token_endpoint"`
	TokenEndpointAuthMethods fwtypes.ListOfString `tfsdk:"token_endpoint_auth_methods"`
}

// Credentials not returned by the AWS API.
type oauth2CredentialProviderWriteOnlyCredentialsModel struct {
	oauth2ProviderClientCredentialsModel
	microsoftOAuth2ProviderTenantIDModel
}
