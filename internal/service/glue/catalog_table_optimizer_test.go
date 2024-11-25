// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCatalogTableOptimizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "compaction"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccCatalogTableOptimizer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_update(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "compaction"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccCatalogTableOptimizerConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "compaction"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCatalogTableOptimizer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, resourceName, &catalogTableOptimizer),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglue.ResourceCatalogTableOptimizer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCatalogTableOptimizerStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s,%s", rs.Primary.Attributes[names.AttrCatalogID], rs.Primary.Attributes[names.AttrDatabaseName],
			rs.Primary.Attributes[names.AttrTableName], rs.Primary.Attributes[names.AttrType]), nil
	}
}

func testAccCheckCatalogTableOptimizerExists(ctx context.Context, resourceName string, catalogTableOptimizer *glue.GetTableOptimizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, resourceName, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, resourceName, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		resp, err := tfglue.FindCatalogTableOptimizer(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID], rs.Primary.Attributes[names.AttrDatabaseName],
			rs.Primary.Attributes[names.AttrTableName], rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, rs.Primary.ID, err)
		}

		*catalogTableOptimizer = *resp

		return nil
	}
}

func testAccCheckCatalogTableOptimizerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_catalog_table_optimizer" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
			_, err := tfglue.FindCatalogTableOptimizer(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID], rs.Primary.Attributes[names.AttrDatabaseName],
				rs.Primary.Attributes[names.AttrTableName], rs.Primary.Attributes[names.AttrType])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Glue, create.ErrActionCheckingDestroyed, tfglue.ResNameCatalogTableOptimizer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCatalogTableOptimizerConfig_baseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
        ]
        Resource = [
          "${aws_s3_bucket.bucket.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.bucket.arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "glue:UpdateTable",
          "glue:GetTable"
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/${aws_glue_catalog_database.test.name}/${aws_glue_catalog_table.test.name}",
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:database/${aws_glue_catalog_database.test.name}",
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:catalog"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws-glue/iceberg-compaction/logs:*"
      }
    ]
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
  table_type    = "EXTERNAL_TABLE"

  open_table_format_input {
    iceberg_input {
      metadata_operation = "CREATE"
      version            = 2
    }
  }

  storage_descriptor {
    location = "s3://${aws_s3_bucket.bucket.bucket}/files/"

    columns {
      name = "my_column_1"
      type = "int"
    }
  }
}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "DELETE", "DESCRIBE"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}`, rName)
}

func testAccCatalogTableOptimizerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName), `
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "compaction"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = true
  }
}
`,
	)
}
func testAccCatalogTableOptimizerConfig_update(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "compaction"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = %[1]t
  }
}
`, enabled))
}
