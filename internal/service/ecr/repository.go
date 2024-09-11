// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_repository", name="Repository")
// @Tags(identifierAttribute="arn")
func resourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreate,
		ReadWithoutTimeout:   resourceRepositoryRead,
		UpdateWithoutTimeout: resourceRepositoryUpdate,
		DeleteWithoutTimeout: resourceRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

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

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ecr.CreateRepositoryInput{
		EncryptionConfiguration: expandRepositoryEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]interface{})),
		ImageTagMutability:      types.ImageTagMutability((d.Get("image_tag_mutability").(string))),
		RepositoryName:          aws.String(name),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input.ImageScanningConfiguration = &types.ImageScanningConfiguration{
			ScanOnPush: tfMap["scan_on_push"].(bool),
		}
	}

	output, err := conn.CreateRepository(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
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
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			return append(diags, resourceRepositoryRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECR Repository (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findRepositoryByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository (%s): %s", d.Id(), err)
	}

	repository := outputRaw.(*types.Repository)

	d.Set(names.AttrARN, repository.RepositoryArn)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_scanning_configuration: %s", err)
	}
	d.Set("image_tag_mutability", repository.ImageTagMutability)
	d.Set(names.AttrName, repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)

	return diags
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	if d.HasChange("image_tag_mutability") {
		input := &ecr.PutImageTagMutabilityInput{
			ImageTagMutability: types.ImageTagMutability((d.Get("image_tag_mutability").(string))),
			RegistryId:         aws.String(d.Get("registry_id").(string)),
			RepositoryName:     aws.String(d.Id()),
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

		if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})
			input.ImageScanningConfiguration.ScanOnPush = tfMap["scan_on_push"].(bool)
		}

		_, err := conn.PutImageScanningConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECR Repository (%s) image scanning configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
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
		return nil, &retry.NotFoundError{
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

func flattenImageScanningConfiguration(isc *types.ImageScanningConfiguration) []map[string]interface{} {
	if isc == nil {
		return nil
	}

	config := make(map[string]interface{})
	config["scan_on_push"] = isc.ScanOnPush

	return []map[string]interface{}{
		config,
	}
}

func expandRepositoryEncryptionConfiguration(data []interface{}) *types.EncryptionConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]interface{})
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

func flattenRepositoryEncryptionConfiguration(ec *types.EncryptionConfiguration) []map[string]interface{} {
	if ec == nil {
		return nil
	}

	config := map[string]interface{}{
		"encryption_type": ec.EncryptionType,
		names.AttrKMSKey:  aws.ToString(ec.KmsKey),
	}

	return []map[string]interface{}{
		config,
	}
}
