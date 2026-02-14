// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCatalogID, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "athena_properties.%", resourceName, "athena_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.%", resourceName, "connection_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.#", resourceName, "physical_connection_requirements.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "match_criteria.#", resourceName, "match_criteria.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccGlueConnectionDataSource_dynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bucketName := "tf-acc-test-" + sdkacctest.RandString(26)
	region := acctest.Region()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_dynamoDB(rName, region, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCatalogID, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "athena_properties.%", resourceName, "athena_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.%", resourceName, "connection_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.#", resourceName, "physical_connection_requirements.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "match_criteria.#", resourceName, "match_criteria.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

}

data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName, jdbcConnectionUrl)
}

func testAccConnectionDataSourceConfig_dynamoDB(rName, region, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[3]q
}

data "aws_partition" "current" {}

resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_type = "DYNAMODB"
  athena_properties = {
    lambda_function_arn      = "arn:${data.aws_partition.current.partition}:lambda:%[2]s:123456789012:function:athenafederatedcatalog_athena_abcdefgh"
    disable_spill_encryption = "false"
    spill_bucket             = aws_s3_bucket.test.bucket
  }
}
data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName, region, bucketName)
}
