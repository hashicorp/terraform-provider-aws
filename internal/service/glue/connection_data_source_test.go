// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

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
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_configuration.#", resourceName, "authentication_configuration.#"),
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
	bucketName := "tf-acc-test-" + acctest.RandString(t, 26)
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
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_configuration.#", resourceName, "authentication_configuration.#"),
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

func TestAccGlueConnectionDataSource_mySQL(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_mySQL(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.%", resourceName, "connection_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.HOST", resourceName, "connection_properties.HOST"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.PORT", resourceName, "connection_properties.PORT"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties.DATABASE", resourceName, "connection_properties.DATABASE"),
					resource.TestCheckResourceAttrPair(datasourceName, "athena_properties.%", resourceName, "athena_properties.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "athena_properties.lambda_function_arn", resourceName, "athena_properties.lambda_function_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "athena_properties.spill_bucket", resourceName, "athena_properties.spill_bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_configuration.#", resourceName, "authentication_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_configuration.0.authentication_type", resourceName, "authentication_configuration.0.authentication_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_configuration.0.secret_arn", resourceName, "authentication_configuration.0.secret_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.#", resourceName, "physical_connection_requirements.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.0.availability_zone", resourceName, "physical_connection_requirements.0.availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.0.security_group_id_list.#", resourceName, "physical_connection_requirements.0.security_group_id_list.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements.0.subnet_id", resourceName, "physical_connection_requirements.0.subnet_id"),
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

func testAccConnectionDataSourceConfig_mySQL(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-glue-connection-mysql-ds"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = "tcp"
    self      = true
    from_port = 1
    to_port   = 65535
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-glue-connection-mysql-ds"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
}

resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    username = "glueusername"
    password = "gluepassword"
  })
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_glue_connection" "test" {
  name = "%[1]s"

  connection_type = "MYSQL"

  connection_properties = {
    HOST     = "testhost"
    PORT     = "3306"
    DATABASE = "gluedatabase"
  }

  athena_properties = {
    lambda_function_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:123456789012:function:athenafederatedcatalog_mysql_abcdefgh"
    spill_bucket        = aws_s3_bucket.test.bucket
  }

  authentication_configuration {
    authentication_type = "BASIC"
    secret_arn          = aws_secretsmanager_secret.test.arn
  }

  physical_connection_requirements {
    availability_zone      = aws_subnet.test.availability_zone
    security_group_id_list = [aws_security_group.test.id]
    subnet_id              = aws_subnet.test.id
  }
}

data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName)
}
