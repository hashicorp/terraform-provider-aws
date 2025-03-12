// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rId, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "quicksight", fmt.Sprintf("datasource/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.amazon_elasticsearch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.athena.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.aurora.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.aurora_postgresql.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.aws_iot_analytics.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.databricks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.jira.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.maria_db.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.mysql.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.oracle.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.postgresql.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.presto.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.rds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.redshift.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.service_now.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.snowflake.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.spark.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.sql_server.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.teradata.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.twitter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.DataSourceTypeS3)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
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

func TestAccQuickSightDataSource_permissions(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_permissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
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
						names.AttrPrincipal: regexache.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
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

func TestAccQuickSightDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSourceConfig_updateName(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "updated-name"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.DataSourceTypeS3)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccQuickSightDataSource_secretARN(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_secret_arn(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AURORA_POSTGRESQL"),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "credentials.0.secret_arn"),
				),
			},
		},
	})
}

func TestAccQuickSightDataSource_s3RoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource awstypes.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	iamRoleResourceNameUpdated := "aws_iam_role.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_s3RoleARN(rId, rName, rName2, iamRoleResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttrPair(resourceName, "parameters.0.s3.0.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.DataSourceTypeS3)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// change the selector update the data source with the new Role
			{
				Config: testAccDataSourceConfig_s3RoleARN(rId, rName, rName2, iamRoleResourceNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttrPair(resourceName, "parameters.0.s3.0.role_arn", iamRoleResourceNameUpdated, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.DataSourceTypeS3)),
				),
			},
		},
	})
}

func testAccCheckDataSourceExists(ctx context.Context, n string, v *awstypes.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		output, err := tfquicksight.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_source_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_data_source" {
				continue
			}

			_, err := tfquicksight.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["data_source_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Data Source (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataSourceConfig_base(rName string) string {
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

func testAccDataSourceConfig_baseNoACL(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_object" "test_data" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "%[1]s-test-data.csv"
  content = <<-EOT
name,sentiment
a,happy
b,happy
EOT
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  content = jsonencode({
    fileLocations = [
      {
        URIs = [
          "https://${aws_s3_bucket.test.id}.s3.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/%[1]s-test-data.csv"
        ]
      }
    ]
    globalUploadSettings = {
      format         = "CSV"
      delimiter      = ","
      textqualifier  = "\""
      containsHeader = true
    }
  })
}
`, rName)
}

func testAccDataSourceConfig_basic(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfig_base(rName),
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

func testAccDataSource_UserConfigMultiple(rName string, count int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" "test" {
  count = %[3]d

  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = "%[1]s-${count.index}"
  email          = %[2]q
  identity_type  = "QUICKSIGHT"
  user_role      = "AUTHOR"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, acctest.DefaultEmailAddress, count)
}

func testAccDataSourceConfig_permissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfig_base(rName),
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
		testAccDataSourceConfig_base(rName),
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

func testAccDataSourceConfig_updateName(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfig_base(rName),
		testAccDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = "updated-name"

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
`, rId))
}

func testAccDataSourceConfig_secret_arn(rId, rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "qs-vpc-connnection-tf-test"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_policy" "allec2" {
  name        = "testec2policy"
  description = "Add AmazonEC2FullAccess"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["ec2:*"]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = ["elasticloadbalancing:*"]
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = ["cloudwatch:*"]
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = ["autoscaling:*"]
        Resource = "*"
      },
      {
        Effect   = "Allow",
        Action   = ["iam:CreateServiceLinkedRole"]
        Resource = "*",
        Condition = {
          StringEquals = {
            "iam:AWSServiceName" = [
              "autoscaling.amazonaws.com",
              "ec2scheduled.amazonaws.com",
              "elasticloadbalancing.amazonaws.com",
              "spot.amazonaws.com",
              "spotfleet.amazonaws.com",
              "transitgateway.amazonaws.com"
            ]
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "allec2" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.allec2.arn
}

data "aws_availability_zones" "available" {
  exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
  state            = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "rds-quicksight-tf-vpc" {
  cidr_block                           = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block     = false
  enable_dns_hostnames                 = true
  enable_dns_support                   = true
  instance_tenancy                     = "default"
  enable_network_address_usage_metrics = false
  tags = {
    Name = "rds-quicksight-tf-vpc"
  }
}

resource "aws_subnet" "rds-quicksight-tf-subnet" {
  depends_on        = [aws_security_group.qs-sg-test]
  count             = 2
  vpc_id            = aws_vpc.rds-quicksight-tf-vpc.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.rds-quicksight-tf-vpc.cidr_block, 8, count.index)
  tags = {
    Name = "rds-quicksight-tf-vpc"
  }
}

resource "aws_route_table" "rds-quicksight-tf-private1-rtb" {
  vpc_id = aws_vpc.rds-quicksight-tf-vpc.id
  tags = {
    name = "rds-quicksight-tf-rtb"
  }
}
resource "aws_route_table_association" "rds-quicksight-tf-private1-rtb-asso" {
  subnet_id      = aws_subnet.rds-quicksight-tf-subnet[0].id
  route_table_id = aws_route_table.rds-quicksight-tf-private1-rtb.id
}

resource "aws_route_table" "rds-quicksight-tf-private2-rtb" {
  vpc_id = aws_vpc.rds-quicksight-tf-vpc.id
  tags = {
    name = "rds-quicksight-tf-rtb"
  }
}
resource "aws_route_table_association" "rds-quicksight-tf-private2-rtb-asso" {
  subnet_id      = aws_subnet.rds-quicksight-tf-subnet[1].id
  route_table_id = aws_route_table.rds-quicksight-tf-private2-rtb.id
}


resource "aws_security_group" "rds-sg-test" {
  name   = "Amazon-QuickSight-RDS-VPC"
  vpc_id = aws_vpc.rds-quicksight-tf-vpc.id
}
resource "aws_security_group" "qs-sg-test" {
  name   = "Amazon-QuickSight-QS-VPC"
  vpc_id = aws_vpc.rds-quicksight-tf-vpc.id
}

resource "aws_vpc_security_group_ingress_rule" "rds-sg-test-ingress" {
  security_group_id            = aws_security_group.rds-sg-test.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "TCP"
  referenced_security_group_id = aws_security_group.qs-sg-test.id
}

resource "aws_vpc_security_group_egress_rule" "rds-sg-test-egress" {
  security_group_id            = aws_security_group.rds-sg-test.id
  from_port                    = 0
  to_port                      = 65535
  ip_protocol                  = "TCP"
  referenced_security_group_id = aws_security_group.qs-sg-test.id
}

resource "aws_vpc_security_group_ingress_rule" "qs-sg-test-ingress" {
  security_group_id            = aws_security_group.qs-sg-test.id
  from_port                    = 0
  to_port                      = 65535
  ip_protocol                  = "TCP"
  referenced_security_group_id = aws_security_group.rds-sg-test.id
}

resource "aws_vpc_security_group_egress_rule" "qs-sg-test-egress" {
  security_group_id            = aws_security_group.qs-sg-test.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "TCP"
  referenced_security_group_id = aws_security_group.rds-sg-test.id
}

resource "aws_rds_cluster" "qs-rds-tf-test-cluster" {
  cluster_identifier      = "quicksight-vpc-tf-test"
  engine                  = "aurora-postgresql"
  engine_version          = 13.12
  database_name           = "qsrdstftestcluster"
  master_username         = "foo"
  master_password         = "must_be_eight_characters"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  vpc_security_group_ids  = [aws_security_group.rds-sg-test.id]
  skip_final_snapshot     = true
  db_subnet_group_name    = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "qs-rds-tf-test-cluster-instance" {
  identifier                   = "aurora-cluster"
  cluster_identifier           = aws_rds_cluster.qs-rds-tf-test-cluster.id
  instance_class               = "db.r5.large"
  engine                       = aws_rds_cluster.qs-rds-tf-test-cluster.engine
  engine_version               = aws_rds_cluster.qs-rds-tf-test-cluster.engine_version
  performance_insights_enabled = false
}

resource "aws_db_subnet_group" "test" {
  depends_on = [aws_security_group.qs-sg-test]
  name       = "quicksight-vpc-connnection-test"
  subnet_ids = aws_subnet.rds-quicksight-tf-subnet[*].id
  tags = {
    Name = "quicksight-vpc-connnection-test"
  }
}

resource "aws_quicksight_vpc_connection" "qs-rds-vpc-conn-test" {
  depends_on        = [aws_security_group.qs-sg-test, aws_iam_role_policy_attachment.allec2, aws_iam_policy.allec2]
  vpc_connection_id = %[2]q
  name              = %[2]q
  role_arn          = aws_iam_role.test.arn
  subnet_ids        = aws_subnet.rds-quicksight-tf-subnet[*].id
  security_group_ids = [
    aws_security_group.qs-sg-test.id,
  ]
}

resource "aws_secretsmanager_secret" "qs-secret-test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.qs-secret-test.id
  secret_string = jsonencode({
    username = "foo",
    password = "must_be_eight_characters"
  })
}

resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q
  vpc_connection_properties {
    vpc_connection_arn = aws_quicksight_vpc_connection.qs-rds-vpc-conn-test.arn
  }
  credentials {
    secret_arn = aws_secretsmanager_secret.qs-secret-test.arn
  }
  parameters {
    rds {
      database    = aws_rds_cluster.qs-rds-tf-test-cluster.database_name
      instance_id = aws_rds_cluster_instance.qs-rds-tf-test-cluster-instance.identifier
    }
  }
  type = "AURORA_POSTGRESQL"
}
`, rId, rName)
}

func testAccDataSourceConfig_s3RoleARN(rId, rName, rName2, iamRoleResourceName string) string {
	return acctest.ConfigCompose(
		testAccDataSourceConfig_baseNoACL(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = %[2]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role" "test2" {
  name = %[3]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "test" {
  name        = %[2]q
  description = "Policy to allow QuickSight access to S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["s3:GetObject"],
        Effect   = "Allow",
        Resource = "${aws_s3_bucket.test.arn}/${aws_s3_object.test.key}"
      },
      {
        Action   = ["s3:ListBucket"],
        Effect   = "Allow",
        Resource = aws_s3_bucket.test.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = aws_iam_policy.test.arn
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "test2" {
  policy_arn = aws_iam_policy.test.arn
  role       = aws_iam_role.test2.name
}

resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_object.test.key
      }
      role_arn = %[4]s.arn
    }
  }

  type = "S3"

  depends_on = [
    aws_iam_role_policy_attachment.test
  ]
}
`, rId, rName, rName2, iamRoleResourceName))
}
