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
	"github.com/hashicorp/terraform-provider-aws/names"
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, "datasource_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "sync_schedule"),
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
				Config: testAccDatasourceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatasourceConfig_tags(rName, acctest.CtKey1, "value1new", acctest.CtKey2, "value2new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, "value1new"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value2new"),
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
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.condition.0.key", "STRING_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.condition.0.operator", "EXISTS"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.condition.0.value.string_list_value.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.document_content_operator", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.target.0.key", "STRING_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.target.0.attribute_value_operator", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.0.target.0.value.string_value", "STRING_VALUE"),

					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.1.condition.0.value.long_value", "1234"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.inline_configuration.1.target.0.value.date_value", "2012-03-25T12:30:10Z"),

					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.lambda_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.s3_bucket"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.key", "STRING_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.operator", "EXISTS"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.pre_extraction_hook_configuration.0.invocation_condition.0.value.string_value", "STRING_VALUE"),

					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.lambda_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.s3_bucket"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.key", "STRING_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.operator", "EXISTS"),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.0.post_extraction_hook_configuration.0.invocation_condition.0.value.string_value", "STRING_VALUE"),
				),
			},
			{
				Config: testAccDatasourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "document_enrichment_configuration.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccQBusinessDatasource_vpcConfiguration(t *testing.T) {
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
				Config: testAccDatasourceConfig_vpcConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.vpc_security_group_ids.0", "sg-12345678"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.0", "subnet-12345678"),
				),
			},
			{
				Config: testAccDatasourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasourceExists(ctx, resourceName, &datasource),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", acctest.Ct0),
				),
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

func testAccDatasourceConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccIndexConfig_basic(rName), fmt.Sprintf(`
data "aws_region" "current" {}
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName))
}

func testAccDatasourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  sync_schedule        = "cron(0 0 * * ? *)"
  description          = %[1]q

  configuration = jsonencode({
    type    = "S3"
    version = "1.0.0"
    connectionConfiguration = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    enableIdentityCrawler = false
    syncMode              = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })
}
`, rName))
}

func testAccDatasourceConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  configuration = jsonencode({
    type = "S3"
    connectionConfiguration = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode = "FULL_CRAWL"
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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDatasourceConfig_documentEnrichmentConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  sync_schedule        = "cron(0 0 * * ? *)"
  description          = %[1]q

  configuration = jsonencode({
    type = "S3"
    connectionConfiguration = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })

  document_enrichment_configuration {

    inline_configuration {
      condition {
        key      = "STRING_VALUE"
        operator = "EXISTS"
        value {
          string_list_value = ["STRING_VALUE", "STRING_VALUE1"]
        }
      }
      document_content_operator = "DELETE"
      target {
        key                      = "STRING_VALUE"
        attribute_value_operator = "DELETE"
        value {
          string_value = "STRING_VALUE"
        }
      }
    }

    inline_configuration {
      condition {
        key      = "STRING_VALUE1"
        operator = "EXISTS"
        value {
          long_value = 1234
        }
      }
      document_content_operator = "DELETE"
      target {
        key                      = "STRING_VALUE"
        attribute_value_operator = "DELETE"
        value {
          date_value = "2012-03-25T12:30:10Z"
        }
      }
    }

    pre_extraction_hook_configuration {
      lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:123456789012:function:my-function"
      role_arn   = aws_iam_role.test.arn
      s3_bucket  = aws_s3_bucket.test.bucket
      invocation_condition {
        key      = "STRING_VALUE"
        operator = "EXISTS"
        value {
          string_value = "STRING_VALUE"
        }
      }
    }

    post_extraction_hook_configuration {
      lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:123456789012:function:my-function"
      role_arn   = aws_iam_role.test.arn
      s3_bucket  = aws_s3_bucket.test.bucket
      invocation_condition {
        key      = "STRING_VALUE"
        operator = "EXISTS"
        value {
          string_value = "STRING_VALUE"
        }
      }
    }

  }
}
`, rName))
}

func testAccDatasourceConfig_vpcConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_datasource" "test" {
  application_id       = aws_qbusiness_app.test.id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
  sync_schedule        = "cron(0 0 * * ? *)"
  description          = %[1]q

  configuration = jsonencode({
    type = "S3"
    connectionConfiguration = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })

  vpc_config {
    vpc_security_group_ids = ["sg-12345678"]
    subnet_ids             = ["subnet-12345678"]
  }
}
`, rName))
}
