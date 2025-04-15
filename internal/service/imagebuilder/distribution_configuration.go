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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_distribution_configuration", name="Distribution Configuration")
// @Tags(identifierAttribute="id")
func resourceDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDistributionConfigurationCreate,
		ReadWithoutTimeout:   resourceDistributionConfigurationRead,
		UpdateWithoutTimeout: resourceDistributionConfigurationUpdate,
		DeleteWithoutTimeout: resourceDistributionConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrDescription: {
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
									names.AttrDescription: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									names.AttrKMSKeyID: {
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
									names.AttrName: {
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
									names.AttrDescription: {
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
												names.AttrRepositoryName: {
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
									names.AttrAccountID: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Required: true,
									},
									names.AttrLaunchTemplate: {
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
									names.AttrAccountID: {
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
						names.AttrRegion: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"s3_export_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_image_format": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.DiskImageFormat](),
									},
									"role_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									names.AttrS3Bucket: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"s3_prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDistributionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateDistributionConfigurationInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
		input.Distributions = expandDistributions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	output, err := conn.CreateDistributionConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Distribution Configuration: %s", err)
	}

	d.SetId(aws.ToString(output.DistributionConfigurationArn))

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	distributionConfiguration, err := findDistributionConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Distribution Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Distribution Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, distributionConfiguration.Arn)
	d.Set("date_created", distributionConfiguration.DateCreated)
	d.Set("date_updated", distributionConfiguration.DateUpdated)
	d.Set(names.AttrDescription, distributionConfiguration.Description)
	if err := d.Set("distribution", flattenDistributions(distributionConfiguration.Distributions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting distribution: %s", err)
	}
	d.Set(names.AttrName, distributionConfiguration.Name)

	setTagsOut(ctx, distributionConfiguration.Tags)

	return diags
}

func resourceDistributionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	if d.HasChanges(names.AttrDescription, "distribution") {
		input := &imagebuilder.UpdateDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("distribution"); ok && v.(*schema.Set).Len() > 0 {
			input.Distributions = expandDistributions(v.(*schema.Set).List())
		}

		_, err := conn.UpdateDistributionConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Distribution Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionConfigurationRead(ctx, d, meta)...)
}

func resourceDistributionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	log.Printf("[DEBUG] Deleting Image Builder Distribution Configuration: %s", d.Id())
	_, err := conn.DeleteDistributionConfiguration(ctx, &imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Distribution Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findDistributionConfigurationByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.DistributionConfiguration, error) {
	input := &imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(arn),
	}

	output, err := conn.GetDistributionConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DistributionConfiguration, nil
}

func expandAMIDistributionConfiguration(tfMap map[string]any) *awstypes.AmiDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AmiDistributionConfiguration{}

	if v, ok := tfMap["ami_tags"].(map[string]any); ok && len(v) > 0 {
		apiObject.AmiTags = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["launch_permission"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchPermission = expandLaunchPermissionConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["target_account_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TargetAccountIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandContainerDistributionConfiguration(tfMap map[string]any) *awstypes.ContainerDistributionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContainerDistributionConfiguration{}

	if v, ok := tfMap["container_tags"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ContainerTags = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["target_repository"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.TargetRepository = expandTargetContainerRepository(v[0].(map[string]any))
	}

	return apiObject
}

func expandLaunchTemplateConfigurations(tfList []any) []awstypes.LaunchTemplateConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

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

func expandDistribution(tfMap map[string]any) *awstypes.Distribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Distribution{}

	if v, ok := tfMap["ami_distribution_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmiDistributionConfiguration = expandAMIDistributionConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["container_distribution_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContainerDistributionConfiguration = expandContainerDistributionConfiguration(v[0].(map[string]any))
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

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	if v, ok := tfMap["s3_export_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3ExportConfiguration = expandS3ExportConfiguration(v[0].(map[string]any))
	}

	return apiObject
}

func expandDistributions(tfList []any) []awstypes.Distribution {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Distribution

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

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

func expandLaunchPermissionConfiguration(tfMap map[string]any) *awstypes.LaunchPermissionConfiguration {
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

func expandTargetContainerRepository(tfMap map[string]any) *awstypes.TargetContainerRepository {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TargetContainerRepository{}

	if v, ok := tfMap[names.AttrRepositoryName].(string); ok && v != "" {
		apiObject.RepositoryName = aws.String(v)
	}

	if v, ok := tfMap["service"].(string); ok && v != "" {
		apiObject.Service = awstypes.ContainerRepositoryService(v)
	}

	return apiObject
}

func expandFastLaunchConfigurations(tfList []any) []awstypes.FastLaunchConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FastLaunchConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

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

func expandFastLaunchConfiguration(tfMap map[string]any) *awstypes.FastLaunchConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FastLaunchConfiguration{}

	if v, ok := tfMap[names.AttrAccountID].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	if v, ok := tfMap[names.AttrLaunchTemplate].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.LaunchTemplate = expandFastLaunchLaunchTemplateSpecification(v[0].(map[string]any))
	}

	if v, ok := tfMap["max_parallel_launches"].(int); ok && v != 0 {
		apiObject.MaxParallelLaunches = aws.Int32(int32(v))
	}

	if v, ok := tfMap["snapshot_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.SnapshotConfiguration = expandFastLaunchSnapshotConfiguration(v[0].(map[string]any))
	}

	return apiObject
}

func expandFastLaunchLaunchTemplateSpecification(tfMap map[string]any) *awstypes.FastLaunchLaunchTemplateSpecification {
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

func expandFastLaunchSnapshotConfiguration(tfMap map[string]any) *awstypes.FastLaunchSnapshotConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FastLaunchSnapshotConfiguration{}

	if v, ok := tfMap["target_resource_count"].(int); ok && v != 0 {
		apiObject.TargetResourceCount = aws.Int32(int32(v))
	}

	return apiObject
}

func expandLaunchTemplateConfiguration(tfMap map[string]any) *awstypes.LaunchTemplateConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateConfiguration{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["default"].(bool); ok {
		apiObject.SetDefaultVersion = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrAccountID].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	return apiObject
}

func expandS3ExportConfiguration(tfMap map[string]any) *awstypes.S3ExportConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3ExportConfiguration{}

	if v, ok := tfMap["disk_image_format"].(string); ok && v != "" {
		apiObject.DiskImageFormat = awstypes.DiskImageFormat(v)
	}

	if v, ok := tfMap["role_name"].(string); ok && v != "" {
		apiObject.RoleName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrS3Bucket].(string); ok && v != "" {
		apiObject.S3Bucket = aws.String(v)
	}

	if v, ok := tfMap["s3_prefix"].(string); ok && v != "" {
		apiObject.S3Prefix = aws.String(v)
	}

	return apiObject
}

func flattenAMIDistributionConfiguration(apiObject *awstypes.AmiDistributionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AmiTags; v != nil {
		tfMap["ami_tags"] = aws.StringMap(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}

	if v := apiObject.LaunchPermission; v != nil {
		tfMap["launch_permission"] = []any{flattenLaunchPermissionConfiguration(v)}
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.TargetAccountIds; v != nil {
		tfMap["target_account_ids"] = aws.StringSlice(v)
	}

	return tfMap
}

func flattenContainerDistributionConfiguration(apiObject *awstypes.ContainerDistributionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ContainerTags; v != nil {
		tfMap["container_tags"] = aws.StringSlice(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.TargetRepository; v != nil {
		tfMap["target_repository"] = []any{flattenTargetContainerRepository(v)}
	}

	return tfMap
}

func flattenLaunchTemplateConfigurations(apiObjects []awstypes.LaunchTemplateConfiguration) []any {
	if apiObjects == nil {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateConfiguration(apiObject))
	}

	return tfList
}

func flattenDistribution(apiObject awstypes.Distribution) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AmiDistributionConfiguration; v != nil {
		tfMap["ami_distribution_configuration"] = []any{flattenAMIDistributionConfiguration(v)}
	}

	if v := apiObject.ContainerDistributionConfiguration; v != nil {
		tfMap["container_distribution_configuration"] = []any{flattenContainerDistributionConfiguration(v)}
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
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObject.S3ExportConfiguration; v != nil {
		tfMap["s3_export_configuration"] = []any{flattenS3ExportConfiguration(v)}
	}

	return tfMap
}

func flattenDistributions(apiObjects []awstypes.Distribution) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenDistribution(apiObject))
	}

	return tfList
}

func flattenLaunchPermissionConfiguration(apiObject *awstypes.LaunchPermissionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

func flattenTargetContainerRepository(apiObject *awstypes.TargetContainerRepository) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.RepositoryName; v != nil {
		tfMap[names.AttrRepositoryName] = aws.ToString(v)
	}

	tfMap["service"] = string(apiObject.Service)

	return tfMap
}

func flattenLaunchTemplateConfiguration(apiObject awstypes.LaunchTemplateConfiguration) map[string]any {
	tfMap := map[string]any{
		"default": apiObject.SetDefaultVersion,
	}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.AccountId; v != nil {
		tfMap[names.AttrAccountID] = aws.ToString(v)
	}

	return tfMap
}

func flattenFastLaunchConfigurations(apiObjects []awstypes.FastLaunchConfiguration) []any {
	if apiObjects == nil {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFastLaunchConfiguration(apiObject))
	}

	return tfList
}

func flattenFastLaunchConfiguration(apiObject awstypes.FastLaunchConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AccountId; v != nil {
		tfMap[names.AttrAccountID] = aws.ToString(v)
	}

	tfMap[names.AttrEnabled] = aws.Bool(apiObject.Enabled)

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap[names.AttrLaunchTemplate] = []any{flattenFastLaunchLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxParallelLaunches; v != nil {
		tfMap["max_parallel_launches"] = aws.ToInt32(v)
	}

	if v := apiObject.SnapshotConfiguration; v != nil {
		tfMap["snapshot_configuration"] = []any{flattenFastLaunchSnapshotConfiguration(v)}
	}

	return tfMap
}

func flattenFastLaunchLaunchTemplateSpecification(apiObject *awstypes.FastLaunchLaunchTemplateSpecification) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

func flattenFastLaunchSnapshotConfiguration(apiObject *awstypes.FastLaunchSnapshotConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.TargetResourceCount; v != nil {
		tfMap["target_resource_count"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenS3ExportConfiguration(apiObject *awstypes.S3ExportConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"disk_image_format": apiObject.DiskImageFormat,
	}

	if v := apiObject.RoleName; v != nil {
		tfMap["role_name"] = aws.ToString(v)
	}

	if v := apiObject.S3Bucket; v != nil {
		tfMap[names.AttrS3Bucket] = aws.ToString(v)
	}

	if v := apiObject.S3Prefix; v != nil {
		tfMap["s3_prefix"] = aws.ToString(v)
	}

	return tfMap
}
