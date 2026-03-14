// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecr

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_repository", name="Repository")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ecr/types;types.Repository")
// @Testing(preIdentityVersion="v6.10.0")
// @Testing(idAttrDuplicates="name")
func resourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreate,
		ReadWithoutTimeout:   resourceRepositoryRead,
		UpdateWithoutTimeout: resourceRepositoryUpdate,
		DeleteWithoutTimeout: resourceRepositoryDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		CustomizeDiff: validateImageTagMutabilityExclusionFilterUsage,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          types.EncryptionTypeAes256,
							ValidateDiagFunc: enum.Validate[types.EncryptionType](),
						},
						names.AttrKMSKey: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				ForceNew:         true,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"image_tag_mutability": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ImageTagMutabilityMutable,
				ValidateDiagFunc: enum.Validate[types.ImageTagMutability](),
			},
			"image_tag_mutability_exclusion_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrFilter: {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.AllDiag(
								validation.ToDiagFunc(validation.StringLenBetween(1, 128)),
								validation.ToDiagFunc(validation.StringMatch(
									regexache.MustCompile(`^[a-zA-Z0-9._*-]+$`),
									"must contain only letters, numbers, and special characters (._*-)",
								)),
								validateImageTagMutabilityExclusionFilter(),
							),
						},
						"filter_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ImageTagMutabilityExclusionFilterType](),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ecr.CreateRepositoryInput{
		EncryptionConfiguration: expandRepositoryEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]any)),
		ImageTagMutability:      types.ImageTagMutability((d.Get("image_tag_mutability").(string))),
		RepositoryName:          aws.String(name),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)
		input.ImageScanningConfiguration = &types.ImageScanningConfiguration{
			ScanOnPush: tfMap["scan_on_push"].(bool),
		}
	}

	if v, ok := d.GetOk("image_tag_mutability_exclusion_filter"); ok && len(v.([]any)) > 0 {
		input.ImageTagMutabilityExclusionFilters = expandImageTagMutabilityExclusionFilters(v.([]any))
	}

	output, err := conn.CreateRepository(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
		input.Tags = nil

		output, err = conn.CreateRepository(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Repository (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Repository.RepositoryName))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.ToString(output.Repository.RepositoryArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
			return append(diags, resourceRepositoryRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECR Repository (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	repository, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*types.Repository, error) {
		return findRepositoryByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ECR Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository (%s): %s", d.Id(), err)
	}

	resourceRepositoryFlatten(ctx, d, repository)

	return diags
}

func resourceRepositoryFlatten(_ context.Context, d *schema.ResourceData, repository *types.Repository) {
	d.Set(names.AttrARN, repository.RepositoryArn)
	d.Set(names.AttrEncryptionConfiguration, flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration))
	d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration))
	d.Set("image_tag_mutability", repository.ImageTagMutability)
	d.Set("image_tag_mutability_exclusion_filter", flattenImageTagMutabilityExclusionFilters(repository.ImageTagMutabilityExclusionFilters))
	d.Set(names.AttrName, repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	if d.HasChanges("image_tag_mutability", "image_tag_mutability_exclusion_filter") {
		input := &ecr.PutImageTagMutabilityInput{
			ImageTagMutability: types.ImageTagMutability((d.Get("image_tag_mutability").(string))),
			RegistryId:         aws.String(d.Get("registry_id").(string)),
			RepositoryName:     aws.String(d.Id()),
		}

		if v, ok := d.GetOk("image_tag_mutability_exclusion_filter"); ok && len(v.([]any)) > 0 {
			input.ImageTagMutabilityExclusionFilters = expandImageTagMutabilityExclusionFilters(v.([]any))
		}

		_, err := conn.PutImageTagMutability(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECR Repository (%s) image tag mutability: %s", d.Id(), err)
		}
	}

	if d.HasChange("image_scanning_configuration") {
		input := &ecr.PutImageScanningConfigurationInput{
			ImageScanningConfiguration: &types.ImageScanningConfiguration{},
			RegistryId:                 aws.String(d.Get("registry_id").(string)),
			RepositoryName:             aws.String(d.Id()),
		}

		if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)
			input.ImageScanningConfiguration.ScanOnPush = tfMap["scan_on_push"].(bool)
		}

		_, err := conn.PutImageScanningConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECR Repository (%s) image scanning configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Repository: %s", d.Id())
	_, err := conn.DeleteRepository(ctx, &ecr.DeleteRepositoryInput{
		Force:          d.Get(names.AttrForceDelete).(bool),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		RepositoryName: aws.String(d.Id()),
	})

	if errs.IsA[*types.RepositoryNotFoundException](err) {
		return diags
	} else if errs.IsA[*types.RepositoryNotEmptyException](err) {
		return sdkdiag.AppendErrorf(diags, "ECR Repository (%s) not empty, consider using force_delete: %s", d.Id(), err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Repository (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (any, error) {
		return findRepositoryByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECR Repository (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findRepositoryByName(ctx context.Context, conn *ecr.Client, name string) (*types.Repository, error) {
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{name},
	}

	output, err := findRepository(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.RepositoryName) != name {
		return nil, &sdkretry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findRepository(ctx context.Context, conn *ecr.Client, input *ecr.DescribeRepositoriesInput) (*types.Repository, error) {
	output, err := findRepositories(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func flattenImageScanningConfiguration(isc *types.ImageScanningConfiguration) []map[string]any {
	if isc == nil {
		return nil
	}

	config := make(map[string]any)
	config["scan_on_push"] = isc.ScanOnPush

	return []map[string]any{
		config,
	}
}

func expandRepositoryEncryptionConfiguration(data []any) *types.EncryptionConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]any)
	config := &types.EncryptionConfiguration{
		EncryptionType: types.EncryptionType((ec["encryption_type"].(string))),
	}
	if v, ok := ec[names.AttrKMSKey]; ok {
		if s := v.(string); s != "" {
			config.KmsKey = aws.String(v.(string))
		}
	}
	return config
}

func flattenRepositoryEncryptionConfiguration(ec *types.EncryptionConfiguration) []map[string]any {
	if ec == nil {
		return nil
	}

	config := map[string]any{
		"encryption_type": ec.EncryptionType,
		names.AttrKMSKey:  aws.ToString(ec.KmsKey),
	}

	return []map[string]any{
		config,
	}
}

func expandImageTagMutabilityExclusionFilters(data []any) []types.ImageTagMutabilityExclusionFilter {
	if len(data) == 0 {
		return nil
	}

	var filters []types.ImageTagMutabilityExclusionFilter
	for _, v := range data {
		tfMap := v.(map[string]any)
		filter := types.ImageTagMutabilityExclusionFilter{
			Filter:     aws.String(tfMap[names.AttrFilter].(string)),
			FilterType: types.ImageTagMutabilityExclusionFilterType(tfMap["filter_type"].(string)),
		}
		filters = append(filters, filter)
	}

	return filters
}

func flattenImageTagMutabilityExclusionFilters(filters []types.ImageTagMutabilityExclusionFilter) []any {
	if len(filters) == 0 {
		return nil
	}

	var tfList []any
	for _, filter := range filters {
		tfMap := map[string]any{
			names.AttrFilter: aws.ToString(filter.Filter),
			"filter_type":    string(filter.FilterType),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func validateImageTagMutabilityExclusionFilter() schema.SchemaValidateDiagFunc {
	return func(v any, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		value := v.(string)

		wildcardCount := strings.Count(value, "*")
		if wildcardCount > 2 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid filter pattern",
				Detail:   "Image tag mutability exclusion filter can contain a maximum of 2 wildcards (*)",
			})
		}

		return diags
	}
}

func validateImageTagMutabilityExclusionFilterUsage(_ context.Context, d *schema.ResourceDiff, meta any) error {
	mutability := d.Get("image_tag_mutability").(string)
	filters := d.Get("image_tag_mutability_exclusion_filter").([]any)

	if len(filters) > 0 && mutability != string(types.ImageTagMutabilityImmutableWithExclusion) && mutability != string(types.ImageTagMutabilityMutableWithExclusion) {
		return fmt.Errorf("image_tag_mutability_exclusion_filter can only be used when image_tag_mutability is set to IMMUTABLE_WITH_EXCLUSION or MUTABLE_WITH_EXCLUSION")
	}

	return nil
}
