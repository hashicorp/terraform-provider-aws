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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_code_repository", name="Code Repository")
// @Tags(identifierAttribute="arn")
func resourceCodeRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCodeRepositoryCreate,
		ReadWithoutTimeout:   resourceCodeRepositoryRead,
		UpdateWithoutTimeout: resourceCodeRepositoryUpdate,
		DeleteWithoutTimeout: resourceCodeRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code_repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"git_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_url": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"branch": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCodeRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("code_repository_name").(string)
	input := &sagemaker.CreateCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
		GitConfig:          expandCodeRepositoryGitConfig(d.Get("git_config").([]any)),
		Tags:               getTagsIn(ctx),
	}

	log.Printf("[DEBUG] sagemaker code repository create config: %#v", *input)
	_, err := conn.CreateCodeRepository(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI code repository: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceCodeRepositoryRead(ctx, d, meta)...)
}

func resourceCodeRepositoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	codeRepository, err := findCodeRepositoryByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find CodeRepository") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker AI code repository (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI code repository (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(codeRepository.CodeRepositoryArn)
	d.Set("code_repository_name", codeRepository.CodeRepositoryName)
	d.Set(names.AttrARN, arn)

	if err := d.Set("git_config", flattenCodeRepositoryGitConfig(codeRepository.GitConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting git_config for sagemaker code repository (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceCodeRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChange("git_config") {
		input := &sagemaker.UpdateCodeRepositoryInput{
			CodeRepositoryName: aws.String(d.Id()),
			GitConfig:          expandCodeRepositoryUpdateGitConfig(d.Get("git_config").([]any)),
		}

		log.Printf("[DEBUG] sagemaker code repository update config: %#v", *input)
		_, err := conn.UpdateCodeRepository(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI code repository: %s", err)
		}
	}

	return append(diags, resourceCodeRepositoryRead(ctx, d, meta)...)
}

func resourceCodeRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteCodeRepositoryInput{
		CodeRepositoryName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteCodeRepository(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find CodeRepository") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI code repository (%s): %s", d.Id(), err)
	}

	return diags
}

func findCodeRepositoryByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeCodeRepositoryOutput, error) {
	input := &sagemaker.DescribeCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
	}

	output, err := conn.DescribeCodeRepository(ctx, input)

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

func expandCodeRepositoryGitConfig(l []any) *awstypes.GitConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.GitConfig{
		RepositoryUrl: aws.String(m["repository_url"].(string)),
	}

	if v, ok := m["branch"].(string); ok && v != "" {
		config.Branch = aws.String(v)
	}

	if v, ok := m["secret_arn"].(string); ok && v != "" {
		config.SecretArn = aws.String(v)
	}

	return config
}

func flattenCodeRepositoryGitConfig(config *awstypes.GitConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"repository_url": aws.ToString(config.RepositoryUrl),
	}

	if config.Branch != nil {
		m["branch"] = aws.ToString(config.Branch)
	}

	if config.SecretArn != nil {
		m["secret_arn"] = aws.ToString(config.SecretArn)
	}

	return []map[string]any{m}
}

func expandCodeRepositoryUpdateGitConfig(l []any) *awstypes.GitConfigForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.GitConfigForUpdate{
		SecretArn: aws.String(m["secret_arn"].(string)),
	}

	return config
}
