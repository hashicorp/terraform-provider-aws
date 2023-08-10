// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBGlobalTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dynamodb_global_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckGlobalTable(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalTableDestroy(ctx),
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
				Config: testAccGlobalTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_dynamodb_global_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckGlobalTable(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckGlobalTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalTableConfig_multipleRegions1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(ctx, resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
			{
				Config:            testAccGlobalTableConfig_multipleRegions1(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalTableConfig_multipleRegions2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
			{
				Config: testAccGlobalTableConfig_multipleRegions1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalTableExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func testAccCheckGlobalTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_global_table" {
				continue
			}

			_, err := tfdynamodb.FindGlobalTableByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Global Table %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGlobalTableExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB Global Table ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)

		_, err := tfdynamodb.FindGlobalTableByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheckGlobalTable(ctx context.Context, t *testing.T) {
	// Region availability for Version 2017.11.29: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GlobalTables.html
	supportedRegions := []string{
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
	acctest.PreCheckRegion(t, supportedRegions...)

	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)

	input := &dynamodb.ListGlobalTablesInput{}

	_, err := conn.ListGlobalTablesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalTableConfig_basic(rName string) string {
	return fmt.Sprintf(`
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

resource "aws_dynamodb_global_table" "test" {
  depends_on = [aws_dynamodb_table.test]

  name = %[1]q

  replica {
    region_name = data.aws_region.current.name
  }
}
`, rName)
}

func testAccGlobalTableConfig_baseMultipleRegions(tableName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
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
`, tableName))
}

func testAccGlobalTableConfig_multipleRegions1(tableName string) string {
	return acctest.ConfigCompose(testAccGlobalTableConfig_baseMultipleRegions(tableName), `
resource "aws_dynamodb_global_table" "test" {
  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.current.name
  }
}
`)
}

func testAccGlobalTableConfig_multipleRegions2(tableName string) string {
	return acctest.ConfigCompose(testAccGlobalTableConfig_baseMultipleRegions(tableName), `
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
`)
}

func testAccGlobalTableConfig_invalidName(tableName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_dynamodb_global_table" "test" {
  name = %[1]q

  replica {
    region_name = data.aws_region.current.name
  }
}
`, tableName))
}
