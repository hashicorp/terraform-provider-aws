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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_app_image_config", name="App Image Config")
// @Tags(identifierAttribute="arn")
func resourceAppImageConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppImageConfigCreate,
		ReadWithoutTimeout:   resourceAppImageConfigRead,
		UpdateWithoutTimeout: resourceAppImageConfigUpdate,
		DeleteWithoutTimeout: resourceAppImageConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_image_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"code_editor_app_image_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_arguments": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"container_entrypoint": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"container_environment_variables": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"file_system_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_gid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      100,
										ValidateFunc: validation.IntInSlice([]int{0, 100}),
									},
									"default_uid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1000,
										ValidateFunc: validation.IntInSlice([]int{0, 1000}),
									},
									"mount_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/home/sagemaker-user",
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^\/.*`), "Must start with `/`."),
										),
									},
								},
							},
						},
					},
				},
			},
			"jupyter_lab_image_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_arguments": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"container_entrypoint": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"container_environment_variables": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"file_system_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_gid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      100,
										ValidateFunc: validation.IntInSlice([]int{0, 100}),
									},
									"default_uid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1000,
										ValidateFunc: validation.IntInSlice([]int{0, 1000}),
									},
									"mount_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/home/sagemaker-user",
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^\/.*`), "Must start with `/`."),
										),
									},
								},
							},
						},
					},
				},
			},
			"kernel_gateway_image_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_system_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_gid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      100,
										ValidateFunc: validation.IntInSlice([]int{0, 100}),
									},
									"default_uid": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1000,
										ValidateFunc: validation.IntInSlice([]int{0, 1000}),
									},
									"mount_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/home/sagemaker-user",
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^\/.*`), "Must start with `/`."),
										),
									},
								},
							},
						},
						"kernel_spec": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDisplayName: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAppImageConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("app_image_config_name").(string)
	input := &sagemaker.CreateAppImageConfigInput{
		AppImageConfigName: aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("code_editor_app_image_config"); ok && len(v.([]any)) > 0 {
		input.CodeEditorAppImageConfig = expandCodeEditorAppImageConfig(v.([]any))
	}

	if v, ok := d.GetOk("jupyter_lab_image_config"); ok && len(v.([]any)) > 0 {
		input.JupyterLabAppImageConfig = expandJupyterLabAppImageConfig(v.([]any))
	}

	if v, ok := d.GetOk("kernel_gateway_image_config"); ok && len(v.([]any)) > 0 {
		input.KernelGatewayImageConfig = expandKernelGatewayImageConfig(v.([]any))
	}

	_, err := conn.CreateAppImageConfig(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI App Image Config %s: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceAppImageConfigRead(ctx, d, meta)...)
}

func resourceAppImageConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	image, err := findAppImageConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI App Image Config (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI App Image Config (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(image.AppImageConfigArn)
	d.Set("app_image_config_name", image.AppImageConfigName)
	d.Set(names.AttrARN, arn)

	if err := d.Set("code_editor_app_image_config", flattenCodeEditorAppImageConfig(image.CodeEditorAppImageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting code_editor_app_image_config: %s", err)
	}

	if err := d.Set("kernel_gateway_image_config", flattenKernelGatewayImageConfig(image.KernelGatewayImageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kernel_gateway_image_config: %s", err)
	}

	if err := d.Set("jupyter_lab_image_config", flattenJupyterLabAppImageConfig(image.JupyterLabAppImageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kernel_gateway_image_config: %s", err)
	}

	return diags
}

func resourceAppImageConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateAppImageConfigInput{
			AppImageConfigName: aws.String(d.Id()),
		}

		if d.HasChange("code_editor_app_image_config") {
			if v, ok := d.GetOk("code_editor_app_image_config"); ok && len(v.([]any)) > 0 {
				input.CodeEditorAppImageConfig = expandCodeEditorAppImageConfig(v.([]any))
			}
		}

		if d.HasChange("kernel_gateway_image_config") {
			if v, ok := d.GetOk("kernel_gateway_image_config"); ok && len(v.([]any)) > 0 {
				input.KernelGatewayImageConfig = expandKernelGatewayImageConfig(v.([]any))
			}
		}

		if d.HasChange("jupyter_lab_image_config") {
			if v, ok := d.GetOk("jupyter_lab_image_config"); ok && len(v.([]any)) > 0 {
				input.JupyterLabAppImageConfig = expandJupyterLabAppImageConfig(v.([]any))
			}
		}

		log.Printf("[DEBUG] SageMaker AI App Image Config update config: %#v", *input)
		_, err := conn.UpdateAppImageConfig(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI App Image Config: %s", err)
		}
	}

	return append(diags, resourceAppImageConfigRead(ctx, d, meta)...)
}

func resourceAppImageConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteAppImageConfigInput{
		AppImageConfigName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteAppImageConfig(ctx, input); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI App Image Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findAppImageConfigByName(ctx context.Context, conn *sagemaker.Client, appImageConfigID string) (*sagemaker.DescribeAppImageConfigOutput, error) {
	input := &sagemaker.DescribeAppImageConfigInput{
		AppImageConfigName: aws.String(appImageConfigID),
	}

	output, err := conn.DescribeAppImageConfig(ctx, input)

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

func expandKernelGatewayImageConfig(l []any) *awstypes.KernelGatewayImageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.KernelGatewayImageConfig{}

	if v, ok := m["kernel_spec"].([]any); ok && len(v) > 0 {
		config.KernelSpecs = expandKernelGatewayImageConfigKernelSpecs(v)
	}

	if v, ok := m["file_system_config"].([]any); ok && len(v) > 0 {
		config.FileSystemConfig = expandFileSystemConfig(v)
	}

	return config
}

func expandFileSystemConfig(l []any) *awstypes.FileSystemConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.FileSystemConfig{
		DefaultGid: aws.Int32(int32(m["default_gid"].(int))),
		DefaultUid: aws.Int32(int32(m["default_uid"].(int))),
		MountPath:  aws.String(m["mount_path"].(string)),
	}

	return config
}

func expandKernelGatewayImageConfigKernelSpecs(tfList []any) []awstypes.KernelSpec {
	if len(tfList) == 0 {
		return nil
	}

	var kernelSpecs []awstypes.KernelSpec

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		kernelSpec := awstypes.KernelSpec{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}

		if v, ok := tfMap[names.AttrDisplayName].(string); ok && v != "" {
			kernelSpec.DisplayName = aws.String(v)
		}

		kernelSpecs = append(kernelSpecs, kernelSpec)
	}

	return kernelSpecs
}

func flattenKernelGatewayImageConfig(config *awstypes.KernelGatewayImageConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.KernelSpecs != nil {
		m["kernel_spec"] = flattenKernelGatewayImageConfigKernelSpecs(config.KernelSpecs)
	}

	if config.FileSystemConfig != nil {
		m["file_system_config"] = flattenFileSystemConfig(config.FileSystemConfig)
	}

	return []map[string]any{m}
}

func flattenFileSystemConfig(config *awstypes.FileSystemConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"mount_path":  aws.ToString(config.MountPath),
		"default_gid": aws.ToInt32(config.DefaultGid),
		"default_uid": aws.ToInt32(config.DefaultUid),
	}

	return []map[string]any{m}
}

func flattenKernelGatewayImageConfigKernelSpecs(kernelSpecs []awstypes.KernelSpec) []map[string]any {
	res := make([]map[string]any, 0, len(kernelSpecs))

	for _, raw := range kernelSpecs {
		kernelSpec := make(map[string]any)

		kernelSpec[names.AttrName] = aws.ToString(raw.Name)

		if raw.DisplayName != nil {
			kernelSpec[names.AttrDisplayName] = aws.ToString(raw.DisplayName)
		}

		res = append(res, kernelSpec)
	}

	return res
}

func expandCodeEditorAppImageConfig(l []any) *awstypes.CodeEditorAppImageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CodeEditorAppImageConfig{}

	if v, ok := m["container_config"].([]any); ok && len(v) > 0 {
		config.ContainerConfig = expandContainerConfig(v)
	}

	if v, ok := m["file_system_config"].([]any); ok && len(v) > 0 {
		config.FileSystemConfig = expandFileSystemConfig(v)
	}

	return config
}

func flattenCodeEditorAppImageConfig(config *awstypes.CodeEditorAppImageConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.ContainerConfig != nil {
		m["container_config"] = flattenContainerConfig(config.ContainerConfig)
	}

	if config.FileSystemConfig != nil {
		m["file_system_config"] = flattenFileSystemConfig(config.FileSystemConfig)
	}

	return []map[string]any{m}
}

func expandJupyterLabAppImageConfig(l []any) *awstypes.JupyterLabAppImageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.JupyterLabAppImageConfig{}

	if v, ok := m["container_config"].([]any); ok && len(v) > 0 {
		config.ContainerConfig = expandContainerConfig(v)
	}

	if v, ok := m["file_system_config"].([]any); ok && len(v) > 0 {
		config.FileSystemConfig = expandFileSystemConfig(v)
	}

	return config
}

func flattenJupyterLabAppImageConfig(config *awstypes.JupyterLabAppImageConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.ContainerConfig != nil {
		m["container_config"] = flattenContainerConfig(config.ContainerConfig)
	}

	if config.FileSystemConfig != nil {
		m["file_system_config"] = flattenFileSystemConfig(config.FileSystemConfig)
	}

	return []map[string]any{m}
}

func expandContainerConfig(l []any) *awstypes.ContainerConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.ContainerConfig{}

	if v, ok := m["container_arguments"].([]any); ok && len(v) > 0 {
		config.ContainerArguments = flex.ExpandStringValueList(v)
	}

	if v, ok := m["container_entrypoint"].([]any); ok && len(v) > 0 {
		config.ContainerEntrypoint = flex.ExpandStringValueList(v)
	}

	if v, ok := m["container_environment_variables"].(map[string]any); ok && len(v) > 0 {
		config.ContainerEnvironmentVariables = flex.ExpandStringValueMap(v)
	}

	return config
}

func flattenContainerConfig(config *awstypes.ContainerConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"container_arguments":             flex.FlattenStringValueList(config.ContainerArguments),
		"container_entrypoint":            flex.FlattenStringValueList(config.ContainerEntrypoint),
		"container_environment_variables": flex.FlattenStringValueMap(config.ContainerEnvironmentVariables),
	}

	return []map[string]any{m}
}
