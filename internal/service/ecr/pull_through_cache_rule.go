// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ecr_pull_through_cache_rule", name="Pull Through Cache Rule")
func resourcePullThroughCacheRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePullThroughCacheRuleCreate,
		ReadWithoutTimeout:   resourcePullThroughCacheRuleRead,
		DeleteWithoutTimeout: resourcePullThroughCacheRuleDelete,
		UpdateWithoutTimeout: resourcePullThroughCacheRuleUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"credential_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"ecr_repository_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 30),
					validation.StringMatch(
						regexache.MustCompile(`(?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*`),
						"must only include alphanumeric, underscore, period, hyphen, or slash characters"),
				),
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upstream_registry_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePullThroughCacheRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.ecr-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	repositoryPrefix := d.Get("ecr_repository_prefix").(string)
	input := &ecr.CreatePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(repositoryPrefix),
		UpstreamRegistryUrl: aws.String(d.Get("upstream_registry_url").(string)),
	}

	if v, ok := d.GetOk("credential_arn"); ok {
		input.CredentialArn = aws.String(v.(string))
	}

	_, err := conn.CreatePullThroughCacheRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Pull Through Cache Rule (%s): %s", repositoryPrefix, err)
	}

	d.SetId(repositoryPrefix)

	return append(diags, resourcePullThroughCacheRuleRead(ctx, d, meta)...)
}

func resourcePullThroughCacheRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	rule, err := findPullThroughCacheRuleByRepositoryPrefix(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Pull Through Cache Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Pull Through Cache Rule (%s): %s", d.Id(), err)
	}

	d.Set("credential_arn", rule.CredentialArn)
	d.Set("ecr_repository_prefix", rule.EcrRepositoryPrefix)
	d.Set("registry_id", rule.RegistryId)
	d.Set("upstream_registry_url", rule.UpstreamRegistryUrl)

	return diags
}

func resourcePullThroughCacheRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	repositoryPrefix := d.Get("ecr_repository_prefix").(string)
	input := &ecr.UpdatePullThroughCacheRuleInput{
		CredentialArn:       aws.String(d.Get("credential_arn").(string)),
		EcrRepositoryPrefix: aws.String(repositoryPrefix),
	}

	_, err := conn.UpdatePullThroughCacheRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ECR Pull Through Cache Rule (%s): %s", repositoryPrefix, err)
	}

	d.SetId(repositoryPrefix)

	return append(diags, resourcePullThroughCacheRuleRead(ctx, d, meta)...)
}

func resourcePullThroughCacheRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Pull Through Cache Rule: %s", d.Id())
	_, err := conn.DeletePullThroughCacheRule(ctx, &ecr.DeletePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(d.Id()),
		RegistryId:          aws.String(d.Get("registry_id").(string)),
	})

	if errs.IsA[*types.PullThroughCacheRuleNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Pull Through Cache Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findPullThroughCacheRuleByRepositoryPrefix(ctx context.Context, conn *ecr.Client, repositoryPrefix string) (*types.PullThroughCacheRule, error) {
	input := &ecr.DescribePullThroughCacheRulesInput{
		EcrRepositoryPrefixes: []string{repositoryPrefix},
	}

	return findPullThroughCacheRule(ctx, conn, input)
}

func findPullThroughCacheRule(ctx context.Context, conn *ecr.Client, input *ecr.DescribePullThroughCacheRulesInput) (*types.PullThroughCacheRule, error) {
	output, err := findPullThroughCacheRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPullThroughCacheRules(ctx context.Context, conn *ecr.Client, input *ecr.DescribePullThroughCacheRulesInput) ([]types.PullThroughCacheRule, error) {
	var output []types.PullThroughCacheRule

	pages := ecr.NewDescribePullThroughCacheRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.PullThroughCacheRuleNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PullThroughCacheRules...)
	}

	return output, nil
}
