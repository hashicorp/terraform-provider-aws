// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCatalogTableOptimizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_update(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "compaction"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccCatalogTableOptimizerConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfglue.ResourceCatalogTableOptimizer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCatalogTableOptimizer_RetentionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_retentionConfiguration(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.snapshot_retention_period_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.number_of_snapshots_to_retain", "3"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.clean_expired_files", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.run_rate_in_hours", "24"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccCatalogTableOptimizerConfig_retentionConfiguration(rName, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.snapshot_retention_period_in_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.number_of_snapshots_to_retain", "3"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.clean_expired_files", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.run_rate_in_hours", "24"),
				),
			},
		},
	})
}

func testAccCatalogTableOptimizer_RetentionConfigurationWithRunRateInHours(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_retentionConfigurationWithRunRateInHours(rName, 7, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.snapshot_retention_period_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.number_of_snapshots_to_retain", "3"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.clean_expired_files", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.run_rate_in_hours", "6"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccCatalogTableOptimizerConfig_retentionConfigurationWithRunRateInHours(rName, 6, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.snapshot_retention_period_in_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.number_of_snapshots_to_retain", "3"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.clean_expired_files", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.retention_configuration.0.iceberg_configuration.0.run_rate_in_hours", "4"),
				),
			},
		},
	})
}

func testAccCatalogTableOptimizer_DeleteOrphanFileConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_orphanFileDeletionConfiguration(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "orphan_file_deletion"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.orphan_file_retention_period_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.location", fmt.Sprintf("s3://%s/files/", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.run_rate_in_hours", "24"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccCatalogTableOptimizerConfig_orphanFileDeletionConfiguration(rName, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "orphan_file_deletion"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.orphan_file_retention_period_in_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.location", fmt.Sprintf("s3://%s/files/", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.run_rate_in_hours", "24"),
				),
			},
		},
	})
}

func testAccCatalogTableOptimizer_DeleteOrphanFileConfigurationWithRunRateInHours(t *testing.T) {
	ctx := acctest.Context(t)
	var catalogTableOptimizer glue.GetTableOptimizerOutput

	resourceName := "aws_glue_catalog_table_optimizer.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCatalogTableOptimizerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTableOptimizerConfig_orphanFileDeletionConfigurationWithRunRateInHours(rName, 7, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "orphan_file_deletion"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.orphan_file_retention_period_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.location", fmt.Sprintf("s3://%s/files/", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.run_rate_in_hours", "6"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    testAccCatalogTableOptimizerStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccCatalogTableOptimizerConfig_orphanFileDeletionConfigurationWithRunRateInHours(rName, 6, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogTableOptimizerExists(ctx, t, resourceName, &catalogTableOptimizer),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "orphan_file_deletion"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.orphan_file_retention_period_in_days", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.location", fmt.Sprintf("s3://%s/files/", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.orphan_file_deletion_configuration.0.iceberg_configuration.0.run_rate_in_hours", "4"),
				),
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

func testAccCheckCatalogTableOptimizerExists(ctx context.Context, t *testing.T, resourceName string, catalogTableOptimizer *glue.GetTableOptimizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, resourceName, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, resourceName, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)
		resp, err := tfglue.FindCatalogTableOptimizer(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID], rs.Primary.Attributes[names.AttrDatabaseName],
			rs.Primary.Attributes[names.AttrTableName], rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return create.Error(names.Glue, create.ErrActionCheckingExistence, tfglue.ResNameCatalogTableOptimizer, rs.Primary.ID, err)
		}

		*catalogTableOptimizer = *resp

		return nil
	}
}

func testAccCheckCatalogTableOptimizerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_catalog_table_optimizer" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)
			_, err := tfglue.FindCatalogTableOptimizer(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID], rs.Primary.Attributes[names.AttrDatabaseName],
				rs.Primary.Attributes[names.AttrTableName], rs.Primary.Attributes[names.AttrType])

			if retry.NotFound(err) {
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
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:table/${aws_glue_catalog_database.test.name}/${aws_glue_catalog_table.test.name}",
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:database/${aws_glue_catalog_database.test.name}",
          "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:catalog"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws-glue/iceberg-compaction/logs:*"
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

func testAccCatalogTableOptimizerConfig_retentionConfiguration(rName string, retentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "retention"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = true

    retention_configuration {
      iceberg_configuration {
        snapshot_retention_period_in_days = %[1]d
        number_of_snapshots_to_retain     = 3
        clean_expired_files               = true
      }
    }
  }
  depends_on = [aws_iam_role_policy.test]
}
`, retentionPeriod))
}

func testAccCatalogTableOptimizerConfig_retentionConfigurationWithRunRateInHours(rName string, retentionPeriod, runRateInHours int) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "retention"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = true

    retention_configuration {
      iceberg_configuration {
        snapshot_retention_period_in_days = %[1]d
        number_of_snapshots_to_retain     = 3
        clean_expired_files               = true
        run_rate_in_hours                 = %[2]d
      }
    }
  }
  depends_on = [aws_iam_role_policy.test]
}
`, retentionPeriod, runRateInHours))
}

func testAccCatalogTableOptimizerConfig_orphanFileDeletionConfiguration(rName string, retentionPeriod int) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "orphan_file_deletion"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = true

    orphan_file_deletion_configuration {
      iceberg_configuration {
        orphan_file_retention_period_in_days = %[1]d
        location                             = "s3://${aws_s3_bucket.bucket.bucket}/files/"
      }
    }
  }
  depends_on = [aws_iam_role_policy.test]
}
`, retentionPeriod))
}

func testAccCatalogTableOptimizerConfig_orphanFileDeletionConfigurationWithRunRateInHours(rName string, retentionPeriod, runRateInHours int) string {
	return acctest.ConfigCompose(
		testAccCatalogTableOptimizerConfig_baseConfig(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_table_optimizer" "test" {
  catalog_id    = data.aws_caller_identity.current.account_id
  database_name = aws_glue_catalog_database.test.name
  table_name    = aws_glue_catalog_table.test.name
  type          = "orphan_file_deletion"

  configuration {
    role_arn = aws_iam_role.test.arn
    enabled  = true

    orphan_file_deletion_configuration {
      iceberg_configuration {
        orphan_file_retention_period_in_days = %[1]d
        location                             = "s3://${aws_s3_bucket.bucket.bucket}/files/"
        run_rate_in_hours                    = %[2]d
      }
    }
  }
  depends_on = [aws_iam_role_policy.test]
}
`, retentionPeriod, runRateInHours))
}
