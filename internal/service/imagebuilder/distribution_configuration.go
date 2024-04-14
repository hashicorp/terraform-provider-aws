// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_distribution_configuration", name="Distribution Configuration")
// @Tags(identifierAttribute="id")
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
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_{-][0-9A-Za-z_\.\s:{}-]+[0-9A-Za-z_}-]$`), "must be a valid output AMI name"),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDistributionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateDistributionConfigurationInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
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

	output, err := conn.CreateDistributionConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Distribution Configuration: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Distribution Configuration: empty response")
	}

	d.SetId(aws.ToString(output.DistributionConfigurationArn))

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetDistributionConfiguration(ctx, input)

	if !d.IsNewResource() && errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
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

	setTagsOut(ctx, distributionConfiguration.Tags)

	return diags
}

func resourceDistributionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

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
		_, err := conn.UpdateDistributionConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Distribution Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteDistributionConfiguration(ctx, input)

	if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Distribution Config (%s): %s", d.Id(), err)
	}

	return diags
}

func expandAMIDistributionConfiguration(tfMap map[string]interface{}) *awstypes.AmiDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AmiDistributionConfiguration{}

	if v, ok := tfMap["ami_tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.AmiTags = flex.ExpandStringValueMap(v)
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
		apiObject.TargetAccountIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandContainerDistributionConfiguration(tfMap map[string]interface{}) *awstypes.ContainerDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContainerDistributionConfiguration{}

	if v, ok := tfMap["container_tags"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ContainerTags = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["target_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.TargetRepository = expandTargetContainerRepository(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLaunchTemplateConfigurations(tfList []interface{}) []awstypes.LaunchTemplateConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDistribution(tfMap map[string]interface{}) *awstypes.Distribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Distribution{}

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
		apiObject.LicenseConfigurationArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["region"].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandDistributions(tfList []interface{}) []awstypes.Distribution {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Distribution

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

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLaunchPermissionConfiguration(tfMap map[string]interface{}) *awstypes.LaunchPermissionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchPermissionConfiguration{}

	if v, ok := tfMap["organization_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.OrganizationArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["organizational_unit_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.OrganizationalUnitArns = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["user_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["user_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserGroups = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandTargetContainerRepository(tfMap map[string]interface{}) *awstypes.TargetContainerRepository {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TargetContainerRepository{}

	if v, ok := tfMap["repository_name"].(string); ok && v != "" {
		apiObject.RepositoryName = aws.String(v)
	}

	if v, ok := tfMap["service"].(string); ok && v != "" {
		apiObject.Service = awstypes.ContainerRepositoryService(v)
	}

	return apiObject
}

func expandFastLaunchConfigurations(tfList []interface{}) []awstypes.FastLaunchConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FastLaunchConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFastLaunchConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFastLaunchConfiguration(tfMap map[string]interface{}) *awstypes.FastLaunchConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FastLaunchConfiguration{}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = v
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandFastLaunchLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_parallel_launches"].(int); ok && v != 0 {
		apiObject.MaxParallelLaunches = aws.Int32(int32(v))
	}

	if v, ok := tfMap["snapshot_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SnapshotConfiguration = expandFastLaunchSnapshotConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFastLaunchLaunchTemplateSpecification(tfMap map[string]interface{}) *awstypes.FastLaunchLaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FastLaunchLaunchTemplateSpecification{}

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

func expandFastLaunchSnapshotConfiguration(tfMap map[string]interface{}) *awstypes.FastLaunchSnapshotConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FastLaunchSnapshotConfiguration{}

	if v, ok := tfMap["target_resource_count"].(int); ok && v != 0 {
		apiObject.TargetResourceCount = aws.Int32(int32(v))
	}

	return apiObject
}

func expandLaunchTemplateConfiguration(tfMap map[string]interface{}) *awstypes.LaunchTemplateConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateConfiguration{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["default"].(bool); ok {
		apiObject.SetDefaultVersion = v
	}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	return apiObject
}

func flattenAMIDistributionConfiguration(apiObject *awstypes.AmiDistributionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmiTags; v != nil {
		tfMap["ami_tags"] = aws.StringMap(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.ToString(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.ToString(v)
	}

	if v := apiObject.LaunchPermission; v != nil {
		tfMap["launch_permission"] = []interface{}{flattenLaunchPermissionConfiguration(v)}
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	if v := apiObject.TargetAccountIds; v != nil {
		tfMap["target_account_ids"] = aws.StringSlice(v)
	}

	return tfMap
}

func flattenContainerDistributionConfiguration(apiObject *awstypes.ContainerDistributionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContainerTags; v != nil {
		tfMap["container_tags"] = aws.StringSlice(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.ToString(v)
	}

	if v := apiObject.TargetRepository; v != nil {
		tfMap["target_repository"] = []interface{}{flattenTargetContainerRepository(v)}
	}

	return tfMap
}

func flattenLaunchTemplateConfigurations(apiObjects []awstypes.LaunchTemplateConfiguration) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateConfiguration(apiObject))
	}

	return tfList
}

func flattenDistribution(apiObject awstypes.Distribution) map[string]interface{} {
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
		tfMap["license_configuration_arns"] = aws.StringSlice(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.ToString(v)
	}

	return tfMap
}

func flattenDistributions(apiObjects []awstypes.Distribution) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDistribution(apiObject))
	}

	return tfList
}

func flattenLaunchPermissionConfiguration(apiObject *awstypes.LaunchPermissionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OrganizationArns; v != nil {
		tfMap["organization_arns"] = aws.StringSlice(v)
	}

	if v := apiObject.OrganizationalUnitArns; v != nil {
		tfMap["organizational_unit_arns"] = aws.StringSlice(v)
	}

	if v := apiObject.UserGroups; v != nil {
		tfMap["user_groups"] = aws.StringSlice(v)
	}

	if v := apiObject.UserIds; v != nil {
		tfMap["user_ids"] = aws.StringSlice(v)
	}

	return tfMap
}

func flattenTargetContainerRepository(apiObject *awstypes.TargetContainerRepository) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RepositoryName; v != nil {
		tfMap["repository_name"] = aws.ToString(v)
	}

	tfMap["service"] = string(apiObject.Service)

	return tfMap
}

func flattenLaunchTemplateConfiguration(apiObject awstypes.LaunchTemplateConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{
		"default": apiObject.SetDefaultVersion,
	}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFastLaunchConfigurations(apiObjects []awstypes.FastLaunchConfiguration) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFastLaunchConfiguration(apiObject))
	}

	return tfList
}

func flattenFastLaunchConfiguration(apiObject awstypes.FastLaunchConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.ToString(v)
	}

	tfMap["enabled"] = aws.Bool(apiObject.Enabled)

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenFastLaunchLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxParallelLaunches; v != nil {
		tfMap["max_parallel_launches"] = aws.ToInt32(v)
	}

	if v := apiObject.SnapshotConfiguration; v != nil {
		tfMap["snapshot_configuration"] = []interface{}{flattenFastLaunchSnapshotConfiguration(v)}
	}

	return tfMap
}

func flattenFastLaunchLaunchTemplateSpecification(apiObject *awstypes.FastLaunchLaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateVersion; v != nil {
		tfMap["launch_template_version"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFastLaunchSnapshotConfiguration(apiObject *awstypes.FastLaunchSnapshotConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TargetResourceCount; v != nil {
		tfMap["target_resource_count"] = aws.ToInt32(v)
	}

	return tfMap
}
