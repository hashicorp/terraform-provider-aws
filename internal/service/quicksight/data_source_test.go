// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("datasource/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "type", quicksight.DataSourceTypeS3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccQuickSightDataSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceDataSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_tags1(rId, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_tags2(rId, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataSourceConfig_tags1(rId, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccQuickSightDataSource_permissions(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_permissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSourcePermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSource"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_updatePermissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSourcePermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DeleteDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSourcePermissions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDataSourceExists(ctx context.Context, resourceName string, dataSource *quicksight.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSourceId, err := tfquicksight.ParseDataSourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)

		input := &quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceId),
		}

		output, err := conn.DescribeDataSourceWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || output.DataSource == nil {
			return fmt.Errorf("QuickSight Data Source (%s) not found", rs.Primary.ID)
		}

		*dataSource = *output.DataSource

		return nil
	}
}

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_data_source" {
				continue
			}

			awsAccountID, dataSourceId, err := tfquicksight.ParseDataSourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			output, err := conn.DescribeDataSourceWithContext(ctx, &quicksight.DescribeDataSourceInput{
				AwsAccountId: aws.String(awsAccountID),
				DataSourceId: aws.String(dataSourceId),
			})

			if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			if output != nil && output.DataSource != nil {
				return fmt.Errorf("QuickSight Data Source (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]

  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_object" "test_data" {
  depends_on = [aws_s3_bucket_acl.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-test-data"
  content = <<EOF
[
	{
		"Column1": "aaa",
		"Column2": 1
	},
	{
		"Column1": "bbb",
		"Column2": 1
	}
]
  EOF
  acl     = "public-read"
}

resource "aws_s3_object" "test" {
  depends_on = [aws_s3_bucket_acl.test]

  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = <<EOF
{
  "fileLocations": [
      {
          "URIs": [
              "https://${aws_s3_bucket.test.bucket}.s3.${data.aws_partition.current.dns_suffix}/%[1]s-test-data"
          ]
      }
  ],
  "globalUploadSettings": {
      "format": "JSON"
  }
}
EOF
  acl     = "public-read"
}
`, rName)
}

func testAccDataSourceConfig_basic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  type = "S3"
}
`, rId, rName))
}

func testAccDataSourceConfig_tags1(rId, rName, key, value string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  tags = {
    %[3]q = %[4]q
  }

  type = "S3"
}
`, rId, rName, key, value))
}

func testAccDataSourceConfig_tags2(rId, rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
  type = "S3"
}
`, rId, rName, key1, value1, key2, value2))
}

func testAccDataSource_UserConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" "test" {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  identity_type  = "QUICKSIGHT"
  user_role      = "AUTHOR"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccDataSourceConfig_permissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  permission {
    actions = [
      "quicksight:DescribeDataSource",
      "quicksight:DescribeDataSourcePermissions",
      "quicksight:PassDataSource"
    ]

    principal = aws_quicksight_user.test.arn
  }

  type = "S3"
}
`, rId, rName))
}

func testAccDataSourceConfig_updatePermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
    }
  }

  permission {
    actions = [
      "quicksight:DescribeDataSource",
      "quicksight:DescribeDataSourcePermissions",
      "quicksight:PassDataSource",
      "quicksight:UpdateDataSource",
      "quicksight:DeleteDataSource",
      "quicksight:UpdateDataSourcePermissions"
    ]

    principal = aws_quicksight_user.test.arn
  }

  type = "S3"
}
`, rId, rName))
}
