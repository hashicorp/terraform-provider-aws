// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_gateway_target", name="Gateway Target")
func newGatewayTargetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayTargetResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type gatewayTargetResource struct {
	framework.ResourceWithModel[gatewayTargetResourceModel]
	framework.WithTimeouts
}

func jsonAttribute(conflictWith string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(conflictWith)),
		},
	}
}

// headerNameValidators returns validators for HTTP header names in metadata_configuration.
// Header names must contain only alphanumeric characters, hyphens, and underscores.
// Certain restricted headers cannot be configured for propagation.
// Headers starting with X-Amzn- are prohibited except X-Amzn-Bedrock-AgentCore-Runtime-Custom-*.
// See: https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-headers.html
func headerNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(
			regexache.MustCompile(`^[a-zA-Z0-9_-]+$`),
			"header names must contain only alphanumeric characters, hyphens, and underscores",
		),
		tfstringvalidator.NoneOfCaseInsensitive(restrictedHeaders()...),
		tfstringvalidator.PrefixNoneOfCaseInsensitive(
			[]string{"X-Amzn-"},
			[]string{"X-Amzn-Bedrock-AgentCore-Runtime-Custom-"},
		),
	}
}

// restrictedHeaders returns the full list of restricted HTTP headers that cannot be
// configured for propagation in metadata_configuration.
func restrictedHeaders() []string {
	return []string{
		// Authentication & Authorization
		"Authorization",
		"Proxy-Authorization",
		"WWW-Authenticate",
		// Content Negotiation
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		// Content Headers
		"Content-Type",
		"Content-Length",
		"Content-Encoding",
		"Content-Language",
		"Content-Location",
		"Content-Range",
		// Caching
		"Cache-Control",
		"ETag",
		"Expires",
		"If-Match",
		"If-Modified-Since",
		"If-None-Match",
		"If-Range",
		"If-Unmodified-Since",
		"Last-Modified",
		"Pragma",
		"Vary",
		// Connection Management
		"Connection",
		"Keep-Alive",
		"Proxy-Connection",
		"Upgrade",
		// Request Context
		"Host",
		"User-Agent",
		"Referer",
		"From",
		// Range Requests
		"Range",
		"Accept-Ranges",
		// Transfer
		"Transfer-Encoding",
		"TE",
		"Trailer",
		// Response
		"Server",
		"Date",
		"Location",
		"Retry-After",
		// Cookies
		"Set-Cookie",
		"Cookie",
		// Security
		"Content-Security-Policy",
		"Content-Security-Policy-Report-Only",
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
		// Cross-Origin
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
		// CORS
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
		"Access-Control-Expose-Headers",
		"Access-Control-Max-Age",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
		"Origin",
		// Client Hints
		"Accept-CH",
		"Accept-CH-Lifetime",
		"DPR",
		"Width",
		"Viewport-Width",
		"Downlink",
		"ECT", //nolint:misspell // https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/ECT
		"RTT",
		"Save-Data",
		// Other Security
		"Clear-Site-Data",
		"Feature-Policy",
		"Expect-CT",
		"Public-Key-Pins",
		"Public-Key-Pins-Report-Only",
		// Proxy & Forwarding
		"X-Forwarded-For",
		"X-Forwarded-Host",
		"X-Forwarded-Proto",
		"X-Real-IP",
		"X-Requested-With",
		"X-CSRF-Token",
		// CDN & Infrastructure
		"CF-Ray",
		"CF-Connecting-IP",
		"X-Amz-Cf-Id",
		"X-Cache",
		"X-Served-By",
		// HTTP/2 Pseudo-Headers
		":method",
		":path",
		":scheme",
		":authority",
		":status",
		// Misc
		"Link",
		// WebSocket
		"Sec-WebSocket-Key",
		"Sec-WebSocket-Accept",
		"Sec-WebSocket-Version",
		"Sec-WebSocket-Protocol",
		"Sec-WebSocket-Extensions",
	}
}

func createLeafItemsBlock[T any](ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrDescription: schema.StringAttribute{Optional: true},
				names.AttrType: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
				},
				"items_json":      jsonAttribute("properties_json"),
				"properties_json": jsonAttribute("items_json"),
			},
		},
	}
}

func createLeafPropertyBlock[T any](ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[T](ctx),
		Validators: []validator.Set{
			setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("items")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName:        schema.StringAttribute{Required: true},
				names.AttrDescription: schema.StringAttribute{Optional: true},
				names.AttrType: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
				},
				"required": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(false),
				},
				"items_json":      jsonAttribute("properties_json"),
				"properties_json": jsonAttribute("items_json"),
			},
		},
	}
}

func schemaDefinitionNestedBlock(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{Optional: true},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
			},
		},
		Blocks: map[string]schema.Block{
			"property": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[schemaPropertyModel](ctx),
				Validators: []validator.Set{
					setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("items")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName:        schema.StringAttribute{Required: true},
						names.AttrDescription: schema.StringAttribute{Optional: true},
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
						},
						"required": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"items": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[schemaItemsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDescription: schema.StringAttribute{Optional: true},
									names.AttrType: schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
									},
								},
								Blocks: map[string]schema.Block{
									"items":    createLeafItemsBlock[schemaItemsLeafModel](ctx),
									"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
								},
							},
						},
						"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
					},
				},
			},
			"items": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[schemaItemsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("property")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{Optional: true},
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SchemaType](),
						},
					},
					Blocks: map[string]schema.Block{
						"items":    createLeafItemsBlock[schemaItemsLeafModel](ctx),
						"property": createLeafPropertyBlock[schemaPropertyLeafModel](ctx),
					},
				},
			},
		},
	}
}

func (r *gatewayTargetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][-]?){1,100}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"credential_provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[credentialProviderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"api_key": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAPIKeyCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("caller_iam_credentials"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
									path.MatchRelative().AtParent().AtName("jwt_passthrough"),
									path.MatchRelative().AtParent().AtName("oauth"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credential_location": schema.StringAttribute{
										Optional:   true,
										CustomType: fwtypes.StringEnumType[awstypes.ApiKeyCredentialLocation](),
									},
									"credential_parameter_name": schema.StringAttribute{
										Optional: true,
									},
									"credential_prefix": schema.StringAttribute{
										Optional: true,
									},
									"provider_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"caller_iam_credentials": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[iamCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
									path.MatchRelative().AtParent().AtName("jwt_passthrough"),
									path.MatchRelative().AtParent().AtName("oauth"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRegion: schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("service")),
											fwvalidators.AWSRegion(),
										},
									},
									"service": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"gateway_iam_role": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[iamCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("caller_iam_credentials"),
									path.MatchRelative().AtParent().AtName("jwt_passthrough"),
									path.MatchRelative().AtParent().AtName("oauth"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRegion: schema.StringAttribute{
										Optional: true,
										Description: "AWS Region used for SigV4 signing of upstream requests. " +
											"Defaults to the gateway's Region when omitted. Only meaningful when `service` is set.",
										Validators: []validator.String{
											stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("service")),
											fwvalidators.AWSRegion(),
										},
									},
									"service": schema.StringAttribute{
										Optional: true,
										Description: "The target AWS service name used for SigV4 signing of upstream requests. " +
											"Required when calling SigV4-protected endpoints such as another Bedrock AgentCore " +
											"Runtime (use `bedrock-agentcore`). Omit for non-SigV4 IAM-role-based authentication.",
									},
								},
							},
						},
						"jwt_passthrough": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[jwtPassthroughCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("caller_iam_credentials"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
									path.MatchRelative().AtParent().AtName("oauth"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								// Empty block - no attributes needed for JWT Passthrough
							},
						},
						"oauth": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[oauthCredentialProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("api_key"),
									path.MatchRelative().AtParent().AtName("gateway_iam_role"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"custom_parameters": schema.MapAttribute{
										CustomType: fwtypes.MapOfStringType,
										Optional:   true,
									},
									"default_return_url": schema.StringAttribute{
										Optional:    true,
										Description: "The URL where the end user's browser is redirected after obtaining the authorization code. Required when grant_type is AUTHORIZATION_CODE.",
									},
									"grant_type": schema.StringAttribute{
										Optional:    true,
										CustomType:  fwtypes.StringEnumType[awstypes.OAuthGrantType](),
										Description: "The OAuth grant type. Valid values are AUTHORIZATION_CODE and CLIENT_CREDENTIALS.",
									},
									"provider_arn": schema.StringAttribute{
										Required: true,
									},
									"scopes": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"metadata_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[metadataConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allowed_query_parameters": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							Description: "A list of URL query parameters that are allowed to be propagated from incoming gateway URL to the target.",
							Validators: []validator.Set{
								setvalidator.SizeAtMost(10),
							},
						},
						"allowed_request_headers": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							Description: "A list of HTTP headers that are allowed to be propagated from incoming client requests to the target.",
							Validators: []validator.Set{
								setvalidator.SizeAtMost(10),
								setvalidator.ValueStringsAre(headerNameValidators()...),
							},
						},
						"allowed_response_headers": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							Description: "A list of HTTP headers that are allowed to be propagated from the target response back to the client.",
							Validators: []validator.Set{
								setvalidator.SizeAtMost(10),
								setvalidator.ValueStringsAre(headerNameValidators()...),
							},
						},
					},
				},
			},
			"private_endpoint": privateEndpointSchema(ctx),
			"target_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("http"),
							path.MatchRelative().AtName("mcp"),
						),
					},
					Blocks: map[string]schema.Block{
						"http": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[httpTargetConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"agentcore_runtime": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[runtimeTargetConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												path.MatchRelative().AtParent().AtName("agentcore_runtime"),
												path.MatchRelative().AtParent().AtName("passthrough"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrARN: schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.ARNType,
												},
												"qualifier": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"passthrough": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[passthroughTargetConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrEndpoint: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile(`^https://.+`),
															"Must start with https://",
														),
													},
												},
												"protocol_type": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.StringEnumType[awstypes.PassthroughProtocolType](),
												},
											},
											Blocks: map[string]schema.Block{
												"schema": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"inline_payload": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[inlinePayloadModel](ctx),
																Validators: []validator.List{
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("inline_payload"),
																		path.MatchRelative().AtParent().AtName("s3"),
																	),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"payload": schema.StringAttribute{
																			Required: true,
																		},
																	},
																},
															},
															"s3": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("inline_payload"),
																		path.MatchRelative().AtParent().AtName("s3"),
																	),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"bucket_owner_account_id": schema.StringAttribute{
																			Optional: true,
																		},
																		names.AttrURI: schema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"stickiness_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[stickinessConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"identifier": schema.StringAttribute{
																Required: true,
															},
															"timeout": schema.Int32Attribute{
																Optional: true,
																Validators: []validator.Int32{
																	int32validator.Between(1, 86400),
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
						},
						"mcp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mcpTargetConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"api_gateway": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[apiGatewayTargetConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"rest_api_id": schema.StringAttribute{
													Required: true,
												},
												names.AttrStage: schema.StringAttribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"api_gateway_tool_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[apiGatewayToolConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"tool_filter": schema.SetNestedBlock{
																CustomType: fwtypes.NewSetNestedObjectTypeOf[apiGatewayToolFilterModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"filter_path": schema.StringAttribute{
																			Required: true,
																		},
																		"methods": schema.SetAttribute{
																			ElementType: fwtypes.StringEnumType[awstypes.RestApiMethod](),
																			Required:    true,
																		},
																	},
																},
															},
															"tool_override": schema.SetNestedBlock{
																CustomType: fwtypes.NewSetNestedObjectTypeOf[apiGatewayToolOverrideModel](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrDescription: schema.StringAttribute{
																			Optional: true,
																		},
																		"method": schema.StringAttribute{
																			CustomType: fwtypes.StringEnumType[awstypes.RestApiMethod](),
																			Required:   true,
																		},
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																		},
																		names.AttrPath: schema.StringAttribute{
																			Required: true,
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
									"lambda": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mcpLambdaTargetConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"lambda_arn": schema.StringAttribute{
													Required: true,
												},
											},
											Blocks: map[string]schema.Block{
												"tool_schema": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[toolSchemaModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"inline_payload": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[toolDefinitionModel](ctx),
																Validators: []validator.List{
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("inline_payload"),
																		path.MatchRelative().AtParent().AtName("s3"),
																	),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrDescription: schema.StringAttribute{
																			Required: true,
																		},
																		names.AttrName: schema.StringAttribute{
																			Required: true,
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"input_schema": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[schemaDefinitionModel](ctx),
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schemaDefinitionNestedBlock(ctx),
																		},
																		"output_schema": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[schemaDefinitionModel](ctx),
																			Validators: []validator.List{
																				listvalidator.SizeAtMost(1),
																			},
																			NestedObject: schemaDefinitionNestedBlock(ctx),
																		},
																	},
																},
															},
															"s3": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.ExactlyOneOf(
																		path.MatchRelative().AtParent().AtName("inline_payload"),
																		path.MatchRelative().AtParent().AtName("s3"),
																	),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"bucket_owner_account_id": schema.StringAttribute{
																			Optional: true,
																		},
																		names.AttrURI: schema.StringAttribute{
																			Optional: true,
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
									"mcp_server": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[mcpServerTargetConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrEndpoint: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile(`^https://.+`),
															"Must start with https://",
														),
													},
												},
												"listing_mode": schema.StringAttribute{
													Optional:   true,
													Computed:   true,
													CustomType: fwtypes.StringEnumType[awstypes.ListingMode](),
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
									"open_api_schema": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"inline_payload": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inlinePayloadModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("inline_payload"),
															path.MatchRelative().AtParent().AtName("s3"),
														),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"payload": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("inline_payload"),
															path.MatchRelative().AtParent().AtName("s3"),
														),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket_owner_account_id": schema.StringAttribute{
																Optional: true,
															},
															names.AttrURI: schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"smithy_model": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[apiSchemaConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"inline_payload": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inlinePayloadModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("inline_payload"),
															path.MatchRelative().AtParent().AtName("s3"),
														),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"payload": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"s3": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("inline_payload"),
															path.MatchRelative().AtParent().AtName("s3"),
														),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket_owner_account_id": schema.StringAttribute{
																Optional: true,
															},
															names.AttrURI: schema.StringAttribute{
																Optional: true,
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

func (r *gatewayTargetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	var input bedrockagentcorecontrol.CreateGatewayTargetInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	var (
		out *bedrockagentcorecontrol.CreateGatewayTargetOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateGatewayTarget(ctx, &input)

		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Gateway execution role lacks permission") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	targetID := aws.ToString(out.TargetId)

	target, err := waitGatewayTargetCreated(ctx, conn, gatewayIdentifier, targetID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("gateway_identifier"), gatewayIdentifier)
		response.State.SetAttribute(ctx, path.Root("target_id"), targetID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, target, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *gatewayTargetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.TargetID)
	out, err := findGatewayTargetByTwoPartKey(ctx, conn, gatewayIdentifier, targetID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithIgnoredFieldNames([]string{"GatewayArn"})))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayTargetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old gatewayTargetResourceModel
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
		gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, new.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, new.TargetID)
		var input bedrockagentcorecontrol.UpdateGatewayTargetInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateGatewayTarget(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
			return
		}

		if _, err := waitGatewayTargetUpdated(ctx, conn, gatewayIdentifier, targetID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *gatewayTargetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, targetID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.TargetID)
	if err := deleteGatewayTarget(ctx, conn, gatewayIdentifier, targetID); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}

	if _, err := waitGatewayTargetDeleted(ctx, conn, gatewayIdentifier, targetID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, targetID)
		return
	}
}

func (r *gatewayTargetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")

	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`Unexpected format for import ID (%s), use: "GatewayIdentifier,TargetId"`, request.ID))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("gateway_identifier"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("target_id"), parts[1]))
}

func (r gatewayTargetResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	// Force replacement if target configuration changes between lambda, smithy_model, and open_api_schema
	var plan, state gatewayTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	planTargetData, diags := plan.TargetConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	stateTargetData, diags := state.TargetConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	planTargetType := planTargetData.GetConfigurationType(ctx)
	stateTargetType := stateTargetData.GetConfigurationType(ctx)

	if planTargetType != stateTargetType {
		response.RequiresReplace = append(response.RequiresReplace, path.Root("target_configuration"))
	}
}

func waitGatewayTargetCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusCreating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TargetStatusUpdating),
		Target:                    enum.Slice(awstypes.TargetStatusReady),
		Refresh:                   statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayTargetDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		// FAILED and SYNCHRONIZING can appear until AWS moves the target to DELETING.
		Pending: enum.Slice(awstypes.TargetStatusDeleting, awstypes.TargetStatusReady, awstypes.TargetStatusFailed, awstypes.TargetStatusSynchronizing),
		Target:  []string{},
		Refresh: statusGatewayTarget(conn, gatewayIdentifier, targetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayTargetOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGatewayTarget(conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGatewayTargetByTwoPartKey(ctx, conn, gatewayIdentifier, targetID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findGatewayTargetByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayTargetInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		TargetId:          aws.String(targetID),
	}

	return findGatewayTarget(ctx, conn, &input)
}

func findGatewayTarget(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetGatewayTargetInput) (*bedrockagentcorecontrol.GetGatewayTargetOutput, error) {
	out, err := conn.GetGatewayTarget(ctx, input)

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

func deleteGatewayTarget(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, targetID string) error {
	input := bedrockagentcorecontrol.DeleteGatewayTargetInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		TargetId:          aws.String(targetID),
	}
	_, err := conn.DeleteGatewayTarget(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return smarterr.NewError(fmt.Errorf("deleting Bedrock AgentCore Gateway (%s) Target (%s): %w", gatewayIdentifier, targetID, err))
	}

	return nil
}

func listGatewayTargets(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListGatewayTargetsInput) iter.Seq2[awstypes.TargetSummary, error] {
	return func(yield func(awstypes.TargetSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListGatewayTargetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.TargetSummary](), fmt.Errorf("listing Bedrock AgentCore Gateway Targets: %w", err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

type gatewayTargetResourceModel struct {
	framework.WithRegionModel
	CredentialProviderConfiguration fwtypes.ListNestedObjectValueOf[credentialProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	Description                     types.String                                                          `tfsdk:"description"`
	GatewayIdentifier               types.String                                                          `tfsdk:"gateway_identifier"`
	MetadataConfiguration           fwtypes.ListNestedObjectValueOf[metadataConfigurationModel]           `tfsdk:"metadata_configuration"`
	Name                            types.String                                                          `tfsdk:"name"`
	PrivateEndpoint                 fwtypes.ListNestedObjectValueOf[privateEndpointModel]                 `tfsdk:"private_endpoint"`
	TargetConfiguration             fwtypes.ListNestedObjectValueOf[targetConfigurationModel]             `tfsdk:"target_configuration"`
	TargetID                        types.String                                                          `tfsdk:"target_id"`
	Timeouts                        timeouts.Value                                                        `tfsdk:"timeouts"`
}

type metadataConfigurationModel struct {
	AllowedQueryParameters fwtypes.SetOfString `tfsdk:"allowed_query_parameters"`
	AllowedRequestHeaders  fwtypes.SetOfString `tfsdk:"allowed_request_headers"`
	AllowedResponseHeaders fwtypes.SetOfString `tfsdk:"allowed_response_headers"`
}

type privateEndpointModel struct {
	ManagedVPCResource         fwtypes.ListNestedObjectValueOf[managedVPCResourceModel]         `tfsdk:"managed_vpc_resource"`
	SelfManagedLatticeResource fwtypes.ListNestedObjectValueOf[selfManagedLatticeResourceModel] `tfsdk:"self_managed_lattice_resource"`
}

var (
	_ fwflex.Expander  = privateEndpointModel{}
	_ fwflex.Flattener = &privateEndpointModel{}
)

func (m *privateEndpointModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PrivateEndpointMemberManagedVpcResource:
		var model managedVPCResourceModel
		model.Tags = tftags.NewMapValueNull() // Tags are not handled by AutoFlex.

		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}

		// Tags are not handled by AutoFlex.
		model.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMap(ctx, t.Value.Tags))

		m.ManagedVPCResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.PrivateEndpointMemberSelfManagedLatticeResource:
		var model selfManagedLatticeResourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.SelfManagedLatticeResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported PrivateEndpoint Type",
			fmt.Sprintf("private endpoint flatten: unexpected type %T", v),
		)
	}
	return diags
}

func (m privateEndpointModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.ManagedVPCResource.IsNull():
		data, d := m.ManagedVPCResource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PrivateEndpointMemberManagedVpcResource
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		// Tags are not handled by AutoFlex.
		r.Value.Tags = fwflex.ExpandFrameworkStringValueMap(ctx, data.Tags)

		return &r, diags

	case !m.SelfManagedLatticeResource.IsNull():
		data, d := m.SelfManagedLatticeResource.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.PrivateEndpointMemberSelfManagedLatticeResource
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type managedVPCResourceModel struct {
	EndpointIPAddressType fwtypes.StringEnum[awstypes.EndpointIpAddressType] `tfsdk:"endpoint_ip_address_type"`
	RoutingDomain         types.String                                       `tfsdk:"routing_domain"`
	SecurityGroupIDs      fwtypes.SetOfString                                `tfsdk:"security_group_ids"`
	SubnetIDs             fwtypes.SetOfString                                `tfsdk:"subnet_ids"`
	Tags                  tftags.Map                                         `tfsdk:"tags"`
	VPCIdentifier         types.String                                       `tfsdk:"vpc_identifier"`
}
type selfManagedLatticeResourceModel struct {
	ResourceConfigurationIdentifier types.String `tfsdk:"resource_configuration_identifier"`
}

var (
	_ fwflex.Expander  = selfManagedLatticeResourceModel{}
	_ fwflex.Flattener = &selfManagedLatticeResourceModel{}
)

func (m *selfManagedLatticeResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier:
		m.ResourceConfigurationIdentifier = fwflex.StringValueToFramework(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported SelfManagedLatticeResource Type",
			fmt.Sprintf("self managed lattice resource flatten: unexpected type %T", v),
		)
	}
	return diags
}

func (m selfManagedLatticeResourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.ResourceConfigurationIdentifier.IsNull():
		var r awstypes.SelfManagedLatticeResourceMemberResourceConfigurationIdentifier
		r.Value = fwflex.StringValueFromFramework(ctx, m.ResourceConfigurationIdentifier)
		return &r, diags
	}

	return nil, diags
}

type credentialProviderConfigurationModel struct {
	ApiKey               fwtypes.ListNestedObjectValueOf[gatewayAPIKeyCredentialProviderModel]  `tfsdk:"api_key"`
	CallerIAMCredentials fwtypes.ListNestedObjectValueOf[iamCredentialProviderModel]            `tfsdk:"caller_iam_credentials"`
	GatewayIAMRole       fwtypes.ListNestedObjectValueOf[iamCredentialProviderModel]            `tfsdk:"gateway_iam_role"`
	JWTPassthrough       fwtypes.ListNestedObjectValueOf[jwtPassthroughCredentialProviderModel] `tfsdk:"jwt_passthrough"`
	OAuth                fwtypes.ListNestedObjectValueOf[oauthCredentialProviderModel]          `tfsdk:"oauth"`
}

var (
	_ fwflex.Expander  = credentialProviderConfigurationModel{}
	_ fwflex.Flattener = &credentialProviderConfigurationModel{}
)

func (m *credentialProviderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CredentialProviderConfiguration:
		switch t.CredentialProviderType {
		case awstypes.CredentialProviderTypeApiKey:
			if apiKeyProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberApiKeyCredentialProvider); ok {
				var model gatewayAPIKeyCredentialProviderModel
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, apiKeyProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
				m.ApiKey = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}

		case awstypes.CredentialProviderTypeCallerIamCredentials:
			if callerIamProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberIamCredentialProvider); ok {
				var model iamCredentialProviderModel
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, callerIamProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
				m.CallerIAMCredentials = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}

		case awstypes.CredentialProviderTypeGatewayIamRole:
			var model iamCredentialProviderModel
			if iamProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberIamCredentialProvider); ok {
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, iamProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
			}
			m.GatewayIAMRole = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		case awstypes.CredentialProviderTypeJwtPassthrough:
			m.JWTPassthrough = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &jwtPassthroughCredentialProviderModel{})

		case awstypes.CredentialProviderTypeOauth:
			if oauthProvider, ok := t.CredentialProvider.(*awstypes.CredentialProviderMemberOauthCredentialProvider); ok {
				var model oauthCredentialProviderModel
				smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, oauthProvider.Value, &model))
				if diags.HasError() {
					return diags
				}
				m.OAuth = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
			}

		default:
			diags.AddError(
				"Unknown Credential Provider Type",
				fmt.Sprintf("Received unknown credential provider type: %s", t.CredentialProviderType),
			)
		}

	default:
		diags.AddError(
			"Invalid Credential Provider Configuration",
			fmt.Sprintf("Received unexpected type: %T", v),
		)
	}
	return diags
}

func (m credentialProviderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var c awstypes.CredentialProviderConfiguration
	switch {
	case !m.ApiKey.IsNull():
		apiKeyCredentialProviderConfigurationData, d := m.ApiKey.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberApiKeyCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, apiKeyCredentialProviderConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeApiKey
		c.CredentialProvider = &r
		return &c, diags

	case !m.CallerIAMCredentials.IsNull():
		callerIAMCredentialsData, d := m.CallerIAMCredentials.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberIamCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, callerIAMCredentialsData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeCallerIamCredentials
		c.CredentialProvider = &r
		return &c, diags

	case !m.GatewayIAMRole.IsNull():
		gatewayIAMRoleData, d := m.GatewayIAMRole.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		c.CredentialProviderType = awstypes.CredentialProviderTypeGatewayIamRole
		if gatewayIAMRoleData != nil && !gatewayIAMRoleData.Service.IsNull() {
			var r awstypes.CredentialProviderMemberIamCredentialProvider
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, gatewayIAMRoleData, &r.Value))
			if diags.HasError() {
				return nil, diags
			}
			c.CredentialProvider = &r
		} else {
			c.CredentialProvider = nil
		}
		return &c, diags

	case !m.JWTPassthrough.IsNull():
		c.CredentialProviderType = awstypes.CredentialProviderTypeJwtPassthrough
		c.CredentialProvider = nil
		return &c, diags

	case !m.OAuth.IsNull():
		oauthCredentialProviderConfigurationData, d := m.OAuth.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialProviderMemberOauthCredentialProvider
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, oauthCredentialProviderConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		c.CredentialProviderType = awstypes.CredentialProviderTypeOauth
		c.CredentialProvider = &r
		return &c, diags

	default:
		diags.AddError(
			"Invalid Credential Provider Configuration",
			"At least one credential provider must be configured: api_key, caller_iam_credentials, gateway_iam_role, jwt_passthrough, or oauth",
		)
		return nil, diags
	}
}

type targetConfigurationModel struct {
	HTTP fwtypes.ListNestedObjectValueOf[httpTargetConfigurationModel] `tfsdk:"http"`
	MCP  fwtypes.ListNestedObjectValueOf[mcpTargetConfigurationModel]  `tfsdk:"mcp"`
}

func (m *targetConfigurationModel) GetConfigurationType(ctx context.Context) string {
	if !m.HTTP.IsNull() {
		httpData, _ := m.HTTP.ToPtr(ctx)
		switch {
		case !httpData.AgentcoreRuntime.IsNull():
			return "http_agentcore_runtime"
		case !httpData.Passthrough.IsNull():
			return "http_passthrough"
		}
	}
	if !m.MCP.IsNull() {
		mcpData, _ := m.MCP.ToPtr(ctx)
		switch {
		case !mcpData.Lambda.IsNull():
			return "lambda"
		case !mcpData.MCPServer.IsNull():
			return "mcp_server"
		case !mcpData.OpenApiSchema.IsNull():
			return "open_api_schema"
		case !mcpData.SmithyModel.IsNull():
			return "smithy_model"
		}
	}
	return "unknown"
}

var (
	_ fwflex.Expander  = targetConfigurationModel{}
	_ fwflex.Flattener = &targetConfigurationModel{}
)

func (m *targetConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.TargetConfigurationMemberHttp:
		var model httpTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.HTTP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.TargetConfigurationMemberMcp:
		var model mcpTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.MCP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("target configuration flatten: %T", v),
		)
	}
	return diags
}

func (m targetConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.HTTP.IsNull():
		httpConfigurationData, d := m.HTTP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.TargetConfigurationMemberHttp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, httpConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MCP.IsNull():
		mcpConfigurationData, d := m.MCP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.TargetConfigurationMemberMcp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, mcpConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type httpTargetConfigurationModel struct {
	AgentcoreRuntime fwtypes.ListNestedObjectValueOf[runtimeTargetConfigurationModel]     `tfsdk:"agentcore_runtime"`
	Passthrough      fwtypes.ListNestedObjectValueOf[passthroughTargetConfigurationModel] `tfsdk:"passthrough"`
}

var (
	_ fwflex.Expander  = httpTargetConfigurationModel{}
	_ fwflex.Flattener = &httpTargetConfigurationModel{}
)

func (m *httpTargetConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.HttpTargetConfigurationMemberAgentcoreRuntime:
		var model runtimeTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.AgentcoreRuntime = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.HttpTargetConfigurationMemberPassthrough:
		var model passthroughTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Passthrough = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported HTTP Target Configuration Type",
			fmt.Sprintf("http configuration flatten: %T", v),
		)
	}

	return diags
}

func (m httpTargetConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.AgentcoreRuntime.IsNull():
		runtimeData, d := m.AgentcoreRuntime.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.HttpTargetConfigurationMemberAgentcoreRuntime
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, runtimeData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.Passthrough.IsNull():
		passthroughData, d := m.Passthrough.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.HttpTargetConfigurationMemberPassthrough
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, passthroughData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type mcpTargetConfigurationModel struct {
	APIGateway    fwtypes.ListNestedObjectValueOf[apiGatewayTargetConfigurationModel] `tfsdk:"api_gateway"`
	Lambda        fwtypes.ListNestedObjectValueOf[mcpLambdaTargetConfigurationModel]  `tfsdk:"lambda"`
	MCPServer     fwtypes.ListNestedObjectValueOf[mcpServerTargetConfigurationModel]  `tfsdk:"mcp_server"`
	SmithyModel   fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel]        `tfsdk:"smithy_model"`
	OpenApiSchema fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel]        `tfsdk:"open_api_schema"`
}

var (
	_ fwflex.Expander  = mcpTargetConfigurationModel{}
	_ fwflex.Flattener = &mcpTargetConfigurationModel{}
)

func (m *mcpTargetConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.McpTargetConfigurationMemberApiGateway:
		var model apiGatewayTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.APIGateway = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberLambda:
		var model mcpLambdaTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Lambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberMcpServer:
		var model mcpServerTargetConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.MCPServer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberOpenApiSchema:
		var model apiSchemaConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.OpenApiSchema = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.McpTargetConfigurationMemberSmithyModel:
		var model apiSchemaConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.SmithyModel = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("mcp configuration flatten: %T", v),
		)
	}
	return diags
}

func (m mcpTargetConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.APIGateway.IsNull():
		apiGatewayMCPConfigurationData, d := m.APIGateway.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberApiGateway
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, apiGatewayMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.Lambda.IsNull():
		lambdaMCPConfigurationData, d := m.Lambda.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberLambda
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, lambdaMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MCPServer.IsNull():
		mcpServerConfigurationData, d := m.MCPServer.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberMcpServer
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, mcpServerConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.OpenApiSchema.IsNull():
		openApiMCPConfigurationData, d := m.OpenApiSchema.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberOpenApiSchema
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, openApiMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SmithyModel.IsNull():
		smithyMCPConfigurationData, d := m.SmithyModel.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.McpTargetConfigurationMemberSmithyModel
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, smithyMCPConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type runtimeTargetConfigurationModel struct {
	ARN       fwtypes.ARN  `tfsdk:"arn"`
	Qualifier types.String `tfsdk:"qualifier"`
}

type passthroughTargetConfigurationModel struct {
	Endpoint                types.String                                                  `tfsdk:"endpoint"`
	ProtocolType            fwtypes.StringEnum[awstypes.PassthroughProtocolType]          `tfsdk:"protocol_type"`
	Schema                  fwtypes.ListNestedObjectValueOf[apiSchemaConfigurationModel]  `tfsdk:"schema"`
	StickinessConfiguration fwtypes.ListNestedObjectValueOf[stickinessConfigurationModel] `tfsdk:"stickiness_configuration"`
}

var (
	_ fwflex.Expander  = passthroughTargetConfigurationModel{}
	_ fwflex.Flattener = &passthroughTargetConfigurationModel{}
)

func (m *passthroughTargetConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PassthroughTargetConfiguration:
		m.Endpoint = fwflex.StringToFramework(ctx, t.Endpoint)
		m.ProtocolType = fwtypes.StringEnumValue(t.ProtocolType)

		if t.Schema != nil && t.Schema.Source != nil {
			var schemaModel apiSchemaConfigurationModel
			d := schemaModel.Flatten(ctx, t.Schema.Source)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Schema = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &schemaModel)
		} else {
			m.Schema = fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx)
		}

		if t.StickinessConfiguration != nil {
			var stickinessModel stickinessConfigurationModel
			d := fwflex.Flatten(ctx, t.StickinessConfiguration, &stickinessModel)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.StickinessConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &stickinessModel)
		} else {
			m.StickinessConfiguration = fwtypes.NewListNestedObjectValueOfNull[stickinessConfigurationModel](ctx)
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("passthrough target configuration flatten: %T", v),
		)
	}
	return diags
}

func (m passthroughTargetConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	r := &awstypes.PassthroughTargetConfiguration{
		Endpoint:     fwflex.StringFromFramework(ctx, m.Endpoint),
		ProtocolType: awstypes.PassthroughProtocolType(m.ProtocolType.ValueString()),
	}

	if !m.Schema.IsNull() {
		schemaData, d := m.Schema.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		source, d := schemaData.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		apiSchemaConfiguration, ok := source.(awstypes.ApiSchemaConfiguration)
		if !ok {
			diags.AddError(
				"Invalid Schema Configuration",
				"Exactly one of \"inline_payload\" or \"s3\" must be configured within \"schema\".",
			)
			return nil, diags
		}
		r.Schema = &awstypes.HttpApiSchemaConfiguration{
			Source: apiSchemaConfiguration,
		}
	}

	if !m.StickinessConfiguration.IsNull() {
		stickinessData, d := m.StickinessConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var sc awstypes.StickinessConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, stickinessData, &sc))
		if diags.HasError() {
			return nil, diags
		}
		r.StickinessConfiguration = &sc
	}

	return r, diags
}

type stickinessConfigurationModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Timeout    types.Int32  `tfsdk:"timeout"`
}

type gatewayAPIKeyCredentialProviderModel struct {
	CredentialLocation      fwtypes.StringEnum[awstypes.ApiKeyCredentialLocation] `tfsdk:"credential_location"`
	CredentialParameterName types.String                                          `tfsdk:"credential_parameter_name"`
	CredentialPrefix        types.String                                          `tfsdk:"credential_prefix"`
	ProviderARN             fwtypes.ARN                                           `tfsdk:"provider_arn"`
}

type iamCredentialProviderModel struct {
	Region  types.String `tfsdk:"region"`
	Service types.String `tfsdk:"service"`
}

type jwtPassthroughCredentialProviderModel struct {
	// Empty struct - JWT Passthrough provider requires no configuration
}

type oauthCredentialProviderModel struct {
	CustomParameters fwtypes.MapOfString                         `tfsdk:"custom_parameters"`
	DefaultReturnURL types.String                                `tfsdk:"default_return_url"`
	GrantType        fwtypes.StringEnum[awstypes.OAuthGrantType] `tfsdk:"grant_type"`
	ProviderARN      fwtypes.ARN                                 `tfsdk:"provider_arn"`
	Scopes           fwtypes.SetOfString                         `tfsdk:"scopes"`
}

type apiGatewayTargetConfigurationModel struct {
	ApiGatewayToolConfiguration fwtypes.ListNestedObjectValueOf[apiGatewayToolConfigurationModel] `tfsdk:"api_gateway_tool_configuration"`
	RestApiID                   types.String                                                      `tfsdk:"rest_api_id"`
	Stage                       types.String                                                      `tfsdk:"stage"`
}

type apiGatewayToolConfigurationModel struct {
	ToolFilter   fwtypes.SetNestedObjectValueOf[apiGatewayToolFilterModel]   `tfsdk:"tool_filter"`
	ToolOverride fwtypes.SetNestedObjectValueOf[apiGatewayToolOverrideModel] `tfsdk:"tool_override"`
}

type apiGatewayToolFilterModel struct {
	FilterPath types.String                                    `tfsdk:"filter_path"`
	Methods    fwtypes.SetOfStringEnum[awstypes.RestApiMethod] `tfsdk:"methods"`
}

type apiGatewayToolOverrideModel struct {
	Description types.String                               `tfsdk:"description"`
	Method      fwtypes.StringEnum[awstypes.RestApiMethod] `tfsdk:"method"`
	Name        types.String                               `tfsdk:"name"`
	Path        types.String                               `tfsdk:"path"`
}

type mcpLambdaTargetConfigurationModel struct {
	LambdaArn  types.String                                     `tfsdk:"lambda_arn"`
	ToolSchema fwtypes.ListNestedObjectValueOf[toolSchemaModel] `tfsdk:"tool_schema"`
}

type toolSchemaModel struct {
	InlinePayload fwtypes.ListNestedObjectValueOf[toolDefinitionModel]  `tfsdk:"inline_payload"`
	S3            fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = toolSchemaModel{}
	_ fwflex.Flattener = &toolSchemaModel{}
)

func (m *toolSchemaModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ToolSchemaMemberInlinePayload:
		var toolDefModels []*toolDefinitionModel
		for _, toolDef := range t.Value {
			var model toolDefinitionModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, toolDef, &model))
			if diags.HasError() {
				return diags
			}
			toolDefModels = append(toolDefModels, &model)
		}
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, toolDefModels)

	case awstypes.ToolSchemaMemberS3:
		var model s3ConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("tool schema configuration flatten: %T", v),
		)
	}
	return diags
}

func (m toolSchemaModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadToolSchemaData, d := m.InlinePayload.ToSlice(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var toolDefs []awstypes.ToolDefinition
		for _, toolDefModel := range inlinePayloadToolSchemaData {
			var toolDef awstypes.ToolDefinition
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, toolDefModel, &toolDef))
			if diags.HasError() {
				return nil, diags
			}
			toolDefs = append(toolDefs, toolDef)
		}

		var r awstypes.ToolSchemaMemberInlinePayload
		r.Value = toolDefs
		return &r, diags

	case !m.S3.IsNull():
		s3ToolSchemaData, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ToolSchemaMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, s3ToolSchemaData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type toolDefinitionModel struct {
	Description  types.String                                           `tfsdk:"description"`
	Name         types.String                                           `tfsdk:"name"`
	InputSchema  fwtypes.ListNestedObjectValueOf[schemaDefinitionModel] `tfsdk:"input_schema"`
	OutputSchema fwtypes.ListNestedObjectValueOf[schemaDefinitionModel] `tfsdk:"output_schema"`
}

type schemaDefinitionCoreModel struct {
	Description types.String                                      `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]           `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsModel] `tfsdk:"items"`
}

type schemaDefinitionModel struct {
	schemaDefinitionCoreModel
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyModel] `tfsdk:"property"`
}

var (
	_ fwflex.Expander  = schemaDefinitionModel{}
	_ fwflex.Flattener = &schemaDefinitionModel{}
)

func (m *schemaDefinitionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaDefinitionCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema definition flatten: %T", v),
		)
	}

	return diags
}

func (m schemaDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(fwflex.Expand(ctx, m.schemaDefinitionCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = properties
		schemaDefinitionData.Required = requiredProps
	}

	return schemaDefinitionData, diags
}

type schemaItemsCoreModel struct {
	Description types.String                                          `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]               `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsLeafModel] `tfsdk:"items"`
}

type schemaItemsModel struct {
	schemaItemsCoreModel
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel] `tfsdk:"property"`
}

var (
	_ fwflex.Expander  = schemaItemsModel{}
	_ fwflex.Flattener = &schemaItemsModel{}
)

func (m *schemaItemsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaItemsCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsLeafModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema items flatten: %T", v),
		)
	}

	return diags
}

func (m schemaItemsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	schemaDefinitionData := &awstypes.SchemaDefinition{}

	diags.Append(fwflex.Expand(ctx, m.schemaItemsCoreModel, schemaDefinitionData)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = properties
		schemaDefinitionData.Required = requiredProps
	}

	return schemaDefinitionData, diags
}

type schemaPropertyCoreModel struct {
	Name        types.String                                      `tfsdk:"name"`
	Description types.String                                      `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType]           `tfsdk:"type"`
	Items       fwtypes.ListNestedObjectValueOf[schemaItemsModel] `tfsdk:"items"`
}

type schemaPropertyModel struct {
	schemaPropertyCoreModel
	Required   types.Bool                                              `tfsdk:"required"`
	Properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel] `tfsdk:"property"`
}

var (
	_ fwflex.Expander  = schemaPropertyModel{}
	_ fwflex.Flattener = &schemaPropertyModel{}
)

func (m *schemaPropertyModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// Ensure Properties is a typed Null when absent to avoid zero-value Set panics during state encoding
	m.Properties = fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx)
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaPropertyCoreModel))
		if diags.HasError() {
			return diags
		}

		// Normalize: when API omits Items, return an empty list (not null)
		if t.Items == nil {
			m.Items = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*schemaItemsModel{})
		}

		if t.Properties != nil {
			properties, d := flattenTargetSchemaLeafProperties(ctx, t.Properties, t.Required)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.Properties = properties
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema property flatten: %T", v),
		)
	}
	return diags
}

func (m schemaPropertyModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var schemaDefinitionLeafData = awstypes.SchemaDefinition{}
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaPropertyCoreModel, &schemaDefinitionLeafData))
	if diags.HasError() {
		return nil, diags
	}

	if !m.Properties.IsNull() {
		properties, requiredProps, d := expandTargetSchemaLeafProperties(ctx, m.Properties)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionLeafData.Properties = properties
		schemaDefinitionLeafData.Required = requiredProps
	}

	return schemaDefinitionLeafData, diags
}

type schemaItemsLeafCoreModel struct {
	Description types.String                            `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType] `tfsdk:"type"`
}

type schemaItemsLeafModel struct {
	schemaItemsLeafCoreModel
	// JSON serialized schema for deeper nesting
	ItemsJSON      types.String `tfsdk:"items_json"`
	PropertiesJSON types.String `tfsdk:"properties_json"`
}

var (
	_ fwflex.Expander  = schemaItemsLeafModel{}
	_ fwflex.Flattener = &schemaItemsLeafModel{}
)

func (m *schemaItemsLeafModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaItemsLeafCoreModel))
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			s, err := tfjson.EncodeToString(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(s)
			}
		} else {
			m.ItemsJSON = types.StringNull()
		}
		// Populate PropertiesJSON
		if t.Properties != nil || len(t.Required) > 0 {
			propObj := awstypes.SchemaDefinition{
				Properties: t.Properties,
				Required:   t.Required,
			}
			jsonProps := convertToJSONSchemaDefinition(&propObj)
			s, err := tfjson.EncodeToString(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(s)
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema items leaf flatten: %T", v),
		)
	}
	return diags
}

func (m schemaItemsLeafModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var sd awstypes.SchemaDefinition
	// Expand core (type/description)
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaItemsLeafCoreModel, &sd))
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		sd.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		sd.Properties = jsd.Properties
		sd.Required = jsd.Required
	}

	return &sd, diags
}

type schemaPropertyLeafCoreModel struct {
	Name        types.String                            `tfsdk:"name"`
	Description types.String                            `tfsdk:"description"`
	Type        fwtypes.StringEnum[awstypes.SchemaType] `tfsdk:"type"`
}

type schemaPropertyLeafModel struct {
	schemaPropertyLeafCoreModel
	Required types.Bool `tfsdk:"required"`
	// JSON serialized schema for deeper nesting
	ItemsJSON      types.String `tfsdk:"items_json"`
	PropertiesJSON types.String `tfsdk:"properties_json"`
}

var (
	_ fwflex.Expander  = schemaPropertyLeafModel{}
	_ fwflex.Flattener = &schemaPropertyLeafModel{}
)

func (m *schemaPropertyLeafModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.SchemaDefinition:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, v, &m.schemaPropertyLeafCoreModel))
		if diags.HasError() {
			return diags
		}
		// Populate ItemsJSON
		if t.Items != nil {
			jsonItems := convertToJSONSchemaDefinition(t.Items)
			s, err := tfjson.EncodeToString(jsonItems)
			if err != nil {
				diags.AddWarning("Failed to marshal items for items_json", err.Error())
				m.ItemsJSON = types.StringNull()
			} else {
				m.ItemsJSON = types.StringValue(strings.TrimSpace(s))
			}
		} else {
			m.ItemsJSON = types.StringNull()
		}
		// Populate PropertiesJSON
		if t.Properties != nil || len(t.Required) > 0 {
			propObj := awstypes.SchemaDefinition{
				Properties: t.Properties,
				Required:   t.Required,
			}
			jsonProps := convertToJSONSchemaDefinition(&propObj)
			s, err := tfjson.EncodeToString(jsonProps)
			if err != nil {
				diags.AddWarning("Failed to marshal properties for properties_json", err.Error())
				m.PropertiesJSON = types.StringNull()
			} else {
				m.PropertiesJSON = types.StringValue(s)
			}
		} else {
			m.PropertiesJSON = types.StringNull()
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("schema property leaf flatten: %T", v),
		)
	}
	return diags
}

func (m schemaPropertyLeafModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var schemaDefinitionData = awstypes.SchemaDefinition{}

	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.schemaPropertyLeafCoreModel, &schemaDefinitionData))
	if diags.HasError() {
		return nil, diags
	}

	if isNonEmpty(m.ItemsJSON) {
		jsd, d := parseJSONSchemaDefinition(m.ItemsJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Items = jsd
	}
	if isNonEmpty(m.PropertiesJSON) {
		jsd, d := parseJSONSchemaDefinition(m.PropertiesJSON.ValueString())
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		schemaDefinitionData.Properties = jsd.Properties
		schemaDefinitionData.Required = jsd.Required
	}
	return schemaDefinitionData, diags
}

type s3ConfigurationModel struct {
	BucketOwnerAccountId types.String `tfsdk:"bucket_owner_account_id"`
	Uri                  types.String `tfsdk:"uri"`
}

type mcpServerTargetConfigurationModel struct {
	Endpoint    types.String                             `tfsdk:"endpoint"`
	ListingMode fwtypes.StringEnum[awstypes.ListingMode] `tfsdk:"listing_mode"`
}

type apiSchemaConfigurationModel struct {
	InlinePayload fwtypes.ListNestedObjectValueOf[inlinePayloadModel]   `tfsdk:"inline_payload"`
	S3            fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = apiSchemaConfigurationModel{}
	_ fwflex.Flattener = &apiSchemaConfigurationModel{}
)

func (m *apiSchemaConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	m.InlinePayload = fwtypes.NewListNestedObjectValueOfNull[inlinePayloadModel](ctx)
	m.S3 = fwtypes.NewListNestedObjectValueOfNull[s3ConfigurationModel](ctx)

	// This is called two ways: directly with the union interface still holding a
	// pointer (the passthrough schema path), and via AutoFlex which delivers a
	// dereferenced value (the mcp_server open_api_schema / smithy_model path).
	// Normalize a pointer to its value so both dispatch to the same cases.
	switch p := v.(type) {
	case *awstypes.ApiSchemaConfigurationMemberInlinePayload:
		v = *p
	case *awstypes.ApiSchemaConfigurationMemberS3:
		v = *p
	}

	switch t := v.(type) {
	case awstypes.ApiSchemaConfigurationMemberInlinePayload:
		var model inlinePayloadModel
		model.Payload = types.StringValue(t.Value)
		m.InlinePayload = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	case awstypes.ApiSchemaConfigurationMemberS3:
		var model s3ConfigurationModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("api schema configuration flatten: %T", v),
		)
	}
	return diags
}

func (m apiSchemaConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.InlinePayload.IsNull():
		inlinePayloadApiSchemaConfigurationData, d := m.InlinePayload.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberInlinePayload
		r.Value = inlinePayloadApiSchemaConfigurationData.Payload.ValueString()
		return &r, diags

	case !m.S3.IsNull():
		s3ApiSchemaConfigurationData, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ApiSchemaConfigurationMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, s3ApiSchemaConfigurationData, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	diags.AddError(
		"Invalid Schema Configuration",
		"Exactly one of \"inline_payload\" or \"s3\" must be configured.",
	)
	return nil, diags
}

type inlinePayloadModel struct {
	Payload types.String `tfsdk:"payload"`
}

// Helper functions for PropertiesJSON map conversion
func flattenTargetSchemaProperties(
	ctx context.Context,
	properties map[string]awstypes.SchemaDefinition,
	required []string,
) (fwtypes.SetNestedObjectValueOf[schemaPropertyModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(properties) == 0 {
		return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx), diags
	}

	requiredSet := map[string]bool{}
	for _, n := range required {
		requiredSet[n] = true
	}

	var propertyModels []*schemaPropertyModel
	for name, schemaDefn := range properties {
		pm := &schemaPropertyModel{}
		d := pm.Flatten(ctx, schemaDefn)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyModel](ctx), diags
		}

		pm.Name = types.StringValue(name)
		pm.Required = types.BoolValue(requiredSet[name])

		propertyModels = append(propertyModels, pm)
	}

	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, propertyModels), diags
}

func expandTargetSchemaProperties(ctx context.Context, properties fwtypes.SetNestedObjectValueOf[schemaPropertyModel]) (map[string]awstypes.SchemaDefinition, []string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]awstypes.SchemaDefinition)
	var requiredProps []string

	propertySlice, d := properties.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, nil, diags
		}

		if schemaDefn, ok := expandedValue.(awstypes.SchemaDefinition); ok {
			name := propertyModel.Name.ValueString()
			result[name] = schemaDefn

			// Since we always set required to explicit boolean, we can check it directly
			if propertyModel.Required.ValueBool() {
				requiredProps = append(requiredProps, name)
			}
		}
	}
	return result, requiredProps, diags
}

// Helper functions for Leaf PropertiesJSON map conversion
func flattenTargetSchemaLeafProperties(ctx context.Context, properties map[string]awstypes.SchemaDefinition, requiredProps []string) (fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	requiredSet := make(map[string]bool)
	for _, prop := range requiredProps {
		requiredSet[prop] = true
	}

	var propertyModels []*schemaPropertyLeafModel
	for name, schemaDefn := range properties {
		pm := &schemaPropertyLeafModel{}
		d := pm.Flatten(ctx, schemaDefn)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return fwtypes.NewSetNestedObjectValueOfNull[schemaPropertyLeafModel](ctx), diags
		}
		pm.Name = types.StringValue(name)
		pm.Required = types.BoolValue(requiredSet[name])
		propertyModels = append(propertyModels, pm)
	}
	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, propertyModels), diags
}

func expandTargetSchemaLeafProperties(ctx context.Context, properties fwtypes.SetNestedObjectValueOf[schemaPropertyLeafModel]) (map[string]awstypes.SchemaDefinition, []string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]awstypes.SchemaDefinition)
	var requiredProps []string

	propertySlice, d := properties.ToSlice(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, nil, diags
	}

	for _, propertyModel := range propertySlice {
		expandedValue, d := propertyModel.Expand(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, nil, diags
		}

		if schemaDefn, ok := expandedValue.(awstypes.SchemaDefinition); ok {
			name := propertyModel.Name.ValueString()
			result[name] = schemaDefn

			if propertyModel.Required.ValueBool() {
				requiredProps = append(requiredProps, name)
			}
		}
	}

	return result, requiredProps, diags
}

func parseJSONSchemaDefinition(s string) (*awstypes.SchemaDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	s = strings.TrimSpace(s)
	if s == "" {
		diags.AddError("Invalid JSON", "JSON schema must be a non-empty string")
		return nil, diags
	}
	var sd awstypes.SchemaDefinition
	if err := tfjson.DecodeFromString(s, &sd); err != nil {
		diags.AddError("Invalid JSON", err.Error())
		return nil, diags
	}
	return &sd, diags
}

func isNonEmpty(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown() && strings.TrimSpace(s.ValueString()) != ""
}

// jsonSchemaDefinition is a helper struct for JSON serialization with lowercase field names
type jsonSchemaDefinition struct {
	Type        string                           `json:"type,omitempty"`
	Description *string                          `json:"description,omitempty"`
	Items       *jsonSchemaDefinition            `json:"items,omitempty"`
	Properties  map[string]*jsonSchemaDefinition `json:"properties,omitempty"`
	Required    []string                         `json:"required,omitempty"`
}

// convertToJSONSchemaDefinition converts AWS SDK SchemaDefinition to our JSON-friendly version
func convertToJSONSchemaDefinition(sd *awstypes.SchemaDefinition) *jsonSchemaDefinition {
	if sd == nil {
		return nil
	}

	jsd := &jsonSchemaDefinition{
		Type: string(sd.Type), // Convert SchemaType enum to string
	}

	// Only set non-nil values to avoid null fields in JSON
	if sd.Description != nil && aws.ToString(sd.Description) != "" {
		jsd.Description = sd.Description
	}
	if len(sd.Required) > 0 {
		jsd.Required = sd.Required
	}

	if sd.Items != nil {
		jsd.Items = convertToJSONSchemaDefinition(sd.Items)
	}

	if sd.Properties != nil {
		jsd.Properties = make(map[string]*jsonSchemaDefinition)
		for k, v := range sd.Properties {
			if converted := convertToJSONSchemaDefinition(&v); converted != nil {
				jsd.Properties[k] = converted
			}
		}
		// If no properties were added, don't include the properties field
		if len(jsd.Properties) == 0 {
			jsd.Properties = nil
		}
	}

	return jsd
}
