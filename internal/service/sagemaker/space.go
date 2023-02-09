package sagemaker

import (
	"context"
	"fmt"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"space_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
										Required: true,
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
			"home_efs_file_system_uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	domainId := d.Get("domain_id").(string)
	spaceName := d.Get("space_name").(string)
	input := &sagemaker.CreateSpaceInput{
		SpaceName: aws.String(spaceName),
		DomainId:  aws.String(domainId),
	}

	if v, ok := d.GetOk("space_settings"); ok && len(v.([]interface{})) > 0 {
		input.SpaceSettings = expandSpaceSettings(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
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
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainID, name, err := decodeSpaceName(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Space (%s): %s", d.Id(), err)
	}

	Space, err := FindSpaceByName(ctx, conn, domainID, name)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Space (%s), removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Space (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(Space.SpaceArn)
	d.Set("space_name", Space.SpaceName)
	d.Set("domain_id", Space.DomainId)
	d.Set("arn", arn)
	d.Set("home_efs_file_system_uid", Space.HomeEfsFileSystemUid)

	if err := d.Set("space_settings", flattenSpaceSettings(Space.SpaceSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting space_settings for SageMaker Space (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Space (%s): %s", d.Id(), err)
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

func resourceSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("space_settings") {
		domainID := d.Get("domain_id").(string)
		name := d.Get("space_name").(string)

		input := &sagemaker.UpdateSpaceInput{
			SpaceName:     aws.String(name),
			DomainId:      aws.String(domainID),
			SpaceSettings: expandSpaceSettings(d.Get("space_settings").([]interface{})),
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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Space (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSpaceRead(ctx, d, meta)...)
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

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

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandDomainKernelGatewayAppSettings(v)
	}

	return config
}

func flattenSpaceSettings(config *sagemaker.SpaceSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	return []map[string]interface{}{m}
}
