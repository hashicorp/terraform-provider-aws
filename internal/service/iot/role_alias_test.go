// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTRoleAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	alias := acctest.RandomWithPrefix(t, "RoleAlias-")
	alias2 := acctest.RandomWithPrefix(t, "RoleAlias2-")
	resourceName := "aws_iot_role_alias.ra"
	resourceName2 := "aws_iot_role_alias.ra2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_basic(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update1(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
					testAccCheckRoleAliasExists(ctx, t, resourceName2),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "43200"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update2(alias, alias2),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRoleAliasExists(ctx, t, resourceName2)),
			},
			{
				Config:      testAccRoleAliasConfig_update3(alias, alias2),
				ExpectError: regexache.MustCompile("Role alias .+? already exists for this account"),
			},
			{
				Config: testAccRoleAliasConfig_update4(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName2),
				),
			},
			{
				Config: testAccRoleAliasConfig_update5(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName2),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName2, names.AttrRoleARN, "iam", regexache.MustCompile("role/"+alias+"/bogus")),
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

func TestAccIoTRoleAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	alias := acctest.RandomWithPrefix(t, "RoleAlias-")
	resourceName := "aws_iot_role_alias.ra"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_basic(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiot.ResourceRoleAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTRoleAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_role_alias.tags"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
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
				Config: testAccRoleAliasConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRoleAliasConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckRoleAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_role_alias" {
				continue
			}

			_, err := tfiot.FindRoleAliasByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Role Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRoleAliasExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindRoleAliasByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccRoleAliasConfig_basic(alias string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  alias    = %[1]q
  role_arn = aws_iam_role.role.arn
}
`, alias)
}

func testAccRoleAliasConfig_update1(alias string, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  alias               = %[1]q
  role_arn            = aws_iam_role.role.arn
  credential_duration = 43200
}

resource "aws_iot_role_alias" "ra2" {
  alias    = %[2]q
  role_arn = aws_iam_role.role.arn
}
`, alias, alias2)
}

func testAccRoleAliasConfig_update2(alias, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  alias    = %[2]q
  role_arn = aws_iam_role.role.arn
}
`, alias, alias2)
}

func testAccRoleAliasConfig_update3(alias, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  alias    = %[2]q
  role_arn = aws_iam_role.role.arn
}

resource "aws_iot_role_alias" "ra3" {
  alias    = %[2]q
  role_arn = aws_iam_role.role.arn
}
`, alias, alias2)
}

func testAccRoleAliasConfig_update4(alias, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  name = %[2]q

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
  alias    = %[1]q
  role_arn = aws_iam_role.role2.arn
}
`, alias, alias2)
}

func testAccRoleAliasConfig_update5(alias, alias2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  name = %[2]q

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
  alias    = %[2]q
  role_arn = "${aws_iam_role.role.arn}/bogus"
}
`, alias, alias2)
}

func testAccRoleAliasConfig_tags1(alias, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = %[1]q

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
  name = %[1]q

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
