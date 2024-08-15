// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataLake_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "S3_MANAGED_KEY"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "s3_bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceDataLake, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataLake_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
			{
				Config: testAccDataLakeConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataLakeConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDataLake_lifeCycle(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_lifeCycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.days", "80"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_lifeCycleUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_lifeCycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.days", "80"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
			{
				Config: testAccDataLakeConfig_lifeCycleUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_replication(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.region_2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_replication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "S3_MANAGED_KEY"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.replication_configuration.0.role_arn", "aws_iam_role.datalake_s3_replication", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.0.regions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.replication_configuration.0.regions.*", acctest.Region()),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccCheckDataLakeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_data_lake" {
				continue
			}

			_, err := tfsecuritylake.FindDataLakeByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Data Lake %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataLakeExists(ctx context.Context, n string, v *types.DataLakeResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindDataLakeByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccDataLakeConfigConfig_base = `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "meta_store_manager" {
  name               = "AmazonSecurityLakeMetaStoreManagerV2"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowLambda",
    "Effect": "Allow",
    "Principal": {
      "Service": [
        "lambda.amazonaws.com"
      ]
    },
    "Action": "sts:AssumeRole"
  }]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "datalake" {
  role       = aws_iam_role.meta_store_manager.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonSecurityLakeMetastoreManager"
}

resource "aws_iam_role" "datalake_s3_replication" {
  name               = "AmazonSecurityLakeS3ReplicationRole"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "s3.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

resource "aws_iam_role_policy" "datalake_s3_replication" {
  name = "AmazonSecurityLakeS3ReplicationRolePolicy"
  role = aws_iam_role.datalake_s3_replication.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowReadS3ReplicationSetting",
      "Action": [
        "s3:ListBucket",
        "s3:GetReplicationConfiguration",
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersion",
        "s3:GetObjectVersionAcl",
        "s3:GetObjectVersionTagging",
        "s3:GetObjectRetention",
        "s3:GetObjectLegalHold"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*",
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:ResourceAccount": [
            "${data.aws_caller_identity.current.account_id}"
          ]
        }
      }
    },
    {
      "Sid": "AllowS3Replication",
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete",
        "s3:ReplicateTags"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:ResourceAccount": [
            "${data.aws_caller_identity.current.account_id}"
          ]
        }
      }
    }
  ]
}
POLICY
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [{
    "Sid": "Enable IAM User Permissions",
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": "kms:*",
    "Resource": "*"
  }]
}
POLICY
}
`

func testAccDataLakeConfig_basic() string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[1]q
  }

  depends_on = [aws_iam_role.meta_store_manager]
}
`, acctest.Region()))
}

func testAccDataLakeConfig_tags1(tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[3]q
  }

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_iam_role.meta_store_manager]
}
`, tag1Key, tag1Value, acctest.Region()))
}

func testAccDataLakeConfig_tags2(tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[5]q
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_iam_role.meta_store_manager]
}
`, tag1Key, tag1Value, tag2Key, tag2Value, acctest.Region()))
}

func testAccDataLakeConfig_lifeCycle(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[2]q

    encryption_configuration {
      kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      transition {
        days          = 80
        storage_class = "ONEZONE_IA"
      }
      expiration {
        days = 300
      }
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role.meta_store_manager]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_lifeCycleUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[2]q

    encryption_configuration {
      kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role.meta_store_manager]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_replication(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_data_lake" "region_2" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[3]q

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }

    replication_configuration {
      role_arn = aws_iam_role.datalake_s3_replication.arn
      regions  = [%[2]q]
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role.meta_store_manager, aws_iam_role.datalake_s3_replication, aws_securitylake_data_lake.test]
}
`, rName, acctest.Region(), acctest.AlternateRegion()))
}
