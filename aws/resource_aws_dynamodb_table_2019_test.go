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

// func testAccCheckInitialAWSDynamoDbTableConf(n string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		log.Printf("[DEBUG] Trying to create initial table state!")
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", n)
// 		}
//
// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("No DynamoDB table name specified!")
// 		}
//
// 		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn
//
// 		params := &dynamodb.DescribeTableInput{
// 			TableName: aws.String(rs.Primary.ID),
// 		}
//
// 		resp, err := conn.DescribeTable(params)
//
// 		if err != nil {
// 			return fmt.Errorf("Problem describing table '%s': %s", rs.Primary.ID, err)
// 		}
//
// 		table := resp.Table
//
// 		log.Printf("[DEBUG] Checking on table %s", rs.Primary.ID)
//
// 		if table.BillingModeSummary != nil && aws.StringValue(table.BillingModeSummary.BillingMode) != dynamodb.BillingModeProvisioned {
// 			return fmt.Errorf("Billing Mode was %s, not %s!", aws.StringValue(table.BillingModeSummary.BillingMode), dynamodb.BillingModeProvisioned)
// 		}
//
// 		if *table.ProvisionedThroughput.WriteCapacityUnits != 2 {
// 			return fmt.Errorf("Provisioned write capacity was %d, not 2!", table.ProvisionedThroughput.WriteCapacityUnits)
// 		}
//
// 		if *table.ProvisionedThroughput.ReadCapacityUnits != 1 {
// 			return fmt.Errorf("Provisioned read capacity was %d, not 1!", table.ProvisionedThroughput.ReadCapacityUnits)
// 		}
//
// 		if table.SSEDescription != nil && *table.SSEDescription.Status != dynamodb.SSEStatusDisabled {
// 			return fmt.Errorf("SSE status was %s, not %s", *table.SSEDescription.Status, dynamodb.SSEStatusDisabled)
// 		}
//
// 		attrCount := len(table.AttributeDefinitions)
// 		gsiCount := len(table.GlobalSecondaryIndexes)
// 		lsiCount := len(table.LocalSecondaryIndexes)
//
// 		if attrCount != 4 {
// 			return fmt.Errorf("There were %d attributes, not 4 like there should have been!", attrCount)
// 		}
//
// 		if gsiCount != 1 {
// 			return fmt.Errorf("There were %d GSIs, not 1 like there should have been!", gsiCount)
// 		}
//
// 		if lsiCount != 1 {
// 			return fmt.Errorf("There were %d LSIs, not 1 like there should have been!", lsiCount)
// 		}
//
// 		attrmap := dynamoDbAttributesToMap(&table.AttributeDefinitions)
// 		if attrmap["TestTableHashKey"] != "S" {
// 			return fmt.Errorf("Test table hash key was of type %s instead of S!", attrmap["TestTableHashKey"])
// 		}
// 		if attrmap["TestTableRangeKey"] != "S" {
// 			return fmt.Errorf("Test table range key was of type %s instead of S!", attrmap["TestTableRangeKey"])
// 		}
// 		if attrmap["TestLSIRangeKey"] != "N" {
// 			return fmt.Errorf("Test table LSI range key was of type %s instead of N!", attrmap["TestLSIRangeKey"])
// 		}
// 		if attrmap["TestGSIRangeKey"] != "S" {
// 			return fmt.Errorf("Test table GSI range key was of type %s instead of S!", attrmap["TestGSIRangeKey"])
// 		}
//
// 		return nil
// 	}
// }

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
