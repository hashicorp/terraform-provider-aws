// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func ResourceDomain() *schema.Resource {
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
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      sagemaker.AppNetworkAccessTypePublicInternetOnly,
				ValidateFunc: validation.StringInSlice(sagemaker.AppNetworkAccessType_Values(), false),
			},
			"app_security_group_management": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AppSecurityGroupManagement_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_mode": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AuthMode_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
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
												"status": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.FeatureStatus_Values(), false),
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
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.DataSourceName_Values(), false),
												},
												"secret_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
												"status": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.FeatureStatus_Values(), false),
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
												"status": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.FeatureStatus_Values(), false),
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
												"status": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.FeatureStatus_Values(), false),
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
												"status": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.FeatureStatus_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RStudioServerProAccessStatus_Values(), false),
									},
									"user_group": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      sagemaker.RStudioServerProUserGroupRStudioUser,
										ValidateFunc: validation.StringInSlice(sagemaker.RStudioServerProUserGroup_Values(), false),
									},
								},
							},
						},
						"security_groups": {
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
										Type:         schema.TypeString,
										Optional:     true,
										Default:      sagemaker.NotebookOutputOptionDisabled,
										ValidateFunc: validation.StringInSlice(sagemaker.NotebookOutputOption_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
					},
				},
			},
			"domain_name": {
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
						"execution_role_identity_config": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ExecutionRoleIdentityConfig_Values(), false),
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
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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
						"security_group_ids": {
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
			"kms_key_id": {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.RetentionType_Values(), false),
							Default:      sagemaker.RetentionTypeRetain,
						},
					},
				},
			},
			"security_group_id_for_domain_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"single_sign_on_managed_application_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 16,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.CreateDomainInput{
		DomainName:           aws.String(d.Get("domain_name").(string)),
		AuthMode:             aws.String(d.Get("auth_mode").(string)),
		VpcId:                aws.String(d.Get("vpc_id").(string)),
		AppNetworkAccessType: aws.String(d.Get("app_network_access_type").(string)),
		SubnetIds:            flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		DefaultUserSettings:  expandDomainDefaultUserSettings(d.Get("default_user_settings").([]interface{})),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("app_security_group_management"); ok {
		input.AppSecurityGroupManagement = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_settings"); ok && len(v.([]interface{})) > 0 {
		input.DomainSettings = expandDomainSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("default_space_settings"); ok && len(v.([]interface{})) > 0 {
		input.DefaultSpaceSettings = expanDefaultSpaceSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] SageMaker Domain create config: %#v", *input)
	output, err := conn.CreateDomainWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Domain: %s", err)
	}

	domainArn := aws.StringValue(output.DomainArn)
	domainID, err := DecodeDomainID(domainArn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Domain (%s): %s", d.Id(), err)
	}

	d.SetId(domainID)

	if _, err := WaitDomainInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Domain (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	domain, err := FindDomainByName(ctx, conn, d.Id())
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Domain (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Domain (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(domain.DomainArn)
	d.Set("domain_name", domain.DomainName)
	d.Set("auth_mode", domain.AuthMode)
	d.Set("app_network_access_type", domain.AppNetworkAccessType)
	d.Set("arn", arn)
	d.Set("home_efs_file_system_id", domain.HomeEfsFileSystemId)
	d.Set("single_sign_on_managed_application_instance_id", domain.SingleSignOnManagedApplicationInstanceId)
	d.Set("url", domain.Url)
	d.Set("vpc_id", domain.VpcId)
	d.Set("kms_key_id", domain.KmsKeyId)
	d.Set("app_security_group_management", domain.AppSecurityGroupManagement)
	d.Set("security_group_id_for_domain_boundary", domain.SecurityGroupIdForDomainBoundary)

	if err := d.Set("subnet_ids", flex.FlattenStringSet(domain.SubnetIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids for SageMaker Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_user_settings", flattenDomainDefaultUserSettings(domain.DefaultUserSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_user_settings for SageMaker Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_space_settings", flattenDefaultSpaceSettings(domain.DefaultSpaceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_space_settings for SageMaker Domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("domain_settings", flattenDomainSettings(domain.DomainSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_settings for SageMaker Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &sagemaker.UpdateDomainInput{
			DomainId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("default_user_settings"); ok && len(v.([]interface{})) > 0 {
			input.DefaultUserSettings = expandDomainDefaultUserSettings(v.([]interface{}))
		}

		if v, ok := d.GetOk("domain_settings"); ok && len(v.([]interface{})) > 0 {
			input.DomainSettingsForUpdate = expandDomainSettingsUpdate(v.([]interface{}))
		}

		if v, ok := d.GetOk("default_space_settings"); ok && len(v.([]interface{})) > 0 {
			input.DefaultSpaceSettings = expanDefaultSpaceSettings(v.([]interface{}))
		}

		log.Printf("[DEBUG] SageMaker Domain update config: %#v", *input)
		_, err := conn.UpdateDomainWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Domain: %s", err)
		}

		if _, err := WaitDomainInService(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Domain (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteDomainInput{
		DomainId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("retention_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetentionPolicy = expandRetentionPolicy(v.([]interface{}))
	}

	if _, err := conn.DeleteDomainWithContext(ctx, input); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker Domain (%s): %s", d.Id(), err)
		}
	}

	if _, err := WaitDomainDeleted(ctx, conn, d.Id()); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Domain (%s) to delete: %s", d.Id(), err)
		}
	}

	return diags
}

func expandDomainSettings(l []interface{}) *sagemaker.DomainSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DomainSettings{}

	if v, ok := m["execution_role_identity_config"].(string); ok && v != "" {
		config.ExecutionRoleIdentityConfig = aws.String(v)
	}

	if v, ok := m["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := m["r_studio_server_pro_domain_settings"].([]interface{}); ok && len(v) > 0 {
		config.RStudioServerProDomainSettings = expandRStudioServerProDomainSettings(v)
	}

	return config
}

func expandRStudioServerProDomainSettings(l []interface{}) *sagemaker.RStudioServerProDomainSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.RStudioServerProDomainSettings{}

	if v, ok := m["domain_execution_role_arn"].(string); ok && v != "" {
		config.DomainExecutionRoleArn = aws.String(v)
	}

	if v, ok := m["r_studio_connect_url"].(string); ok && v != "" {
		config.RStudioConnectUrl = aws.String(v)
	}

	if v, ok := m["r_studio_packageManager_url"].(string); ok && v != "" {
		config.RStudioPackageManagerUrl = aws.String(v)
	}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandDomainDefaultResourceSpec(v)
	}

	return config
}

func expandDomainSettingsUpdate(l []interface{}) *sagemaker.DomainSettingsForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DomainSettingsForUpdate{}

	if v, ok := m["execution_role_identity_config"].(string); ok && v != "" {
		config.ExecutionRoleIdentityConfig = aws.String(v)
	}

	return config
}

func expandRetentionPolicy(l []interface{}) *sagemaker.RetentionPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.RetentionPolicy{}

	if v, ok := m["home_efs_file_system"].(string); ok && v != "" {
		config.HomeEfsFileSystem = aws.String(v)
	}

	return config
}

func expandDomainDefaultUserSettings(l []interface{}) *sagemaker.UserSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.UserSettings{}

	if v, ok := m["canvas_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.CanvasAppSettings = expandCanvasAppSettings(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["r_session_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.RSessionAppSettings = expandRSessionAppSettings(v)
	}

	if v, ok := m["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = flex.ExpandStringSet(v)
	}

	if v, ok := m["sharing_settings"].([]interface{}); ok && len(v) > 0 {
		config.SharingSettings = expandDomainShareSettings(v)
	}

	if v, ok := m["tensor_board_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.TensorBoardAppSettings = expandDomainTensorBoardAppSettings(v)
	}

	if v, ok := m["r_studio_server_pro_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.RStudioServerProAppSettings = expandRStudioServerProAppSettings(v)
	}

	return config
}

func expandRStudioServerProAppSettings(l []interface{}) *sagemaker.RStudioServerProAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.RStudioServerProAppSettings{}

	if v, ok := m["access_status"].(string); ok && v != "" {
		config.AccessStatus = aws.String(v)

		if v == sagemaker.RStudioServerProAccessStatusEnabled {
			if g, ok := m["user_group"].(string); ok && g != "" {
				config.UserGroup = aws.String(g)
			}
		}
	}

	return config
}

func expandDomainJupyterServerAppSettings(l []interface{}) *sagemaker.JupyterServerAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.JupyterServerAppSettings{}

	if v, ok := m["code_repository"].(*schema.Set); ok && v.Len() > 0 {
		config.CodeRepositories = expandCodeRepositories(v.List())
	}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandDomainDefaultResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringSet(v)
	}

	return config
}

func expandDomainKernelGatewayAppSettings(l []interface{}) *sagemaker.KernelGatewayAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.KernelGatewayAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandDomainDefaultResourceSpec(v)
	}

	if v, ok := m["lifecycle_config_arns"].(*schema.Set); ok && v.Len() > 0 {
		config.LifecycleConfigArns = flex.ExpandStringSet(v)
	}

	if v, ok := m["custom_image"].([]interface{}); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	return config
}

func expandRSessionAppSettings(l []interface{}) *sagemaker.RSessionAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.RSessionAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandDomainDefaultResourceSpec(v)
	}

	if v, ok := m["custom_image"].([]interface{}); ok && len(v) > 0 {
		config.CustomImages = expandDomainCustomImages(v)
	}

	return config
}

func expandDomainTensorBoardAppSettings(l []interface{}) *sagemaker.TensorBoardAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.TensorBoardAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandDomainDefaultResourceSpec(v)
	}

	return config
}

func expandDomainDefaultResourceSpec(l []interface{}) *sagemaker.ResourceSpec {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.ResourceSpec{}

	if v, ok := m["instance_type"].(string); ok && v != "" {
		config.InstanceType = aws.String(v)
	}

	if v, ok := m["lifecycle_config_arn"].(string); ok && v != "" {
		config.LifecycleConfigArn = aws.String(v)
	}

	if v, ok := m["sagemaker_image_arn"].(string); ok && v != "" {
		config.SageMakerImageArn = aws.String(v)
	}

	if v, ok := m["sagemaker_image_version_arn"].(string); ok && v != "" {
		config.SageMakerImageVersionArn = aws.String(v)
	}

	return config
}

func expandDomainShareSettings(l []interface{}) *sagemaker.SharingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SharingSettings{
		NotebookOutputOption: aws.String(m["notebook_output_option"].(string)),
	}

	if v, ok := m["s3_kms_key_id"].(string); ok && v != "" {
		config.S3KmsKeyId = aws.String(v)
	}

	if v, ok := m["s3_output_path"].(string); ok && v != "" {
		config.S3OutputPath = aws.String(v)
	}

	return config
}

func expandCanvasAppSettings(l []interface{}) *sagemaker.CanvasAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CanvasAppSettings{
		IdentityProviderOAuthSettings: expandIdentityProviderOAuthSettings(m["identity_provider_oauth_settings"].([]interface{})),
		DirectDeploySettings:          expandDirectDeploySettings(m["direct_deploy_settings"].([]interface{})),
		KendraSettings:                expandKendraSettings(m["kendra_settings"].([]interface{})),
		ModelRegisterSettings:         expandModelRegisterSettings(m["model_register_settings"].([]interface{})),
		TimeSeriesForecastingSettings: expandTimeSeriesForecastingSettings(m["time_series_forecasting_settings"].([]interface{})),
		WorkspaceSettings:             expandWorkspaceSettings(m["workspace_settings"].([]interface{})),
	}

	return config
}

func expandKendraSettings(l []interface{}) *sagemaker.KendraSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.KendraSettings{}

	if v, ok := m["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandDirectDeploySettings(l []interface{}) *sagemaker.DirectDeploySettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DirectDeploySettings{}

	if v, ok := m["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandIdentityProviderOAuthSettings(l []interface{}) []*sagemaker.IdentityProviderOAuthSetting {
	providers := make([]*sagemaker.IdentityProviderOAuthSetting, 0, len(l))

	for _, eRaw := range l {
		data := eRaw.(map[string]interface{})

		provider := &sagemaker.IdentityProviderOAuthSetting{}

		if v, ok := data["data_source_name"].(string); ok && v != "" {
			provider.DataSourceName = aws.String(v)
		}

		if v, ok := data["secret_arn"].(string); ok && v != "" {
			provider.SecretArn = aws.String(v)
		}

		if v, ok := data["status"].(string); ok && v != "" {
			provider.Status = aws.String(v)
		}

		providers = append(providers, provider)
	}

	return providers
}

func expandModelRegisterSettings(l []interface{}) *sagemaker.ModelRegisterSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.ModelRegisterSettings{}

	if v, ok := m["model_register_settings"].(string); ok && v != "" {
		config.CrossAccountModelRegisterRoleArn = aws.String(v)
	}

	if v, ok := m["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandTimeSeriesForecastingSettings(l []interface{}) *sagemaker.TimeSeriesForecastingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.TimeSeriesForecastingSettings{}

	if v, ok := m["amazon_forecast_role_arn"].(string); ok && v != "" {
		config.AmazonForecastRoleArn = aws.String(v)
	}

	if v, ok := m["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandWorkspaceSettings(l []interface{}) *sagemaker.WorkspaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.WorkspaceSettings{}

	if v, ok := m["s3_artifact_path"].(string); ok && v != "" {
		config.S3ArtifactPath = aws.String(v)
	}

	if v, ok := m["s3_kms_key_id"].(string); ok && v != "" {
		config.S3KmsKeyId = aws.String(v)
	}

	return config
}

func expandDomainCustomImages(l []interface{}) []*sagemaker.CustomImage {
	images := make([]*sagemaker.CustomImage, 0, len(l))

	for _, eRaw := range l {
		data := eRaw.(map[string]interface{})

		image := &sagemaker.CustomImage{
			AppImageConfigName: aws.String(data["app_image_config_name"].(string)),
			ImageName:          aws.String(data["image_name"].(string)),
		}

		if v, ok := data["image_version_number"].(int); ok {
			image.ImageVersionNumber = aws.Int64(int64(v))
		}

		images = append(images, image)
	}

	return images
}

func flattenDomainDefaultUserSettings(config *sagemaker.UserSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.CanvasAppSettings != nil {
		m["canvas_app_settings"] = flattenCanvasAppSettings(config.CanvasAppSettings)
	}

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.StringValue(config.ExecutionRole)
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
		m["security_groups"] = flex.FlattenStringSet(config.SecurityGroups)
	}

	if config.SharingSettings != nil {
		m["sharing_settings"] = flattenDomainShareSettings(config.SharingSettings)
	}

	if config.TensorBoardAppSettings != nil {
		m["tensor_board_app_settings"] = flattenDomainTensorBoardAppSettings(config.TensorBoardAppSettings)
	}

	if config.RStudioServerProAppSettings != nil {
		m["r_studio_server_pro_app_settings"] = flattenRStudioServerProAppSettings(config.RStudioServerProAppSettings)
	}

	return []map[string]interface{}{m}
}

func flattenRStudioServerProAppSettings(config *sagemaker.RStudioServerProAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.AccessStatus != nil {
		m["access_status"] = aws.StringValue(config.AccessStatus)
	}

	if config.UserGroup != nil {
		m["user_group"] = aws.StringValue(config.UserGroup)
	}

	return []map[string]interface{}{m}
}

func flattenDomainDefaultResourceSpec(config *sagemaker.ResourceSpec) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.InstanceType != nil {
		m["instance_type"] = aws.StringValue(config.InstanceType)
	}

	if config.LifecycleConfigArn != nil {
		m["lifecycle_config_arn"] = aws.StringValue(config.LifecycleConfigArn)
	}

	if config.SageMakerImageArn != nil {
		m["sagemaker_image_arn"] = aws.StringValue(config.SageMakerImageArn)
	}

	if config.SageMakerImageVersionArn != nil {
		m["sagemaker_image_version_arn"] = aws.StringValue(config.SageMakerImageVersionArn)
	}

	return []map[string]interface{}{m}
}

func flattenDomainTensorBoardAppSettings(config *sagemaker.TensorBoardAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]interface{}{m}
}

func flattenDomainJupyterServerAppSettings(config *sagemaker.JupyterServerAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.CodeRepositories != nil {
		m["code_repository"] = flattenCodeRepositories(config.CodeRepositories)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringSet(config.LifecycleConfigArns)
	}

	return []map[string]interface{}{m}
}

func flattenDomainKernelGatewayAppSettings(config *sagemaker.KernelGatewayAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	if config.LifecycleConfigArns != nil {
		m["lifecycle_config_arns"] = flex.FlattenStringSet(config.LifecycleConfigArns)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	return []map[string]interface{}{m}
}

func flattenRSessionAppSettings(config *sagemaker.RSessionAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenDomainCustomImages(config.CustomImages)
	}

	return []map[string]interface{}{m}
}

func flattenDomainShareSettings(config *sagemaker.SharingSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notebook_output_option": aws.StringValue(config.NotebookOutputOption),
	}

	if config.S3KmsKeyId != nil {
		m["s3_kms_key_id"] = aws.StringValue(config.S3KmsKeyId)
	}

	if config.S3OutputPath != nil {
		m["s3_output_path"] = aws.StringValue(config.S3OutputPath)
	}

	return []map[string]interface{}{m}
}

func flattenCanvasAppSettings(config *sagemaker.CanvasAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"direct_deploy_settings":           flattenDirectDeploySettings(config.DirectDeploySettings),
		"identity_provider_oauth_settings": flattenIdentityProviderOAuthSettings(config.IdentityProviderOAuthSettings),
		"kendra_settings":                  flattenKendraSettings(config.KendraSettings),
		"time_series_forecasting_settings": flattenTimeSeriesForecastingSettings(config.TimeSeriesForecastingSettings),
		"model_register_settings":          flattenModelRegisterSettings(config.ModelRegisterSettings),
		"workspace_settings":               flattenWorkspaceSettings(config.WorkspaceSettings),
	}

	return []map[string]interface{}{m}
}

func flattenDirectDeploySettings(config *sagemaker.DirectDeploySettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"status": aws.StringValue(config.Status),
	}

	return []map[string]interface{}{m}
}

func flattenKendraSettings(config *sagemaker.KendraSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"status": aws.StringValue(config.Status),
	}

	return []map[string]interface{}{m}
}

func flattenIdentityProviderOAuthSettings(config []*sagemaker.IdentityProviderOAuthSetting) []map[string]interface{} {
	providers := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		provider := make(map[string]interface{})

		if raw.DataSourceName != nil {
			provider["data_source_name"] = aws.StringValue(raw.DataSourceName)
		}

		if raw.SecretArn != nil {
			provider["secret_arn"] = aws.StringValue(raw.SecretArn)
		}

		if raw.Status != nil {
			provider["status"] = aws.StringValue(raw.Status)
		}

		providers = append(providers, provider)
	}

	return providers
}

func flattenModelRegisterSettings(config *sagemaker.ModelRegisterSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cross_account_model_register_role_arn": aws.StringValue(config.CrossAccountModelRegisterRoleArn),
		"status":                                aws.StringValue(config.Status),
	}

	return []map[string]interface{}{m}
}

func flattenTimeSeriesForecastingSettings(config *sagemaker.TimeSeriesForecastingSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"amazon_forecast_role_arn": aws.StringValue(config.AmazonForecastRoleArn),
		"status":                   aws.StringValue(config.Status),
	}

	return []map[string]interface{}{m}
}

func flattenWorkspaceSettings(config *sagemaker.WorkspaceSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_artifact_path": aws.StringValue(config.S3ArtifactPath),
		"s3_kms_key_id":    aws.StringValue(config.S3KmsKeyId),
	}

	return []map[string]interface{}{m}
}

func flattenDomainSettings(config *sagemaker.DomainSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"execution_role_identity_config":      aws.StringValue(config.ExecutionRoleIdentityConfig),
		"r_studio_server_pro_domain_settings": flattenRStudioServerProDomainSettings(config.RStudioServerProDomainSettings),
		"security_group_ids":                  flex.FlattenStringSet(config.SecurityGroupIds),
	}

	return []map[string]interface{}{m}
}

func flattenRStudioServerProDomainSettings(config *sagemaker.RStudioServerProDomainSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"r_studio_connect_url":         aws.StringValue(config.RStudioConnectUrl),
		"domain_execution_role_arn":    aws.StringValue(config.DomainExecutionRoleArn),
		"r_studio_package_manager_url": aws.StringValue(config.RStudioPackageManagerUrl),
		"default_resource_spec":        flattenDomainDefaultResourceSpec(config.DefaultResourceSpec),
	}

	return []map[string]interface{}{m}
}

func flattenDomainCustomImages(config []*sagemaker.CustomImage) []map[string]interface{} {
	images := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		image := make(map[string]interface{})

		image["app_image_config_name"] = aws.StringValue(raw.AppImageConfigName)
		image["image_name"] = aws.StringValue(raw.ImageName)

		if raw.ImageVersionNumber != nil {
			image["image_version_number"] = aws.Int64Value(raw.ImageVersionNumber)
		}

		images = append(images, image)
	}

	return images
}

func DecodeDomainID(id string) (string, error) {
	domainArn, err := arn.Parse(id)
	if err != nil {
		return "", err
	}

	domainName := strings.TrimPrefix(domainArn.Resource, "domain/")
	return domainName, nil
}

func expanDefaultSpaceSettings(l []interface{}) *sagemaker.DefaultSpaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DefaultSpaceSettings{}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = flex.ExpandStringSet(v)
	}

	return config
}

func flattenDefaultSpaceSettings(config *sagemaker.DefaultSpaceSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.StringValue(config.ExecutionRole)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.SecurityGroups != nil {
		m["security_groups"] = flex.FlattenStringSet(config.SecurityGroups)
	}

	return []map[string]interface{}{m}
}

func expandCodeRepository(tfMap map[string]interface{}) *sagemaker.CodeRepository {
	if tfMap == nil {
		return nil
	}

	apiObject := &sagemaker.CodeRepository{
		RepositoryUrl: aws.String(tfMap["repository_url"].(string)),
	}

	return apiObject
}

func expandCodeRepositories(tfList []interface{}) []*sagemaker.CodeRepository {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*sagemaker.CodeRepository

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCodeRepository(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCodeRepository(apiObject *sagemaker.CodeRepository) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.RepositoryUrl != nil {
		tfMap["repository_url"] = aws.StringValue(apiObject.RepositoryUrl)
	}

	return tfMap
}

func flattenCodeRepositories(apiObjects []*sagemaker.CodeRepository) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCodeRepository(apiObject))
	}

	return tfList
}
