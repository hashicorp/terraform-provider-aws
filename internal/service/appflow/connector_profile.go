package appflow

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnectorProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorProfileCreate,
		ReadWithoutTimeout:   resourceConnectorProfileRead,
		UpdateWithoutTimeout: resourceConnectorProfileUpdate,
		DeleteWithoutTimeout: resourceConnectorProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`[\w/!@#+=.-]+`), "must match [\\w/!@#+=.-]+"),
				),
			},
			"connection_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appflow.ConnectionMode_Values(), false),
			},
			"connector_label": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][\w!@#.-]+`), "must contain only alphanumeric, exclamation point (!), at sign (@), number sign (#), period (.), and hyphen (-) characters"),
					validation.StringLenBetween(1, 256),
				),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"secret_key": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"api_secret_key": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
														},
													},
												},
												"authentication_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appflow.AuthenticationType_Values(), false),
												},
												"basic": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"password": {
																Type:         schema.TypeString,
																Required:     true,
																Sensitive:    true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"username": {
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
																ValidateDiagFunc: allDiagFunc(
																	validation.MapKeyLenBetween(1, 128),
																	validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
																),
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(0, 2048),
																		validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
															"custom_authentication_type": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringLenBetween(1, 2048),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"client_id": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"client_secret": {
																Type:      schema.TypeString,
																Optional:  true,
																Sensitive: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																				validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																		"redirect_uri": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 512),
																				validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"application_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"datakey": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"secret_access_key": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"user_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_credentials_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
															"password": {
																Type:         schema.TypeString,
																Required:     true,
																Sensitive:    true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"username": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"client_id": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"client_secret": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																				validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																			),
																		},
																		"redirect_uri": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ValidateFunc: validation.All(
																				validation.StringLenBetween(1, 512),
																				validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													Sensitive:    true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appflow.OAuth2GrantType_Values(), false),
															},
															"token_url": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexp.MustCompile(`^(https?)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`), "must provide a valid HTTPS url"),
																),
															},
															"token_url_custom_properties": {
																Type:     schema.TypeMap,
																Optional: true,
																ValidateDiagFunc: allDiagFunc(
																	validation.MapKeyLenBetween(1, 128),
																	validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
																),
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(0, 2048),
																		validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
														},
													},
												},
												"profile_properties": {
													Type:     schema.TypeMap,
													Optional: true,
													ValidateDiagFunc: allDiagFunc(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type: schema.TypeString,
														ValidateFunc: validation.All(
															validation.StringLenBetween(0, 2048),
															validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"database_url": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"role_arn": {
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`^(https?)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`), "must provide a valid HTTPS url"),
													),
												},
												"application_service_path": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"client_number": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 3),
														validation.StringMatch(regexp.MustCompile(`^\d{3}$`), "must consist of exactly three digits"),
													),
												},
												"logon_language": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 2),
														validation.StringMatch(regexp.MustCompile(` ^[a-zA-Z0-9_]*$`), "must contain only alphanumeric characters and the underscore (_) character"),
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
																	validation.StringMatch(regexp.MustCompile(`^(https?)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`), "must provide a valid HTTPS url"),
																),
															},
															"oauth_scopes": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(1, 128),
																		validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
																	),
																},
															},
															"token_url": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 256),
																	validation.StringMatch(regexp.MustCompile(`^(https?)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`), "must provide a valid HTTPS url"),
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
														validation.StringMatch(regexp.MustCompile(`^$|com.amazonaws.vpce.[\w/!:@#.\-]+`), "must be a valid AWS VPC endpoint address"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"private_link_service_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`^$|com.amazonaws.vpce.[\w/!:@#.\-]+`), "must be a valid AWS VPC endpoint address"),
													),
												},
												"region": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 64),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"stage": {
													Type:     schema.TypeString,
													Required: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														return old == new || old == "@"+new
													},
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
													),
												},
												"warehouse": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 512),
														validation.StringMatch(regexp.MustCompile(`[\s\w/!@#+=.-]*`), "must match [\\s\\w/!@#+=.-]*"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
														validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"),
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"credentials_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceConnectorProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn
	name := d.Get("name").(string)

	createConnectorProfileInput := appflow.CreateConnectorProfileInput{
		ConnectionMode:         aws.String(d.Get("connection_mode").(string)),
		ConnectorProfileConfig: expandConnectorProfileConfig(d.Get("connector_profile_config").([]interface{})[0].(map[string]interface{})),
		ConnectorProfileName:   aws.String(name),
		ConnectorType:          aws.String(d.Get("connector_type").(string)),
	}

	if v, ok := d.Get("connector_label").(string); ok && len(v) > 0 {
		createConnectorProfileInput.ConnectorLabel = aws.String(v)
	}

	if v, ok := d.Get("kms_arn").(string); ok && len(v) > 0 {
		createConnectorProfileInput.KmsArn = aws.String(v)
	}

	out, err := conn.CreateConnectorProfile(&createConnectorProfileInput)

	if err != nil {
		return diag.Errorf("creating AppFlow Connector Profile: %s", err)
	}

	if out == nil || out.ConnectorProfileArn == nil {
		return diag.Errorf("creating Appflow Connector Profile (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.StringValue(out.ConnectorProfileArn))

	return resourceConnectorProfileRead(ctx, d, meta)
}

func resourceConnectorProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	connectorProfile, err := FindConnectorProfileByArn(context.Background(), conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppFlow Connector Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	// Credentials are not returned by any API operation. Instead, a
	// "credentials_arn" property is returned.
	//
	// It may be possible to implement a function that reads from this
	// credentials resource -- but it is not documented in the API reference.
	// (https://docs.aws.amazon.com/appflow/1.0/APIReference/API_ConnectorProfile.html#appflow-Type-ConnectorProfile-credentialsArn)
	credentials := d.Get("connector_profile_config.0.connector_profile_credentials").([]interface{})

	d.Set("connection_mode", connectorProfile.ConnectionMode)
	d.Set("connector_label", connectorProfile.ConnectorLabel)
	d.Set("arn", connectorProfile.ConnectorProfileArn)
	d.Set("name", connectorProfile.ConnectorProfileName)
	d.Set("connector_profile_config", flattenConnectorProfileConfig(connectorProfile.ConnectorProfileProperties, credentials))
	d.Set("connector_type", connectorProfile.ConnectorType)
	d.Set("credentials_arn", connectorProfile.CredentialsArn)

	return nil
}

func resourceConnectorProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn
	name := d.Get("name").(string)

	updateConnectorProfileInput := appflow.UpdateConnectorProfileInput{
		ConnectionMode:         aws.String(d.Get("connection_mode").(string)),
		ConnectorProfileConfig: expandConnectorProfileConfig(d.Get("connector_profile_config").([]interface{})[0].(map[string]interface{})),
		ConnectorProfileName:   aws.String(name),
	}

	log.Printf("[DEBUG] Updating AppFlow Connector Profile (%s): %#v", d.Id(), updateConnectorProfileInput)
	_, err := conn.UpdateConnectorProfile(&updateConnectorProfileInput)

	if err != nil {
		return diag.Errorf("updating AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	return resourceConnectorProfileRead(ctx, d, meta)
}

func resourceConnectorProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	out, _ := FindConnectorProfileByArn(ctx, conn, d.Id())

	log.Printf("[INFO] Deleting AppFlow Flow %s", d.Id())

	_, err := conn.DeleteConnectorProfileWithContext(ctx, &appflow.DeleteConnectorProfileInput{
		ConnectorProfileName: out.ConnectorProfileName,
	})

	if err != nil {
		return diag.Errorf("deleting AppFlow Connector Profile (%s): %s", d.Id(), err)
	}

	return nil
}

func expandConnectorProfileConfig(m map[string]interface{}) *appflow.ConnectorProfileConfig {
	cpc := appflow.ConnectorProfileConfig{
		ConnectorProfileCredentials: expandConnectorProfileCredentials(m["connector_profile_credentials"].([]interface{})[0].(map[string]interface{})),
		ConnectorProfileProperties:  expandConnectorProfileProperties(m["connector_profile_properties"].([]interface{})[0].(map[string]interface{})),
	}

	return &cpc
}

func expandConnectorProfileCredentials(m map[string]interface{}) *appflow.ConnectorProfileCredentials {
	cpc := appflow.ConnectorProfileCredentials{}

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

	return &cpc
}

func expandAmplitudeConnectorProfileCredentials(m map[string]interface{}) *appflow.AmplitudeConnectorProfileCredentials {
	credentials := appflow.AmplitudeConnectorProfileCredentials{
		ApiKey:    aws.String(m["api_key"].(string)),
		SecretKey: aws.String(m["secret_key"].(string)),
	}

	return &credentials
}

func expandCustomConnectorProfileCredentials(m map[string]interface{}) *appflow.CustomConnectorProfileCredentials {
	credentials := appflow.CustomConnectorProfileCredentials{
		AuthenticationType: aws.String(m["authentication_type"].(string)),
	}

	if v, ok := m["api_key"].([]interface{}); ok && len(v) > 0 {
		credentials.ApiKey = expandApiKeyCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := m["basic"].([]interface{}); ok && len(v) > 0 {
		credentials.Basic = expandBasicAuthCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := m["custom"].([]interface{}); ok && len(v) > 0 {
		credentials.Custom = expandCustomAuthCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := m["oauth2_credentials"].([]interface{}); ok && len(v) > 0 {
		credentials.Oauth2 = expandOAuth2Credentials(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandDatadogConnectorProfileCredentials(m map[string]interface{}) *appflow.DatadogConnectorProfileCredentials {
	credentials := appflow.DatadogConnectorProfileCredentials{
		ApiKey:         aws.String(m["api_key"].(string)),
		ApplicationKey: aws.String(m["application_key"].(string)),
	}

	return &credentials
}

func expandDynatraceConnectorProfileCredentials(m map[string]interface{}) *appflow.DynatraceConnectorProfileCredentials {
	credentials := appflow.DynatraceConnectorProfileCredentials{
		ApiToken: aws.String(m["api_token"].(string)),
	}

	return &credentials
}

func expandGoogleAnalyticsConnectorProfileCredentials(m map[string]interface{}) *appflow.GoogleAnalyticsConnectorProfileCredentials {
	credentials := appflow.GoogleAnalyticsConnectorProfileCredentials{
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandHoneycodeConnectorProfileCredentials(m map[string]interface{}) *appflow.HoneycodeConnectorProfileCredentials {
	credentials := appflow.HoneycodeConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandInforNexusConnectorProfileCredentials(m map[string]interface{}) *appflow.InforNexusConnectorProfileCredentials {
	credentials := appflow.InforNexusConnectorProfileCredentials{
		AccessKeyId:     aws.String(m["access_key_id"].(string)),
		Datakey:         aws.String(m["datakey"].(string)),
		SecretAccessKey: aws.String(m["secret_access_key"].(string)),
		UserId:          aws.String(m["user_id"].(string)),
	}

	return &credentials
}

func expandMarketoConnectorProfileCredentials(m map[string]interface{}) *appflow.MarketoConnectorProfileCredentials {
	credentials := appflow.MarketoConnectorProfileCredentials{
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandRedshiftConnectorProfileCredentials(m map[string]interface{}) *appflow.RedshiftConnectorProfileCredentials {
	credentials := appflow.RedshiftConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandSalesforceConnectorProfileCredentials(m map[string]interface{}) *appflow.SalesforceConnectorProfileCredentials {
	credentials := appflow.SalesforceConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["client_credentials_arn"].(string); ok && v != "" {
		credentials.ClientCredentialsArn = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandSAPODataConnectorProfileCredentials(m map[string]interface{}) *appflow.SAPODataConnectorProfileCredentials {
	credentials := appflow.SAPODataConnectorProfileCredentials{}

	if v, ok := m["basic_auth_credentials"].([]interface{}); ok && len(v) > 0 {
		credentials.BasicAuthCredentials = expandBasicAuthCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := m["oauth_credentials"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthCredentials = expandOAuthCredentials(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandServiceNowConnectorProfileCredentials(m map[string]interface{}) *appflow.ServiceNowConnectorProfileCredentials {
	credentials := appflow.ServiceNowConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandSingularConnectorProfileCredentials(m map[string]interface{}) *appflow.SingularConnectorProfileCredentials {
	credentials := appflow.SingularConnectorProfileCredentials{
		ApiKey: aws.String(m["api_key"].(string)),
	}

	return &credentials
}

func expandSlackConnectorProfileCredentials(m map[string]interface{}) *appflow.SlackConnectorProfileCredentials {
	credentials := appflow.SlackConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandSnowflakeConnectorProfileCredentials(m map[string]interface{}) *appflow.SnowflakeConnectorProfileCredentials {
	credentials := appflow.SnowflakeConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandTrendmicroConnectorProfileCredentials(m map[string]interface{}) *appflow.TrendmicroConnectorProfileCredentials {
	credentials := appflow.TrendmicroConnectorProfileCredentials{
		ApiSecretKey: aws.String(m["api_secret_key"].(string)),
	}

	return &credentials
}

func expandVeevaConnectorProfileCredentials(m map[string]interface{}) *appflow.VeevaConnectorProfileCredentials {
	credentials := appflow.VeevaConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandZendeskConnectorProfileCredentials(m map[string]interface{}) *appflow.ZendeskConnectorProfileCredentials {
	credentials := appflow.ZendeskConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandOAuthRequest(m map[string]interface{}) *appflow.ConnectorOAuthRequest {
	r := appflow.ConnectorOAuthRequest{}

	if v, ok := m["auth_code"].(string); ok && v != "" {
		r.AuthCode = aws.String(v)
	}

	if v, ok := m["redirect_uri"].(string); ok && v != "" {
		r.RedirectUri = aws.String(v)
	}

	return &r
}

func expandApiKeyCredentials(m map[string]interface{}) *appflow.ApiKeyCredentials {
	credentials := appflow.ApiKeyCredentials{}

	if v, ok := m["api_key"].(string); ok && v != "" {
		credentials.ApiKey = aws.String(v)
	}

	if v, ok := m["api_secret_key"].(string); ok && v != "" {
		credentials.ApiSecretKey = aws.String(v)
	}

	return &credentials
}

func expandBasicAuthCredentials(m map[string]interface{}) *appflow.BasicAuthCredentials {
	credentials := appflow.BasicAuthCredentials{}

	if v, ok := m["password"].(string); ok && v != "" {
		credentials.Password = aws.String(v)
	}

	if v, ok := m["username"].(string); ok && v != "" {
		credentials.Username = aws.String(v)
	}

	return &credentials
}

func expandCustomAuthCredentials(m map[string]interface{}) *appflow.CustomAuthCredentials {
	credentials := appflow.CustomAuthCredentials{}

	if v, ok := m["credentials_map"].(map[string]interface{}); ok && len(v) > 0 {
		credentials.CredentialsMap = flex.ExpandStringMap(v)
	}

	if v, ok := m["custom_authentication_type"].(string); ok && v != "" {
		credentials.CustomAuthenticationType = aws.String(v)
	}

	return &credentials
}

func expandOAuthCredentials(m map[string]interface{}) *appflow.OAuthCredentials {
	credentials := appflow.OAuthCredentials{
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandOAuth2Credentials(m map[string]interface{}) *appflow.OAuth2Credentials {
	credentials := appflow.OAuth2Credentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["client_id"].(string); ok && v != "" {
		credentials.ClientId = aws.String(v)
	}

	if v, ok := m["client_secret"].(string); ok && v != "" {
		credentials.ClientSecret = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandConnectorProfileProperties(m map[string]interface{}) *appflow.ConnectorProfileProperties {
	cpc := appflow.ConnectorProfileProperties{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 {
		cpc.Amplitude = v[0].(*appflow.AmplitudeConnectorProfileProperties)
	}

	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Datadog = expandDatadogConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Dynatrace = expandDynatraceConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 {
		cpc.GoogleAnalytics = v[0].(*appflow.GoogleAnalyticsConnectorProfileProperties)
	}

	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 {
		cpc.Honeycode = v[0].(*appflow.HoneycodeConnectorProfileProperties)
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

	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 {
		cpc.Singular = v[0].(*appflow.SingularConnectorProfileProperties)
	}

	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Slack = expandSlackConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Snowflake = expandSnowflakeConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 {
		cpc.Trendmicro = v[0].(*appflow.TrendmicroConnectorProfileProperties)
	}

	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 {
		cpc.Veeva = expandVeevaConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Zendesk = expandZendeskConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	return &cpc
}

func expandDatadogConnectorProfileProperties(m map[string]interface{}) *appflow.DatadogConnectorProfileProperties {
	properties := appflow.DatadogConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandDynatraceConnectorProfileProperties(m map[string]interface{}) *appflow.DynatraceConnectorProfileProperties {
	properties := appflow.DynatraceConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandInforNexusConnectorProfileProperties(m map[string]interface{}) *appflow.InforNexusConnectorProfileProperties {
	properties := appflow.InforNexusConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandMarketoConnectorProfileProperties(m map[string]interface{}) *appflow.MarketoConnectorProfileProperties {
	properties := appflow.MarketoConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandRedshiftConnectorProfileProperties(m map[string]interface{}) *appflow.RedshiftConnectorProfileProperties {
	properties := appflow.RedshiftConnectorProfileProperties{
		BucketName: aws.String(m["bucket_name"].(string)),
		RoleArn:    aws.String(m["role_arn"].(string)),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["database_url"].(string); ok && v != "" {
		properties.DatabaseUrl = aws.String(v)
	}

	return &properties
}

func expandServiceNowConnectorProfileProperties(m map[string]interface{}) *appflow.ServiceNowConnectorProfileProperties {
	properties := appflow.ServiceNowConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandSalesforceConnectorProfileProperties(m map[string]interface{}) *appflow.SalesforceConnectorProfileProperties {
	properties := appflow.SalesforceConnectorProfileProperties{}

	if v, ok := m["instance_url"].(string); ok && v != "" {
		properties.InstanceUrl = aws.String(v)
	}

	if v, ok := m["is_sandbox_environment"].(bool); ok {
		properties.IsSandboxEnvironment = aws.Bool(v)
	}

	return &properties
}

func expandSAPODataConnectorProfileProperties(m map[string]interface{}) *appflow.SAPODataConnectorProfileProperties {
	properties := appflow.SAPODataConnectorProfileProperties{
		ApplicationHostUrl:     aws.String(m["application_host_url"].(string)),
		ApplicationServicePath: aws.String(m["application_service_path"].(string)),
		ClientNumber:           aws.String(m["client_number"].(string)),
		PortNumber:             aws.Int64(int64(m["port_number"].(int))),
	}

	if v, ok := m["logon_language"].(string); ok && v != "" {
		properties.LogonLanguage = aws.String(v)
	}

	if v, ok := m["oauth_properties"].([]interface{}); ok && len(v) > 0 {
		properties.OAuthProperties = expandOAuthProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["private_link_service_name"].(string); ok && v != "" {
		properties.PrivateLinkServiceName = aws.String(v)
	}

	return &properties
}

func expandSlackConnectorProfileProperties(m map[string]interface{}) *appflow.SlackConnectorProfileProperties {
	properties := appflow.SlackConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandSnowflakeConnectorProfileProperties(m map[string]interface{}) *appflow.SnowflakeConnectorProfileProperties {
	properties := appflow.SnowflakeConnectorProfileProperties{
		BucketName: aws.String(m["bucket_name"].(string)),
		Stage:      aws.String(m["stage"].(string)),
		Warehouse:  aws.String(m["warehouse"].(string)),
	}

	if v, ok := m["account_name"].(string); ok && v != "" {
		properties.AccountName = aws.String(v)
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["private_link_service_name"].(string); ok && v != "" {
		properties.PrivateLinkServiceName = aws.String(v)
	}

	if v, ok := m["region"].(string); ok && v != "" {
		properties.Region = aws.String(v)
	}

	return &properties
}

func expandVeevaConnectorProfileProperties(m map[string]interface{}) *appflow.VeevaConnectorProfileProperties {
	properties := appflow.VeevaConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandZendeskConnectorProfileProperties(m map[string]interface{}) *appflow.ZendeskConnectorProfileProperties {
	properties := appflow.ZendeskConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandOAuthProperties(m map[string]interface{}) *appflow.OAuthProperties {
	properties := appflow.OAuthProperties{
		AuthCodeUrl: aws.String(m["auth_code_url"].(string)),
		OAuthScopes: flex.ExpandStringList(m["oauth_scopes"].([]interface{})),
		TokenUrl:    aws.String(m["token_url"].(string)),
	}

	return &properties
}

func flattenConnectorProfileConfig(cpp *appflow.ConnectorProfileProperties, cpc []interface{}) []interface{} {
	m := make(map[string]interface{})

	m["connector_profile_properties"] = flattenConnectorProfileProperties(cpp)
	m["connector_profile_credentials"] = cpc

	return []interface{}{m}
}

func flattenConnectorProfileProperties(cpp *appflow.ConnectorProfileProperties) []interface{} {
	result := make(map[string]interface{})
	m := make(map[string]interface{})

	if cpp.Amplitude != nil {
		result["amplitude"] = []interface{}{m}
	}
	if cpp.Datadog != nil {
		m["instance_url"] = aws.StringValue(cpp.Datadog.InstanceUrl)
		result["datadog"] = []interface{}{m}
	}
	if cpp.Dynatrace != nil {
		m["instance_url"] = aws.StringValue(cpp.Dynatrace.InstanceUrl)
		result["dynatrace"] = []interface{}{m}
	}
	if cpp.GoogleAnalytics != nil {
		result["google_analytics"] = []interface{}{m}
	}
	if cpp.Honeycode != nil {
		result["honeycode"] = []interface{}{m}
	}
	if cpp.InforNexus != nil {
		m["instance_url"] = aws.StringValue(cpp.InforNexus.InstanceUrl)
		result["infor_nexus"] = []interface{}{m}
	}
	if cpp.Marketo != nil {
		m["instance_url"] = aws.StringValue(cpp.Marketo.InstanceUrl)
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
		m["instance_url"] = aws.StringValue(cpp.ServiceNow.InstanceUrl)
		result["service_now"] = []interface{}{m}
	}
	if cpp.Singular != nil {
		result["singular"] = []interface{}{m}
	}
	if cpp.Slack != nil {
		m["instance_url"] = aws.StringValue(cpp.Slack.InstanceUrl)
		result["slack"] = []interface{}{m}
	}
	if cpp.Snowflake != nil {
		result["snowflake"] = flattenSnowflakeConnectorProfileProperties(cpp.Snowflake)
	}
	if cpp.Trendmicro != nil {
		result["trendmicro"] = []interface{}{m}
	}
	if cpp.Veeva != nil {
		m["instance_url"] = aws.StringValue(cpp.Veeva.InstanceUrl)
		result["veeva"] = []interface{}{m}
	}
	if cpp.Zendesk != nil {
		m["instance_url"] = aws.StringValue(cpp.Zendesk.InstanceUrl)
		result["zendesk"] = []interface{}{m}
	}

	return []interface{}{result}
}

func flattenRedshiftConnectorProfileProperties(properties *appflow.RedshiftConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	m["bucket_name"] = aws.StringValue(properties.BucketName)

	if properties.BucketPrefix != nil {
		m["bucket_prefix"] = aws.StringValue(properties.BucketPrefix)
	}

	if properties.DatabaseUrl != nil {
		m["database_url"] = aws.StringValue(properties.DatabaseUrl)
	}

	m["role_arn"] = aws.StringValue(properties.RoleArn)

	return []interface{}{m}
}

func flattenSalesforceConnectorProfileProperties(properties *appflow.SalesforceConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	if properties.InstanceUrl != nil {
		m["instance_url"] = aws.StringValue(properties.InstanceUrl)
	}

	if properties.IsSandboxEnvironment != nil {
		m["is_sandbox_environment"] = aws.BoolValue(properties.IsSandboxEnvironment)
	}

	return []interface{}{m}
}

func flattenSAPODataConnectorProfileProperties(properties *appflow.SAPODataConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	m["application_host_url"] = aws.StringValue(properties.ApplicationHostUrl)
	m["application_service_path"] = aws.StringValue(properties.ApplicationServicePath)
	m["client_number"] = aws.StringValue(properties.ClientNumber)
	m["port_number"] = aws.Int64Value(properties.PortNumber)

	if properties.LogonLanguage != nil {
		m["logon_language"] = aws.StringValue(properties.LogonLanguage)
	}

	if properties.OAuthProperties != nil {
		m["oauth_properties"] = flattenOAuthProperties(properties.OAuthProperties)
	}

	if properties.PrivateLinkServiceName != nil {
		m["private_link_service_name"] = aws.StringValue(properties.PrivateLinkServiceName)
	}

	return []interface{}{m}
}

func flattenSnowflakeConnectorProfileProperties(properties *appflow.SnowflakeConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})
	if properties.AccountName != nil {
		m["account_name"] = aws.StringValue(properties.AccountName)
	}

	m["bucket_name"] = aws.StringValue(properties.BucketName)

	if properties.BucketPrefix != nil {
		m["bucket_prefix"] = aws.StringValue(properties.BucketPrefix)
	}

	if properties.Region != nil {
		m["region"] = aws.StringValue(properties.Region)
	}

	m["stage"] = aws.StringValue(properties.Stage)
	m["warehouse"] = aws.StringValue(properties.Warehouse)

	return []interface{}{m}
}

func flattenOAuthProperties(properties *appflow.OAuthProperties) []interface{} {
	m := make(map[string]interface{})

	m["auth_code_url"] = aws.StringValue(properties.AuthCodeUrl)
	m["oauth_scopes"] = aws.StringValueSlice(properties.OAuthScopes)
	m["token_url"] = aws.StringValue(properties.TokenUrl)

	return []interface{}{m}
}
