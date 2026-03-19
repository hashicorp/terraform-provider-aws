// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSProxySecret_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy_secret.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxySecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxySecretConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxySecretExists(ctx, t, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "db_proxy_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "secret_arn", "aws_secretsmanager_secret.additional", names.AttrARN),
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

func TestAccRDSProxySecret_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy_secret.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxySecretDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxySecretConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxySecretExists(ctx, t, resourceName, &dbProxy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceProxySecret(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProxySecretDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_secret" {
				continue
			}

			dbProxyName := rs.Primary.Attributes["db_proxy_name"]
			secretARN := rs.Primary.Attributes["secret_arn"]

			dbProxy, err := tfrds.FindDBProxyByName(ctx, conn, dbProxyName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			for _, auth := range dbProxy.Auth {
				if aws.ToString(auth.SecretArn) == secretARN {
					return fmt.Errorf("RDS DB Proxy Secret %s still exists", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckProxySecretExists(ctx context.Context, t *testing.T, n string, v *types.DBProxy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		dbProxyName := rs.Primary.Attributes["db_proxy_name"]
		secretARN := rs.Primary.Attributes["secret_arn"]

		dbProxy, err := tfrds.FindDBProxyByName(ctx, conn, dbProxyName)
		if err != nil {
			return err
		}

		for _, auth := range dbProxy.Auth {
			if aws.ToString(auth.SecretArn) == secretARN {
				*v = *dbProxy
				return nil
			}
		}

		return fmt.Errorf("RDS DB Proxy Secret %s not found in proxy auth configuration", rs.Primary.ID)
	}
}

func testAccProxySecretConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "additional" {
  name                    = "%[1]s-additional"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "additional" {
  secret_id     = aws_secretsmanager_secret.additional.id
  secret_string = "{\"username\":\"db_user2\",\"password\":\"db_user2_password\"}"
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
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName))
}

func testAccProxySecretConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxySecretConfig_base(rName), `
resource "aws_db_proxy_secret" "test" {
  db_proxy_name = aws_db_proxy.test.name
  secret_arn    = aws_secretsmanager_secret.additional.arn

  depends_on = [aws_secretsmanager_secret_version.additional]
}
`)
}
