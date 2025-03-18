// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"errors"
	"log"
	"time"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_image", name="Image")
// @Tags(identifierAttribute="id")
func resourceImage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageCreate,
		ReadWithoutTimeout:   resourceImageRead,
		UpdateWithoutTimeout: resourceImageUpdate,
		DeleteWithoutTimeout: resourceImageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"execution_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"image_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ecr_configuration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									names.AttrRepositoryName: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"image_scanning_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  true,
						},
						"timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      720,
							ValidateFunc: validation.IntBetween(60, 1440),
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amis": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAccountID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrDescription: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"image": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"containers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_uris": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workflow": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_failure": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OnWorkflowFailure](),
						},
						"parallel_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-_+#]{0,99}$`), "valid parallel group string must be provider"),
						},
						names.AttrParameter: {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"workflow_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceImageCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateImageInput{
		ClientToken:                  aws.String(id.UniqueId()),
		EnhancedImageMetadataEnabled: aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("container_recipe_arn"); ok {
		input.ContainerRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role"); ok {
		input.ExecutionRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_recipe_arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ImageScanningConfiguration = expandImageScanningConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("image_tests_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ImageTestsConfiguration = expandImageTestConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("infrastructure_configuration_arn"); ok {
		input.InfrastructureConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workflow"); ok && len(v.(*schema.Set).List()) > 0 {
		input.Workflows = expandWorkflowConfigurations(v.(*schema.Set).List())
	}

	output, err := conn.CreateImage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image: %s", err)
	}

	d.SetId(aws.ToString(output.ImageBuildVersionArn))

	if _, err := waitImageStatusAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Image Builder Image (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	image, err := findImageByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Image (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, image.Arn)
	if image.ContainerRecipe != nil {
		d.Set("container_recipe_arn", image.ContainerRecipe.Arn)
	}
	d.Set("date_created", image.DateCreated)
	if image.DistributionConfiguration != nil {
		d.Set("distribution_configuration_arn", image.DistributionConfiguration.Arn)
	}
	d.Set("enhanced_image_metadata_enabled", image.EnhancedImageMetadataEnabled)
	d.Set("execution_role", image.ExecutionRole)
	if image.ImageRecipe != nil {
		d.Set("image_recipe_arn", image.ImageRecipe.Arn)
	}
	if image.ImageScanningConfiguration != nil {
		if err := d.Set("image_scanning_configuration", []any{flattenImageScanningConfiguration(image.ImageScanningConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting image_scanning_configuration: %s", err)
		}
	} else {
		d.Set("image_scanning_configuration", nil)
	}
	if image.ImageTestsConfiguration != nil {
		if err := d.Set("image_tests_configuration", []any{flattenImageTestsConfiguration(image.ImageTestsConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting image_tests_configuration: %s", err)
		}
	} else {
		d.Set("image_tests_configuration", nil)
	}
	if image.InfrastructureConfiguration != nil {
		d.Set("infrastructure_configuration_arn", image.InfrastructureConfiguration.Arn)
	}
	d.Set(names.AttrName, image.Name)
	d.Set("os_version", image.OsVersion)
	if image.OutputResources != nil {
		if err := d.Set("output_resources", []any{flattenOutputResources(image.OutputResources)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting output_resources: %s", err)
		}
	} else {
		d.Set("output_resources", nil)
	}
	d.Set("platform", image.Platform)
	d.Set(names.AttrVersion, image.Version)
	if image.Workflows != nil {
		if err := d.Set("workflow", flattenWorkflowConfigurations(image.Workflows)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting workflow: %s", err)
		}
	} else {
		d.Set("workflow", nil)
	}

	setTagsOut(ctx, image.Tags)

	return diags
}

func resourceImageUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	log.Printf("[DEBUG] Deleting Image Builder Image: %s", d.Id())
	_, err := conn.DeleteImage(ctx, &imagebuilder.DeleteImageInput{
		ImageBuildVersionArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Image (%s): %s", d.Id(), err)
	}

	return diags
}

func findImageByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.Image, error) {
	input := &imagebuilder.GetImageInput{
		ImageBuildVersionArn: aws.String(arn),
	}

	output, err := conn.GetImage(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Image == nil || output.Image.State == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Image, nil
}

func statusImage(ctx context.Context, conn *imagebuilder.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findImageByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, string(awstypes.ImageStatusPending), nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Status), nil
	}
}

func waitImageStatusAvailable(ctx context.Context, conn *imagebuilder.Client, arn string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ImageStatusBuilding,
			awstypes.ImageStatusCreating,
			awstypes.ImageStatusDistributing,
			awstypes.ImageStatusIntegrating,
			awstypes.ImageStatusPending,
			awstypes.ImageStatusTesting,
		),
		Target:  enum.Slice(awstypes.ImageStatusAvailable),
		Refresh: statusImage(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Image); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.State.Reason)))

		return output, err
	}

	return nil, err
}

func flattenOutputResources(apiObject *awstypes.OutputResources) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Amis; v != nil {
		tfMap["amis"] = flattenAMIs(v)
	}

	if v := apiObject.Containers; v != nil {
		tfMap["containers"] = flattenContainers(v)
	}

	return tfMap
}

func flattenAMI(apiObject awstypes.Ami) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AccountId; v != nil {
		tfMap[names.AttrAccountID] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.Image; v != nil {
		tfMap["image"] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenAMIs(apiObjects []awstypes.Ami) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAMI(apiObject))
	}

	return tfList
}

func flattenContainer(apiObject awstypes.Container) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.ImageUris; v != nil {
		tfMap["image_uris"] = aws.StringSlice(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenContainers(apiObjects []awstypes.Container) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenContainer(apiObject))
	}

	return tfList
}
