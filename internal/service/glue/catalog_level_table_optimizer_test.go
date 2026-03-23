// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccGlueCatalogLevelTableOptimizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_level_table_optimizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "glue"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogLevelTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogLevelTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogLevelTableOptimizerExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "catalog_id"),
					resource.TestCheckResourceAttr(resourceName, "iceberg_optimization.#", "1"),
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

func TestAccGlueCatalogLevelTableOptimizer_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_catalog_level_table_optimizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "glue"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogLevelTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogLevelTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogLevelTableOptimizerExists(ctx, resourceName),
				),
			},
			{
				Config: testAccCatalogLevelTableOptimizerConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogLevelTableOptimizerExists(ctx, resourceName),
				),
			},
		},
	})
}

func testAccCheckCatalogLevelTableOptimizerDestroy(ctx context.Context) resource.TestCheckDestroyFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_catalog_level_table_optimizer" {
				continue
			}

			out, err := conn.GetCatalog(ctx, &glue.GetCatalogInput{
				CatalogId: &rs.Primary.ID,
			})
			if err != nil {
				return nil
			}
			if out.Catalog.CatalogProperties == nil {
				return nil
			}
			if out.Catalog.CatalogProperties.IcebergOptimizationProperties == nil {
				return nil
			}
			if out.Catalog.CatalogProperties.IcebergOptimizationProperties.RoleArn == nil {
				return nil
			}
			return fmt.Errorf("Glue Catalog Level Table Optimizer %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckCatalogLevelTableOptimizerExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		out, err := conn.GetCatalog(ctx, &glue.GetCatalogInput{
			CatalogId: &rs.Primary.ID,
		})
		if err != nil {
			return err
		}
		if out.Catalog.CatalogProperties == nil || out.Catalog.CatalogProperties.IcebergOptimizationProperties == nil {
			return fmt.Errorf("Glue Catalog Level Table Optimizer %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCatalogLevelTableOptimizerConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "glue.amazonaws.com" }
    }]
  })
}

resource "aws_glue_catalog_level_table_optimizer" "test" {
  catalog_id = data.aws_caller_identity.current.account_id

  iceberg_optimization {
    role_arn = aws_iam_role.test.arn

    compaction = {
      enabled = "true"
    }
  }
}
`, rName)
}

func testAccCatalogLevelTableOptimizerConfig_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "glue.amazonaws.com" }
    }]
  })
}

resource "aws_glue_catalog_level_table_optimizer" "test" {
  catalog_id = data.aws_caller_identity.current.account_id

  iceberg_optimization {
    role_arn = aws_iam_role.test.arn

    compaction = {
      enabled = "true"
    }

    retention = {
      enabled = "true"
    }
  }
}
`, rName)
}
