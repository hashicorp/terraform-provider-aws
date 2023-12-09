// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ecr_pull_through_cache_rule")
func ResourcePullThroughCacheRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePullThroughCacheRuleCreate,
		ReadWithoutTimeout:   resourcePullThroughCacheRuleRead,
		DeleteWithoutTimeout: resourcePullThroughCacheRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"ecr_repository_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 20),
					validation.StringMatch(
						regexache.MustCompile(`^[0-9a-z]+(?:[._-][0-9a-z]+)*$`),
						"must only include alphanumeric, underscore, period, or hyphen characters"),
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

	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	repositoryPrefix := d.Get("ecr_repository_prefix").(string)
	input := &ecr.CreatePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(repositoryPrefix),
		UpstreamRegistryUrl: aws.String(d.Get("upstream_registry_url").(string)),
	}

	log.Printf("[DEBUG] Creating ECR Pull Through Cache Rule: %s", input)
	_, err := conn.CreatePullThroughCacheRuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Pull Through Cache Rule (%s): %s", repositoryPrefix, err)
	}

	d.SetId(repositoryPrefix)

	return append(diags, resourcePullThroughCacheRuleRead(ctx, d, meta)...)
}

func resourcePullThroughCacheRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	rule, err := FindPullThroughCacheRuleByRepositoryPrefix(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Pull Through Cache Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Pull Through Cache Rule (%s): %s", d.Id(), err)
	}

	d.Set("ecr_repository_prefix", rule.EcrRepositoryPrefix)
	d.Set("registry_id", rule.RegistryId)
	d.Set("upstream_registry_url", rule.UpstreamRegistryUrl)

	return diags
}

func resourcePullThroughCacheRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	log.Printf("[DEBUG] Deleting ECR Pull Through Cache Rule: (%s)", d.Id())
	_, err := conn.DeletePullThroughCacheRuleWithContext(ctx, &ecr.DeletePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(d.Id()),
		RegistryId:          aws.String(d.Get("registry_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, ecr.ErrCodePullThroughCacheRuleNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Pull Through Cache Rule (%s): %s", d.Id(), err)
	}

	return diags
}
