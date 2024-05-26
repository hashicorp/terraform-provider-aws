// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSProxy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "engine_family", "MYSQL"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`db-proxy:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"auth_scheme":               "SECRETS",
						"client_password_auth_type": "MYSQL_NATIVE_PASSWORD",
						names.AttrDescription:       "test",
						"iam_auth":                  "DISABLED",
					}),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrEndpoint, regexache.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "1800"),
					resource.TestCheckResourceAttr(resourceName, "require_tls", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", names.AttrID),
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

func TestAccRDSProxy_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_name(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nName),
				),
			},
		},
	})
}

func TestAccRDSProxy_debugLogging(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_debugLogging(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_debugLogging(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSProxy_idleClientTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_idleClientTimeout(rName, 900),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "900"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_idleClientTimeout(rName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "3600"),
				),
			},
		},
	})
}

func TestAccRDSProxy_requireTLS(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_requireTLS(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_requireTLS(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRDSProxy_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_roleARN(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSProxy_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_vpcSecurityGroupIDs(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test2", names.AttrID),
				),
			},
		},
	})
}

func TestAccRDSProxy_authDescription(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						names.AttrDescription: "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_authDescription(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						names.AttrDescription: description,
					}),
				),
			},
		},
	})
}

func TestAccRDSProxy_authIAMAuth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamAuth := "REQUIRED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"iam_auth": "DISABLED",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_authIAMAuth(rName, iamAuth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"iam_auth": iamAuth,
					}),
				),
			},
		},
	})
}

func TestAccRDSProxy_authSecretARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "auth.*.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_authSecretARN(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "auth.*.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "auth.*.secret_arn", "aws_secretsmanager_secret.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccRDSProxy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProxyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSProxy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceProxy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccDBProxyPreCheck checks if a call to describe db proxies errors out meaning feature not supported
func testAccDBProxyPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBProxiesInput{}
	_, err := conn.DescribeDBProxies(ctx, input)

	if tfawserr.ErrCodeEquals(err, tfrds.ErrCodeInvalidAction) {
		t.Skipf("skipping acceptance test, RDS Proxy not supported: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckProxyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy" {
				continue
			}

			_, err := tfrds.FindDBProxyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Proxy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyExists(ctx context.Context, n string, v *types.DBProxy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBProxyByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyConfig_base(rName string) string {
	return fmt.Sprintf(`
# Secrets Manager setup

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

# IAM setup

resource "aws_iam_role" "test" {
  name               = "%[1]s"
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

# VPC setup

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "%[1]s"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-${count.index}"
  }
}
`, rName)
}

func testAccProxyConfig_basic(rName string) string {
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

func testAccProxyConfig_name(rName, nName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
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
`, nName))
}

func testAccProxyConfig_debugLogging(rName string, debugLogging bool) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = %[1]q
  debug_logging  = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, debugLogging))
}

func testAccProxyConfig_idleClientTimeout(rName string, idleClientTimeout int) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                = %[1]q
  idle_client_timeout = %[2]d
  engine_family       = "MYSQL"
  role_arn            = aws_iam_role.test.arn
  vpc_subnet_ids      = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, idleClientTimeout))
}

func testAccProxyConfig_requireTLS(rName string, requireTls bool) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = %[1]q
  require_tls    = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, requireTls))
}

func testAccProxyConfig_roleARN(rName, nName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test2
  ]

  name           = %[1]q
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test2.arn
  vpc_subnet_ids = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

resource "aws_iam_role" "test2" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName, nName))
}

func testAccProxyConfig_vpcSecurityGroupIDs(rName, nName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test2.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

resource "aws_security_group" "test2" {
  name   = %[2]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[2]q
  }
}
`, rName, nName))
}

func testAccProxyConfig_authDescription(rName, description string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = %[2]q
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, description))
}

func testAccProxyConfig_authIAMAuth(rName, iamAuth string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  require_tls            = true
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = %[2]q
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, iamAuth))
}

func testAccProxyConfig_authSecretARN(rName, nName string) string {
	return acctest.ConfigCompose(testAccProxyConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme               = "SECRETS"
    client_password_auth_type = "MYSQL_NATIVE_PASSWORD"
    description               = "user with read/write access to the database."
    iam_auth                  = "DISABLED"
    secret_arn                = aws_secretsmanager_secret.test.arn
  }

  auth {
    auth_scheme               = "SECRETS"
    client_password_auth_type = "MYSQL_NATIVE_PASSWORD"
    description               = "user with read only access to the database."
    iam_auth                  = "DISABLED"
    secret_arn                = aws_secretsmanager_secret.test2.arn
  }
}

resource "aws_secretsmanager_secret" "test2" {
  name                    = %[2]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}
`, rName, nName))
}

func testAccProxyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccProxyConfig_base(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProxyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccProxyConfig_base(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
