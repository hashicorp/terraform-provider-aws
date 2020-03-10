package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func init() {
	resource.AddTestSweepers("aws_dynamodb_table_2019", &resource.Sweeper{
		Name: "aws_dynamodb_table_2019",
		F:    testSweepDynamoDbTables2019,
	})
}

func testSweepDynamoDbTables2019(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dynamodbconn

	err = conn.ListTablesPages(&dynamodb.ListTablesInput{}, func(out *dynamodb.ListTablesOutput, lastPage bool) bool {
		for _, tableName := range out.TableNames {
			log.Printf("[INFO] Deleting DynamoDB Table: %s", *tableName)

			err := deleteAwsDynamoDbTable(*tableName, conn)
			if err != nil {
				log.Printf("[ERROR] Failed to delete DynamoDB Table %s: %s", *tableName, err)
				continue
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DynamoDB Table sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DynamoDB Tables: %s", err)
	}

	return nil
}

func TestAccAWSDynamoDbTable2019_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table_2019.test"
	tableName := acctest.RandomWithPrefix("TerraformTestTable2019-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbReplicaUpdates(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "attribute.2990477658.name", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "attribute.2990477658.type", "S"),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replica.0.region", "us-west-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbReplicaDeletes(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "attribute.2990477658.name", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "attribute.2990477658.type", "S"),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "0"),
				),
			},
		},
	})
}

func testAccAWSDynamoDbReplicaUpdates(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table_2019" "test" {
  name         = "%s"
  hash_key     = "TestTableHashKey"
	billing_mode = "PAY_PER_REQUEST"
	stream_enabled = true
	stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

	replica {
	  region = "us-west-1"
	}
}
`, rName)
}

func testAccAWSDynamoDbReplicaDeletes(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table_2019" "test" {
  name         = "%s"
  hash_key     = "TestTableHashKey"
	billing_mode = "PAY_PER_REQUEST"
	stream_enabled = true
	stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}
