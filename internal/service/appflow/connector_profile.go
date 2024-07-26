// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appflow

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appflow"
	"github.com/aws/aws-sdk-go-v2/service/appflow/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appflow_connector_profile", name="Connector Profile")
func resourceConnectorProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorProfileCreate,
		ReadWithoutTimeout:   resourceConnectorProfileRead,
		UpdateWithoutTimeout: resourceConnectorProfileUpdate,
		DeleteWithoutTimeout: resourceConnectorProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connector_label": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][\w!@#.-]+`), "must contain only alphanumeric, exclamation point (!), at sign (@), number sign (#), period (.), and hyphen (-) characters"),
					validation.StringLenBetween(1, 256),
				),
			},
			"connection_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ConnectionMode](),
			},
			"connector_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_profile_credentials": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrSecretKey: {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"custom_connector": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"api_key": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"api_secret_key": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"authentication_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.AuthenticationType](),
												},
												"basic": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrPassword: {
																Type:         schema.TypeString,
																Required:     true,
																Sensitive:    true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															names.AttrUsername: {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
														},
													},
												},
												"custom": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"credentials_map": {
																Type:      schema.TypeMap,
																Optional:  true,
																Sensitive: true,
																ValidateDiagFunc: validation.AllDiag(
																	validation.MapKeyLenBetween(1, 128),
																	validation.MapKeyMatch(regexache.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
																),
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(0, 2048),
																		validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
															"custom_authentication_type": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
															},
														},
													},
												},
												"oauth2": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"access_token": {
																Type:      schema.TypeString,
																Optional:  true,
																Sensitive: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 4096),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															names.AttrClientID: {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															names.AttrClientSecret: {
																Type:      schema.TypeString,
																Optional:  true,
																Sensitive: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"oauth_request": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"auth_code": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 4096),
																				validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																		"redirect_uri": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 512),
																				validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																	},
																},
															},
															"refresh_token": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 4096),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
											},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"application_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_token": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientID: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientSecret: {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_key_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"datakey": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"secret_access_key": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"user_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientID: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientSecret: {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPassword: {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrUsername: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_credentials_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"jwt_token": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 8000),
												},
												"oauth2_grant_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.OAuth2GrantType](),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"sapo_data": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"basic_auth_credentials": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrPassword: {
																Type:         schema.TypeString,
																Required:     true,
																Sensitive:    true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															names.AttrUsername: {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"oauth_credentials": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"access_token": {
																Type:      schema.TypeString,
																Optional:  true,
																Sensitive: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															names.AttrClientID: {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															names.AttrClientSecret: {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"oauth_request": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"auth_code": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 2048),
																				validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																		"redirect_uri": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 512),
																				validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																	},
																},
															},
															"refresh_token": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 1024),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPassword: {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrUsername: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientID: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientSecret: {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPassword: {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrUsername: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_secret_key": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPassword: {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrUsername: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 2048),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientID: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrClientSecret: {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
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
						"connector_profile_properties": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"custom_connector": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"oauth2_properties": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"oauth2_grant_type": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[types.OAuth2GrantType](),
															},
															"token_url": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexache.MustCompile(`^(https?)://[0-9A-Za-z-+&@#/%?=~_|!:,.;]*[0-9A-Za-z-+&@#/%=~_|]`), "must provide a valid HTTPS url"),
																),
															},
															"token_url_custom_properties": {
																Type:     schema.TypeMap,
																Optional: true,
																ValidateDiagFunc: validation.AllDiag(
																	validation.MapKeyLenBetween(1, 128),
																	validation.MapKeyMatch(regexache.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
																),
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(0, 2048),
																		validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
														},
													},
												},
												"profile_properties": {
													Type:     schema.TypeMap,
													Optional: true,
													ValidateDiagFunc: validation.AllDiag(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexache.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type: schema.TypeString,
														ValidateFunc: validation.All(
															validation.StringLenBetween(0, 2048),
															validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
														),
													},
												},
											},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketName: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrClusterIdentifier: {
													Type:     schema.TypeString,
													Optional: true,
												},
												"data_api_role_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrDatabaseName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												"database_url": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												names.AttrRoleARN: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"is_sandbox_environment": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"sapo_data": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"application_host_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`^(https?)://[0-9A-Za-z-+&@#/%?=~_|!:,.;]*[0-9A-Za-z-+&@#/%=~_|]`), "must provide a valid HTTPS url"),
													),
												},
												"application_service_path": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_number": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 3),
														validation.StringMatch(regexache.MustCompile(`^\d{3}$`), "must consist of exactly three digits"),
													),
												},
												"logon_language": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 2),
														validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]*$`), "must contain only alphanumeric characters and the underscore (_) character"),
													),
												},
												"oauth_properties": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code_url": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexache.MustCompile(`^(https?)://[0-9A-Za-z-+&@#/%?=~_|!:,.;]*[0-9A-Za-z-+&@#/%=~_|]`), "must provide a valid HTTPS url"),
																),
															},
															"oauth_scopes": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(1, 128),
																		validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
															"token_url": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexache.MustCompile(`^(https?)://[0-9A-Za-z-+&@#/%?=~_|!:,.;]*[0-9A-Za-z-+&@#/%=~_|]`), "must provide a valid HTTPS url"),
																),
															},
														},
													},
												},
												"port_number": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 65535),
												},
												"private_link_service_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`^$|com.amazonaws.vpce.[\w/!:@#.\-]+`), "must be a valid AWS VPC endpoint address"),
													),
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"account_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrBucketName: {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"private_link_service_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`^$|com.amazonaws.vpce.[\w/!:@#.\-]+`), "must be a valid AWS VPC endpoint address"),
													),
												},
												names.AttrRegion: {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 64),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												names.AttrStage: {
													Type:     schema.TypeString,
													Required: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														return old == new || old == "@"+new
													},
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"warehouse": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 512),
														validation.StringMatch(regexache.MustCompile(`[\s\w/!@#+=.-]*`), "must match [\\s\\w/!@#+=.-]*"),
													),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
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
			"connector_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ConnectorType](),
			},
			"credentials_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`[\w/!@#+=.-]+`), "must match [\\w/!@#+=.-]+"),
				),
			},
		},
	}
}

func resourceConnectorProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appflow.CreateConnectorProfileInput{
		ConnectionMode:       types.ConnectionMode(d.Get("connection_mode").(string)),
		ConnectorProfileName: aws.String(name),
		ConnectorType:        types.ConnectorType(d.Get("connector_type").(string)),
	}

	if v, ok := d.Get("connector_label").(string); ok && len(v) > 0 {
		input.ConnectorLabel = aws.String(v)
	}

	if v, ok := d.GetOk("connector_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConnectorProfileConfig = expandConnectorProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.Get("kms_arn").(string); ok && len(v) > 0 {
		input.KmsArn = aws.String(v)
	}

	output, err := conn.CreateConnectorProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppFlow Connector Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ConnectorProfileArn))

	return append(diags, resourceConnectorProfileRead(ctx, d, meta)...)
}

func resourceConnectorProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	connectorProfile, err := findConnectorProfileByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppFlow Connector Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	// Credentials are not returned by any API operation. Instead, a
	// "credentials_arn" property is returned.
	//
	// It may be possible to implement a function that reads from this
	// credentials resource -- but it is not documented in the API reference.
	// (https://docs.aws.amazon.com/appflow/1.0/APIReference/API_ConnectorProfile.html#appflow-Type-ConnectorProfile-credentialsArn)
	credentials := d.Get("connector_profile_config.0.connector_profile_credentials").([]interface{})
	d.Set(names.AttrARN, connectorProfile.ConnectorProfileArn)
	d.Set("connection_mode", connectorProfile.ConnectionMode)
	d.Set("connector_label", connectorProfile.ConnectorLabel)
	d.Set("connector_profile_config", flattenConnectorProfileConfig(connectorProfile.ConnectorProfileProperties, credentials))
	d.Set("connector_type", connectorProfile.ConnectorType)
	d.Set("credentials_arn", connectorProfile.CredentialsArn)
	d.Set(names.AttrName, connectorProfile.ConnectorProfileName)

	return diags
}

func resourceConnectorProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appflow.UpdateConnectorProfileInput{
		ConnectionMode:       types.ConnectionMode(d.Get("connection_mode").(string)),
		ConnectorProfileName: aws.String(name),
	}

	if v, ok := d.GetOk("connector_profile_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConnectorProfileConfig = expandConnectorProfileConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateConnectorProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConnectorProfileRead(ctx, d, meta)...)
}

func resourceConnectorProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	log.Printf("[INFO] Deleting AppFlow Connector Profile: %s", d.Id())
	_, err := conn.DeleteConnectorProfile(ctx, &appflow.DeleteConnectorProfileInput{
		ConnectorProfileName: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findConnectorProfileByARN(ctx context.Context, conn *appflow.Client, arn string) (*types.ConnectorProfile, error) {
	input := &appflow.DescribeConnectorProfilesInput{}

	pages := appflow.NewDescribeConnectorProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ConnectorProfileDetails {
			if aws.ToString(v.ConnectorProfileArn) == arn {
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func expandConnectorProfileConfig(m map[string]interface{}) *types.ConnectorProfileConfig {
	cpc := &types.ConnectorProfileConfig{}

	if v, ok := m["connector_profile_credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ConnectorProfileCredentials = expandConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["connector_profile_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ConnectorProfileProperties = expandConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	return cpc
}

func expandConnectorProfileCredentials(m map[string]interface{}) *types.ConnectorProfileCredentials {
	cpc := &types.ConnectorProfileCredentials{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Amplitude = expandAmplitudeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["custom_connector"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.CustomConnector = expandCustomConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Datadog = expandDatadogConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Dynatrace = expandDynatraceConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.GoogleAnalytics = expandGoogleAnalyticsConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Honeycode = expandHoneycodeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.InforNexus = expandInforNexusConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Marketo = expandMarketoConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["redshift"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Redshift = expandRedshiftConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Salesforce = expandSalesforceConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["sapo_data"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.SAPOData = expandSAPODataConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ServiceNow = expandServiceNowConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Singular = expandSingularConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Slack = expandSlackConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Snowflake = expandSnowflakeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Trendmicro = expandTrendmicroConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Veeva = expandVeevaConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Zendesk = expandZendeskConnectorProfileCredentials(v[0].(map[string]interface{}))
	}

	return cpc
}

func expandAmplitudeConnectorProfileCredentials(m map[string]interface{}) *types.AmplitudeConnectorProfileCredentials {
	credentials := &types.AmplitudeConnectorProfileCredentials{
		ApiKey:    aws.String(m["api_key"].(string)),
		SecretKey: aws.String(m[names.AttrSecretKey].(string)),
	}

	return credentials
}

func expandCustomConnectorProfileCredentials(m map[string]interface{}) *types.CustomConnectorProfileCredentials {
	credentials := &types.CustomConnectorProfileCredentials{
		AuthenticationType: types.AuthenticationType(m["authentication_type"].(string)),
	}

	if v, ok := m["api_key"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.ApiKey = expandAPIKeyCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["basic"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.Basic = expandBasicAuthCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["custom"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.Custom = expandCustomAuthCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["oauth2"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.Oauth2 = expandOAuth2Credentials(v[0].(map[string]interface{}))
	}

	return credentials
}

func expandDatadogConnectorProfileCredentials(m map[string]interface{}) *types.DatadogConnectorProfileCredentials {
	credentials := &types.DatadogConnectorProfileCredentials{
		ApiKey:         aws.String(m["api_key"].(string)),
		ApplicationKey: aws.String(m["application_key"].(string)),
	}

	return credentials
}

func expandDynatraceConnectorProfileCredentials(m map[string]interface{}) *types.DynatraceConnectorProfileCredentials {
	credentials := &types.DynatraceConnectorProfileCredentials{
		ApiToken: aws.String(m["api_token"].(string)),
	}

	return credentials
}

func expandGoogleAnalyticsConnectorProfileCredentials(m map[string]interface{}) *types.GoogleAnalyticsConnectorProfileCredentials {
	credentials := &types.GoogleAnalyticsConnectorProfileCredentials{
		ClientId:     aws.String(m[names.AttrClientID].(string)),
		ClientSecret: aws.String(m[names.AttrClientSecret].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}
	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return credentials
}

func expandHoneycodeConnectorProfileCredentials(m map[string]interface{}) *types.HoneycodeConnectorProfileCredentials {
	credentials := &types.HoneycodeConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}
	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return credentials
}

func expandInforNexusConnectorProfileCredentials(m map[string]interface{}) *types.InforNexusConnectorProfileCredentials {
	credentials := &types.InforNexusConnectorProfileCredentials{
		AccessKeyId:     aws.String(m["access_key_id"].(string)),
		Datakey:         aws.String(m["datakey"].(string)),
		SecretAccessKey: aws.String(m["secret_access_key"].(string)),
		UserId:          aws.String(m["user_id"].(string)),
	}

	return credentials
}

func expandMarketoConnectorProfileCredentials(m map[string]interface{}) *types.MarketoConnectorProfileCredentials {
	credentials := &types.MarketoConnectorProfileCredentials{
		ClientId:     aws.String(m[names.AttrClientID].(string)),
		ClientSecret: aws.String(m[names.AttrClientSecret].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return credentials
}

func expandRedshiftConnectorProfileCredentials(m map[string]interface{}) *types.RedshiftConnectorProfileCredentials {
	credentials := &types.RedshiftConnectorProfileCredentials{
		Password: aws.String(m[names.AttrPassword].(string)),
		Username: aws.String(m[names.AttrUsername].(string)),
	}

	return credentials
}

func expandSalesforceConnectorProfileCredentials(m map[string]interface{}) *types.SalesforceConnectorProfileCredentials {
	credentials := &types.SalesforceConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m["client_credentials_arn"].(string); ok && v != "" {
		credentials.ClientCredentialsArn = aws.String(v)
	}
	if v, ok := m["jwt_token"].(string); ok && v != "" {
		credentials.JwtToken = aws.String(v)
	}
	if v, ok := m["oauth2_grant_type"].(string); ok && v != "" {
		credentials.OAuth2GrantType = types.OAuth2GrantType(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}
	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return credentials
}

func expandSAPODataConnectorProfileCredentials(m map[string]interface{}) *types.SAPODataConnectorProfileCredentials {
	credentials := &types.SAPODataConnectorProfileCredentials{}

	if v, ok := m["basic_auth_credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.BasicAuthCredentials = expandBasicAuthCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["oauth_credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthCredentials = expandOAuthCredentials(v[0].(map[string]interface{}))
	}

	return credentials
}

func expandServiceNowConnectorProfileCredentials(m map[string]interface{}) *types.ServiceNowConnectorProfileCredentials {
	credentials := &types.ServiceNowConnectorProfileCredentials{
		Password: aws.String(m[names.AttrPassword].(string)),
		Username: aws.String(m[names.AttrUsername].(string)),
	}

	return credentials
}

func expandSingularConnectorProfileCredentials(m map[string]interface{}) *types.SingularConnectorProfileCredentials {
	credentials := &types.SingularConnectorProfileCredentials{
		ApiKey: aws.String(m["api_key"].(string)),
	}

	return credentials
}

func expandSlackConnectorProfileCredentials(m map[string]interface{}) *types.SlackConnectorProfileCredentials {
	credentials := &types.SlackConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m[names.AttrClientID].(string)),
		ClientSecret: aws.String(m[names.AttrClientSecret].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return credentials
}

func expandSnowflakeConnectorProfileCredentials(m map[string]interface{}) *types.SnowflakeConnectorProfileCredentials {
	credentials := &types.SnowflakeConnectorProfileCredentials{
		Password: aws.String(m[names.AttrPassword].(string)),
		Username: aws.String(m[names.AttrUsername].(string)),
	}

	return credentials
}

func expandTrendmicroConnectorProfileCredentials(m map[string]interface{}) *types.TrendmicroConnectorProfileCredentials {
	credentials := &types.TrendmicroConnectorProfileCredentials{
		ApiSecretKey: aws.String(m["api_secret_key"].(string)),
	}

	return credentials
}

func expandVeevaConnectorProfileCredentials(m map[string]interface{}) *types.VeevaConnectorProfileCredentials {
	credentials := &types.VeevaConnectorProfileCredentials{
		Password: aws.String(m[names.AttrPassword].(string)),
		Username: aws.String(m[names.AttrUsername].(string)),
	}

	return credentials
}

func expandZendeskConnectorProfileCredentials(m map[string]interface{}) *types.ZendeskConnectorProfileCredentials {
	credentials := &types.ZendeskConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m[names.AttrClientID].(string)),
		ClientSecret: aws.String(m[names.AttrClientSecret].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return credentials
}

func expandOAuthRequest(m map[string]interface{}) *types.ConnectorOAuthRequest {
	r := &types.ConnectorOAuthRequest{}

	if v, ok := m["auth_code"].(string); ok && v != "" {
		r.AuthCode = aws.String(v)
	}

	if v, ok := m["redirect_uri"].(string); ok && v != "" {
		r.RedirectUri = aws.String(v)
	}

	return r
}

func expandAPIKeyCredentials(m map[string]interface{}) *types.ApiKeyCredentials {
	credentials := &types.ApiKeyCredentials{}

	if v, ok := m["api_key"].(string); ok && v != "" {
		credentials.ApiKey = aws.String(v)
	}

	if v, ok := m["api_secret_key"].(string); ok && v != "" {
		credentials.ApiSecretKey = aws.String(v)
	}

	return credentials
}

func expandBasicAuthCredentials(m map[string]interface{}) *types.BasicAuthCredentials {
	credentials := &types.BasicAuthCredentials{}

	if v, ok := m[names.AttrPassword].(string); ok && v != "" {
		credentials.Password = aws.String(v)
	}

	if v, ok := m[names.AttrUsername].(string); ok && v != "" {
		credentials.Username = aws.String(v)
	}

	return credentials
}

func expandCustomAuthCredentials(m map[string]interface{}) *types.CustomAuthCredentials {
	credentials := &types.CustomAuthCredentials{}

	if v, ok := m["credentials_map"].(map[string]interface{}); ok && len(v) > 0 {
		credentials.CredentialsMap = flex.ExpandStringValueMap(v)
	}

	if v, ok := m["custom_authentication_type"].(string); ok && v != "" {
		credentials.CustomAuthenticationType = aws.String(v)
	}

	return credentials
}

func expandOAuthCredentials(m map[string]interface{}) *types.OAuthCredentials {
	credentials := &types.OAuthCredentials{
		ClientId:     aws.String(m[names.AttrClientID].(string)),
		ClientSecret: aws.String(m[names.AttrClientSecret].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}
	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return credentials
}

func expandOAuth2Credentials(m map[string]interface{}) *types.OAuth2Credentials {
	credentials := &types.OAuth2Credentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}
	if v, ok := m[names.AttrClientID].(string); ok && v != "" {
		credentials.ClientId = aws.String(v)
	}
	if v, ok := m[names.AttrClientSecret].(string); ok && v != "" {
		credentials.ClientSecret = aws.String(v)
	}
	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}
	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return credentials
}

func expandConnectorProfileProperties(m map[string]interface{}) *types.ConnectorProfileProperties {
	cpc := &types.ConnectorProfileProperties{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Amplitude = v[0].(*types.AmplitudeConnectorProfileProperties)
	}
	if v, ok := m["custom_connector"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.CustomConnector = expandCustomConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Datadog = expandDatadogConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Dynatrace = expandDynatraceConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.GoogleAnalytics = v[0].(*types.GoogleAnalyticsConnectorProfileProperties)
	}
	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Honeycode = v[0].(*types.HoneycodeConnectorProfileProperties)
	}
	if v, ok := m["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.InforNexus = expandInforNexusConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Marketo = expandMarketoConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["redshift"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Redshift = expandRedshiftConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Salesforce = expandSalesforceConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["sapo_data"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.SAPOData = expandSAPODataConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ServiceNow = expandServiceNowConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Singular = v[0].(*types.SingularConnectorProfileProperties)
	}
	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Slack = expandSlackConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Snowflake = expandSnowflakeConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Trendmicro = v[0].(*types.TrendmicroConnectorProfileProperties)
	}
	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Veeva = expandVeevaConnectorProfileProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Zendesk = expandZendeskConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	return cpc
}

func expandDatadogConnectorProfileProperties(m map[string]interface{}) *types.DatadogConnectorProfileProperties {
	properties := &types.DatadogConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandDynatraceConnectorProfileProperties(m map[string]interface{}) *types.DynatraceConnectorProfileProperties {
	properties := &types.DynatraceConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandInforNexusConnectorProfileProperties(m map[string]interface{}) *types.InforNexusConnectorProfileProperties {
	properties := &types.InforNexusConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandMarketoConnectorProfileProperties(m map[string]interface{}) *types.MarketoConnectorProfileProperties {
	properties := &types.MarketoConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandRedshiftConnectorProfileProperties(m map[string]interface{}) *types.RedshiftConnectorProfileProperties {
	properties := &types.RedshiftConnectorProfileProperties{
		BucketName:        aws.String(m[names.AttrBucketName].(string)),
		ClusterIdentifier: aws.String(m[names.AttrClusterIdentifier].(string)),
		RoleArn:           aws.String(m[names.AttrRoleARN].(string)),
		DataApiRoleArn:    aws.String(m["data_api_role_arn"].(string)),
		DatabaseName:      aws.String(m[names.AttrDatabaseName].(string)),
	}

	if v, ok := m[names.AttrBucketPrefix].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["database_url"].(string); ok && v != "" {
		properties.DatabaseUrl = aws.String(v)
	}

	return properties
}

func expandServiceNowConnectorProfileProperties(m map[string]interface{}) *types.ServiceNowConnectorProfileProperties {
	properties := &types.ServiceNowConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandSalesforceConnectorProfileProperties(m map[string]interface{}) *types.SalesforceConnectorProfileProperties {
	properties := &types.SalesforceConnectorProfileProperties{}

	if v, ok := m["instance_url"].(string); ok && v != "" {
		properties.InstanceUrl = aws.String(v)
	}

	if v, ok := m["is_sandbox_environment"].(bool); ok {
		properties.IsSandboxEnvironment = v
	}

	return properties
}

func expandCustomConnectorProfileProperties(m map[string]interface{}) *types.CustomConnectorProfileProperties {
	properties := &types.CustomConnectorProfileProperties{}

	if v, ok := m["oauth2_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		properties.OAuth2Properties = expandOAuth2Properties(v[0].(map[string]interface{}))
	}
	if v, ok := m["profile_properties"].(map[string]interface{}); ok && len(v) > 0 {
		properties.ProfileProperties = flex.ExpandStringValueMap(v)
	}

	return properties
}

func expandSAPODataConnectorProfileProperties(m map[string]interface{}) *types.SAPODataConnectorProfileProperties {
	properties := &types.SAPODataConnectorProfileProperties{
		ApplicationHostUrl:     aws.String(m["application_host_url"].(string)),
		ApplicationServicePath: aws.String(m["application_service_path"].(string)),
		ClientNumber:           aws.String(m["client_number"].(string)),
		PortNumber:             aws.Int32(int32(m["port_number"].(int))),
	}

	if v, ok := m["logon_language"].(string); ok && v != "" {
		properties.LogonLanguage = aws.String(v)
	}
	if v, ok := m["oauth_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		properties.OAuthProperties = expandOAuthProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["private_link_service_name"].(string); ok && v != "" {
		properties.PrivateLinkServiceName = aws.String(v)
	}

	return properties
}

func expandSlackConnectorProfileProperties(m map[string]interface{}) *types.SlackConnectorProfileProperties {
	properties := &types.SlackConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandSnowflakeConnectorProfileProperties(m map[string]interface{}) *types.SnowflakeConnectorProfileProperties {
	properties := &types.SnowflakeConnectorProfileProperties{
		BucketName: aws.String(m[names.AttrBucketName].(string)),
		Stage:      aws.String(m[names.AttrStage].(string)),
		Warehouse:  aws.String(m["warehouse"].(string)),
	}

	if v, ok := m["account_name"].(string); ok && v != "" {
		properties.AccountName = aws.String(v)
	}

	if v, ok := m[names.AttrBucketPrefix].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["private_link_service_name"].(string); ok && v != "" {
		properties.PrivateLinkServiceName = aws.String(v)
	}

	if v, ok := m[names.AttrRegion].(string); ok && v != "" {
		properties.Region = aws.String(v)
	}

	return properties
}

func expandVeevaConnectorProfileProperties(m map[string]interface{}) *types.VeevaConnectorProfileProperties {
	properties := &types.VeevaConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandZendeskConnectorProfileProperties(m map[string]interface{}) *types.ZendeskConnectorProfileProperties {
	properties := &types.ZendeskConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return properties
}

func expandOAuthProperties(m map[string]interface{}) *types.OAuthProperties {
	properties := &types.OAuthProperties{
		AuthCodeUrl: aws.String(m["auth_code_url"].(string)),
		OAuthScopes: flex.ExpandStringValueList(m["oauth_scopes"].([]interface{})),
		TokenUrl:    aws.String(m["token_url"].(string)),
	}

	return properties
}

func expandOAuth2Properties(m map[string]interface{}) *types.OAuth2Properties {
	properties := &types.OAuth2Properties{
		OAuth2GrantType: types.OAuth2GrantType(m["oauth2_grant_type"].(string)),
		TokenUrl:        aws.String(m["token_url"].(string)),
	}

	if v, ok := m["token_url_custom_properties"].(map[string]interface{}); ok && len(v) > 0 {
		properties.TokenUrlCustomProperties = flex.ExpandStringValueMap(v)
	}

	return properties
}

func flattenConnectorProfileConfig(cpp *types.ConnectorProfileProperties, cpc []interface{}) []interface{} {
	m := make(map[string]interface{})

	m["connector_profile_properties"] = flattenConnectorProfileProperties(cpp)
	m["connector_profile_credentials"] = cpc

	return []interface{}{m}
}

func flattenConnectorProfileProperties(cpp *types.ConnectorProfileProperties) []interface{} {
	result := make(map[string]interface{})
	m := make(map[string]interface{})

	if cpp.Amplitude != nil {
		result["amplitude"] = []interface{}{m}
	}
	if cpp.CustomConnector != nil {
		result["custom_connector"] = flattenCustomConnectorProfileProperties(cpp.CustomConnector)
	}
	if cpp.Datadog != nil {
		m["instance_url"] = aws.ToString(cpp.Datadog.InstanceUrl)
		result["datadog"] = []interface{}{m}
	}
	if cpp.Dynatrace != nil {
		m["instance_url"] = aws.ToString(cpp.Dynatrace.InstanceUrl)
		result["dynatrace"] = []interface{}{m}
	}
	if cpp.GoogleAnalytics != nil {
		result["google_analytics"] = []interface{}{m}
	}
	if cpp.Honeycode != nil {
		result["honeycode"] = []interface{}{m}
	}
	if cpp.InforNexus != nil {
		m["instance_url"] = aws.ToString(cpp.InforNexus.InstanceUrl)
		result["infor_nexus"] = []interface{}{m}
	}
	if cpp.Marketo != nil {
		m["instance_url"] = aws.ToString(cpp.Marketo.InstanceUrl)
		result["marketo"] = []interface{}{m}
	}
	if cpp.Redshift != nil {
		result["redshift"] = flattenRedshiftConnectorProfileProperties(cpp.Redshift)
	}
	if cpp.SAPOData != nil {
		result["sapo_data"] = flattenSAPODataConnectorProfileProperties(cpp.SAPOData)
	}
	if cpp.Salesforce != nil {
		result["salesforce"] = flattenSalesforceConnectorProfileProperties(cpp.Salesforce)
	}
	if cpp.ServiceNow != nil {
		m["instance_url"] = aws.ToString(cpp.ServiceNow.InstanceUrl)
		result["service_now"] = []interface{}{m}
	}
	if cpp.Singular != nil {
		result["singular"] = []interface{}{m}
	}
	if cpp.Slack != nil {
		m["instance_url"] = aws.ToString(cpp.Slack.InstanceUrl)
		result["slack"] = []interface{}{m}
	}
	if cpp.Snowflake != nil {
		result["snowflake"] = flattenSnowflakeConnectorProfileProperties(cpp.Snowflake)
	}
	if cpp.Trendmicro != nil {
		result["trendmicro"] = []interface{}{m}
	}
	if cpp.Veeva != nil {
		m["instance_url"] = aws.ToString(cpp.Veeva.InstanceUrl)
		result["veeva"] = []interface{}{m}
	}
	if cpp.Zendesk != nil {
		m["instance_url"] = aws.ToString(cpp.Zendesk.InstanceUrl)
		result["zendesk"] = []interface{}{m}
	}

	return []interface{}{result}
}

func flattenRedshiftConnectorProfileProperties(properties *types.RedshiftConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	m[names.AttrBucketName] = aws.ToString(properties.BucketName)

	if properties.BucketPrefix != nil {
		m[names.AttrBucketPrefix] = aws.ToString(properties.BucketPrefix)
	}

	if properties.DatabaseUrl != nil {
		m["database_url"] = aws.ToString(properties.DatabaseUrl)
	}

	m[names.AttrRoleARN] = aws.ToString(properties.RoleArn)
	m[names.AttrClusterIdentifier] = aws.ToString(properties.ClusterIdentifier)
	m["data_api_role_arn"] = aws.ToString(properties.DataApiRoleArn)
	m[names.AttrDatabaseName] = aws.ToString(properties.DatabaseName)

	return []interface{}{m}
}

func flattenCustomConnectorProfileProperties(properties *types.CustomConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	if properties.OAuth2Properties != nil {
		m["oauth2_properties"] = flattenOAuth2Properties(properties.OAuth2Properties)
	}

	if properties.ProfileProperties != nil {
		m["profile_properties"] = properties.ProfileProperties
	}

	return []interface{}{m}
}

func flattenSalesforceConnectorProfileProperties(properties *types.SalesforceConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	if properties.InstanceUrl != nil {
		m["instance_url"] = aws.ToString(properties.InstanceUrl)
	}
	m["is_sandbox_environment"] = properties.IsSandboxEnvironment

	return []interface{}{m}
}

func flattenSAPODataConnectorProfileProperties(properties *types.SAPODataConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	m["application_host_url"] = aws.ToString(properties.ApplicationHostUrl)
	m["application_service_path"] = aws.ToString(properties.ApplicationServicePath)
	m["client_number"] = aws.ToString(properties.ClientNumber)
	m["port_number"] = aws.ToInt32(properties.PortNumber)

	if properties.LogonLanguage != nil {
		m["logon_language"] = aws.ToString(properties.LogonLanguage)
	}

	if properties.OAuthProperties != nil {
		m["oauth_properties"] = flattenOAuthProperties(properties.OAuthProperties)
	}

	if properties.PrivateLinkServiceName != nil {
		m["private_link_service_name"] = aws.ToString(properties.PrivateLinkServiceName)
	}

	return []interface{}{m}
}

func flattenSnowflakeConnectorProfileProperties(properties *types.SnowflakeConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})
	if properties.AccountName != nil {
		m["account_name"] = aws.ToString(properties.AccountName)
	}

	m[names.AttrBucketName] = aws.ToString(properties.BucketName)

	if properties.BucketPrefix != nil {
		m[names.AttrBucketPrefix] = aws.ToString(properties.BucketPrefix)
	}

	if properties.Region != nil {
		m[names.AttrRegion] = aws.ToString(properties.Region)
	}

	m[names.AttrStage] = aws.ToString(properties.Stage)
	m["warehouse"] = aws.ToString(properties.Warehouse)

	return []interface{}{m}
}

func flattenOAuthProperties(properties *types.OAuthProperties) []interface{} {
	m := make(map[string]interface{})

	m["auth_code_url"] = aws.ToString(properties.AuthCodeUrl)
	m["oauth_scopes"] = properties.OAuthScopes
	m["token_url"] = aws.ToString(properties.TokenUrl)

	return []interface{}{m}
}

func flattenOAuth2Properties(properties *types.OAuth2Properties) []interface{} {
	m := make(map[string]interface{})

	m["oauth2_grant_type"] = properties.OAuth2GrantType
	m["token_url"] = aws.ToString(properties.TokenUrl)
	m["token_url_custom_properties"] = properties.TokenUrlCustomProperties

	return []interface{}{m}
}
