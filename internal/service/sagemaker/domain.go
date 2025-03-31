// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_network_access_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AppNetworkAccessTypePublicInternetOnly,
				ValidateDiagFunc: enum.Validate[awstypes.AppNetworkAccessType](),
			},
			"app_security_group_management": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AppSecurityGroupManagement](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_mode": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthMode](),
			},
			"default_space_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"jupyter_server_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code_repository": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_url": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
						"kernel_gateway_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"jupyter_lab_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_lifecycle_management": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_settings": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"lifecycle_management": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.LifecycleManagement](),
															},
															"max_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"min_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
														},
													},
												},
											},
										},
									},
									"built_in_lifecycle_config_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"code_repository": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_url": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"emr_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"assumable_role_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
												"execution_role_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
						"space_storage_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_ebs_storage_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_ebs_volume_size_in_gb": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"maximum_ebs_volume_size_in_gb": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"custom_file_system_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"efs_file_system_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrFileSystemID: {
													Type:     schema.TypeString,
													Required: true,
												},
												"file_system_path": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"custom_posix_user_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gid": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1001),
									},
									"uid": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(10000),
									},
								},
							},
						},
					},
				},
			},
			"default_user_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_mount_home_efs": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AutoMountHomeEFS](),
						},
						"canvas_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"direct_deploy_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"emr_serverless_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrExecutionRoleARN: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"generative_ai_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"amazon_bedrock_role_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"identity_provider_oauth_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 20,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data_source_name": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.DataSourceName](),
												},
												"secret_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"kendra_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"model_register_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cross_account_model_register_role_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"time_series_forecasting_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"amazon_forecast_role_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												names.AttrStatus: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
												},
											},
										},
									},
									"workspace_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_artifact_path": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
														validation.StringLenBetween(1, 1024),
													),
												},
												"s3_kms_key_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"code_editor_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_lifecycle_management": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_settings": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"lifecycle_management": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.LifecycleManagement](),
															},
															"max_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"min_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
														},
													},
												},
											},
										},
									},
									"built_in_lifecycle_config_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"custom_file_system_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"efs_file_system_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrFileSystemID: {
													Type:     schema.TypeString,
													Required: true,
												},
												"file_system_path": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"custom_posix_user_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gid": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1001),
									},
									"uid": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(10000),
									},
								},
							},
						},
						"default_landing_uri": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"jupyter_lab_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_lifecycle_management": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_settings": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"lifecycle_management": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.LifecycleManagement](),
															},
															"max_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
															"min_idle_timeout_in_minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(60, 525600),
															},
														},
													},
												},
											},
										},
									},
									"built_in_lifecycle_config_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"code_repository": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_url": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"emr_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"assumable_role_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
												"execution_role_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
						"jupyter_server_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code_repository": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_url": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
						"kernel_gateway_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"lifecycle_config_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"r_studio_server_pro_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_status": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RStudioServerProAccessStatus](),
									},
									"user_group": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.RStudioServerProUserGroupUser,
										ValidateDiagFunc: enum.Validate[awstypes.RStudioServerProUserGroup](),
									},
								},
							},
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"r_session_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 200,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"sharing_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notebook_output_option": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.NotebookOutputOptionDisabled,
										ValidateDiagFunc: enum.Validate[awstypes.NotebookOutputOption](),
									},
									"s3_kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"s3_output_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"studio_web_portal": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.StudioWebPortal](),
						},
						"space_storage_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_ebs_storage_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_ebs_volume_size_in_gb": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"maximum_ebs_volume_size_in_gb": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"tensor_board_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						"studio_web_portal_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hidden_app_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.AppType](),
										},
									},
									"hidden_instance_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
										},
									},
									"hidden_ml_tools": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.MlTools](),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"domain_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"docker_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_docker_access": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FeatureStatus](),
									},
									"vpc_only_trusted_accounts": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidAccountID,
										},
										MaxItems: 20,
									},
								},
							},
						},
						"execution_role_identity_config": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ExecutionRoleIdentityConfig](),
						},
						"r_studio_server_pro_domain_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrInstanceType: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
												},
												"lifecycle_config_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"sagemaker_image_version_alias": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sagemaker_image_version_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"domain_execution_role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"r_studio_connect_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"r_studio_package_manager_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							MaxItems: 3,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"home_efs_file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"retention_policy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"home_efs_file_system": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.RetentionTypeRetain,
							ValidateDiagFunc: enum.Validate[awstypes.RetentionType](),
						},
					},
				},
			},
			"security_group_id_for_domain_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"single_sign_on_application_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"single_sign_on_managed_application_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 16,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag_propagation": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.TagPropagationDisabled,
				ValidateDiagFunc: enum.Validate[awstypes.TagPropagation](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.CreateDomainInput{
		DomainName:           aws.String(d.Get(names.AttrDomainName).(string)),
		AuthMode:             awstypes.AuthMode(d.Get("auth_mode").(string)),
		VpcId:                aws.String(d.Get(names.AttrVPCID).(string)),
		AppNetworkAccessType: awstypes.AppNetworkAccessType(d.Get("app_network_access_type").(string)),
		SubnetIds:            flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		DefaultUserSettings:  expandUserSettings(d.Get("default_user_settings").([]any)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("app_security_group_management"); ok && rstudioDomainEnabled(d.Get("domain_settings").([]any)) {
		input.AppSecurityGroupManagement = awstypes.AppSecurityGroupManagement(v.(string))
	}

	if v, ok := d.GetOk("domain_settings"); ok && len(v.([]any)) > 0 {
		input.DomainSettings = expandDomainSettings(v.([]any))
	}

	if v, ok := d.GetOk("default_space_settings"); ok && len(v.([]any)) > 0 {
		input.DefaultSpaceSettings = expanDefaultSpaceSettings(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_propagation"); ok {
		input.TagPropagation = awstypes.TagPropagation(v.(string))
	}

	log.Printf("[DEBUG] SageMaker AI Domain create config: %#v", *input)
	output, err := conn.CreateDomain(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Domain: %s", err)
	}

	domainID := aws.ToString(output.DomainId)
	d.SetId(domainID)

	if err := waitDomainInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Domain (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	domain, err := findDomainByName(ctx, conn, d.Id())
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker AI Domain (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Domain (%s): %s", d.Id(), err)
	}

	d.Set("app_network_access_type", domain.AppNetworkAccessType)
	d.Set("app_security_group_management", domain.AppSecurityGroupManagement)
	d.Set(names.AttrARN, domain.DomainArn)
	d.Set("auth_mode", domain.AuthMode)
	d.Set(names.AttrDomainName, domain.DomainName)
	d.Set("home_efs_file_system_id", domain.HomeEfsFileSystemId)
	d.Set(names.AttrKMSKeyID, domain.KmsKeyId)
	d.Set("security_group_id_for_domain_boundary", domain.SecurityGroupIdForDomainBoundary)
	d.Set("single_sign_on_managed_application_instance_id", domain.SingleSignOnManagedApplicationInstanceId)
	d.Set("single_sign_on_application_arn", domain.SingleSignOnApplicationArn)
	d.Set("tag_propagation", domain.TagPropagation)
	d.Set(names.AttrURL, domain.Url)
	d.Set(names.AttrVPCID, domain.VpcId)

	if err := d.Set(names.AttrSubnetIDs, flex.FlattenStringValueSet(domain.SubnetIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids for SageMaker AI Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_user_settings", flattenUserSettings(domain.DefaultUserSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_user_settings for SageMaker AI Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_space_settings", flattenDefaultSpaceSettings(domain.DefaultSpaceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_space_settings for SageMaker AI Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("domain_settings", flattenDomainSettings(domain.DomainSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_settings for SageMaker AI Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateDomainInput{
			DomainId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("app_network_access_type"); ok {
			input.AppNetworkAccessType = awstypes.AppNetworkAccessType(v.(string))
		}

		if v, ok := d.GetOk("app_security_group_management"); ok && rstudioDomainEnabled(d.Get("domain_settings").([]any)) {
			input.AppSecurityGroupManagement = awstypes.AppSecurityGroupManagement(v.(string))
		}

		if v, ok := d.GetOk("default_user_settings"); ok && len(v.([]any)) > 0 {
			input.DefaultUserSettings = expandUserSettings(v.([]any))
		}

		if v, ok := d.GetOk("domain_settings"); ok && len(v.([]any)) > 0 {
			input.DomainSettingsForUpdate = expandDomainSettingsUpdate(v.([]any))
		}

		if v, ok := d.GetOk("default_space_settings"); ok && len(v.([]any)) > 0 {
			input.DefaultSpaceSettings = expanDefaultSpaceSettings(v.([]any))
		}

		if v, ok := d.GetOk("tag_propagation"); ok {
			input.TagPropagation = awstypes.TagPropagation(v.(string))
		}

		log.Printf("[DEBUG] SageMaker AI Domain update config: %#v", *input)
		_, err := conn.UpdateDomain(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Domain: %s", err)
		}

		if err := waitDomainInService(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Domain (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteDomainInput{
		DomainId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("retention_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RetentionPolicy = expandRetentionPolicy(v.([]any))
	}

	if _, err := conn.DeleteDomain(ctx, input); err != nil {
		if !errs.IsA[*awstypes.ResourceNotFound](err) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Domain (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitDomainDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Domain (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findDomainByName(ctx context.Context, conn *sagemaker.Client, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	input := &sagemaker.DescribeDomainInput{
		DomainId: aws.String(domainID),
	}

	output, err := conn.DescribeDomain(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandDomainSettings(l []any) *awstypes.DomainSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DomainSettings{}

	if v, ok := m["docker_settings"].([]any); ok && len(v) > 0 {
		config.DockerSettings = expandDockerSettings(v)
	}

	if v, ok := m["execution_role_identity_config"].(string); ok && v != "" {
		config.ExecutionRoleIdentityConfig = awstypes.ExecutionRoleIdentityConfig(v)
	}

	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["r_studio_server_pro_domain_settings"].([]any); ok && len(v) > 0 {
		config.RStudioServerProDomainSettings = expandRStudioServerProDomainSettings(v)
	}

	return config
}

func expandDockerSettings(l []any) *awstypes.DockerSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DockerSettings{}

	if v, ok := m["enable_docker_access"].(string); ok && v != "" {
		config.EnableDockerAccess = awstypes.FeatureStatus(v)
	}

	if v, ok := m["vpc_only_trusted_accounts"].(*schema.Set); ok && v.Len() > 0 {
		config.VpcOnlyTrustedAccounts = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandRStudioServerProDomainSettings(l []any) *awstypes.RStudioServerProDomainSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.RStudioServerProDomainSettings{}

	if v, ok := m["domain_execution_role_arn"].(string); ok && v != "" {
		config.DomainExecutionRoleArn = aws.String(v)
	}

	if v, ok := m["r_studio_connect_url"].(string); ok && v != "" {
		config.RStudioConnectUrl = aws.String(v)
	}

	if v, ok := m["r_studio_package_manager_url"].(string); ok && v != "" {
		config.RStudioPackageManagerUrl = aws.String(v)
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func expandDomainSettingsUpdate(l []any) *awstypes.DomainSettingsForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DomainSettingsForUpdate{}

	if v, ok := m["docker_settings"].([]any); ok && len(v) > 0 {
		config.DockerSettings = expandDockerSettings(v)
	}

	if v, ok := m["execution_role_identity_config"].(string); ok && v != "" {
		config.ExecutionRoleIdentityConfig = awstypes.ExecutionRoleIdentityConfig(v)
	}

	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["r_studio_server_pro_domain_settings"].([]any); ok && len(v) > 0 {
		config.RStudioServerProDomainSettingsForUpdate = expandRStudioServerProDomainSettingsUpdate(v)
	}

	return config
}

// rstudioDomainEnabled takes domain_settings and returns true if rstudio is enabled
func rstudioDomainEnabled(domainSettings []any) bool {
	if len(domainSettings) == 0 || domainSettings[0] == nil {
		return false
	}

	m := domainSettings[0].(map[string]any)

	v, ok := m["r_studio_server_pro_domain_settings"].([]any)
	if !ok || len(v) < 1 {
		return false
	}

	rsspds, ok := v[0].(map[string]any)
	if !ok || len(rsspds) == 0 {
		return false
	}

	domainExecutionRoleArn, ok := rsspds["domain_execution_role_arn"].(string)
	if !ok || domainExecutionRoleArn == "" {
		return false
	}

	return true
}

func expandRStudioServerProDomainSettingsUpdate(l []any) *awstypes.RStudioServerProDomainSettingsForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.RStudioServerProDomainSettingsForUpdate{}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["domain_execution_role_arn"].(string); ok && v != "" {
		config.DomainExecutionRoleArn = aws.String(v)
	}

	if v, ok := m["r_studio_connect_url"].(string); ok && v != "" {
		config.RStudioConnectUrl = aws.String(v)
	}

	if v, ok := m["r_studio_package_manager_url"].(string); ok && v != "" {
		config.RStudioPackageManagerUrl = aws.String(v)
	}

	return config
}

func expandRetentionPolicy(l []any) *awstypes.RetentionPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.RetentionPolicy{}

	if v, ok := m["home_efs_file_system"].(string); ok && v != "" {
		config.HomeEfsFileSystem = awstypes.RetentionType(v)
	}

	return config
}

func expandUserSettings(l []any) *awstypes.UserSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.UserSettings{}

	if v, ok := m["auto_mount_home_efs"].(string); ok && v != "" {
		config.AutoMountHomeEFS = awstypes.AutoMountHomeEFS(v)
	}

	if v, ok := m["canvas_app_settings"].([]any); ok && len(v) > 0 {
		config.CanvasAppSettings = expandCanvasAppSettings(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["default_landing_uri"].(string); ok && v != "" {
		config.DefaultLandingUri = aws.String(v)
	}

	if v, ok := m["code_editor_app_settings"].([]any); ok && len(v) > 0 {
		config.CodeEditorAppSettings = expandDomainCodeEditorAppSettings(v)
	}

	if v, ok := m["custom_file_system_config"].([]any); ok && len(v) > 0 {
		config.CustomFileSystemConfigs = expandCustomFileSystemConfigs(v)
	}

	if v, ok := m["custom_posix_user_config"].([]any); ok && len(v) > 0 {
		config.CustomPosixUserConfig = expandCustomPOSIXUserConfig(v)
	}

	if v, ok := m["jupyter_lab_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterLabAppSettings = expandDomainJupyterLabAppSettings(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]any); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["r_session_app_settings"].([]any); ok && len(v) > 0 {
		config.RSessionAppSettings = expandRSessionAppSettings(v)
	}

	if v, ok := m[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["sharing_settings"].([]any); ok && len(v) > 0 {
		config.SharingSettings = expandDomainShareSettings(v)
	}

	if v, ok := m["studio_web_portal"].(string); ok && v != "" {
		config.StudioWebPortal = awstypes.StudioWebPortal(v)
	}

	if v, ok := m["space_storage_settings"].([]any); ok && len(v) > 0 {
		config.SpaceStorageSettings = expandDefaultSpaceStorageSettings(v)
	}

	if v, ok := m["tensor_board_app_settings"].([]any); ok && len(v) > 0 {
		config.TensorBoardAppSettings = expandDomainTensorBoardAppSettings(v)
	}

	if v, ok := m["r_studio_server_pro_app_settings"].([]any); ok && len(v) > 0 {
		config.RStudioServerProAppSettings = expandRStudioServerProAppSettings(v)
	}

	if v, ok := m["studio_web_portal_settings"].([]any); ok && len(v) > 0 {
		config.StudioWebPortalSettings = expandStudioWebPortalSettings(v)
	}

	return config
}

func expandRStudioServerProAppSettings(l []any) *awstypes.RStudioServerProAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.RStudioServerProAppSettings{}

	if v, ok := m["access_status"].(string); ok && v != "" {
		config.AccessStatus = awstypes.RStudioServerProAccessStatus(v)

		if v == string(awstypes.RStudioServerProAccessStatusEnabled) {
			if g, ok := m["user_group"].(string); ok && g != "" {
				config.UserGroup = awstypes.RStudioServerProUserGroup(g)
			}
		}
	}

	return config
}

func expandCustomPOSIXUserConfig(l []any) *awstypes.CustomPosixUserConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CustomPosixUserConfig{}

	if v, ok := m["gid"].(int); ok {
		config.Gid = aws.Int64(int64(v))
	}

	if v, ok := m["uid"].(int); ok {
		config.Uid = aws.Int64(int64(v))
	}

	return config
}

func expandDomainCodeEditorAppSettings(l []any) *awstypes.CodeEditorAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CodeEditorAppSettings{}

	if v, ok := m["app_lifecycle_management"].([]any); ok && len(v) > 0 {
		config.AppLifecycleManagement = expandAppLifecycleManagement(v)
	}

	if v, ok := m["built_in_lifecycle_config_arn"].(string); ok && v != "" {
		config.BuiltInLifecycleConfigArn = aws.String(v)
	}

	if v, ok := m["custom_image"].([]any); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandDomainJupyterLabAppSettings(l []any) *awstypes.JupyterLabAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.JupyterLabAppSettings{}

	if v, ok := m["app_lifecycle_management"].([]any); ok && len(v) > 0 {
		config.AppLifecycleManagement = expandAppLifecycleManagement(v)
	}

	if v, ok := m["built_in_lifecycle_config_arn"].(string); ok && v != "" {
		config.BuiltInLifecycleConfigArn = aws.String(v)
	}

	if v, ok := m["code_repository"].(*schema.Set); ok && v.Len() > 0 {
		config.CodeRepositories = expandCodeRepositories(v.List())
	}

	if v, ok := m["custom_image"].([]any); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["emr_settings"].([]any); ok && len(v) > 0 {
		config.EmrSettings = expandEMRSettings(v)
	}

	return config
}

func expandAppLifecycleManagement(l []any) *awstypes.AppLifecycleManagement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.AppLifecycleManagement{}

	if v, ok := m["idle_settings"].([]any); ok && len(v) > 0 {
		config.IdleSettings = expandIdleSettings(v)
	}

	return config
}

func expandIdleSettings(l []any) *awstypes.IdleSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.IdleSettings{}

	if v, ok := m["idle_timeout_in_minutes"].(int); ok {
		config.IdleTimeoutInMinutes = aws.Int32(int32(v))
	}

	if v, ok := m["lifecycle_management"].(string); ok && v != "" {
		config.LifecycleManagement = awstypes.LifecycleManagement(v)
	}

	if v, ok := m["max_idle_timeout_in_minutes"].(int); ok {
		config.MaxIdleTimeoutInMinutes = aws.Int32(int32(v))
	}

	if v, ok := m["min_idle_timeout_in_minutes"].(int); ok {
		config.MinIdleTimeoutInMinutes = aws.Int32(int32(v))
	}

	return config
}

func expandDomainJupyterServerAppSettings(l []any) *awstypes.JupyterServerAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.JupyterServerAppSettings{}

	if v, ok := m["code_repository"].(*schema.Set); ok && v.Len() > 0 {
		config.CodeRepositories = expandCodeRepositories(v.List())
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandDomainKernelGatewayAppSettings(l []any) *awstypes.KernelGatewayAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.KernelGatewayAppSettings{}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["custom_image"].([]any); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	return config
}

func expandRSessionAppSettings(l []any) *awstypes.RSessionAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.RSessionAppSettings{}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	if v, ok := m["custom_image"].([]any); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	return config
}

func expandDefaultSpaceStorageSettings(l []any) *awstypes.DefaultSpaceStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DefaultSpaceStorageSettings{}

	if v, ok := m["default_ebs_storage_settings"].([]any); ok && len(v) > 0 {
		config.DefaultEbsStorageSettings = expandDefaultEBSStorageSettings(v)
	}

	return config
}

func expandDefaultEBSStorageSettings(l []any) *awstypes.DefaultEbsStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DefaultEbsStorageSettings{}

	if v, ok := m["default_ebs_volume_size_in_gb"].(int); ok {
		config.DefaultEbsVolumeSizeInGb = aws.Int32(int32(v))
	}

	if v, ok := m["maximum_ebs_volume_size_in_gb"].(int); ok {
		config.MaximumEbsVolumeSizeInGb = aws.Int32(int32(v))
	}

	return config
}

func expandDomainTensorBoardAppSettings(l []any) *awstypes.TensorBoardAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.TensorBoardAppSettings{}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func expandResourceSpec(l []any) *awstypes.ResourceSpec {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.ResourceSpec{}

	if v, ok := m[names.AttrInstanceType].(string); ok && v != "" {
		config.InstanceType = awstypes.AppInstanceType(v)
	}

	if v, ok := m["lifecycle_config_arn"].(string); ok && v != "" {
		config.LifecycleConfigArn = aws.String(v)
	}

	if v, ok := m["sagemaker_image_arn"].(string); ok && v != "" {
		config.SageMakerImageArn = aws.String(v)
	}

	if v, ok := m["sagemaker_image_version_alias"].(string); ok && v != "" {
		config.SageMakerImageVersionAlias = aws.String(v)
	}

	if v, ok := m["sagemaker_image_version_arn"].(string); ok && v != "" {
		config.SageMakerImageVersionArn = aws.String(v)
	}

	return config
}

func expandEMRSettings(l []any) *awstypes.EmrSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.EmrSettings{}

	if v, ok := m["assumable_role_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.AssumableRoleArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["execution_role_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.ExecutionRoleArns = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandDomainShareSettings(l []any) *awstypes.SharingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SharingSettings{
		NotebookOutputOption: awstypes.NotebookOutputOption(m["notebook_output_option"].(string)),
	}

	if v, ok := m["s3_kms_key_id"].(string); ok && v != "" {
		config.S3KmsKeyId = aws.String(v)
	}

	if v, ok := m["s3_output_path"].(string); ok && v != "" {
		config.S3OutputPath = aws.String(v)
	}

	return config
}

func expandCanvasAppSettings(l []any) *awstypes.CanvasAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CanvasAppSettings{}

	if v, ok := m["direct_deploy_settings"].([]any); ok {
		config.DirectDeploySettings = expandDirectDeploySettings(v)
	}

	if v, ok := m["emr_serverless_settings"].([]any); ok {
		config.EmrServerlessSettings = expandEMRServerlessSettings(v)
	}

	if v, ok := m["generative_ai_settings"].([]any); ok {
		config.GenerativeAiSettings = expandGenerativeAiSettings(v)
	}
	if v, ok := m["identity_provider_oauth_settings"].([]any); ok {
		config.IdentityProviderOAuthSettings = expandIdentityProviderOAuthSettings(v)
	}
	if v, ok := m["kendra_settings"].([]any); ok {
		config.KendraSettings = expandKendraSettings(v)
	}
	if v, ok := m["model_register_settings"].([]any); ok {
		config.ModelRegisterSettings = expandModelRegisterSettings(v)
	}
	if v, ok := m["time_series_forecasting_settings"].([]any); ok {
		config.TimeSeriesForecastingSettings = expandTimeSeriesForecastingSettings(v)
	}
	if v, ok := m["workspace_settings"].([]any); ok {
		config.WorkspaceSettings = expandWorkspaceSettings(v)
	}

	return config
}

func expandEMRServerlessSettings(l []any) *awstypes.EmrServerlessSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.EmrServerlessSettings{}

	if v, ok := m[names.AttrExecutionRoleARN].(string); ok && v != "" {
		config.ExecutionRoleArn = aws.String(v)
	}

	if v, ok := m[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.FeatureStatus(v)
	}

	return config
}

func expandKendraSettings(l []any) *awstypes.KendraSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.KendraSettings{}

	if v, ok := m[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.FeatureStatus(v)
	}

	return config
}

func expandDirectDeploySettings(l []any) *awstypes.DirectDeploySettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DirectDeploySettings{}

	if v, ok := m[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.FeatureStatus(v)
	}

	return config
}

func expandGenerativeAiSettings(l []any) *awstypes.GenerativeAiSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.GenerativeAiSettings{}

	if v, ok := m["amazon_bedrock_role_arn"].(string); ok && v != "" {
		config.AmazonBedrockRoleArn = aws.String(v)
	}

	return config
}

func expandIdentityProviderOAuthSettings(l []any) []awstypes.IdentityProviderOAuthSetting {
	providers := make([]awstypes.IdentityProviderOAuthSetting, 0, len(l))

	for _, eRaw := range l {
		data := eRaw.(map[string]any)

		provider := awstypes.IdentityProviderOAuthSetting{}

		if v, ok := data["data_source_name"].(string); ok && v != "" {
			provider.DataSourceName = awstypes.DataSourceName(v)
		}

		if v, ok := data["secret_arn"].(string); ok && v != "" {
			provider.SecretArn = aws.String(v)
		}

		if v, ok := data[names.AttrStatus].(string); ok && v != "" {
			provider.Status = awstypes.FeatureStatus(v)
		}

		providers = append(providers, provider)
	}

	return providers
}

func expandModelRegisterSettings(l []any) *awstypes.ModelRegisterSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.ModelRegisterSettings{}

	if v, ok := m["model_register_settings"].(string); ok && v != "" {
		config.CrossAccountModelRegisterRoleArn = aws.String(v)
	}

	if v, ok := m[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.FeatureStatus(v)
	}

	return config
}

func expandTimeSeriesForecastingSettings(l []any) *awstypes.TimeSeriesForecastingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.TimeSeriesForecastingSettings{}

	if v, ok := m["amazon_forecast_role_arn"].(string); ok && v != "" {
		config.AmazonForecastRoleArn = aws.String(v)
	}

	if v, ok := m[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.FeatureStatus(v)
	}

	return config
}

func expandWorkspaceSettings(l []any) *awstypes.WorkspaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.WorkspaceSettings{}

	if v, ok := m["s3_artifact_path"].(string); ok && v != "" {
		config.S3ArtifactPath = aws.String(v)
	}

	if v, ok := m["s3_kms_key_id"].(string); ok && v != "" {
		config.S3KmsKeyId = aws.String(v)
	}

	return config
}

func expandDomainCustomImages(l []any) []awstypes.CustomImage {
	images := make([]awstypes.CustomImage, 0, len(l))

	for _, eRaw := range l {
		data := eRaw.(map[string]any)

		image := awstypes.CustomImage{
			AppImageConfigName: aws.String(data["app_image_config_name"].(string)),
			ImageName:          aws.String(data["image_name"].(string)),
		}

		if v, ok := data["image_version_number"].(int); ok {
			image.ImageVersionNumber = aws.Int32(int32(v))
		}

		images = append(images, image)
	}

	return images
}

func expandStudioWebPortalSettings(l []any) *awstypes.StudioWebPortalSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.StudioWebPortalSettings{}

	if v, ok := m["hidden_app_types"].(*schema.Set); ok && v.Len() > 0 {
		config.HiddenAppTypes = flex.ExpandStringyValueSet[awstypes.AppType](v)
	}

	if v, ok := m["hidden_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		config.HiddenInstanceTypes = flex.ExpandStringyValueSet[awstypes.AppInstanceType](v)
	}

	if v, ok := m["hidden_ml_tools"].(*schema.Set); ok && v.Len() > 0 {
		config.HiddenMlTools = flex.ExpandStringyValueSet[awstypes.MlTools](v)
	}

	return config
}

func flattenUserSettings(config *awstypes.UserSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	m["auto_mount_home_efs"] = config.AutoMountHomeEFS

	if config.CanvasAppSettings != nil {
		m["canvas_app_settings"] = flattenCanvasAppSettings(config.CanvasAppSettings)
	}

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.ToString(config.ExecutionRole)
	}

	if config.CustomFileSystemConfigs != nil {
		m["custom_file_system_config"] = flattenCustomFileSystemConfigs(config.CustomFileSystemConfigs)
	}

	if config.CustomPosixUserConfig != nil {
		m["custom_posix_user_config"] = flattenCustomPOSIXUserConfig(config.CustomPosixUserConfig)
	}

	if config.CodeEditorAppSettings != nil {
		m["code_editor_app_settings"] = flattenDomainCodeEditorAppSettings(config.CodeEditorAppSettings)
	}

	if config.DefaultLandingUri != nil {
		m["default_landing_uri"] = aws.ToString(config.DefaultLandingUri)
	}

	if config.JupyterLabAppSettings != nil {
		m["jupyter_lab_app_settings"] = flattenDomainJupyterLabAppSettings(config.JupyterLabAppSettings)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.RSessionAppSettings != nil {
		m["r_session_app_settings"] = flattenRSessionAppSettings(config.RSessionAppSettings)
	}

	if config.SecurityGroups != nil {
		m[names.AttrSecurityGroups] = flex.FlattenStringValueSet(config.SecurityGroups)
	}

	if config.SharingSettings != nil {
		m["sharing_settings"] = flattenDomainShareSettings(config.SharingSettings)
	}

	m["studio_web_portal"] = config.StudioWebPortal

	if config.SpaceStorageSettings != nil {
		m["space_storage_settings"] = flattenDefaultSpaceStorageSettings(config.SpaceStorageSettings)
	}

	if config.TensorBoardAppSettings != nil {
		m["tensor_board_app_settings"] = flattenDomainTensorBoardAppSettings(config.TensorBoardAppSettings)
	}

	if config.RStudioServerProAppSettings != nil {
		m["r_studio_server_pro_app_settings"] = flattenRStudioServerProAppSettings(config.RStudioServerProAppSettings)
	}

	if config.StudioWebPortalSettings != nil {
		m["studio_web_portal_settings"] = flattenStudioWebPortalSettings(config.StudioWebPortalSettings)
	}

	return []map[string]any{m}
}

func flattenRStudioServerProAppSettings(config *awstypes.RStudioServerProAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"access_status": config.AccessStatus,
		"user_group":    config.UserGroup,
	}

	return []map[string]any{m}
}

func flattenResourceSpec(config *awstypes.ResourceSpec) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrInstanceType: config.InstanceType,
	}

	if config.LifecycleConfigArn != nil {
		m["lifecycle_config_arn"] = aws.ToString(config.LifecycleConfigArn)
	}

	if config.SageMakerImageArn != nil {
		m["sagemaker_image_arn"] = aws.ToString(config.SageMakerImageArn)
	}

	if config.SageMakerImageVersionAlias != nil {
		m["sagemaker_image_version_alias"] = aws.ToString(config.SageMakerImageVersionAlias)
	}

	if config.SageMakerImageVersionArn != nil {
		m["sagemaker_image_version_arn"] = aws.ToString(config.SageMakerImageVersionArn)
	}

	return []map[string]any{m}
}

func flattenAppLifecycleManagement(config *awstypes.AppLifecycleManagement) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.IdleSettings != nil {
		m["idle_settings"] = flattenIdleSettings(config.IdleSettings)
	}

	return []map[string]any{m}
}

func flattenIdleSettings(config *awstypes.IdleSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.IdleTimeoutInMinutes != nil {
		m["idle_timeout_in_minutes"] = aws.ToInt32(config.IdleTimeoutInMinutes)
	}

	m["lifecycle_management"] = config.LifecycleManagement

	if config.MaxIdleTimeoutInMinutes != nil {
		m["max_idle_timeout_in_minutes"] = aws.ToInt32(config.MaxIdleTimeoutInMinutes)
	}

	if config.MinIdleTimeoutInMinutes != nil {
		m["min_idle_timeout_in_minutes"] = aws.ToInt32(config.MinIdleTimeoutInMinutes)
	}

	return []map[string]any{m}
}

func flattenDefaultSpaceStorageSettings(config *awstypes.DefaultSpaceStorageSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.DefaultEbsStorageSettings != nil {
		m["default_ebs_storage_settings"] = flattenDefaultEBSStorageSettings(config.DefaultEbsStorageSettings)
	}

	return []map[string]any{m}
}

func flattenEMRSettings(config *awstypes.EmrSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AssumableRoleArns != nil {
		m["assumable_role_arns"] = flex.FlattenStringValueSet(config.AssumableRoleArns)
	}

	if config.ExecutionRoleArns != nil {
		m["execution_role_arns"] = flex.FlattenStringValueSet(config.ExecutionRoleArns)
	}

	return []map[string]any{m}
}

func flattenDefaultEBSStorageSettings(config *awstypes.DefaultEbsStorageSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.DefaultEbsVolumeSizeInGb != nil {
		m["default_ebs_volume_size_in_gb"] = aws.ToInt32(config.DefaultEbsVolumeSizeInGb)
	}

	if config.MaximumEbsVolumeSizeInGb != nil {
		m["maximum_ebs_volume_size_in_gb"] = aws.ToInt32(config.MaximumEbsVolumeSizeInGb)
	}

	return []map[string]any{m}
}

func flattenDomainTensorBoardAppSettings(config *awstypes.TensorBoardAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]any{m}
}

func flattenCustomPOSIXUserConfig(config *awstypes.CustomPosixUserConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.Gid != nil {
		m["gid"] = aws.ToInt64(config.Gid)
	}

	if config.Uid != nil {
		m["uid"] = aws.ToInt64(config.Uid)
	}

	return []map[string]any{m}
}

func flattenDomainCodeEditorAppSettings(config *awstypes.CodeEditorAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AppLifecycleManagement != nil {
		m["app_lifecycle_management"] = flattenAppLifecycleManagement(config.AppLifecycleManagement)
	}

	if config.BuiltInLifecycleConfigArn != nil {
		m["built_in_lifecycle_config_arn"] = aws.ToString(config.BuiltInLifecycleConfigArn)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringValueSet(config.LifecycleConfigArns)
	}

	return []map[string]any{m}
}

func flattenDomainJupyterLabAppSettings(config *awstypes.JupyterLabAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AppLifecycleManagement != nil {
		m["app_lifecycle_management"] = flattenAppLifecycleManagement(config.AppLifecycleManagement)
	}

	if config.BuiltInLifecycleConfigArn != nil {
		m["built_in_lifecycle_config_arn"] = aws.ToString(config.BuiltInLifecycleConfigArn)
	}

	if config.CodeRepositories != nil {
		m["code_repository"] = flattenCodeRepositories(config.CodeRepositories)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringValueSet(config.LifecycleConfigArns)
	}

	if config.EmrSettings != nil {
		m["emr_settings"] = flattenEMRSettings(config.EmrSettings)
	}

	return []map[string]any{m}
}

func flattenDomainJupyterServerAppSettings(config *awstypes.JupyterServerAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.CodeRepositories != nil {
		m["code_repository"] = flattenCodeRepositories(config.CodeRepositories)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringValueSet(config.LifecycleConfigArns)
	}

	return []map[string]any{m}
}

func flattenDomainKernelGatewayAppSettings(config *awstypes.KernelGatewayAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringValueSet(config.LifecycleConfigArns)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	return []map[string]any{m}
}

func flattenRSessionAppSettings(config *awstypes.RSessionAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	return []map[string]any{m}
}

func flattenDomainShareSettings(config *awstypes.SharingSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"notebook_output_option": config.NotebookOutputOption,
	}

	if config.S3KmsKeyId != nil {
		m["s3_kms_key_id"] = aws.ToString(config.S3KmsKeyId)
	}

	if config.S3OutputPath != nil {
		m["s3_output_path"] = aws.ToString(config.S3OutputPath)
	}

	return []map[string]any{m}
}

func flattenCanvasAppSettings(config *awstypes.CanvasAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"direct_deploy_settings":           flattenDirectDeploySettings(config.DirectDeploySettings),
		"emr_serverless_settings":          flattenEMRServerlessSettings(config.EmrServerlessSettings),
		"generative_ai_settings":           flattenGenerativeAiSettings(config.GenerativeAiSettings),
		"identity_provider_oauth_settings": flattenIdentityProviderOAuthSettings(config.IdentityProviderOAuthSettings),
		"kendra_settings":                  flattenKendraSettings(config.KendraSettings),
		"time_series_forecasting_settings": flattenTimeSeriesForecastingSettings(config.TimeSeriesForecastingSettings),
		"model_register_settings":          flattenModelRegisterSettings(config.ModelRegisterSettings),
		"workspace_settings":               flattenWorkspaceSettings(config.WorkspaceSettings),
	}

	return []map[string]any{m}
}

func flattenDirectDeploySettings(config *awstypes.DirectDeploySettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrStatus: config.Status,
	}

	return []map[string]any{m}
}

func flattenEMRServerlessSettings(config *awstypes.EmrServerlessSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrExecutionRoleARN: aws.ToString(config.ExecutionRoleArn),
		names.AttrStatus:           config.Status,
	}

	return []map[string]any{m}
}

func flattenGenerativeAiSettings(config *awstypes.GenerativeAiSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"amazon_bedrock_role_arn": aws.ToString(config.AmazonBedrockRoleArn),
	}

	return []map[string]any{m}
}

func flattenKendraSettings(config *awstypes.KendraSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrStatus: config.Status,
	}

	return []map[string]any{m}
}

func flattenIdentityProviderOAuthSettings(config []awstypes.IdentityProviderOAuthSetting) []map[string]any {
	providers := make([]map[string]any, 0, len(config))

	for _, raw := range config {
		provider := make(map[string]any)

		provider["data_source_name"] = raw.DataSourceName

		if raw.SecretArn != nil {
			provider["secret_arn"] = aws.ToString(raw.SecretArn)
		}

		provider[names.AttrStatus] = raw.Status

		providers = append(providers, provider)
	}

	return providers
}

func flattenModelRegisterSettings(config *awstypes.ModelRegisterSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"cross_account_model_register_role_arn": aws.ToString(config.CrossAccountModelRegisterRoleArn),
		names.AttrStatus:                        config.Status,
	}

	return []map[string]any{m}
}

func flattenTimeSeriesForecastingSettings(config *awstypes.TimeSeriesForecastingSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"amazon_forecast_role_arn": aws.ToString(config.AmazonForecastRoleArn),
		names.AttrStatus:           config.Status,
	}

	return []map[string]any{m}
}

func flattenWorkspaceSettings(config *awstypes.WorkspaceSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"s3_artifact_path": aws.ToString(config.S3ArtifactPath),
		"s3_kms_key_id":    aws.ToString(config.S3KmsKeyId),
	}

	return []map[string]any{m}
}

func flattenDomainSettings(config *awstypes.DomainSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"docker_settings":                     flattenDockerSettings(config.DockerSettings),
		"execution_role_identity_config":      config.ExecutionRoleIdentityConfig,
		"r_studio_server_pro_domain_settings": flattenRStudioServerProDomainSettings(config.RStudioServerProDomainSettings),
		names.AttrSecurityGroupIDs:            flex.FlattenStringValueSet(config.SecurityGroupIds),
	}

	return []map[string]any{m}
}

func flattenDockerSettings(config *awstypes.DockerSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.EnableDockerAccess != "" {
		m["enable_docker_access"] = config.EnableDockerAccess
	}

	if config.VpcOnlyTrustedAccounts != nil {
		m["vpc_only_trusted_accounts"] = flex.FlattenStringValueSet(config.VpcOnlyTrustedAccounts)
	}

	return []map[string]any{m}
}

func flattenRStudioServerProDomainSettings(config *awstypes.RStudioServerProDomainSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"r_studio_connect_url":         aws.ToString(config.RStudioConnectUrl),
		"domain_execution_role_arn":    aws.ToString(config.DomainExecutionRoleArn),
		"r_studio_package_manager_url": aws.ToString(config.RStudioPackageManagerUrl),
		"default_resource_spec":        flattenResourceSpec(config.DefaultResourceSpec),
	}

	return []map[string]any{m}
}

func flattenDomainCustomImages(config []awstypes.CustomImage) []map[string]any {
	images := make([]map[string]any, 0, len(config))

	for _, raw := range config {
		image := make(map[string]any)

		image["app_image_config_name"] = aws.ToString(raw.AppImageConfigName)
		image["image_name"] = aws.ToString(raw.ImageName)

		if raw.ImageVersionNumber != nil {
			image["image_version_number"] = aws.ToInt32(raw.ImageVersionNumber)
		}

		images = append(images, image)
	}

	return images
}

func expanDefaultSpaceSettings(l []any) *awstypes.DefaultSpaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.DefaultSpaceSettings{}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]any); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["jupyter_lab_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterLabAppSettings = expandDomainJupyterLabAppSettings(v)
	}

	if v, ok := m["space_storage_settings"].([]any); ok && len(v) > 0 {
		config.SpaceStorageSettings = expandDefaultSpaceStorageSettings(v)
	}

	if v, ok := m["custom_file_system_config"].([]any); ok && len(v) > 0 {
		config.CustomFileSystemConfigs = expandCustomFileSystemConfigs(v)
	}

	if v, ok := m["custom_posix_user_config"].([]any); ok && len(v) > 0 {
		config.CustomPosixUserConfig = expandCustomPOSIXUserConfig(v)
	}

	return config
}

func flattenDefaultSpaceSettings(config *awstypes.DefaultSpaceSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.ToString(config.ExecutionRole)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.SecurityGroups != nil {
		m[names.AttrSecurityGroups] = flex.FlattenStringValueSet(config.SecurityGroups)
	}

	if config.JupyterLabAppSettings != nil {
		m["jupyter_lab_app_settings"] = flattenDomainJupyterLabAppSettings(config.JupyterLabAppSettings)
	}

	if config.SpaceStorageSettings != nil {
		m["space_storage_settings"] = flattenDefaultSpaceStorageSettings(config.SpaceStorageSettings)
	}

	if config.CustomFileSystemConfigs != nil {
		m["custom_file_system_config"] = flattenCustomFileSystemConfigs(config.CustomFileSystemConfigs)
	}

	if config.CustomPosixUserConfig != nil {
		m["custom_posix_user_config"] = flattenCustomPOSIXUserConfig(config.CustomPosixUserConfig)
	}

	return []map[string]any{m}
}

func expandCodeRepository(tfMap map[string]any) awstypes.CodeRepository {
	apiObject := awstypes.CodeRepository{
		RepositoryUrl: aws.String(tfMap["repository_url"].(string)),
	}

	return apiObject
}

func expandCodeRepositories(tfList []any) []awstypes.CodeRepository {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CodeRepository

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandCodeRepository(tfMap))
	}

	return apiObjects
}

func flattenCodeRepository(apiObject awstypes.CodeRepository) map[string]any {
	tfMap := map[string]any{}

	if apiObject.RepositoryUrl != nil {
		tfMap["repository_url"] = aws.ToString(apiObject.RepositoryUrl)
	}

	return tfMap
}

func flattenCodeRepositories(apiObjects []awstypes.CodeRepository) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCodeRepository(apiObject))
	}

	return tfList
}

func expandCustomFileSystemConfig(tfMap map[string]any) awstypes.CustomFileSystemConfig {
	apiObject := &awstypes.CustomFileSystemConfigMemberEFSFileSystemConfig{}

	if v, ok := tfMap["efs_file_system_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandEFSFileSystemConfig(v[0].(map[string]any))
	}

	return apiObject
}

func expandCustomFileSystemConfigs(tfList []any) []awstypes.CustomFileSystemConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CustomFileSystemConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandCustomFileSystemConfig(tfMap))
	}

	return apiObjects
}

func expandEFSFileSystemConfig(tfMap map[string]any) awstypes.EFSFileSystemConfig {
	apiObject := awstypes.EFSFileSystemConfig{}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok {
		apiObject.FileSystemId = aws.String(v)
	}

	if v, ok := tfMap["file_system_path"].(string); ok {
		apiObject.FileSystemPath = aws.String(v)
	}

	return apiObject
}

func flattenCustomFileSystemConfig(apiObject awstypes.CustomFileSystemConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject, ok := apiObject.(*awstypes.CustomFileSystemConfigMemberEFSFileSystemConfig); ok {
		tfMap["efs_file_system_config"] = flattenEFSFileSystemConfig(apiObject.Value)
	}

	return tfMap
}

func flattenCustomFileSystemConfigs(apiObjects []awstypes.CustomFileSystemConfig) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCustomFileSystemConfig(apiObject))
	}

	return tfList
}

func flattenEFSFileSystemConfig(apiObject awstypes.EFSFileSystemConfig) []map[string]any {
	tfMap := map[string]any{}

	if apiObject.FileSystemId != nil {
		tfMap[names.AttrFileSystemID] = aws.ToString(apiObject.FileSystemId)
	}

	if apiObject.FileSystemPath != nil {
		tfMap["file_system_path"] = aws.ToString(apiObject.FileSystemPath)
	}

	return []map[string]any{tfMap}
}

func flattenStudioWebPortalSettings(config *awstypes.StudioWebPortalSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.HiddenAppTypes != nil {
		m["hidden_app_types"] = flex.FlattenStringyValueSet[awstypes.AppType](config.HiddenAppTypes)
	}

	if config.HiddenInstanceTypes != nil {
		m["hidden_instance_types"] = flex.FlattenStringyValueSet[awstypes.AppInstanceType](config.HiddenInstanceTypes)
	}

	if config.HiddenMlTools != nil {
		m["hidden_ml_tools"] = flex.FlattenStringyValueSet[awstypes.MlTools](config.HiddenMlTools)
	}

	return []map[string]any{m}
}
