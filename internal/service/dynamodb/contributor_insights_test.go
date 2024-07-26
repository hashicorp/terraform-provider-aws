// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBContributorInsights_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf dynamodb.DescribeContributorInsightsOutput
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	indexName := fmt.Sprintf("%s-index", rName)
	resourceName := "aws_dynamodb_contributor_insights.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightsConfig_basic(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContributorInsightsConfig_basic(rName, indexName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
				),
			},
		},
	})
}

func TestAccDynamoDBContributorInsights_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf dynamodb.DescribeContributorInsightsOutput
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_dynamodb_contributor_insights.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorInsightsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightsConfig_basic(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceContributorInsights(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContributorInsightsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 2
  write_capacity = 2
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }

  global_secondary_index {
    name            = "%[1]s-index"
    hash_key        = %[1]q
    projection_type = "ALL"
    read_capacity   = 1
    write_capacity  = 1
  }
}
`, rName)
}

func testAccContributorInsightsConfig_basic(rName, indexName string) string {
	return acctest.ConfigCompose(testAccContributorInsightsBaseConfig(rName), fmt.Sprintf(`
resource "aws_dynamodb_contributor_insights" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
}
`, rName, indexName))
}

func testAccCheckContributorInsightsExists(ctx context.Context, n string, v *dynamodb.DescribeContributorInsightsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		tableName, indexName, err := tfdynamodb.ContributorInsightsParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfdynamodb.FindContributorInsightsByTwoPartKey(ctx, conn, tableName, indexName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckContributorInsightsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_contributor_insights" {
				continue
			}

			tableName, indexName, err := tfdynamodb.ContributorInsightsParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfdynamodb.FindContributorInsightsByTwoPartKey(ctx, conn, tableName, indexName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Contributor Insights %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
