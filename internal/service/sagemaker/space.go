// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_space", name="Space")
// @Tags(identifierAttribute="arn")
func ResourceSpace() *schema.Resource {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.AppType_Values(), false),
						},
						"code_editor_app_settings": {
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.SharingType_Values(), false),
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	domainId := d.Get("domain_id").(string)
	spaceName := d.Get("space_name").(string)
	input := &sagemaker.CreateSpaceInput{
		SpaceName: aws.String(spaceName),
		DomainId:  aws.String(domainId),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ownership_settings"); ok && len(v.([]interface{})) > 0 {
		input.OwnershipSettings = expandOwnershipSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("space_settings"); ok && len(v.([]interface{})) > 0 {
		input.SpaceSettings = expandSpaceSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("space_sharing_settings"); ok && len(v.([]interface{})) > 0 {
		input.SpaceSharingSettings = expandSpaceSharingSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("space_display_name"); ok {
		input.SpaceDisplayName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] SageMaker Space create config: %#v", *input)
	out, err := conn.CreateSpaceWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Space: %s", err)
	}

	d.SetId(aws.StringValue(out.SpaceArn))

	if _, err := WaitSpaceInService(ctx, conn, domainId, spaceName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Space (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceSpaceRead(ctx, d, meta)...)
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	domainID, name, err := decodeSpaceName(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Space (%s): %s", d.Id(), err)
	}

	space, err := FindSpaceByName(ctx, conn, domainID, name)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Space (%s), removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Space (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(space.SpaceArn)
	d.Set(names.AttrARN, arn)
	d.Set("domain_id", space.DomainId)
	d.Set("home_efs_file_system_uid", space.HomeEfsFileSystemUid)
	d.Set("space_display_name", space.SpaceDisplayName)
	d.Set("space_name", space.SpaceName)
	d.Set(names.AttrURL, space.Url)

	if err := d.Set("ownership_settings", flattenOwnershipSettings(space.OwnershipSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ownership_settings for SageMaker Space (%s): %s", d.Id(), err)
	}

	if err := d.Set("space_settings", flattenSpaceSettings(space.SpaceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting space_settings for SageMaker Space (%s): %s", d.Id(), err)
	}

	if err := d.Set("space_sharing_settings", flattenSpaceSharingSettings(space.SpaceSharingSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting space_sharing_settings for SageMaker Space (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		domainID := d.Get("domain_id").(string)
		name := d.Get("space_name").(string)

		input := &sagemaker.UpdateSpaceInput{
			SpaceName: aws.String(name),
			DomainId:  aws.String(domainID),
		}

		if d.HasChanges("space_settings") {
			input.SpaceSettings = expandSpaceSettings(d.Get("space_settings").([]interface{}))
		}

		if d.HasChange("space_display_name") {
			input.SpaceDisplayName = aws.String(d.Get("space_display_name").(string))
		}

		log.Printf("[DEBUG] SageMaker Space update config: %#v", *input)
		_, err := conn.UpdateSpaceWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Space: %s", err)
		}

		if _, err := WaitSpaceInService(ctx, conn, domainID, name); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Space (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSpaceRead(ctx, d, meta)...)
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("space_name").(string)
	domainID := d.Get("domain_id").(string)

	input := &sagemaker.DeleteSpaceInput{
		SpaceName: aws.String(name),
		DomainId:  aws.String(domainID),
	}

	if _, err := conn.DeleteSpaceWithContext(ctx, input); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker Space (%s): %s", d.Id(), err)
		}
	}

	if _, err := WaitSpaceDeleted(ctx, conn, domainID, name); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Space (%s) to delete: %s", d.Id(), err)
		}
	}

	return diags
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

func expandSpaceSettings(l []interface{}) *sagemaker.SpaceSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SpaceSettings{}

	if v, ok := m["app_type"].(string); ok && len(v) > 0 {
		config.AppType = aws.String(v)
	}

	if v, ok := m["code_editor_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.CodeEditorAppSettings = expandSpaceCodeEditorAppSettings(v)
	}

	if v, ok := m["custom_file_system"].([]interface{}); ok && len(v) > 0 {
		config.CustomFileSystems = expandCustomFileSystems(v)
	}

	if v, ok := m["jupyter_lab_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterLabAppSettings = expandSpaceJupyterLabAppSettings(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["space_storage_settings"].([]interface{}); ok && len(v) > 0 {
		config.SpaceStorageSettings = expandSpaceStorageSettings(v)
	}

	return config
}

func flattenSpaceSettings(config *sagemaker.SpaceSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.AppType != nil {
		m["app_type"] = aws.StringValue(config.AppType)
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

	return []map[string]interface{}{m}
}

func expandSpaceCodeEditorAppSettings(l []interface{}) *sagemaker.SpaceCodeEditorAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SpaceCodeEditorAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func flattenSpaceCodeEditorAppSettings(config *sagemaker.SpaceCodeEditorAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]interface{}{m}
}

func expandSpaceJupyterLabAppSettings(l []interface{}) *sagemaker.SpaceJupyterLabAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SpaceJupyterLabAppSettings{}

	if v, ok := m["code_repository"].(*schema.Set); ok && v.Len() > 0 {
		config.CodeRepositories = expandCodeRepositories(v.List())
	}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandResourceSpec(v)
	}

	return config
}

func flattenSpaceJupyterLabAppSettings(config *sagemaker.SpaceJupyterLabAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.CodeRepositories != nil {
		m["code_repository"] = flattenCodeRepositories(config.CodeRepositories)
	}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]interface{}{m}
}

func expandSpaceStorageSettings(l []interface{}) *sagemaker.SpaceStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SpaceStorageSettings{}

	if v, ok := m["ebs_storage_settings"].([]interface{}); ok && len(v) > 0 {
		config.EbsStorageSettings = expandEBSStorageSettings(v)
	}

	return config
}

func flattenSpaceStorageSettings(config *sagemaker.SpaceStorageSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.EbsStorageSettings != nil {
		m["ebs_storage_settings"] = flattenEBSStorageSettings(config.EbsStorageSettings)
	}

	return []map[string]interface{}{m}
}

func expandEBSStorageSettings(l []interface{}) *sagemaker.EbsStorageSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.EbsStorageSettings{}

	if v, ok := m["ebs_volume_size_in_gb"].(int); ok {
		config.EbsVolumeSizeInGb = aws.Int64(int64(v))
	}

	return config
}

func flattenEBSStorageSettings(config *sagemaker.EbsStorageSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.EbsVolumeSizeInGb != nil {
		m["ebs_volume_size_in_gb"] = aws.Int64Value(config.EbsVolumeSizeInGb)
	}

	return []map[string]interface{}{m}
}

func expandCustomFileSystem(tfMap map[string]interface{}) *sagemaker.CustomFileSystem {
	if tfMap == nil {
		return nil
	}

	apiObject := &sagemaker.CustomFileSystem{}

	if v, ok := tfMap["efs_file_system"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EFSFileSystem = expandEFSFileSystem(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCustomFileSystems(tfList []interface{}) []*sagemaker.CustomFileSystem {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*sagemaker.CustomFileSystem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

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

func expandEFSFileSystem(tfMap map[string]interface{}) *sagemaker.EFSFileSystem {
	if tfMap == nil {
		return nil
	}

	apiObject := &sagemaker.EFSFileSystem{}

	if v, ok := tfMap[names.AttrFileSystemID].(string); ok {
		apiObject.FileSystemId = aws.String(v)
	}

	return apiObject
}

func flattenCustomFileSystem(apiObject *sagemaker.CustomFileSystem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.EFSFileSystem != nil {
		tfMap["efs_file_system"] = flattenEFSFileSystem(apiObject.EFSFileSystem)
	}

	return tfMap
}

func flattenCustomFileSystems(apiObjects []*sagemaker.CustomFileSystem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCustomFileSystem(apiObject))
	}

	return tfList
}

func flattenEFSFileSystem(apiObject *sagemaker.EFSFileSystem) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FileSystemId != nil {
		tfMap[names.AttrFileSystemID] = aws.StringValue(apiObject.FileSystemId)
	}

	return []map[string]interface{}{tfMap}
}

func expandOwnershipSettings(l []interface{}) *sagemaker.OwnershipSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OwnershipSettings{}

	if v, ok := m["owner_user_profile_name"].(string); ok {
		config.OwnerUserProfileName = aws.String(v)
	}

	return config
}

func flattenOwnershipSettings(config *sagemaker.OwnershipSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.OwnerUserProfileName != nil {
		m["owner_user_profile_name"] = aws.StringValue(config.OwnerUserProfileName)
	}

	return []map[string]interface{}{m}
}

func expandSpaceSharingSettings(l []interface{}) *sagemaker.SpaceSharingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SpaceSharingSettings{}

	if v, ok := m["sharing_type"].(string); ok {
		config.SharingType = aws.String(v)
	}

	return config
}

func flattenSpaceSharingSettings(config *sagemaker.SpaceSharingSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.SharingType != nil {
		m["sharing_type"] = aws.StringValue(config.SharingType)
	}

	return []map[string]interface{}{m}
}
