// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_ecr_pull_through_cache_rule")
func DataSourcePullThroughCacheRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePullThroughCacheRuleRead,
		Schema: map[string]*schema.Schema{
			"ecr_repository_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 20),
					validation.StringMatch(
						regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`),
						"must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upstream_registry_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePullThroughCacheRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECRConn(ctx)

	repositoryPrefix := d.Get("ecr_repository_prefix").(string)

	rule, err := FindPullThroughCacheRuleByRepositoryPrefix(ctx, conn, repositoryPrefix)

	if err != nil {
		return diag.Errorf("reading ECR Pull Through Cache Rule (%s): %s", repositoryPrefix, err)
	}

	d.SetId(aws.StringValue(rule.EcrRepositoryPrefix))
	d.Set("ecr_repository_prefix", rule.EcrRepositoryPrefix)
	d.Set("registry_id", rule.RegistryId)
	d.Set("upstream_registry_url", rule.UpstreamRegistryUrl)

	return nil
}
