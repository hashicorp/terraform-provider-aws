// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBVectorIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var index awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "dimensions", "2"),
					resource.TestCheckResourceAttr(resourceName, "distance_function", "COSINE"),
					resource.TestCheckResourceAttr(resourceName, "vector_attribute.attribute_name", "embedding"),
					resource.TestCheckResourceAttr(resourceName, "projection.0.projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_name", rNameTable),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_type", "S"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.type", "HASH"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccVectorIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var index awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdynamodb.ResourceVectorIndex, resourceName),
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

func testAccCheckVectorIndexDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_vector_index" {
				continue
			}

			tableName := rs.Primary.Attributes[names.AttrTableName]
			indexName := rs.Primary.Attributes["index_name"]

			_, err := tfdynamodb.FindVectorIndexByTwoPartKey(ctx, conn, tableName, indexName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Vector Index %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVectorIndexExists(ctx context.Context, t *testing.T, n string, v *awstypes.VectorIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)
		tableName := rs.Primary.Attributes[names.AttrTableName]
		indexName := rs.Primary.Attributes["index_name"]

		output, err := tfdynamodb.FindVectorIndexByTwoPartKey(ctx, conn, tableName, indexName)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVectorIndexImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		parts := []string{
			rs.Primary.Attributes[names.AttrTableName],
			rs.Primary.Attributes["index_name"],
		}

		return strings.Join(parts, intflex.ResourceIdSeparator), nil
	}
}

func testAccVectorIndexConfig_basic(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName)
}
