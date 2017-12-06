package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDynamoDbBackupSnapshot_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamoDbBackupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbBackupExists("aws_dynamodb_backup.test"),
				),
			},
		},
	})
}

func testAccCheckDynamoDbBackupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		request := &dynamodb.DescribeBackupInput{
			BackupArn: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeBackup(request)
		if err == nil {
			if response.BackupDescription.BackupDetails.BackupStatus == aws.String("COMPLETE") {
				return nil
			}
		}
		return fmt.Errorf("Error finding RDS DB Snapshot %s", rs.Primary.ID)
	}
}

func testAccAwsDynamoDbBackupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 10
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
}

resource "aws_dynamodb_backup" "test" {
	table_name = "${aws_dynamodb_table.basic-dynamodb-table.id}"
	backup_name = "terraformtesting"
}
`, rName)
}
