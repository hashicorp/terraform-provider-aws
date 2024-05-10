// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTRoleAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	alias := sdkacctest.RandomWithPrefix("RoleAlias-")
	alias2 := sdkacctest.RandomWithPrefix("RoleAlias2-")

	resourceName := "aws_iot_role_alias.ra"
	resourceName2 := "aws_iot_role_alias.ra2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_basic(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update1(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName),
					testAccCheckRoleAliasExists(ctx, resourceName2),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "43200"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update2(alias2),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRoleAliasExists(ctx, resourceName2)),
			},
			{
				Config: testAccRoleAliasConfig_update3(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName2),
				),
				ExpectError: regexache.MustCompile("Role alias .+? already exists for this account"),
			},
			{
				Config: testAccRoleAliasConfig_update4(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName2),
				),
			},
			{
				Config: testAccRoleAliasConfig_update5(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName2),
					acctest.MatchResourceAttrGlobalARN(resourceName2, names.AttrRoleARN, "iam", regexache.MustCompile("role/rolebogus")),
				),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoleAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_role_alias" {
				continue
			}

			_, err := tfiot.GetRoleAliasDescription(ctx, conn, rs.Primary.ID)

			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				continue
			}

			return fmt.Errorf("IoT Role Alias (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckRoleAliasExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)
		role_arn := rs.Primary.Attributes[names.AttrRoleARN]

		roleAliasDescription, err := tfiot.GetRoleAliasDescription(ctx, conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error: Failed to get role alias %s for role %s (%s): %s", rs.Primary.ID, role_arn, n, err)
		}

		if roleAliasDescription == nil {
			return fmt.Errorf("Error: Role alias %s is not attached to role (%s)", rs.Primary.ID, role_arn)
		}

		return nil
	}
}

func TestAccIoTRoleAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_role_alias.tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName),
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
				Config: testAccRoleAliasConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoleAliasConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccRoleAliasConfig_basic(alias string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias)
}

func testAccRoleAliasConfig_update1(alias string, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra" {
  alias               = "%s"
  role_arn            = aws_iam_role.role.arn
  credential_duration = 43200
}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias, alias2)
}

func testAccRoleAliasConfig_update2(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias2)
}

func testAccRoleAliasConfig_update3(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}

resource "aws_iot_role_alias" "ra3" {
  alias    = "%s"
  role_arn = aws_iam_role.role.arn
}
`, alias2, alias2)
}

func testAccRoleAliasConfig_update4(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iam_role" "role2" {
  name = "role2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = aws_iam_role.role2.arn
}
`, alias2)
}

func testAccRoleAliasConfig_update5(alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iam_role" "role2" {
  name = "role2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "ra2" {
  alias    = "%s"
  role_arn = "${aws_iam_role.role.arn}bogus"
}
`, alias2)
}

func testAccRoleAliasConfig_tags1(alias, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "tag_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "tags" {
  alias    = %[1]q
  role_arn = aws_iam_role.role.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, alias, tagKey1, tagValue1)
}

func testAccRoleAliasConfig_tags2(alias, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "tag_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {
      "Service": "credentials.iot.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
}
EOF

}

resource "aws_iot_role_alias" "tags" {
  alias    = %[1]q
  role_arn = aws_iam_role.role.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, alias, tagKey1, tagValue1, tagKey2, tagValue2)
}
