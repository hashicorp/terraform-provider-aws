package sagemaker

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Update: resourceDomainUpdate,
		Delete: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"auth_mode": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AuthMode_Values(), false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 16,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"app_network_access_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      sagemaker.AppNetworkAccessTypePublicInternetOnly,
				ValidateFunc: validation.StringInSlice(sagemaker.AppNetworkAccessType_Values(), false),
			},
			"default_user_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sharing_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
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
							ForceNew: true,
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
						"jupyter_server_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
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
								},
							},
						},
						"kernel_gateway_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
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
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"single_sign_on_managed_application_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_efs_file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] sagemaker domain create config: %#v", *input)
	output, err := conn.CreateDomain(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker domain: %w", err)
	}

	domainArn := aws.StringValue(output.DomainArn)
	domainID, err := DecodeDomainID(domainArn)
	if err != nil {
		return err
	}

	d.SetId(domainID)

	if _, err := WaitDomainInService(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker domain (%s) to create: %w", d.Id(), err)
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domain, err := FindDomainByName(conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker domain (%s), removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker domain (%s): %w", d.Id(), err)
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

	if err := d.Set("subnet_ids", flex.FlattenStringSet(domain.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids for SageMaker domain (%s): %w", d.Id(), err)
	}

	if err := d.Set("default_user_settings", flattenDomainDefaultUserSettings(domain.DefaultUserSettings)); err != nil {
		return fmt.Errorf("error setting default_user_settings for SageMaker domain (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Domain (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("default_user_settings") {
		input := &sagemaker.UpdateDomainInput{
			DomainId:            aws.String(d.Id()),
			DefaultUserSettings: expandDomainDefaultUserSettings(d.Get("default_user_settings").([]interface{})),
		}

		log.Printf("[DEBUG] sagemaker domain update config: %#v", *input)
		_, err := conn.UpdateDomain(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker domain: %w", err)
		}

		if _, err := WaitDomainInService(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for SageMaker domain (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker domain (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteDomainInput{
		DomainId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("retention_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetentionPolicy = expandRetentionPolicy(v.([]interface{}))
	}

	if _, err := conn.DeleteDomain(input); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return fmt.Errorf("error deleting SageMaker domain (%s): %w", d.Id(), err)
		}
	}

	if _, err := WaitDomainDeleted(conn, d.Id()); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return fmt.Errorf("error waiting for SageMaker domain (%s) to delete: %w", d.Id(), err)
		}
	}

	return nil
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

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = flex.ExpandStringSet(v)
	}

	if v, ok := m["tensor_board_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.TensorBoardAppSettings = expandDomainTensorBoardAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["sharing_settings"].([]interface{}); ok && len(v) > 0 {
		config.SharingSettings = expandDomainShareSettings(v)
	}

	return config
}

func expandDomainJupyterServerAppSettings(l []interface{}) *sagemaker.JupyterServerAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.JupyterServerAppSettings{}

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

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.StringValue(config.ExecutionRole)
	}

	if config.SecurityGroups != nil {
		m["security_groups"] = flex.FlattenStringSet(config.SecurityGroups)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.TensorBoardAppSettings != nil {
		m["tensor_board_app_settings"] = flattenDomainTensorBoardAppSettings(config.TensorBoardAppSettings)
	}

	if config.SharingSettings != nil {
		m["sharing_settings"] = flattenDomainShareSettings(config.SharingSettings)
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
