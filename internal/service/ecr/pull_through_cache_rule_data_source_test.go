// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRPullThroughCacheRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "public.ecr.aws"),
					acctest.CheckResourceAttrAccountID(dataSource, "registry_id"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRuleDataSource_repositoryPrefixWithSlash(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test/" + sdkacctest.RandString(22)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_repositoryPrefixWithSlash(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "public.ecr.aws"),
					acctest.CheckResourceAttrAccountID(dataSource, "registry_id"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRuleDataSource_credential(t *testing.T) {
	ctx := acctest.Context(t)
	upstreamRegistryUrl := "registry-1.docker.io"
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_credentialArn(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", upstreamRegistryUrl),
					acctest.CheckResourceAttrAccountID(dataSource, "registry_id"),
					resource.TestCheckResourceAttrSet(dataSource, "credential_arn"),
				),
			},
		},
	})
}

func testAccPullThroughCacheRuleDataSourceConfig_basic() string {
	return `
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`
}

func testAccPullThroughCacheRuleDataSourceConfig_repositoryPrefixWithSlash(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleDataSourceConfig_credentialArn() string {
	return `
resource "aws_secretsmanager_secret" "test" {
  name                    = "ecr-pullthroughcache/docker-hub"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test"
}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "registry-1.docker.io"
  credential_arn        = aws_secretsmanager_secret.test.arn
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`
}
