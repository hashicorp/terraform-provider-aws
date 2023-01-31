package sagemaker

import (
	"context"
	"log"
	"regexp"
	"strings"

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
)

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
										MaxItems: 30,
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
										MaxItems: 30,
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
										MaxItems: 30,
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
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
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
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
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &sagemaker.CreateDomainInput{
		DomainName:           aws.String(d.Get("domain_name").(string)),
		AuthMode:             aws.String(d.Get("auth_mode").(string)),
		VpcId:                aws.String(d.Get("vpc_id").(string)),
		AppNetworkAccessType: aws.String(d.Get("app_network_access_type").(string)),
		SubnetIds:            flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		DefaultUserSettings:  expandDomainDefaultUserSettings(d.Get("default_user_settings").([]interface{})),
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
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
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Domain (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

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
		TimeSeriesForecastingSettings: expandTimeSeriesForecastingSettings(m["time_series_forecasting_settings"].([]interface{})),
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
		"time_series_forecasting_settings": flattenTimeSeriesForecastingSettings(config.TimeSeriesForecastingSettings),
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

func flattenDomainSettings(config *sagemaker.DomainSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"execution_role_identity_config": aws.StringValue(config.ExecutionRoleIdentityConfig),
		"security_group_ids":             flex.FlattenStringSet(config.SecurityGroupIds),
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
