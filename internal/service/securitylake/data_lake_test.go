// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeDataLake_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn", "tags"},
			},
		},
	})
}

func TestAccSecurityLakeDataLake_lifeCycle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_lifeCycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.encryption_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configurations.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.1.days", "80"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.1.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn", "tags"},
			},
		},
	})
}

func TestAccSecurityLakeDataLake_lifeCycleUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_lifeCycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.encryption_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configurations.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.1.days", "80"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.1.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn", "tags"},
			},
			{
				Config: testAccDataLakeConfig_lifeCycleUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.encryption_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configurations.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn", "tags"},
			},
		},
	})
}

func TestAccSecurityLakeDataLake_replication(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.region_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_replication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.transitions.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.lifecycle_configuration.0.expiration.0.days", "300"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.replication_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configurations.0.replication_configuration.0.role_arn", "aws_iam_role.datalake_s3_replication", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configurations.0.replication_configuration.0.regions.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn", "tags"},
			},
		},
	})
}

func TestAccSecurityLakeDataLake_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLake),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, resourceName, &datalake),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceDataLake, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			_, err := tfsecuritylake.FindDataLakeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameDataLake, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDataLakeExists(ctx context.Context, name string, datalake *types.DataLakeResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)
		resp, err := tfsecuritylake.FindDataLakeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameDataLake, rs.Primary.ID, err)
		}

		*datalake = *resp

		return nil
	}
}

func testAccDataLakeConfigBaseConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`




data "aws_caller_identity" "current" {}

resource "aws_iam_role" "meta_store_manager" {
  name               = "AmazonSecurityLakeMetaStoreManager"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
"Version": "2012-10-17",
"Statement": [
	{
	"Sid": "AllowLambda",
	"Effect": "Allow",
	"Principal": {
		"Service": [
		"lambda.amazonaws.com"
		]
	},
	"Action": "sts:AssumeRole"
	}
]
}
POLICY
  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "meta_store_manager" {
  name = "AmazonSecurityLakeMetaStoreManagerPolicy"
  role = aws_iam_role.meta_store_manager.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowWriteLambdaLogs",
		"Effect": "Allow",
		"Action": [
			"logs:CreateLogStream",
			"logs:PutLogEvents"
		],
		"Resource": [
			"arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/SecurityLake_Glue_Partition_Updater_Lambda*"
		]
		},
		{
		"Sid": "AllowCreateAwsCloudWatchLogGroup",
		"Effect": "Allow",
		"Action": [
			"logs:CreateLogGroup"
		],
		"Resource": [
			"arn:aws:logs:*:${data.aws_caller_identity.current.account_id}:/aws/lambda/SecurityLake_Glue_Partition_Updater_Lambda*"
		]
		},
		{
		"Sid": "AllowGlueManage",
		"Effect": "Allow",
		"Action": [
			"glue:CreatePartition",
			"glue:BatchCreatePartition"
		],
		"Resource": [
			"arn:aws:glue:*:*:table/amazon_security_lake_glue_db*/*",
			"arn:aws:glue:*:*:database/amazon_security_lake_glue_db*",
			"arn:aws:glue:*:*:catalog"
		]
		},
		{
		"Sid": "AllowToReadFromSqs",
		"Effect": "Allow",
		"Action": [
			"sqs:ReceiveMessage",
			"sqs:DeleteMessage",
			"sqs:GetQueueAttributes"
		],
		"Resource": [
			"arn:aws:sqs:*:${data.aws_caller_identity.current.account_id}:SecurityLake*"
		]
		}
	]
}
  EOF
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
  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "datalake_s3_replication" {
  name = "AmazonSecurityLakeS3ReplicationRolePolicy"
  role = aws_iam_role.datalake_s3_replication.name

  policy = <<EOF
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
				"arn:aws:s3:::aws-security-data-lake*",
				"arn:aws:s3:::aws-security-data-lake*/*"
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
				"arn:aws:s3:::aws-security-data-lake*/*"
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
  EOF
}

resource "aws_kms_key" "test" {
	deletion_window_in_days = 7
  
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Id": "kms-tf-1",
	"Statement": [
		{
		"Sid": "Enable IAM User Permissions",
		"Effect": "Allow",
		"Principal": {
			"AWS": "*"
		},
		"Action": "kms:*",
		"Resource": "*"
		}
	]
	}
	POLICY
}
`, rName)
}

func testAccDataLakeConfig_basic(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return acctest.ConfigCompose(testAccDataLakeConfigBaseConfig(rName), fmt.Sprintf(`

resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configurations {
    region = "eu-west-1"
	
  }
  tags = {
    Name = %[1]q
  }
  depends_on = [aws_iam_role.meta_store_manager]
}
`, rName))
}

func testAccDataLakeConfig_lifeCycle(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return acctest.ConfigCompose(testAccDataLakeConfigBaseConfig(rName), fmt.Sprintf(`


resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configurations {
    region = "eu-west-1"

    encryption_configuration {
		kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transitions {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      transitions {
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
`, rName))
}

func testAccDataLakeConfig_lifeCycleUpdate(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return acctest.ConfigCompose(testAccDataLakeConfigBaseConfig(rName), fmt.Sprintf(`

resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configurations {
    region = "eu-west-1"

    encryption_configuration {
		kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transitions {
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
`, rName))
}

func testAccDataLakeConfig_replication(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(rName), fmt.Sprintf(`


resource "aws_securitylake_data_lake" "region_2" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configurations {
    region = "eu-west-2"

    lifecycle_configuration {
      transitions {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }
    replication_configuration {
      role_arn = aws_iam_role.datalake_s3_replication.arn
      regions  = ["eu-west-1"]
    }
  }
  tags = {
    Name = %[1]q
  }
  depends_on = [aws_iam_role.meta_store_manager, aws_iam_role.datalake_s3_replication, aws_securitylake_data_lake.test]
}
`, rName))
}
