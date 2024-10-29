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
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
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
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
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
	resourceName := "aws_ecr_pull_through_cache_rule.test"

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
				Config: testAccPullThroughCacheRuleConfig_failWhenAlreadyExist(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(ctx, resourceName),
				),
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
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", "public.ecr.aws"),
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

func testAccPullThroughCacheRuleConfig_failWhenAlreadyExist(repositoryPrefix string) string {
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
