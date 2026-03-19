// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSProxyPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_db_proxy_policy_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "db_proxy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy_document"),
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

func TestAccRDSProxyPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_db_proxy_policy_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceProxyPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSProxyPolicyAttachment_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_db_proxy_policy_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyPolicyAttachmentExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccProxyPolicyAttachmentConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyPolicyAttachmentExists(ctx, t, resourceName),
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

func testAccCheckProxyPolicyAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)
		iamConn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_policy_attachment" {
				continue
			}

			dbProxyName := rs.Primary.Attributes["db_proxy_name"]
			policyName := rs.Primary.Attributes["policy_name"]

			roleName, err := tfrds.FindDBProxyRoleName(ctx, conn, dbProxyName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			_, err = tfrds.FindRolePolicyForProxy(ctx, iamConn, roleName, policyName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Proxy Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyPolicyAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)
		iamConn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		dbProxyName := rs.Primary.Attributes["db_proxy_name"]
		policyName := rs.Primary.Attributes["policy_name"]

		roleName, err := tfrds.FindDBProxyRoleName(ctx, conn, dbProxyName)
		if err != nil {
			return err
		}

		_, err = tfrds.FindRolePolicyForProxy(ctx, iamConn, roleName, policyName)

		return err
	}
}

func testAccProxyPolicyAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
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

func testAccProxyPolicyAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxyPolicyAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_policy_attachment" "test" {
  db_proxy_name = aws_db_proxy.test.name
  policy_name   = %[1]q

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
        ]
        Resource = [
          aws_secretsmanager_secret.test.arn,
        ]
      },
    ]
  })
}
`, rName))
}

func testAccProxyPolicyAttachmentConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccProxyPolicyAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_policy_attachment" "test" {
  db_proxy_name = aws_db_proxy.test.name
  policy_name   = %[1]q

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
        ]
        Resource = [
          aws_secretsmanager_secret.test.arn,
        ]
      },
    ]
  })
}
`, rName))
}
