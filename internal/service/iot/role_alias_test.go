package iot_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
)

func TestAccIoTRoleAlias_basic(t *testing.T) {
	alias := sdkacctest.RandomWithPrefix("RoleAlias-")
	alias2 := sdkacctest.RandomWithPrefix("RoleAlias2-")

	resourceName := "aws_iot_role_alias.ra"
	resourceName2 := "aws_iot_role_alias.ra2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoleAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAliasConfig_basic(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "3600"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update1(alias, alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(resourceName),
					testAccCheckRoleAliasExists(resourceName2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("rolealias/%s", alias)),
					resource.TestCheckResourceAttr(resourceName, "credential_duration", "43200"),
				),
			},
			{
				Config: testAccRoleAliasConfig_update2(alias2),
				Check:  resource.ComposeTestCheckFunc(testAccCheckRoleAliasExists(resourceName2)),
			},
			{
				Config: testAccRoleAliasConfig_update3(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(resourceName2),
				),
				ExpectError: regexp.MustCompile("Role alias .+? already exists for this account"),
			},
			{
				Config: testAccRoleAliasConfig_update4(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(resourceName2),
				),
			},
			{
				Config: testAccRoleAliasConfig_update5(alias2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleAliasExists(resourceName2),
					acctest.MatchResourceAttrGlobalARN(resourceName2, "role_arn", "iam", regexp.MustCompile("role/rolebogus")),
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

func testAccCheckRoleAliasDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_role_alias" {
			continue
		}

		_, err := tfiot.GetRoleAliasDescription(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			continue
		}

		return fmt.Errorf("IoT Role Alias (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRoleAliasExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
		role_arn := rs.Primary.Attributes["role_arn"]

		roleAliasDescription, err := tfiot.GetRoleAliasDescription(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error: Failed to get role alias %s for role %s (%s): %s", rs.Primary.ID, role_arn, n, err)
		}

		if roleAliasDescription == nil {
			return fmt.Errorf("Error: Role alias %s is not attached to role (%s)", rs.Primary.ID, role_arn)
		}

		return nil
	}
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
