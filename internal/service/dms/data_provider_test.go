// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSDataProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "data_provider_arn"),
					resource.TestCheckResourceAttr(resourceName, "data_provider_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "postgres"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"settings"},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "data_provider_arn"),
				ImportStateVerifyIdentifierAttribute: "data_provider_arn",
			},
		},
	})
}

func TestAccDMSDataProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.0.port", "5432"),
				),
			},
			{
				Config: testAccDataProviderConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.0.port", "5433"),
				),
			},
		},
	})
}

func TestAccDMSDataProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceDataProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckDataProviderExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		_, err := tfdms.FindDataProviderByARN(ctx, conn, rs.Primary.Attributes["data_provider_arn"])

		return err
	}
}

func testAccCheckDataProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_data_provider" {
				continue
			}

			_, err := tfdms.FindDataProviderByARN(ctx, conn, rs.Primary.Attributes["data_provider_arn"])

			if errs.IsA[*retry.NotFoundError](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Data Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  data_provider_name = %[1]q
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "%[1]s.example.com"
      port          = 5432
      database_name = "testdb"
      ssl_mode      = "none"
    }
  }
}
`, rName)
}

func testAccDataProviderConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  data_provider_name = %[1]q
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "%[1]s.example.com"
      port          = 5433
      database_name = "testdb"
      ssl_mode      = "none"
    }
  }
}
`, rName)
}
