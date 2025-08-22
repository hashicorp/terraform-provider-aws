// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOptIn_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var optin lakeformation.ListLakeFormationOptInsOutput
	resourceName := "aws_lakeformation_opt_in.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	roleName := "aws_iam_role.test"
	databaseName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptInConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					resource.TestCheckResourceAttr(resourceName, "principal.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "principal.0.data_lake_principal_identifier", roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_data.0.database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_data.0.database.0.name", databaseName, names.AttrName),
				),
			},
		},
	})
}

func testAccOptIn_table(t *testing.T) {
	ctx := acctest.Context(t)

	var optin lakeformation.ListLakeFormationOptInsOutput
	resourceName := "aws_lakeformation_opt_in.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	roleName := "aws_iam_role.test"
	databaseName := "aws_glue_catalog_database.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptInConfig_Table(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					resource.TestCheckResourceAttr(resourceName, "principal.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "principal.0.data_lake_principal_identifier", roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_data.0.table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_data.0.table.0.name", tableName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_data.0.table.0.database_name", databaseName, names.AttrName),
				),
			},
			{
				Config: testAccOptInConfig_Table_wildcard(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					resource.TestCheckResourceAttr(resourceName, "principal.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "principal.0.data_lake_principal_identifier", roleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "resource_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_data.0.table.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_data.0.table.0.wildcard", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "resource_data.0.table.0.database_name", databaseName, names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccOptIn_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var optin lakeformation.ListLakeFormationOptInsOutput
	resourceName := "aws_lakeformation_opt_in.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptInConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceOptIn, resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckOptInDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_opt_in" {
				continue
			}

			principalID := rs.Primary.Attributes["principal.0.data_lake_principal_identifier"]
			in := lakeformation.ListLakeFormationOptInsInput{
				Principal: &awstypes.DataLakePrincipal{
					DataLakePrincipalIdentifier: aws.String(principalID),
				},
				Resource: &awstypes.Resource{},
			}

			in.Resource = constructOptInResource(rs)

			_, err := tflakeformation.FindOptInByID(ctx, conn, principalID, in.Resource)
			if err != nil {
				if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Insufficient Lake Formation permission(s) on Catalog") {
					return nil
				}
				if errs.IsAErrorMessageContains[*awstypes.EntityNotFoundException](err, "") {
					return nil
				}
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameOptIn, rs.Primary.ID, err)
			}
			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameOptIn, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOptInExists(ctx context.Context, name string, optin *lakeformation.ListLakeFormationOptInsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		principalID := rs.Primary.ID
		in := &lakeformation.ListLakeFormationOptInsInput{}

		in.Resource = constructOptInResource(rs)

		if in.Resource == nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, name, errors.New("no valid resource found in state"))
		}

		out, err := conn.ListLakeFormationOptIns(ctx, in)
		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, principalID, err)
		}

		*optin = *out

		return nil
	}
}

func constructOptInResource(rs *terraform.ResourceState) *awstypes.Resource {
	type resourceConstructor func(*terraform.ResourceState) *awstypes.Resource

	resourceConstructors := map[string]resourceConstructor{
		"resource_data.0.catalog.0.id": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{Catalog: &awstypes.CatalogResource{Id: aws.String(rs.Primary.Attributes["resource_data.0.catalog.0.id"])}}
		},
		"resource_data.0.database.0.name": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				Database: &awstypes.DatabaseResource{
					Name:      aws.String(rs.Primary.Attributes["resource_data.0.database.0.name"]),
					CatalogId: aws.String(rs.Primary.Attributes["resource_data.0.database.0.catalog_id"]),
				},
			}
		},
		"resource_data.0.data_cells_filter.0.name": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				DataCellsFilter: &awstypes.DataCellsFilterResource{
					Name:           aws.String(rs.Primary.Attributes["resource_data.0.data_cells_filter.0.name"]),
					DatabaseName:   aws.String(rs.Primary.Attributes["resource_data.0.data_cells_filter.0.database_name"]),
					TableCatalogId: aws.String(rs.Primary.Attributes["resource_data.0.data_cells_filter.0.table_catalog_id"]),
				},
			}
		},
		"resource_data.0.data_location.0.resource_arn": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				DataLocation: &awstypes.DataLocationResource{
					ResourceArn: aws.String(rs.Primary.Attributes["resource_data.0.data_location.0.resource_arn"]),
					CatalogId:   aws.String(rs.Primary.Attributes["resource_data.0.data_location.0.catalog_id"]),
				},
			}
		},
		"resource_data.0.lf_tag.0.key": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{LFTag: &awstypes.LFTagKeyResource{TagKey: aws.String(rs.Primary.Attributes["resource_data.0.lf_tag.0.key"])}}
		},
		"resource_data.0.lf_tag_expression.0.name": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				LFTagExpression: &awstypes.LFTagExpressionResource{
					Name:      aws.String(rs.Primary.Attributes["resource_data.0.lf_tag_expression.0.name"]),
					CatalogId: aws.String(rs.Primary.Attributes["resource_data.0.lf_tag_expression.0.catalog_id"]),
				},
			}
		},
		"resource_data.0.lf_tag_policy.0.resource_type": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				LFTagPolicy: &awstypes.LFTagPolicyResource{
					ResourceType:   awstypes.ResourceType(rs.Primary.Attributes["resource_data.0.lf_tag_policy.0.resource_type"]),
					CatalogId:      aws.String(rs.Primary.Attributes["resource_data.0.lf_tag_policy.0.catalog_id"]),
					ExpressionName: aws.String(rs.Primary.Attributes["resource_data.0.lf_tag_policy.0.expression_name"]),
				},
			}
		},
		"resource_data.0.table.0.name": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				Table: &awstypes.TableResource{
					DatabaseName: aws.String(rs.Primary.Attributes["resource_data.0.table.0.database_name"]),
					Name:         aws.String(rs.Primary.Attributes["resource_data.0.table.0.name"]),
				},
			}
		},
		"resource_data.0.table.0.wildcard": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				Table: &awstypes.TableResource{
					DatabaseName:  aws.String(rs.Primary.Attributes["resource_data.0.table.0.database_name"]),
					TableWildcard: &awstypes.TableWildcard{},
				},
			}
		},
		"resource_data.0.table_with_columns.0.name": func(rs *terraform.ResourceState) *awstypes.Resource {
			return &awstypes.Resource{
				TableWithColumns: &awstypes.TableWithColumnsResource{
					Name:         aws.String(rs.Primary.Attributes["resource_data.0.table_with_columns.0.name"]),
					DatabaseName: aws.String(rs.Primary.Attributes["resource_data.0.table_with_columns.0.database_name"]),
				},
			}
		},
	}

	var resource *awstypes.Resource
	for path, constructor := range resourceConstructors {
		if v, ok := rs.Primary.Attributes[path]; ok && v != "" {
			resource = constructor(rs)
			break
		}
	}

	return resource
}

func testAccOptInConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "glue.${data.aws_partition.current.dns_suffix}",
            "lakeformation.amazonaws.com",
            "s3.amazonaws.com"
          ]
        }
      }
    ]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy" "test" {
  name = "test_policy"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lakeformation:*",
          "glue:*",
          "s3:GetBucketLocation",
          "s3:ListAllMyBuckets",
          "s3:GetObjectVersion",
          "s3:GetBucketAcl",
          "s3:GetObject",
          "s3:GetObjectACL",
          "s3:PutObject",
          "s3:PutObjectAcl",
          "iam:ListUsers",
          "iam:ListRoles",
          "iam:GetRole",
          "iam:GetRolePolicy"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_iam_session_context.current.issuer_arn,
    aws_iam_role.test.arn
  ]
  depends_on = [aws_iam_role_policy.test]

  lifecycle {
    ignore_changes = [admins]
  }
}

resource "aws_lakeformation_resource" "test" {
  arn        = aws_s3_bucket.test.arn
  role_arn   = aws_iam_role.test.arn
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccOptInConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOptInConfig_base(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name       = %[1]q
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_opt_in" "test" {
  principal {
    data_lake_principal_identifier = aws_iam_role.test.arn
  }

  resource_data {
    database {
      name       = aws_glue_catalog_database.test.name
      catalog_id = data.aws_caller_identity.current.account_id
    }
  }
}
`, rName))
}

func testAccOptInConfig_Table(rName string) string {
	return acctest.ConfigCompose(testAccOptInConfig_base(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name       = %[1]q
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    columns {
      name    = "my_column_23"
      type    = "string"
      comment = "my_column23_comment2"
    }
  }
}

resource "aws_lakeformation_opt_in" "test" {
  principal {
    data_lake_principal_identifier = aws_iam_role.test.arn
  }

  resource_data {
    table {
      database_name = aws_glue_catalog_database.test.name
      catalog_id    = data.aws_caller_identity.current.account_id
      name          = aws_glue_catalog_table.test.name
    }
  }
}
`, rName))
}

func testAccOptInConfig_Table_wildcard(rName string) string {
	return acctest.ConfigCompose(testAccOptInConfig_base(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name       = %[1]q
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name    = "my_column_12"
      type    = "date"
      comment = "my_column1_comment2"
    }

    columns {
      name    = "my_column_22"
      type    = "timestamp"
      comment = "my_column2_comment2"
    }

    columns {
      name    = "my_column_23"
      type    = "string"
      comment = "my_column23_comment2"
    }
  }
}

resource "aws_lakeformation_opt_in" "test" {
  principal {
    data_lake_principal_identifier = aws_iam_role.test.arn
  }

  resource_data {
    table {
      database_name = aws_glue_catalog_database.test.name
      catalog_id    = data.aws_caller_identity.current.account_id
      wildcard      = true
    }
  }
}
`, rName))
}
