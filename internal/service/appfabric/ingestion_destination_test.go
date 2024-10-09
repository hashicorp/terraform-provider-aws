// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIngestionDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestiondestination awstypes.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttrSet(resourceName, "app_bundle_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.bucket_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.prefix"),
					resource.TestCheckResourceAttrSet(resourceName, "ingestion_arn"),
					resource.TestCheckResourceAttr(resourceName, "processing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "processing_configuration.0.audit_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "processing_configuration.0.audit_log.0.format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "processing_configuration.0.audit_log.0.schema", "raw"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccIngestionDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestiondestination awstypes.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestionDestination, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIngestionDestination_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestiondestination awstypes.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIngestionDestinationConfig_tags2(rName, tenantID, serviceAccountToken, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIngestionDestinationConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccIngestionDestination_update(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestiondestination awstypes.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.bucket_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.prefix"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIngestionDestinationConfig_s3Prefix(rName, tenantID, serviceAccountToken, "testing"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.0.prefix", "testing"),
				),
			},
		},
	})
}

func testAccIngestionDestination_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestiondestination awstypes.IngestionDestination
	resourceName := "aws_appfabric_ingestion_destination.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionDestinationConfig_firehose(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionDestinationExists(ctx, resourceName, &ingestiondestination),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "destination_configuration.0.audit_log.0.destination.0.firehose_stream.0.stream_name"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.audit_log.0.destination.0.s3_bucket.#", acctest.Ct0),
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

func testAccCheckIngestionDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion_destination" {
				continue
			}

			_, err := tfappfabric.FindIngestionDestinationByThreePartKey(ctx, conn, rs.Primary.Attributes["app_bundle_arn"], rs.Primary.Attributes["ingestion_arn"], rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppFabric Ingestion Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIngestionDestinationExists(ctx context.Context, n string, v *awstypes.IngestionDestination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		output, err := tfappfabric.FindIngestionDestinationByThreePartKey(ctx, conn, rs.Primary.Attributes["app_bundle_arn"], rs.Primary.Attributes["ingestion_arn"], rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken string) string {
	return acctest.ConfigCompose(testAccIngestionConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
  app            = aws_appfabric_app_authorization_connection.test.app
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  tenant_id      = %[2]q
  ingestion_type = "auditLog"

  tags = {
    Name = %[1]q
  }
}
`, rName, tenantID))
}

func testAccIngestionDestinationConfig_basic(rName, tenantID, serviceAccountToken string) string {
	return acctest.ConfigCompose(testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  ingestion_arn  = aws_appfabric_ingestion.test.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        s3_bucket {
          bucket_name = aws_s3_bucket.test.bucket
        }
      }
    }
  }
}
`, rName))
}

func testAccIngestionDestinationConfig_s3Prefix(rName, tenantID, serviceAccountToken, prefix string) string {
	return acctest.ConfigCompose(testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  ingestion_arn  = aws_appfabric_ingestion.test.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        s3_bucket {
          bucket_name = aws_s3_bucket.test.bucket
          prefix      = %[2]q
        }
      }
    }
  }
}
`, rName, prefix))
}

func testAccIngestionDestinationConfig_tags1(rName, tenantID, serviceAccountToken, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  ingestion_arn  = aws_appfabric_ingestion.test.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        s3_bucket {
          bucket_name = aws_s3_bucket.test.bucket
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccIngestionDestinationConfig_tags2(rName, tenantID, serviceAccountToken, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  ingestion_arn  = aws_appfabric_ingestion.test.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        s3_bucket {
          bucket_name = aws_s3_bucket.test.bucket
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccIngestionDestinationConfig_baseFirehose(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }

  tags = {
    AWSAppFabricManaged = "placeholder"
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to AWSAppFabricManaged tag as API adds this tag when ingestion destination is created
      tags["AWSAppFabricManaged"],
    ]
  }
}
`, rName)
}

func testAccIngestionDestinationConfig_firehose(rName, tenantID, serviceAccountToken string) string {
	return acctest.ConfigCompose(testAccIngestionDestinationConfig_base(rName, tenantID, serviceAccountToken), testAccIngestionDestinationConfig_baseFirehose(rName), fmt.Sprintf(`
resource "aws_appfabric_ingestion_destination" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  ingestion_arn  = aws_appfabric_ingestion.test.arn

  processing_configuration {
    audit_log {
      format = "json"
      schema = "raw"
    }
  }

  destination_configuration {
    audit_log {
      destination {
        firehose_stream {
          stream_name = aws_kinesis_firehose_delivery_stream.test.name
        }
      }
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
