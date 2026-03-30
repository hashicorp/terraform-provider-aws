// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRReplicationConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccReplicationConfiguration_basic,
		acctest.CtDisappears: testAccReplicationConfiguration_disappears,
		"repositoryFilter":   testAccReplicationConfiguration_repositoryFilter,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccReplicationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_replication_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckReplicationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_basic(acctest.AlternateRegion(), "current"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationConfigurationConfig_basic(acctest.AlternateRegion(), "alternate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id", "data.aws_caller_identity.alternate", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
			{
				Config: testAccReplicationConfigurationConfig_multipleRegion(acctest.AlternateRegion(), acctest.ThirdRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.1.region", acctest.ThirdRegion()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "replication_configuration.0.rule.0.destination.1.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
			{
				Config: testAccReplicationConfigurationConfig_basic(acctest.AlternateRegion(), "current"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
		},
	})
}

func testAccReplicationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_replication_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckReplicationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_basic(acctest.AlternateRegion(), "current"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecr.ResourceReplicationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccReplicationConfiguration_repositoryFilter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_replication_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_repositoryFilter(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationConfigurationConfig_repositoryFilterMultiple(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.1.filter", "a-second-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.1.filter_type", "PREFIX_MATCH"),
				),
			},
			{
				Config: testAccReplicationConfigurationConfig_repositoryFilter(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		_, err := tfecr.FindReplicationConfiguration(ctx, conn)

		return err
	}
}

func testAccCheckReplicationConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_replication_configuration" {
				continue
			}

			_, err := tfecr.FindReplicationConfiguration(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Replication Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReplicationConfigurationConfig_basic(region, account string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_caller_identity" "alternate" {
  provider = awsalternate
}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.%[2]s.account_id
      }
    }
  }
}
`, region, account))
}

func testAccReplicationConfigurationConfig_multipleRegion(region1, region2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }


      destination {
        region      = %[2]q
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`, region1, region2)
}

func testAccReplicationConfigurationConfig_repositoryFilter(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }

      repository_filter {
        filter      = "a-prefix"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
`, region)
}

func testAccReplicationConfigurationConfig_repositoryFilterMultiple(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }

      repository_filter {
        filter      = "a-prefix"
        filter_type = "PREFIX_MATCH"
      }

      repository_filter {
        filter      = "a-second-prefix"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
`, region)
}
