package dynamodb_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccContributorInsights_basic(t *testing.T) {
	var conf dynamodb.DescribeContributorInsightsOutput
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	indexName := fmt.Sprintf("%s-index", rName)
	resourceName := "aws_dynamodb_contributor_insights.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContributorInsightsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightsBasicConfig(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContributorInsightsBasicConfig(rName, indexName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "index_name", indexName),
				),
			},
		},
	})
}

func TestAccContributorInsights_disappears(t *testing.T) {
	var conf dynamodb.DescribeContributorInsightsOutput
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_dynamodb_contributor_insights.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContributorInsightsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContributorInsightsBasicConfig(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContributorInsightsExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfdynamodb.ResourceContributorInsights(), resourceName),
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

func testAccContributorInsightsBasicConfig(rName, indexName string) string {
	return acctest.ConfigCompose(testAccContributorInsightsBaseConfig(rName), fmt.Sprintf(`
resource "aws_dynamodb_contributor_insights" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q
}
`, rName, indexName))
}

func testAccCheckContributorInsightsExists(n string, ci *dynamodb.DescribeContributorInsightsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no DynamodDB Contributor Insights ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

		tableName, indexName, err := tfdynamodb.DecodeContributorInsightsID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfdynamodb.FindContributorInsights(context.Background(), conn, tableName, indexName)
		if err != nil {
			return err
		}

		ci = output

		return nil
	}
}

func testAccCheckContributorInsightsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_contributor_insights" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB Contributor Insights %s exists", rs.Primary.ID)

		tableName, indexName, err := tfdynamodb.DecodeContributorInsightsID(rs.Primary.ID)
		if err != nil {
			return err
		}

		in := &dynamodb.DescribeContributorInsightsInput{
			TableName: aws.String(tableName),
		}

		if indexName != "" {
			in.IndexName = aws.String(indexName)
		}

		_, err = tfdynamodb.FindContributorInsights(context.Background(), conn, tableName, indexName)
		if err == nil {
			return fmt.Errorf("the DynamoDB Contributor Insights %s still exists. Failing", rs.Primary.ID)
		}

		// Verify the error is what we want
		if tfresource.NotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
