// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPullThroughCacheRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_basic(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "credential_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", "public.ecr.aws"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_credentialARN(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_credentialARN(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "credential_arn"),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", "registry-1.docker.io"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_basic(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecr.ResourcePullThroughCacheRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_failWhenAlreadyExists(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("ECR Pull Through Cache Rule is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPullThroughCacheRuleConfig_failWhenAlreadyExists(repositoryPrefix),
				ExpectError: regexache.MustCompile(`PullThroughCacheRuleAlreadyExistsException`),
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_repositoryPrefixWithSlash(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test/" + sdkacctest.RandString(22)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_basic(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", "public.ecr.aws"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_privateRepositorySelfAccount(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullThroughCacheRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_privateRepositorySelfAccount(repositoryPrefix, "ROOT", acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					testAccCheckRepositoryUpstreamRegistryURL(ctx, resourceName, acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					resource.TestCheckResourceAttr(resourceName, "upstream_repository_prefix", "ROOT"),
				),
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_privateRepositoryCrossAccount(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

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
				Config: testAccPullThroughCacheRuleConfig_privateRepositoryCrossAccount(repositoryPrefix, "ROOT", acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					testAccCheckRepositoryUpstreamRegistryURLCrossAccount(resourceName, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					resource.TestCheckResourceAttr(resourceName, "upstream_repository_prefix", "ROOT"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				Config: testAccPullThroughCacheRuleConfig_privateRepositoryCrossAccountUpdated(repositoryPrefix, "ROOT", acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					testAccCheckRepositoryUpstreamRegistryURLCrossAccount(resourceName, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					resource.TestCheckResourceAttr(resourceName, "upstream_repository_prefix", "ROOT"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_role_arn", "aws_iam_role.test_updated", names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckPullThroughCacheRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_pull_through_cache_rule" {
				continue
			}

			_, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Pull Through Cache Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPullThroughCacheRuleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		_, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckRepositoryUpstreamRegistryURL(ctx context.Context, resourceName, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", acctest.AccountID(ctx), region) //lintignore:AWSR001
		return resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", attributeValue)(s)
	}
}

func testAccCheckRepositoryUpstreamRegistryURLCrossAccount(resourceName, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		callerIdentityDataSourceName := "data.aws_caller_identity.alternate"
		callerIdentity, err := acctest.PrimaryInstanceState(s, callerIdentityDataSourceName)
		if err != nil {
			return err
		}

		anotherAccountId, ok := callerIdentity.Attributes[names.AttrAccountID]
		if !ok {
			return fmt.Errorf("account_id not found in %s", callerIdentityDataSourceName)
		}

		attributeValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", anotherAccountId, region) //lintignore:AWSR001
		return resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", attributeValue)(s)
	}
}

func testAccPullThroughCacheRuleConfig_basic(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleConfig_credentialARN(repositoryPrefix string) string {
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
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleConfig_failWhenAlreadyExists(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}

resource "aws_ecr_pull_through_cache_rule" "duplicate" {
  depends_on            = [aws_ecr_pull_through_cache_rule.test]
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleConfig_privateRepositorySelfAccount(ecrRepositoryPrefix, upstreamRepositoryPrefix, anotherRegion string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix      = %[1]q
  upstream_repository_prefix = %[2]q
  upstream_registry_url      = "${data.aws_caller_identity.current.account_id}.dkr.ecr.%[3]s.amazonaws.com"
}
`, ecrRepositoryPrefix, upstreamRepositoryPrefix, anotherRegion)
}

func testAccPullThroughCacheRuleConfig_privateRepositoryCrossAccount(ecrRepositoryPrefix, upstreamRepositoryPrefix, region string) string {
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
`, ecrRepositoryPrefix, upstreamRepositoryPrefix, region),
	)
}

func testAccPullThroughCacheRuleConfig_privateRepositoryCrossAccountUpdated(ecrRepositoryPrefix, upstreamRepositoryPrefix, region string) string {
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

resource "aws_iam_role" "test_updated" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}
resource "aws_iam_role_policy" "test_updated" {
  role   = aws_iam_role.test_updated.name
  policy = data.aws_iam_policy_document.role_policy.json
}

resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix      = %[1]q
  upstream_repository_prefix = %[2]q
  upstream_registry_url      = "${data.aws_caller_identity.alternate.account_id}.dkr.ecr.%[3]s.amazonaws.com"
  custom_role_arn            = aws_iam_role.test_updated.arn
  depends_on                 = [aws_ecr_registry_policy.test, aws_iam_role_policy.test_updated, aws_iam_role.test_updated]
}
`, ecrRepositoryPrefix, upstreamRepositoryPrefix, region),
	)
}
