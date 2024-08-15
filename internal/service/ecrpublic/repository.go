// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecrpublic_repository", name="Repository")
// @Tags(identifierAttribute="arn")
func ResourceRepository() *schema.Resource {
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
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 205),
					validation.StringMatch(regexache.MustCompile(`(?:[0-9a-z]+(?:[._-][0-9a-z]+)*/)*[0-9a-z]+(?:[._-][0-9a-z]+)*`), "see: https://docs.aws.amazon.com/AmazonECRPublic/latest/APIReference/API_CreateRepository.html#API_CreateRepository_RequestSyntax"),
				),
			},
			"catalog_data": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"about_text": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 10240),
						},
						"architectures": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrDescription: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"logo_image_blob": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"operating_systems": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"usage_text": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 10240),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_uri": {
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
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	input := ecrpublic.CreateRepositoryInput{
		RepositoryName: aws.String(d.Get(names.AttrRepositoryName).(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("catalog_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CatalogData = expandRepositoryCatalogData(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateRepository(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Public repository: %s", err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Public Repository: empty response")
	}

	repository := out.Repository

	log.Printf("[DEBUG] ECR Public repository created: %q", aws.ToString(repository.RepositoryArn))

	d.SetId(aws.ToString(repository.RepositoryName))

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	log.Printf("[DEBUG] Reading ECR Public repository %s", d.Id())
	var out *ecrpublic.DescribeRepositoriesOutput
	input := &ecrpublic.DescribeRepositoriesInput{
		RepositoryNames: []string{d.Id()},
	}

	var err error
	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		out, err = conn.DescribeRepositories(ctx, input)
		if d.IsNewResource() && errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.DescribeRepositories(ctx, input)
	}

	if !d.IsNewResource() && errs.IsA[*awstypes.RepositoryNotFoundException](err) {
		log.Printf("[WARN] ECR Public Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Public repository: %s", err)
	}

	if out == nil || len(out.Repositories) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading ECR Public Repository (%s): empty response", d.Id())
	}

	repository := out.Repositories[0]

	d.Set(names.AttrRepositoryName, d.Id())
	d.Set("registry_id", repository.RegistryId)
	d.Set(names.AttrARN, repository.RepositoryArn)
	d.Set("repository_uri", repository.RepositoryUri)

	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		d.Set(names.AttrForceDestroy, v.(bool))
	} else {
		d.Set(names.AttrForceDestroy, false)
	}

	var catalogOut *ecrpublic.GetRepositoryCatalogDataOutput
	catalogInput := &ecrpublic.GetRepositoryCatalogDataInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     repository.RegistryId,
	}

	catalogOut, err = conn.GetRepositoryCatalogData(ctx, catalogInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading catalog data for ECR Public repository: %s", err)
	}

	if catalogOut != nil {
		flatCatalogData := flattenRepositoryCatalogData(catalogOut)
		if catalogData, ok := d.GetOk("catalog_data"); ok && len(catalogData.([]interface{})) > 0 && catalogData.([]interface{})[0] != nil {
			catalogDataMap := catalogData.([]interface{})[0].(map[string]interface{})
			if v, ok := catalogDataMap["logo_image_blob"].(string); ok && len(v) > 0 {
				flatCatalogData["logo_image_blob"] = v
			}
		}
		d.Set("catalog_data", []interface{}{flatCatalogData})
	} else {
		d.Set("catalog_data", nil)
	}

	return diags
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	deleteInput := &ecrpublic.DeleteRepositoryInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		force := v.(bool)
		deleteInput.Force = aws.ToBool(&force)
	}

	log.Printf("[DEBUG] Deleting ECR Public Repository: (%s)", d.Id())
	_, err := conn.DeleteRepository(ctx, deleteInput)

	if err != nil {
		if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECR Public repository: %s", err)
	}

	log.Printf("[DEBUG] Waiting for ECR Public Repository %q to be deleted", d.Id())
	input := &ecrpublic.DescribeRepositoriesInput{
		RepositoryNames: []string{d.Id()},
	}
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err = conn.DescribeRepositories(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}

		return retry.RetryableError(fmt.Errorf("%q: Timeout while waiting for the ECR Public Repository to be deleted", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DescribeRepositories(ctx, input)
	}

	if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Public repository: %s", err)
	}

	log.Printf("[DEBUG] repository %q deleted.", d.Get(names.AttrRepositoryName).(string))

	return diags
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	if d.HasChange("catalog_data") {
		if err := resourceRepositoryUpdateCatalogData(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECR Public Repository (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func flattenRepositoryCatalogData(apiObject *ecrpublic.GetRepositoryCatalogDataOutput) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	catalogData := apiObject.CatalogData

	tfMap := map[string]interface{}{}

	if v := catalogData.AboutText; v != nil {
		tfMap["about_text"] = aws.ToString(v)
	}

	if v := catalogData.Architectures; v != nil {
		tfMap["architectures"] = v
	}

	if v := catalogData.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := catalogData.OperatingSystems; v != nil {
		tfMap["operating_systems"] = v
	}

	if v := catalogData.UsageText; v != nil {
		tfMap["usage_text"] = aws.ToString(v)
	}

	return tfMap
}

func expandRepositoryCatalogData(tfMap map[string]interface{}) *awstypes.RepositoryCatalogDataInput {
	if tfMap == nil {
		return nil
	}

	repositoryCatalogDataInput := &awstypes.RepositoryCatalogDataInput{}

	if v, ok := tfMap["about_text"].(string); ok && v != "" {
		repositoryCatalogDataInput.AboutText = aws.String(v)
	}

	if v, ok := tfMap["architectures"].(*schema.Set); ok {
		architectures := make([]string, v.Len())
		for i, val := range v.List() {
			architectures[i] = val.(string)
		}
		repositoryCatalogDataInput.Architectures = architectures
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		repositoryCatalogDataInput.Description = aws.String(v)
	}

	if v, ok := tfMap["logo_image_blob"].(string); ok && len(v) > 0 {
		repositoryCatalogDataInput.LogoImageBlob = itypes.MustBase64Decode(v)
	}

	if v, ok := tfMap["operating_systems"].(*schema.Set); ok {
		operatingSystems := make([]string, v.Len())
		for i, val := range v.List() {
			operatingSystems[i] = val.(string)
		}
		repositoryCatalogDataInput.OperatingSystems = operatingSystems
	}

	if v, ok := tfMap["usage_text"].(string); ok && v != "" {
		repositoryCatalogDataInput.UsageText = aws.String(v)
	}

	return repositoryCatalogDataInput
}

func resourceRepositoryUpdateCatalogData(ctx context.Context, conn *ecrpublic.Client, d *schema.ResourceData) error {
	if d.HasChange("catalog_data") {
		if v, ok := d.GetOk("catalog_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input := ecrpublic.PutRepositoryCatalogDataInput{
				RepositoryName: aws.String(d.Id()),
				RegistryId:     aws.String(d.Get("registry_id").(string)),
				CatalogData:    expandRepositoryCatalogData(v.([]interface{})[0].(map[string]interface{})),
			}

			_, err := conn.PutRepositoryCatalogData(ctx, &input)

			if err != nil {
				return fmt.Errorf("updating catalog data for repository(%s): %s", d.Id(), err)
			}
		}
	}

	return nil
}
