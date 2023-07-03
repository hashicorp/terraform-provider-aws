// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`user/.+`)),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "server_id", "aws_transfer_server.test", "id"),
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

func testAccUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userConf transfer.DescribedUser
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &userConf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUser_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
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
				Config: testAccUserConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccUserConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccUser_posix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_posix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.gid", "1000"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.uid", "1000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_posixUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.secondary_gids.#", "2"),
				),
			},
		},
	})
}

func testAccUser_modifyWithOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_options(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/home/tftestuser"),
				),
			},
			{
				Config: testAccUserConfig_modify(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/test"),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
				),
			},
			{
				Config: testAccUserConfig_forceNew(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/home/tftestuser2"),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "tftestuser2"),
				),
			},
		},
	})
}

func testAccUser_UserName_Validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserConfig_nameValidation(rName, "!@#$%^"),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:      testAccUserConfig_nameValidation(rName, sdkacctest.RandString(2)),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:             testAccUserConfig_nameValidation(rName, sdkacctest.RandString(33)),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
			{
				Config:      testAccUserConfig_nameValidation(rName, sdkacctest.RandString(101)),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:      testAccUserConfig_nameValidation(rName, "-abcdef"),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:             testAccUserConfig_nameValidation(rName, "valid_username"),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func testAccUser_homeDirectoryMappings(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedUser
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_user.test"
	entry1 := "/your-personal-report.pdf"
	target1 := "/bucket3/customized-reports/tftestuser.pdf"
	entry2 := "/your-personal-report2.pdf"
	target2 := "/bucket3/customized-reports2/tftestuser.pdf"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_homeDirectoryMappings(rName, entry1, target1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.0.entry", entry1),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.0.target", target1),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "LOGICAL"),
				),
			},
			{
				Config: testAccUserConfig_homeDirectoryMappingsUpdate(rName, entry1, target1, entry2, target2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.0.entry", entry1),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.0.target", target1),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.1.entry", entry2),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.1.target", target2),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "LOGICAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_homeDirectoryMappingsRemove(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
				),
			},
		},
	})
}

func testAccCheckUserExists(ctx context.Context, n string, v *transfer.DescribedUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer User ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		serverID, userName, err := tftransfer.UserParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tftransfer.FindUserByTwoPartKey(ctx, conn, serverID, userName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_user" {
				continue
			}

			serverID, userName, err := tftransfer.UserParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tftransfer.FindUserByTwoPartKey(ctx, conn, serverID, userName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccUserConfig_baseRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccUserConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_baseRole(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}
`, rName))
}

func testAccUserConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), `
resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn
}
`)
}

func testAccUserConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccUserConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccUserConfig_nameValidation(rName, username string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = %[1]q
  role      = aws_iam_role.test.arn
}
`, username))
}

func testAccUserConfig_options(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObjectVersion",
      "s3:DeleteObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/home/tftestuser"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccUserConfig_modify(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/test"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccUserConfig_forceNew(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObjectVersion",
      "s3:DeleteObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser2"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/home/tftestuser2"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccUserConfig_homeDirectoryMappings(rName, entry, target string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  home_directory_type = "LOGICAL"
  role                = aws_iam_role.test.arn
  server_id           = aws_transfer_server.test.id
  user_name           = "tftestuser"

  home_directory_mappings {
    entry  = %[2]q
    target = %[3]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, entry, target))
}

func testAccUserConfig_homeDirectoryMappingsUpdate(rName, entry1, target1, entry2, target2 string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  home_directory_type = "LOGICAL"
  role                = aws_iam_role.test.arn
  server_id           = aws_transfer_server.test.id
  user_name           = "tftestuser"

  home_directory_mappings {
    entry  = %[2]q
    target = %[3]q
  }

  home_directory_mappings {
    entry  = %[4]q
    target = %[5]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, entry1, target1, entry2, target2))
}

func testAccUserConfig_homeDirectoryMappingsRemove(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  role      = aws_iam_role.test.arn
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccUserConfig_posix(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_baseRole(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  domain = "EFS"

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  posix_profile {
    gid = 1000
    uid = 1000
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccUserConfig_posixUpdated(rName string) string {
	return acctest.ConfigCompose(testAccUserConfig_baseRole(rName), fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  domain = "EFS"

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  posix_profile {
    gid            = 1001
    uid            = 1001
    secondary_gids = [1000, 1002]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
