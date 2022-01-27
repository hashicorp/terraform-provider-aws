package dynamodb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

//TODO: Fix Cyclic dependency of attribute and global_secondary_index update
// After applying this test step and performing a `terraform refresh`, the plan was not empty.
func TestAccAWSDynamoDbTableGsi_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table_gsi.gsi"
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	gsiName := fmt.Sprintf("test-gsi-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.table", "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "name", gsiName),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestGsiHashKey"),
					resource.TestCheckResourceAttr(resourceName, "projection_type", "KEYS_ONLY"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_updateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	gsiName := fmt.Sprintf("test-gsi-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbGsiConfigBasicMoreCapacity(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "5"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "5"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_updateNonAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_dynamodb_table_gsi.gsi"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiInitialOtherAttributes(tableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr(resourceName, "non_key_attributes.0", "SomeId"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbGsiConfigUpdatedOtherAttributes(tableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr(resourceName, "non_key_attributes.0", "RandomAttribute"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_delete(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	gsiName := fmt.Sprintf("test-gsi-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.table", "name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbGsiConfigBasicDelete(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialTableExists("aws_dynamodb_table.table", &conf),
					func(*terraform.State) error {
						if len(conf.Table.GlobalSecondaryIndexes) > 0 {
							return fmt.Errorf("expected to find no global secondary indexes, instead found: %v", conf.Table.GlobalSecondaryIndexes)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccBasicDynamoTable(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name                         = "%s"
  read_capacity                = 1
  write_capacity               = 1
  hash_key                     = "id"
  manage_index_as_own_resource = true

  attribute {
    name = "id"
    type = "S"
  }
}
`, tableName)
}

func testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName string) string {
	return testAccBasicDynamoTable(tableName) + fmt.Sprintf(`
resource "aws_dynamodb_table_gsi" "gsi" {
  table_name      = aws_dynamodb_table.table.id
  name            = "%s"
  hash_key        = "TestGsiHashKey"
  write_capacity  = 1
  read_capacity   = 1
  projection_type = "KEYS_ONLY"
  attribute {
    name = "TestGsiHashKey"
    type = "S"
  }
}
`, gsiName)
}

func testAccAWSDynamoDbGsiConfigBasicMoreCapacity(tableName, gsiName string) string {
	return testAccBasicDynamoTable(tableName) + fmt.Sprintf(`
resource "aws_dynamodb_table_gsi" "gsi" {
  table_name      = aws_dynamodb_table.table.id
  name            = "%s"
  hash_key        = "TestGsiHashKey"
  write_capacity  = 5
  read_capacity   = 5
  projection_type = "KEYS_ONLY"

  attribute {
    name = "TestGsiHashKey"
    type = "S"
  }
}
`, gsiName)
}

func testAccAWSDynamoDbGsiConfigBasicDelete(tableName string) string {
	return testAccBasicDynamoTable(tableName)
}

func testAccAWSDynamoDbGsiInitialOtherAttributes(tableName string) string {
	return testAccBasicDynamoTable(tableName) + (`
resource "aws_dynamodb_table_gsi" "gsi" {
  table_name         = aws_dynamodb_table.table.id
  name               = "gsi1-index"
  hash_key           = "att1"
  write_capacity     = "1"
  read_capacity      = "1"
  projection_type    = "INCLUDE"
  non_key_attributes = ["SomeId"]
  attribute {
    name = "att1"
    type = "S"
  }
}
`)
}

func testAccAWSDynamoDbGsiConfigUpdatedOtherAttributes(tableName string) string {
	return testAccBasicDynamoTable(tableName) + `
resource "aws_dynamodb_table_gsi" "gsi" {
  table_name         = aws_dynamodb_table.table.id
  name               = "gsi1-index"
  hash_key           = "att1"
  write_capacity     = "1"
  read_capacity      = "1"
  projection_type    = "INCLUDE"
  non_key_attributes = ["RandomAttribute"]
  attribute {
    name = "att1"
    type = "S"
  }
}
`
}
