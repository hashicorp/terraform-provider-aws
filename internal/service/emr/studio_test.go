// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRStudio_sso(t *testing.T) {
	ctx := acctest.Context(t)
	var studio awstypes.Studio
	resourceName := "aws_emr_studio.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_sso(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticmapreduce", regexache.MustCompile(`studio/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "SSO"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "user_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStudioConfig_sso(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticmapreduce", regexache.MustCompile(`studio/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "SSO"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "user_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccEMRStudio_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var studio awstypes.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_iam(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_key_arn", ""),
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

func TestAccEMRStudio_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var studio awstypes.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_sso(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudio(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceStudio(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRStudio_workspaceStorageEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var studio awstypes.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_workspaceStorageEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "engine_security_group_id", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_key_arn", "aws_kms_key.test", names.AttrARN),
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

func TestAccEMRStudio_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var studio awstypes.Studio
	resourceName := "aws_emr_studio.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStudioDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStudioConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStudioConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStudioConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStudioExists(ctx, resourceName, &studio),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckStudioExists(ctx context.Context, n string, v *awstypes.Studio) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRClient(ctx)

		output, err := tfemr.FindStudioByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStudioDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_studio" {
				continue
			}

			_, err := tfemr.FindStudioByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EMR Studio %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStudioConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy" "emr" {
  name = "AmazonEMRServicePolicy_v2"
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.test.name
  policy_arn = data.aws_iam_policy.emr.arn
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*",
		"${aws_s3_bucket.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccStudioConfig_sso(rName, name string) string {
	return acctest.ConfigCompose(testAccStudioConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_studio" "test" {
  auth_mode                   = "SSO"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = aws_subnet.test[*].id
  user_role                   = aws_iam_role.test.arn
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}
`, name))
}

func testAccStudioConfig_iam(rName string) string {
	return acctest.ConfigCompose(testAccStudioConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_studio" "test" {
  auth_mode                   = "IAM"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = aws_subnet.test[*].id
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccStudioConfig_workspaceStorageEncryption(rName string) string {
	return acctest.ConfigCompose(testAccStudioConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "AllowEMRStudioServiceRoleAccess",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_role.test.arn}"
      },
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey",
        "kms:ReEncryptFrom",
        "kms:ReEncryptTo",
        "kms:DescribeKey"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "kms:CallerAccount": "${data.aws_caller_identity.current.account_id}",
          "kms:EncryptionContext:aws:s3:arn": "${aws_s3_bucket.test.arn}",
          "kms:ViaService": "s3.${data.aws_region.current.name}.amazonaws.com"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test_kms" {
  name   = "%[1]s-kms"
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey",
        "kms:ReEncryptFrom",
        "kms:ReEncryptTo",
        "kms:DescribeKey"
      ],
      "Resource": "${aws_kms_key.test.arn}"
    }
  ]
}
EOF
}

resource "aws_emr_studio" "test" {
  auth_mode                   = "IAM"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  encryption_key_arn          = aws_kms_key.test.arn
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = aws_subnet.test[*].id
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccStudioConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStudioConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_studio" "test" {
  auth_mode                   = "IAM"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = aws_subnet.test[*].id
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccStudioConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStudioConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_studio" "test" {
  auth_mode                   = "IAM"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = aws_subnet.test[*].id
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
