// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPullThroughCacheRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "public.ecr.aws"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_repositoryPrefixWithSlash(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "public.ecr.aws"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRuleDataSource_credential(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_credentialARN(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSource, "credential_arn"),
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "registry-1.docker.io"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRuleDataSource_privateRepositorySelfAccount(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_privateRepositorySelfAccount(repositoryPrefix, "ROOT", acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					testAccCheckRepositoryUpstreamRegistryURL(ctx, dataSource, acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(dataSource, "ecr_repository_prefix", repositoryPrefix),
					resource.TestCheckResourceAttr(dataSource, "upstream_repository_prefix", "ROOT"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRuleDataSource_privateRepositoryCrossAccount(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_privateRepositoryCrossAccount(repositoryPrefix, "ROOT", acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					testAccCheckRepositoryUpstreamRegistryURLCrossAccount(dataSource, acctest.Region()),
					resource.TestCheckResourceAttr(dataSource, "ecr_repository_prefix", repositoryPrefix),
					resource.TestCheckResourceAttr(dataSource, "upstream_repository_prefix", "ROOT"),
					resource.TestCheckResourceAttrPair(dataSource, "custom_role_arn", "aws_iam_role.test", names.AttrARN),
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

func testAccPullThroughCacheRuleDataSourceConfig_credentialARN(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = "ecr-pullthroughcache/%[1]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test"
}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "registry-1.docker.io"
  credential_arn        = aws_secretsmanager_secret.test.arn
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleDataSourceConfig_privateRepositorySelfAccount(ecrRepositoryPrefix, upstreamRepositoryPrefix, anotherRegion string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix      = %[1]q
  upstream_repository_prefix = %[2]q
  upstream_registry_url      = "${data.aws_caller_identity.current.account_id}.dkr.ecr.%[3]s.amazonaws.com"
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`, ecrRepositoryPrefix, upstreamRepositoryPrefix, anotherRegion)
}

func testAccPullThroughCacheRuleDataSourceConfig_privateRepositoryCrossAccount(ecrRepositoryPrefix, upstreamRepositoryPrefix, region string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = awsalternate
}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

data "aws_iam_policy_document" "registry_policy" {
  statement {
    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      ]
      type = "AWS"
    }
    actions = [
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchImportUpstreamImage",
      "ecr:GetImageCopyStatus"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:ecr:%[3]s:${data.aws_caller_identity.alternate.account_id}:repository/*",
    ]
  }
}

resource "aws_ecr_registry_policy" "test" {
  provider = awsalternate
  policy   = data.aws_iam_policy_document.registry_policy.json
}

data "aws_iam_policy_document" "role_policy" {
  statement {
    actions = [
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetAuthorizationToken",
      "ecr:BatchImportUpstreamImage",
      "ecr:BatchGetImage",
      "ecr:GetImageCopyStatus",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage"
    ]
    resources = [
      "*"
    ]
  }
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["pullthroughcache.ecr.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}
resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.role_policy.json
}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix      = %[1]q
  upstream_repository_prefix = %[2]q
  upstream_registry_url      = "${data.aws_caller_identity.alternate.account_id}.dkr.ecr.%[3]s.amazonaws.com"
  custom_role_arn            = aws_iam_role.test.arn
  depends_on                 = [aws_ecr_registry_policy.test, aws_iam_role_policy.test, aws_iam_role.test]
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`, ecrRepositoryPrefix, upstreamRepositoryPrefix, region),
	)
}
