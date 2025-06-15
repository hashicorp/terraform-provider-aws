// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSProxyAuthItem_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.UserAuthConfig
	resourceName := "aws_db_proxy_auth_item.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyAuthItemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyAuthItem_basic(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyAuthItemExists(ctx, resourceName, &v),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn", "aws_secretsmanager_secret.test2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auth_scheme", "SECRETS"),
					resource.TestCheckResourceAttr(resourceName, "iam_auth", "DISABLED"),
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

func testAccCheckProxyAuthItemDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_auth_item" {
				continue
			}

			_, err := tfrds.FindDBProxyAuthItemByArn(ctx, conn, rs.Primary.Attributes["db_proxy_name"], rs.Primary.Attributes["secret_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Auth Item %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyAuthItemExists(ctx context.Context, n string, v *types.UserAuthConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBProxyAuthItemByArn(ctx, conn, rs.Primary.Attributes["db_proxy_name"], rs.Primary.Attributes["secret_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyAuthItem_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "secretsmanager:GetRandomPassword",
      "secretsmanager:CreateSecret",
      "secretsmanager:ListSecrets",
    ]
    resources = ["*"]
  }

  statement {
    actions   = ["secretsmanager:*"]
    resources = [aws_secretsmanager_secret.test.arn]
  }
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = aws_security_group.test[*].id
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }

  lifecycle {
    ignore_changes = [
      auth
    ]
  }
}
`, rName))
}

func testAccProxyAuthItem_basic(rName string, nName string) string {
	return acctest.ConfigCompose(testAccProxyAuthItem_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test2" {
  name                    = "%[1]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

resource "aws_db_proxy_auth_item" "test" {
  db_proxy_name = aws_db_proxy.test.name
  secret_arn    = aws_secretsmanager_secret.test2.arn
}
`, nName))
}
