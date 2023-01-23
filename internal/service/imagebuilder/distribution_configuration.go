package imagebuilder

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDistributionConfigurationCreate,
		ReadWithoutTimeout:   resourceDistributionConfigurationRead,
		UpdateWithoutTimeout: resourceDistributionConfigurationUpdate,
		DeleteWithoutTimeout: resourceDistributionConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"distribution": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami_distribution_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ami_tags": tftags.TagsSchema(),
									"description": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"launch_permission": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"organization_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
												"organizational_unit_arns": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
												"user_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 1024),
													},
												},
												"user_ids": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidAccountID,
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 127),
											validation.StringMatch(regexp.MustCompile(`^[-_A-Za-z0-9{][-_A-Za-z0-9\s:{}]+[-_A-Za-z0-9}]$`), "must contain only alphanumeric characters, underscores, and hyphens"),
										),
									},
									"target_account_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidAccountID,
										},
									},
								},
							},
						},
						"container_distribution_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 1024),
										},
									},
									"description": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"target_repository": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
												"service": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"ECR"}, false),
												},
											},
										},
									},
								},
							},
						},
						"fast_launch_configuration": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1000,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"launch_template": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"launch_template_id": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidLaunchTemplateID,
												},
												"launch_template_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidLaunchTemplateName,
												},
												"launch_template_version": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"max_parallel_launches": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									"snapshot_configuration": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_resource_count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 10000),
												},
											},
										},
									},
								},
							},
						},
						"launch_template_configuration": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									"default": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"launch_template_id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidLaunchTemplateID,
									},
								},
							},
						},
						"license_configuration_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
						"region": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDistributionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &imagebuilder.CreateDistributionConfigurationInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
		input.Distributions = expandDistributions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateDistributionConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Distribution Configuration: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Distribution Configuration: empty response")
	}

	d.SetId(aws.StringValue(output.DistributionConfigurationArn))

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetDistributionConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Distribution Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Distribution Configuration (%s): %s", d.Id(), err)
	}

	if output == nil || output.DistributionConfiguration == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Distribution Configuration (%s): empty response", d.Id())
	}

	distributionConfiguration := output.DistributionConfiguration

	d.Set("arn", distributionConfiguration.Arn)
	d.Set("date_created", distributionConfiguration.DateCreated)
	d.Set("date_updated", distributionConfiguration.DateUpdated)
	d.Set("description", distributionConfiguration.Description)
	d.Set("distribution", flattenDistributions(distributionConfiguration.Distributions))
	d.Set("name", distributionConfiguration.Name)
	tags := KeyValueTags(distributionConfiguration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDistributionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()

	if d.HasChanges("description", "distribution") {
		input := &imagebuilder.UpdateDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
			input.Distributions = expandDistributions(v.(*schema.Set).List())
		}

		log.Printf("[DEBUG] UpdateDistributionConfiguration: %#v", input)
		_, err := conn.UpdateDistributionConfigurationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Distribution Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for Image Builder Distribution Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn()

	input := &imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteDistributionConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Distribution Config (%s): %s", d.Id(), err)
	}

	return diags
}

func expandAMIDistributionConfiguration(tfMap map[string]interface{}) *imagebuilder.AmiDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.AmiDistributionConfiguration{}

	if v, ok := tfMap["ami_tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.AmiTags = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["launch_permission"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchPermission = expandLaunchPermissionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["target_account_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TargetAccountIds = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandContainerDistributionConfiguration(tfMap map[string]interface{}) *imagebuilder.ContainerDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.ContainerDistributionConfiguration{}

	if v, ok := tfMap["container_tags"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ContainerTags = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["target_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.TargetRepository = expandTargetContainerRepository(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLaunchTemplateConfigurations(tfList []interface{}) []*imagebuilder.LaunchTemplateConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.LaunchTemplateConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDistribution(tfMap map[string]interface{}) *imagebuilder.Distribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.Distribution{}

	if v, ok := tfMap["ami_distribution_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmiDistributionConfiguration = expandAMIDistributionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["container_distribution_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContainerDistributionConfiguration = expandContainerDistributionConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["fast_launch_configuration"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FastLaunchConfigurations = expandFastLaunchConfigurations(v.List())
	}

	if v, ok := tfMap["launch_template_configuration"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LaunchTemplateConfigurations = expandLaunchTemplateConfigurations(v.List())
	}

	if v, ok := tfMap["license_configuration_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LicenseConfigurationArns = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["region"].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandDistributions(tfList []interface{}) []*imagebuilder.Distribution {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.Distribution

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDistribution(tfMap)

		if apiObject == nil {
			continue
		}

		// Prevent error: InvalidParameter: 1 validation error(s) found.
		//  - missing required field, UpdateDistributionConfigurationInput.Distributions[0].Region
		// Reference: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
		if apiObject.Region == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchPermissionConfiguration(tfMap map[string]interface{}) *imagebuilder.LaunchPermissionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.LaunchPermissionConfiguration{}

	if v, ok := tfMap["organization_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.OrganizationArns = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["organizational_unit_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.OrganizationalUnitArns = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["user_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["user_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserGroups = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandTargetContainerRepository(tfMap map[string]interface{}) *imagebuilder.TargetContainerRepository {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.TargetContainerRepository{}

	if v, ok := tfMap["repository_name"].(string); ok && v != "" {
		apiObject.RepositoryName = aws.String(v)
	}

	if v, ok := tfMap["service"].(string); ok && v != "" {
		apiObject.Service = aws.String(v)
	}

	return apiObject
}

func expandFastLaunchConfigurations(tfList []interface{}) []*imagebuilder.FastLaunchConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.FastLaunchConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFastLaunchConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandFastLaunchConfiguration(tfMap map[string]interface{}) *imagebuilder.FastLaunchConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.FastLaunchConfiguration{}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandFastLaunchLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_parallel_launches"].(int); ok && v != 0 {
		apiObject.MaxParallelLaunches = aws.Int64(int64(v))
	}

	if v, ok := tfMap["snapshot_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SnapshotConfiguration = expandFastLaunchSnapshotConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFastLaunchLaunchTemplateSpecification(tfMap map[string]interface{}) *imagebuilder.FastLaunchLaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.FastLaunchLaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["launch_template_version"].(string); ok && v != "" {
		apiObject.LaunchTemplateVersion = aws.String(v)
	}

	return apiObject
}

func expandFastLaunchSnapshotConfiguration(tfMap map[string]interface{}) *imagebuilder.FastLaunchSnapshotConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.FastLaunchSnapshotConfiguration{}

	if v, ok := tfMap["target_resource_count"].(int); ok && v != 0 {
		apiObject.TargetResourceCount = aws.Int64(int64(v))
	}

	return apiObject
}

func expandLaunchTemplateConfiguration(tfMap map[string]interface{}) *imagebuilder.LaunchTemplateConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.LaunchTemplateConfiguration{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["default"].(bool); ok {
		apiObject.SetDefaultVersion = aws.Bool(v)
	}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	return apiObject
}

func flattenAMIDistributionConfiguration(apiObject *imagebuilder.AmiDistributionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmiTags; v != nil {
		tfMap["ami_tags"] = aws.StringValueMap(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchPermission; v != nil {
		tfMap["launch_permission"] = []interface{}{flattenLaunchPermissionConfiguration(v)}
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.TargetAccountIds; v != nil {
		tfMap["target_account_ids"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenContainerDistributionConfiguration(apiObject *imagebuilder.ContainerDistributionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContainerTags; v != nil {
		tfMap["container_tags"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.TargetRepository; v != nil {
		tfMap["target_repository"] = []interface{}{flattenTargetContainerRepository(v)}
	}

	return tfMap
}

func flattenLaunchTemplateConfigurations(apiObjects []*imagebuilder.LaunchTemplateConfiguration) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateConfiguration(apiObject))
	}

	return tfList
}

func flattenDistribution(apiObject *imagebuilder.Distribution) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmiDistributionConfiguration; v != nil {
		tfMap["ami_distribution_configuration"] = []interface{}{flattenAMIDistributionConfiguration(v)}
	}

	if v := apiObject.ContainerDistributionConfiguration; v != nil {
		tfMap["container_distribution_configuration"] = []interface{}{flattenContainerDistributionConfiguration(v)}
	}

	if v := apiObject.FastLaunchConfigurations; v != nil {
		tfMap["fast_launch_configuration"] = flattenFastLaunchConfigurations(v)
	}

	if v := apiObject.LaunchTemplateConfigurations; v != nil {
		tfMap["launch_template_configuration"] = flattenLaunchTemplateConfigurations(v)
	}

	if v := apiObject.LicenseConfigurationArns; v != nil {
		tfMap["license_configuration_arns"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenDistributions(apiObjects []*imagebuilder.Distribution) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenDistribution(apiObject))
	}

	return tfList
}

func flattenLaunchPermissionConfiguration(apiObject *imagebuilder.LaunchPermissionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OrganizationArns; v != nil {
		tfMap["organization_arns"] = aws.StringValueSlice(v)
	}

	if v := apiObject.OrganizationalUnitArns; v != nil {
		tfMap["organizational_unit_arns"] = aws.StringValueSlice(v)
	}

	if v := apiObject.UserGroups; v != nil {
		tfMap["user_groups"] = aws.StringValueSlice(v)
	}

	if v := apiObject.UserIds; v != nil {
		tfMap["user_ids"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenTargetContainerRepository(apiObject *imagebuilder.TargetContainerRepository) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RepositoryName; v != nil {
		tfMap["repository_name"] = aws.StringValue(v)
	}

	if v := apiObject.Service; v != nil {
		tfMap["service"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateConfiguration(apiObject *imagebuilder.LaunchTemplateConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.SetDefaultVersion; v != nil {
		tfMap["default"] = aws.BoolValue(v)
	}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenFastLaunchConfigurations(apiObjects []*imagebuilder.FastLaunchConfiguration) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenFastLaunchConfiguration(apiObject))
	}

	return tfList
}

func flattenFastLaunchConfiguration(apiObject *imagebuilder.FastLaunchConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.StringValue(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenFastLaunchLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxParallelLaunches; v != nil {
		tfMap["max_parallel_launches"] = aws.Int64Value(v)
	}

	if v := apiObject.SnapshotConfiguration; v != nil {
		tfMap["snapshot_configuration"] = []interface{}{flattenFastLaunchSnapshotConfiguration(v)}
	}

	return tfMap
}

func flattenFastLaunchLaunchTemplateSpecification(apiObject *imagebuilder.FastLaunchLaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateVersion; v != nil {
		tfMap["launch_template_version"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenFastLaunchSnapshotConfiguration(apiObject *imagebuilder.FastLaunchSnapshotConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TargetResourceCount; v != nil {
		tfMap["target_resource_count"] = aws.Int64Value(v)
	}

	return tfMap
}
