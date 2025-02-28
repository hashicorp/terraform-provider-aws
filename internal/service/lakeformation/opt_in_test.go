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
	"github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLakeFormationOptIn_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var optin lakeformation.ListLakeFormationOptInsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_opt_in.test"
	databaseName := "aws_glue_catalog_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptInConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource.0.principal.data_lake_principal_identifier", databaseName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccLakeFormationOptIn_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var optin lakeformation.ListLakeFormationOptInsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_opt_in.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptInDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptInConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptInExists(ctx, resourceName, &optin),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceOptIn = newResourceOptIn
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceOptIn, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			// Extract principal from state
			principalID := rs.Primary.Attributes["principal.0.data_lake_principal_identifier"]
			if principalID == "" {
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameOptIn, rs.Primary.ID, errors.New("principal identifier not found in state"))
			}

			// Create resource based on what's in state
			input := &lakeformation.ListLakeFormationOptInsInput{
				Resource: &awstypes.Resource{},
				Principal: &awstypes.DataLakePrincipal{
					DataLakePrincipalIdentifier: aws.String(principalID),
				},
			}

			// Check each possible resource type
			if v, ok := rs.Primary.Attributes["resource.0.catalog.0.id"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					Catalog: &awstypes.CatalogResource{Id: aws.String(v)},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.database.0.name"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					Database: &awstypes.DatabaseResource{
						Name: aws.String(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.data_cells_filter.0.name"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					DataCellsFilter: &awstypes.DataCellsFilterResource{
						Name: aws.String(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.data_location.0.resource_arn"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					DataLocation: &awstypes.DataLocationResource{
						ResourceArn: aws.String(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag.0.key"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					LFTag: &awstypes.LFTagKeyResource{
						TagKey: aws.String(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag_expression.0.name"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					LFTagExpression: &awstypes.LFTagExpressionResource{
						Name: aws.String(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag_policy.0.resource_type"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					LFTagPolicy: &awstypes.LFTagPolicyResource{
						ResourceType: awstypes.ResourceType(v),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.table.0.name"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					Table: &awstypes.TableResource{
						DatabaseName: aws.String(rs.Primary.Attributes["resource.0.table.0.database_name"]),
					},
				}
			} else if v, ok := rs.Primary.Attributes["resource.0.table_with_columns.0.name"]; ok && v != "" {
				input.Resource = &awstypes.Resource{
					TableWithColumns: &awstypes.TableWithColumnsResource{
						Name:         aws.String(v),
						DatabaseName: aws.String(rs.Primary.Attributes["resource.0.table_with_columns.0.database_name"]),
					},
				}
			}

			_, err := conn.ListLakeFormationOptIns(ctx, input)

			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				continue
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "not found") {
				continue
			}

			if err != nil {
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

		// Extract principal from state
		principalID := rs.Primary.Attributes["principal.0.data_lake_principal_identifier"]
		if principalID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, name, errors.New("principal identifier not set"))
		}

		// Create input with resource based on what's in state
		in := &lakeformation.ListLakeFormationOptInsInput{}
		var resource *types.Resource

		// Check each possible resource type
		if v, ok := rs.Primary.Attributes["resource.0.catalog.0.id"]; ok && v != "" {
			resource = &types.Resource{
				Catalog: &types.CatalogResource{Id: aws.String(v)},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.database.0.name"]; ok && v != "" {
			resource = &types.Resource{
				Database: &types.DatabaseResource{
					Name: aws.String(v),
					// CatalogId: aws.String(rs.Primary.Attributes["resource.0.database.0.catalog_id"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.data_cells_filter.0.name"]; ok && v != "" {
			resource = &types.Resource{
				DataCellsFilter: &types.DataCellsFilterResource{
					Name: aws.String(v),
					// DatabaseName:   aws.String(rs.Primary.Attributes["resource.0.data_cells_filter.0.database_name"]),
					// TableCatalogId: aws.String(rs.Primary.Attributes["resource.0.data_cells_filter.0.table_catalog_id"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.data_location.0.resource_arn"]; ok && v != "" {
			resource = &types.Resource{
				DataLocation: &types.DataLocationResource{
					ResourceArn: aws.String(v),
					// CatalogId:   aws.String(rs.Primary.Attributes["resource.0.data_location.0.catalog_id"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag.0.key"]; ok && v != "" {
			resource = &types.Resource{
				LFTag: &types.LFTagKeyResource{
					TagKey: aws.String(v),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag_expression.0.name"]; ok && v != "" {
			resource = &types.Resource{
				LFTagExpression: &types.LFTagExpressionResource{
					Name: aws.String(v),
					// CatalogId: aws.String(rs.Primary.Attributes["resource.0.lf_tag_expression.0.catalog_id"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.lf_tag_policy.0.resource_type"]; ok && v != "" {
			resource = &types.Resource{
				LFTagPolicy: &types.LFTagPolicyResource{
					ResourceType: types.ResourceType(v),
					// CatalogId:      aws.String(rs.Primary.Attributes["resource.0.lf_tag_policy.0.catalog_id"]),
					// ExpressionName: aws.String(rs.Primary.Attributes["resource.0.lf_tag_policy.0.expression_name"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.table.0.name"]; ok && v != "" {
			resource = &types.Resource{
				Table: &types.TableResource{
					// Name:         aws.String(v),
					DatabaseName: aws.String(rs.Primary.Attributes["resource.0.table.0.database_name"]),
				},
			}
		} else if v, ok := rs.Primary.Attributes["resource.0.table_with_columns.0.name"]; ok && v != "" {
			resource = &types.Resource{
				TableWithColumns: &types.TableWithColumnsResource{
					Name:         aws.String(v),
					DatabaseName: aws.String(rs.Primary.Attributes["resource.0.table_with_columns.0.database_name"]),
				},
			}
		}

		if resource == nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, name, errors.New("no valid resource found in state"))
		}

		in.Resource = resource

		out, err := tflakeformation.FindOptInByID(ctx, conn, principalID, resource)
		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameOptIn, principalID, err)
		}

		*optin = lakeformation.ListLakeFormationOptInsOutput{
			LakeFormationOptInsInfoList: []types.LakeFormationOptInsInfo{*out},
		}

		return nil
	}
}

func testAccOptInConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_opt_in" "test" {
  principal {
    data_lake_principal_identifier = aws_iam_role.test.arn
  }

  resource {
    database {
      name = aws_glue_catalog_database.test.name
      catalog_id = data.aws_caller_identity.current.account_id
    }
  }
}`, rName)
}
