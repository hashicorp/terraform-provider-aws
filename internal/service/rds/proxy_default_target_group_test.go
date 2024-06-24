// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccRDSProxyDefaultTargetGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`target-group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "connection_pool_config.*", map[string]string{
						"connection_borrow_timeout":    "120",
						"init_query":                   "",
						"max_connections_percent":      "100",
						"max_idle_connections_percent": "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", acctest.Ct0),
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

func TestAccRDSProxyDefaultTargetGroup_emptyConnectionPool(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_emptyConnectionPoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`target-group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "connection_pool_config.*", map[string]string{
						"connection_borrow_timeout":    "120",
						"init_query":                   "",
						"max_connections_percent":      "100",
						"max_idle_connections_percent": "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", acctest.Ct0),
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

func TestAccRDSProxyDefaultTargetGroup_connectionBorrowTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_connectionBorrowTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.connection_borrow_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDefaultTargetGroupConfig_connectionBorrowTimeout(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.connection_borrow_timeout", "90"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_initQuery(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_initQuery(rName, "SET x=1, y=2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.init_query", "SET x=1, y=2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDefaultTargetGroupConfig_initQuery(rName, "SET a=2, b=1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.init_query", "SET a=2, b=1"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_maxConnectionsPercent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxConnectionsPercent(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_connections_percent", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxConnectionsPercent(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_connections_percent", "75"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_maxIdleConnectionsPercent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxIdleConnectionsPercent(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_idle_connections_percent", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxIdleConnectionsPercent(rName, 33),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_idle_connections_percent", "33"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_sessionPinningFilters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup types.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sessionPinningFilters := "EXCLUDE_VARIABLE_SETS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDefaultTargetGroupConfig_sessionPinningFilters(rName, sessionPinningFilters),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(ctx, resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.0", sessionPinningFilters),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxy
	dbProxyResourceName := "aws_db_proxy.test"
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(ctx, resourceName, &v),
					// DB Proxy default Target Group implicitly exists so it cannot be removed.
					// Verify disappearance handling for DB Proxy removal instead.
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceProxy(), dbProxyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProxyTargetGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_default_target_group" {
				continue
			}

			_, err := tfrds.FindDefaultDBProxyTargetGroupByDBProxyName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Proxy Default Target Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyTargetGroupExists(ctx context.Context, n string, v *types.DBProxyTargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDefaultDBProxyTargetGroupByDBProxyName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyDefaultTargetGroupConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

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
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccProxyDefaultTargetGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), `
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name
}
`)
}

func testAccProxyDefaultTargetGroupConfig_emptyConnectionPoolConfig(rName string) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), `
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {}
}
`)
}

func testAccProxyDefaultTargetGroupConfig_connectionBorrowTimeout(rName string, connectionBorrowTimeout int) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    connection_borrow_timeout = %[1]d
  }
}
`, connectionBorrowTimeout))
}

func testAccProxyDefaultTargetGroupConfig_initQuery(rName, initQuery string) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    init_query = %[1]q
  }
}
`, initQuery))
}

func testAccProxyDefaultTargetGroupConfig_maxConnectionsPercent(rName string, maxConnectionsPercent int) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_connections_percent = %[1]d
  }
}
`, maxConnectionsPercent))
}

func testAccProxyDefaultTargetGroupConfig_maxIdleConnectionsPercent(rName string, maxIdleConnectionsPercent int) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_idle_connections_percent = %[1]d
  }
}
`, maxIdleConnectionsPercent))
}

func testAccProxyDefaultTargetGroupConfig_sessionPinningFilters(rName, sessionPinningFilters string) string {
	return acctest.ConfigCompose(testAccProxyDefaultTargetGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    session_pinning_filters = [%[1]q]
  }
}
`, sessionPinningFilters))
}
