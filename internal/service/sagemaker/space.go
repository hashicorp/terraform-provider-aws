// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_space", name="Space")
// @Tags(identifierAttribute="arn")
func resourceSpace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSpaceCreate,
		ReadWithoutTimeout:   resourceSpaceRead,
		UpdateWithoutTimeout: resourceSpaceUpdate,
		DeleteWithoutTimeout: resourceSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"home_efs_file_system_uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ownership_settings": {
				Type:         schema.TypeList,
				RequiredWith: []string{"space_sharing_settings"},
				Optional:     true,
				MaxItems:     1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner_user_profile_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"space_display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"space_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AppType](),
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
														},
													},
												},
											},
										},
									},
									"default_resource_spec": {
										Type:     schema.TypeList,
										Required: true,
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
						"custom_file_system": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"efs_file_system": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrFileSystemID: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
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
														},
													},
												},
											},
										},
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
									"default_resource_spec": {
										Type:     schema.TypeList,
										Required: true,
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
										Required: true,
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
										Required: true,
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
						"space_storage_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ebs_storage_settings": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"ebs_volume_size_in_gb": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(5, 16384),
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
			"space_sharing_settings": {
				Type:         schema.TypeList,
				RequiredWith: []string{"ownership_settings"},
				Optional:     true,
				MaxItems:     1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sharing_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SharingType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	domainId := d.Get("domain_id").(string)
	spaceName := d.Get("space_name").(string)
	input := &sagemaker.CreateSpaceInput{
		SpaceName: aws.String(spaceName),
		DomainId:  aws.String(domainId),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ownership_settings"); ok && len(v.([]any)) > 0 {
		input.OwnershipSettings = expandOwnershipSettings(v.([]any))
	}

	if v, ok := d.GetOk("space_settings"); ok && len(v.([]any)) > 0 {
		input.SpaceSettings = expandSpaceSettings(v.([]any))
	}

	if v, ok := d.GetOk("space_sharing_settings"); ok && len(v.([]any)) > 0 {
		input.SpaceSharingSettings = expandSpaceSharingSettings(v.([]any))
	}

	if v, ok := d.GetOk("space_display_name"); ok {
		input.SpaceDisplayName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] SageMaker AI Space create config: %#v", *input)
	out, err := conn.CreateSpace(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Space: %s", err)
	}

	d.SetId(aws.ToString(out.SpaceArn))

	if err := waitSpaceInService(ctx, conn, domainId, spaceName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Space (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceSpaceRead(ctx, d, meta)...)
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	domainID, name, err := decodeSpaceName(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Space (%s): %s", d.Id(), err)
	}

	space, err := findSpaceByName(ctx, conn, domainID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Space (%s), removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Space (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, space.SpaceArn)
	d.Set("domain_id", space.DomainId)
	d.Set("home_efs_file_system_uid", space.HomeEfsFileSystemUid)
	d.Set("space_display_name", space.SpaceDisplayName)
	d.Set("space_name", space.SpaceName)
	d.Set(names.AttrURL, space.Url)

	if err := d.Set("ownership_settings", flattenOwnershipSettings(space.OwnershipSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ownership_settings for SageMaker AI Space (%s): %s", d.Id(), err)
	}

	if err := d.Set("space_settings", flattenSpaceSettings(space.SpaceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting space_settings for SageMaker AI Space (%s): %s", d.Id(), err)
	}

	if err := d.Set("space_sharing_settings", flattenSpaceSharingSettings(space.SpaceSharingSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting space_sharing_settings for SageMaker AI Space (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		domainID := d.Get("domain_id").(string)
		name := d.Get("space_name").(string)

		input := &sagemaker.UpdateSpaceInput{
			SpaceName: aws.String(name),
			DomainId:  aws.String(domainID),
		}

		if d.HasChanges("space_settings") {
			input.SpaceSettings = expandSpaceSettings(d.Get("space_settings").([]any))
		}

		if d.HasChange("space_display_name") {
			input.SpaceDisplayName = aws.String(d.Get("space_display_name").(string))
		}

		log.Printf("[DEBUG] SageMaker AI Space update config: %#v", *input)
		_, err := conn.UpdateSpace(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Space: %s", err)
		}

		if err := waitSpaceInService(ctx, conn, domainID, name); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Space (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSpaceRead(ctx, d, meta)...)
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("space_name").(string)
	domainID := d.Get("domain_id").(string)

	input := &sagemaker.DeleteSpaceInput{
		SpaceName: aws.String(name),
		DomainId:  aws.String(domainID),
	}

	if _, err := conn.DeleteSpace(ctx, input); err != nil {
		if !errs.IsA[*awstypes.ResourceNotFound](err) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Space (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitSpaceDeleted(ctx, conn, domainID, name); err != nil {
		if !errs.IsA[*awstypes.ResourceNotFound](err) {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Space (%s) to delete: %s", d.Id(), err)
		}
	}

	return diags
}

func findSpaceByName(ctx context.Context, conn *sagemaker.Client, domainId, name string) (*sagemaker.DescribeSpaceOutput, error) {
	input := &sagemaker.DescribeSpaceInput{
		SpaceName: aws.String(name),
		DomainId:  aws.String(domainId),
	}

	output, err := conn.DescribeSpace(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "RecordNotFound") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

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

func decodeSpaceName(id string) (string, string, error) {
	userProfileARN, err := arn.Parse(id)
	if err != nil {
		return "", "", err
	}

	userProfileResourceNameName := strings.TrimPrefix(userProfileARN.Resource, "space/")
	parts := strings.Split(userProfileResourceNameName, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected DOMAIN-ID/SPACE-NAME", userProfileResourceNameName)
	}

	domainID := parts[0]
	spaceName := parts[1]

	return domainID, spaceName, nil
}

func expandSpaceSettings(l []any) *awstypes.SpaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceSettings{}

	if v, ok := m["app_type"].(string); ok && len(v) > 0 {
		config.AppType = awstypes.AppType(v)
	}

	if v, ok := m["code_editor_app_settings"].([]any); ok && len(v) > 0 {
		config.CodeEditorAppSettings = expandSpaceCodeEditorAppSettings(v)
	}

	if v, ok := m["custom_file_system"].([]any); ok && len(v) > 0 {
		config.CustomFileSystems = expandCustomFileSystems(v)
	}

	if v, ok := m["jupyter_lab_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterLabAppSettings = expandSpaceJupyterLabAppSettings(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]any); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]any); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["space_storage_settings"].([]any); ok && len(v) > 0 {
		config.SpaceStorageSettings = expandSpaceStorageSettings(v)
	}

	return config
}

func flattenSpaceSettings(config *awstypes.SpaceSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"app_type": config.AppType,
	}

	if config.CodeEditorAppSettings != nil {
		m["code_editor_app_settings"] = flattenSpaceCodeEditorAppSettings(config.CodeEditorAppSettings)
	}

	if config.CustomFileSystems != nil {
		m["custom_file_system"] = flattenCustomFileSystems(config.CustomFileSystems)
	}

	if config.JupyterLabAppSettings != nil {
		m["jupyter_lab_app_settings"] = flattenSpaceJupyterLabAppSettings(config.JupyterLabAppSettings)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.SpaceStorageSettings != nil {
		m["space_storage_settings"] = flattenSpaceStorageSettings(config.SpaceStorageSettings)
	}

	return []map[string]any{m}
}

func expandSpaceCodeEditorAppSettings(l []any) *awstypes.SpaceCodeEditorAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceCodeEditorAppSettings{}

	if v, ok := m["app_lifecycle_management"].([]any); ok && len(v) > 0 {
		config.AppLifecycleManagement = expandSpaceAppLifecycleManagement(v)
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func flattenSpaceCodeEditorAppSettings(config *awstypes.SpaceCodeEditorAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AppLifecycleManagement != nil {
		m["app_lifecycle_management"] = flattenSpaceAppLifecycleManagement(config.AppLifecycleManagement)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]any{m}
}

func expandSpaceJupyterLabAppSettings(l []any) *awstypes.SpaceJupyterLabAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceJupyterLabAppSettings{}

	if v, ok := m["app_lifecycle_management"].([]any); ok && len(v) > 0 {
		config.AppLifecycleManagement = expandSpaceAppLifecycleManagement(v)
	}

	if v, ok := m["code_repository"].(*schema.Set); ok && v.Len() > 0 {
		config.CodeRepositories = expandCodeRepositories(v.List())
	}

	if v, ok := m["default_resource_spec"].([]any); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func flattenSpaceJupyterLabAppSettings(config *awstypes.SpaceJupyterLabAppSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.AppLifecycleManagement != nil {
		m["app_lifecycle_management"] = flattenSpaceAppLifecycleManagement(config.AppLifecycleManagement)
	}

	if config.CodeRepositories != nil {
		m["code_repository"] = flattenCodeRepositories(config.CodeRepositories)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]any{m}
}

func expandSpaceStorageSettings(l []any) *awstypes.SpaceStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceStorageSettings{}

	if v, ok := m["ebs_storage_settings"].([]any); ok && len(v) > 0 {
		config.EbsStorageSettings = expandEBSStorageSettings(v)
	}

	return config
}

func flattenSpaceStorageSettings(config *awstypes.SpaceStorageSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.EbsStorageSettings != nil {
		m["ebs_storage_settings"] = flattenEBSStorageSettings(config.EbsStorageSettings)
	}

	return []map[string]any{m}
}

func expandEBSStorageSettings(l []any) *awstypes.EbsStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.EbsStorageSettings{}

	if v, ok := m["ebs_volume_size_in_gb"].(int); ok {
		config.EbsVolumeSizeInGb = aws.Int32(int32(v))
	}

	return config
}

func flattenEBSStorageSettings(config *awstypes.EbsStorageSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.EbsVolumeSizeInGb != nil {
		m["ebs_volume_size_in_gb"] = aws.ToInt32(config.EbsVolumeSizeInGb)
	}

	return []map[string]any{m}
}

func expandCustomFileSystem(tfMap map[string]any) awstypes.CustomFileSystem {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomFileSystemMemberEFSFileSystem{}

	if v, ok := tfMap["efs_file_system"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandEFSFileSystem(v[0].(map[string]any))
	}

	return apiObject
}

func expandCustomFileSystems(tfList []any) []awstypes.CustomFileSystem {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CustomFileSystem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandCustomFileSystem(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandEFSFileSystem(tfMap map[string]any) awstypes.EFSFileSystem {
	apiObject := awstypes.EFSFileSystem{}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok {
		apiObject.FileSystemId = aws.String(v)
	}

	return apiObject
}

func flattenCustomFileSystem(apiObject awstypes.CustomFileSystem) map[string]any {
	tfMap := map[string]any{}

	if apiObject, ok := apiObject.(*awstypes.CustomFileSystemMemberEFSFileSystem); ok {
		tfMap["efs_file_system"] = flattenEFSFileSystem(apiObject)
	}

	return tfMap
}

func flattenCustomFileSystems(apiObjects []awstypes.CustomFileSystem) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCustomFileSystem(apiObject))
	}

	return tfList
}

func flattenEFSFileSystem(apiObject *awstypes.CustomFileSystemMemberEFSFileSystem) []map[string]any {
	tfMap := map[string]any{}

	if apiObject.Value.FileSystemId != nil {
		tfMap[names.AttrFileSystemID] = aws.ToString(apiObject.Value.FileSystemId)
	}

	return []map[string]any{tfMap}
}

func expandOwnershipSettings(l []any) *awstypes.OwnershipSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OwnershipSettings{}

	if v, ok := m["owner_user_profile_name"].(string); ok {
		config.OwnerUserProfileName = aws.String(v)
	}

	return config
}

func flattenOwnershipSettings(config *awstypes.OwnershipSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.OwnerUserProfileName != nil {
		m["owner_user_profile_name"] = aws.ToString(config.OwnerUserProfileName)
	}

	return []map[string]any{m}
}

func expandSpaceSharingSettings(l []any) *awstypes.SpaceSharingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceSharingSettings{}

	if v, ok := m["sharing_type"].(string); ok {
		config.SharingType = awstypes.SharingType(v)
	}

	return config
}

func flattenSpaceSharingSettings(config *awstypes.SpaceSharingSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"sharing_type": config.SharingType,
	}

	return []map[string]any{m}
}

func expandSpaceAppLifecycleManagement(l []any) *awstypes.SpaceAppLifecycleManagement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceAppLifecycleManagement{}

	if v, ok := m["idle_settings"].([]any); ok && len(v) > 0 {
		config.IdleSettings = expandSpaceIdleSettings(v)
	}

	return config
}

func expandSpaceIdleSettings(l []any) *awstypes.SpaceIdleSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SpaceIdleSettings{}

	if v, ok := m["idle_timeout_in_minutes"].(int); ok {
		config.IdleTimeoutInMinutes = aws.Int32(int32(v))
	}

	return config
}

func flattenSpaceAppLifecycleManagement(config *awstypes.SpaceAppLifecycleManagement) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.IdleSettings != nil {
		m["idle_settings"] = flattenSpaceIdleSettings(config.IdleSettings)
	}

	return []map[string]any{m}
}

func flattenSpaceIdleSettings(config *awstypes.SpaceIdleSettings) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.IdleTimeoutInMinutes != nil {
		m["idle_timeout_in_minutes"] = aws.ToInt32(config.IdleTimeoutInMinutes)
	}

	return []map[string]any{m}
}
