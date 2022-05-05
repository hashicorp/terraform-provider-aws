package dynamodb_test

import (
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDynamoDBGlobalTable_basic(t *testing.T) {
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckGlobalTable(t)
			testAccGlobalTablePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGlobalTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalTableConfig_invalidName(sdkacctest.RandString(2)),
				ExpectError: regexp.MustCompile("name length must be between 3 and 255 characters"),
			},
			{
				Config:      testAccGlobalTableConfig_invalidName(sdkacctest.RandString(256)),
				ExpectError: regexp.MustCompile("name length must be between 3 and 255 characters"),
			},
			{
				Config:      testAccGlobalTableConfig_invalidName("!!!!"),
				ExpectError: regexp.MustCompile("name must only include alphanumeric, underscore, period, or hyphen characters"),
			},
			{
				Config: testAccGlobalTableConfig_basic(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDynamoDBGlobalTable_multipleRegions(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckGlobalTable(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccGlobalTablePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckGlobalTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalTableConfig_multipleRegions1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
				),
			},
			{
				Config:            testAccGlobalTableConfig_multipleRegions1(tableName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalTableConfig_multipleRegions2(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
			{
				Config: testAccGlobalTableConfig_multipleRegions1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func testAccCheckGlobalTableDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_global_table" {
			continue
		}

		input := &dynamodb.DescribeGlobalTableInput{
			GlobalTableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeGlobalTable(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeGlobalTableNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected DynamoDB Global Table to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckGlobalTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccPreCheckGlobalTable(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

	input := &dynamodb.ListGlobalTablesInput{}

	_, err := conn.ListGlobalTables(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// testAccGlobalTablePreCheck checks if aws_dynamodb_global_table (version 2017.11.29) can be used and skips test if not.
// Region availability for Version 2017.11.29: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GlobalTables.html
func testAccGlobalTablePreCheck(t *testing.T) {
	supportRegionsSort := []string{
		endpoints.ApNortheast1RegionID,
		endpoints.ApNortheast2RegionID,
		endpoints.ApSoutheast1RegionID,
		endpoints.ApSoutheast2RegionID,
		endpoints.EuCentral1RegionID,
		endpoints.EuWest1RegionID,
		endpoints.EuWest2RegionID,
		endpoints.UsEast1RegionID,
		endpoints.UsEast2RegionID,
		endpoints.UsWest1RegionID,
		endpoints.UsWest2RegionID,
	}

	if acctest.Region() != supportRegionsSort[sort.SearchStrings(supportRegionsSort, acctest.Region())] {
		t.Skipf("skipping test; aws_dynamodb_global_table (DynamoDB v2017.11.29) not supported in region %s", acctest.Region())
	}
}

func testAccGlobalTableConfig_basic(tableName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
}

resource "aws_dynamodb_table" "test" {
  hash_key         = "myAttribute"
  name             = "%s"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_global_table" "test" {
  depends_on = [aws_dynamodb_table.test]

  name = "%s"

  replica {
    region_name = data.aws_region.current.name
  }
}
`, tableName, tableName)
}

func testAccGlobalTableConfig_MultipleRegions_tables(tableName string) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "current" {}

resource "aws_dynamodb_table" "test" {
  hash_key         = "myAttribute"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_table" "alternate" {
  provider = "awsalternate"

  hash_key         = "myAttribute"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}
`, tableName)
}

func testAccGlobalTableConfig_multipleRegions1(tableName string) string {
	return testAccGlobalTableConfig_MultipleRegions_tables(tableName) + `
resource "aws_dynamodb_global_table" "test" {
  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.current.name
  }
}
`
}

func testAccGlobalTableConfig_multipleRegions2(tableName string) string {
	return testAccGlobalTableConfig_MultipleRegions_tables(tableName) + `
resource "aws_dynamodb_global_table" "test" {
  depends_on = [aws_dynamodb_table.alternate]

  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name = data.aws_region.current.name
  }
}
`
}

func testAccGlobalTableConfig_invalidName(tableName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_dynamodb_global_table" "test" {
  name = "%s"

  replica {
    region_name = data.aws_region.current.name
  }
}
`, tableName))
}
