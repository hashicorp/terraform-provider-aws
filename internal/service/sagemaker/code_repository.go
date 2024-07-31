// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_code_repository", name="Code Repository")
// @Tags(identifierAttribute="arn")
func ResourceCodeRepository() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCodeRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("code_repository_name").(string)
	input := &sagemaker.CreateCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
		GitConfig:          expandCodeRepositoryGitConfig(d.Get("git_config").([]interface{})),
		Tags:               getTagsIn(ctx),
	}

	log.Printf("[DEBUG] sagemaker code repository create config: %#v", *input)
	_, err := conn.CreateCodeRepositoryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker code repository: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceCodeRepositoryRead(ctx, d, meta)...)
}

func resourceCodeRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	codeRepository, err := FindCodeRepositoryByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "Cannot find CodeRepository") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker code repository (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker code repository (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(codeRepository.CodeRepositoryArn)
	d.Set("code_repository_name", codeRepository.CodeRepositoryName)
	d.Set(names.AttrARN, arn)

	if err := d.Set("git_config", flattenCodeRepositoryGitConfig(codeRepository.GitConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting git_config for sagemaker code repository (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceCodeRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChange("git_config") {
		input := &sagemaker.UpdateCodeRepositoryInput{
			CodeRepositoryName: aws.String(d.Id()),
			GitConfig:          expandCodeRepositoryUpdateGitConfig(d.Get("git_config").([]interface{})),
		}

		log.Printf("[DEBUG] sagemaker code repository update config: %#v", *input)
		_, err := conn.UpdateCodeRepositoryWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker code repository: %s", err)
		}
	}

	return append(diags, resourceCodeRepositoryRead(ctx, d, meta)...)
}

func resourceCodeRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteCodeRepositoryInput{
		CodeRepositoryName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteCodeRepositoryWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "Cannot find CodeRepository") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker code repository (%s): %s", d.Id(), err)
	}

	return diags
}

func expandCodeRepositoryGitConfig(l []interface{}) *sagemaker.GitConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfig{
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

func flattenCodeRepositoryGitConfig(config *sagemaker.GitConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"repository_url": aws.StringValue(config.RepositoryUrl),
	}

	if config.Branch != nil {
		m["branch"] = aws.StringValue(config.Branch)
	}

	if config.SecretArn != nil {
		m["secret_arn"] = aws.StringValue(config.SecretArn)
	}

	return []map[string]interface{}{m}
}

func expandCodeRepositoryUpdateGitConfig(l []interface{}) *sagemaker.GitConfigForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfigForUpdate{
		SecretArn: aws.String(m["secret_arn"].(string)),
	}

	return config
}
