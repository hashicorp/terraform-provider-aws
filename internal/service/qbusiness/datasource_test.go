// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccQBusinessDatasource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var datasource qbusiness.GetDataSourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_datasource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDatasource(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, "datasource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
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

func TestAccQBusinessDatasource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var datasource qbusiness.GetDataSourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_datasource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDatasource(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceDatasource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessDatasource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var datasource qbusiness.GetDataSourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_datasource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDatasource(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceConfig_tags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatasourceConfig_tags(rName, "key1", "value1new", "key2", "value2new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1new"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2new"),
				),
			},
		},
	})
}

func TestAccQBusinessDatasource_documentEnrichmentConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var datasource qbusiness.GetDataSourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_datasource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDatasource(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceConfig_documentEnrichmentConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configurations.0.configuration.0.condition.0.key", "STRING_VALUE"),
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

func testAccPreCheckDatasource(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

	input := &qbusiness.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckDatasourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_datasource" {
				continue
			}

			_, err := tfqbusiness.FindDatasourceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q Datasource %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatasourceExists(ctx context.Context, n string, v *qbusiness.GetDataSourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindDatasourceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDatasourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  configuration        = jsonencode({
    type                     = "S3"
    connectionConfiguration  = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode                 = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  force_destroy = true
}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
    {
    "Action": "sts:AssumeRole",
    "Principal": {
        "Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow",
    "Sid": ""
    }
    ]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
}
`, rName)
}

func testAccDatasourceConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  configuration        = jsonencode({
    type                     = "S3"
    connectionConfiguration  = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode                 = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  force_destroy = true
}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
    {
    "Action": "sts:AssumeRole",
    "Principal": {
        "Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow",
    "Sid": ""
    }
    ]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDatasourceConfig_documentEnrichmentConfiguration(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  configuration        = jsonencode({
    type                     = "S3"
    connectionConfiguration  = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode                 = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })

  document_enrichment_configuration {
    inline_configurations {
      configuration {

        condition {
          key = "STRING_VALUE"
          operator = "EXISTS"
          value {
            string_list_value = ["STRING_VALUE", "STRING_VALUE1"]
          }
        }

        document_content_operator = "DELETE"

        target {
          key = "STRING_VALUE"
          attribute_value_operator = "DELETE"
          value {
            string_value = "STRING_VALUE"
          }
        }

      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  force_destroy = true
}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
    {
    "Action": "sts:AssumeRole",
    "Principal": {
        "Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
    },
    "Effect": "Allow",
    "Sid": ""
    }
    ]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
}
`, rName)
}
