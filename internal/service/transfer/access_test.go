// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccess_s3_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessConfig_s3Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrExternalID, "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/"+rName+"/"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRole),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRole},
			},
			{
				Config: testAccAccessConfig_s3Updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrExternalID, "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/"+rName+"/test"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
				),
			},
		},
	})
}

func testAccAccess_efs_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessConfig_efsBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrExternalID, "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttrSet(resourceName, "home_directory"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRole),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrRole},
			},
			{
				Config: testAccAccessConfig_efsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrExternalID, "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttrSet(resourceName, "home_directory"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRole),
				),
			},
		},
	})
}

func testAccAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessConfig_s3Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceAccess(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccess_s3_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessConfig_s3ScopeDownPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccCheckAccessExists(ctx context.Context, n string, v *awstypes.DescribedAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		output, err := tftransfer.FindAccessByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes[names.AttrExternalID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_access" {
				continue
			}

			_, err := tftransfer.FindAccessByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes[names.AttrExternalID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Access %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccessBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_subnet" "test2" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"

  vpc_settings {
    vpc_id = aws_vpc.test.id

    subnet_ids = [
      aws_subnet.test.id,
      aws_subnet.test2.id
    ]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}`, rName))
}

func testAccAccessBaseConfig_S3(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_DIRECTORY_SERVICE"
  directory_id           = aws_directory_service_directory.test.id
  logging_role           = aws_iam_role.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
   "Version":"2012-10-17",
   "Statement":[
      {
         "Sid":"AllowFullAccesstoCloudWatchLogs",
         "Effect":"Allow",
         "Action":[
            "logs:*"
         ],
         "Resource":"*"
      },
      {
         "Sid":"AllowFullAccesstoS3",
         "Effect":"Allow",
         "Action":[
            "s3:*"
         ],
         "Resource":"*"
      }
   ]
}
POLICY
}
`, rName)
}

func testAccAccessConfig_s3Basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessBaseConfig(rName),
		testAccAccessBaseConfig_S3(rName),
		`
resource "aws_transfer_access" "test" {
  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
  server_id   = aws_transfer_server.test.id
  role        = aws_iam_role.test.arn

  home_directory      = "/${aws_s3_bucket.test.id}/"
  home_directory_type = "PATH"
}
`)
}

func testAccAccessConfig_s3Updated(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessBaseConfig(rName),
		testAccAccessBaseConfig_S3(rName),
		`
resource "aws_transfer_access" "test" {
  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
  server_id   = aws_transfer_server.test.id
  role        = aws_iam_role.test.arn

  home_directory      = "/${aws_s3_bucket.test.id}/test"
  home_directory_type = "PATH"
}
`)
}

func testAccAccessConfig_s3ScopeDownPolicy(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessBaseConfig(rName),
		testAccAccessBaseConfig_S3(rName),
		`
resource "aws_transfer_access" "test" {
  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
  server_id   = aws_transfer_server.test.id
  role        = aws_iam_role.test.arn

  home_directory      = "/${aws_s3_bucket.test.id}/"
  home_directory_type = "PATH"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowListingOfUserFolder",
            "Action": [
                "s3:ListBucket"
            ],
            "Effect": "Allow",
            "Resource": [
                "arn:${data.aws_partition.current.partition}:s3:::$${transfer:HomeBucket}"
            ]
        },
        {
            "Sid": "HomeDirObjectAccess",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject",
                "s3:DeleteObjectVersion",
                "s3:GetObjectVersion",
                "s3:GetObjectACL",
                "s3:PutObjectACL"
            ],
            "Resource": "arn:${data.aws_partition.current.partition}:s3:::$${transfer:HomeDirectory}/*"
        }
    ]
}
EOF
}`)
}

func testAccAccessBaseConfig_efs(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_DIRECTORY_SERVICE"
  directory_id           = aws_directory_service_directory.test.id
  logging_role           = aws_iam_role.test.arn
  domain                 = "EFS"
}

resource "aws_efs_file_system" "test" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  }]
}
POLICY
}
`, rName)
}

func testAccAccessConfig_efsBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessBaseConfig(rName),
		testAccAccessBaseConfig_efs(rName),
		`
resource "aws_transfer_access" "test" {
  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
  server_id   = aws_transfer_server.test.id
  role        = aws_iam_role.test.arn

  home_directory      = "/${aws_efs_file_system.test.id}/"
  home_directory_type = "PATH"

  posix_profile {
    gid = 1000
    uid = 1000
  }
}
`)
}

func testAccAccessConfig_efsUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccAccessBaseConfig(rName),
		testAccAccessBaseConfig_efs(rName),
		`
resource "aws_transfer_access" "test" {
  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
  server_id   = aws_transfer_server.test.id
  role        = aws_iam_role.test.arn

  home_directory      = "/${aws_efs_file_system.test.id}/test"
  home_directory_type = "PATH"

  posix_profile {
    gid = 1000
    uid = 1000
  }
}
`)
}
