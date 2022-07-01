package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestAccRDSProxyDefaultTargetGroup_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`target-group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "connection_pool_config.*", map[string]string{
						"connection_borrow_timeout":    "120",
						"init_query":                   "",
						"max_connections_percent":      "100",
						"max_idle_connections_percent": "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "0"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_emptyConnectionPoolConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`target-group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "connection_pool_config.*", map[string]string{
						"connection_borrow_timeout":    "120",
						"init_query":                   "",
						"max_connections_percent":      "100",
						"max_idle_connections_percent": "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "0"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_connectionBorrowTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
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
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.connection_borrow_timeout", "90"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_initQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_initQuery(rName, "SET x=1, y=2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
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
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.init_query", "SET a=2, b=1"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_maxConnectionsPercent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxConnectionsPercent(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
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
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_connections_percent", "75"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_maxIdleConnectionsPercent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_maxIdleConnectionsPercent(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
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
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_idle_connections_percent", "33"),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_sessionPinningFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sessionPinningFilters := "EXCLUDE_VARIABLE_SETS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "0"),
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
					testAccCheckProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.0", sessionPinningFilters),
				),
			},
		},
	})
}

func TestAccRDSProxyDefaultTargetGroup_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBProxy
	dbProxyResourceName := "aws_db_proxy.test"
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDefaultTargetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &v),
					// DB Proxy default Target Group implicitly exists so it cannot be removed.
					// Verify disappearance handling for DB Proxy removal instead.
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceProxy(), dbProxyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProxyTargetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy_default_target_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBProxyTargetGroups(
			&rds.DescribeDBProxyTargetGroupsInput{
				DBProxyName:     aws.String(rs.Primary.ID),
				TargetGroupName: aws.String("default"),
			})

		if err == nil {
			if len(resp.TargetGroups) != 0 &&
				*resp.TargetGroups[0].DBProxyName == rs.Primary.ID {
				return fmt.Errorf("DB Proxy Target Group still exists")
			}
		}

		if !tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
			return err
		}
	}

	return nil
}

func testAccCheckProxyTargetGroupExists(n string, v *rds.DBProxyTargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBProxyTargetGroupsInput{
			DBProxyName:     aws.String(rs.Primary.ID),
			TargetGroupName: aws.String("default"),
		}

		resp, err := conn.DescribeDBProxyTargetGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.TargetGroups) != 1 ||
			*resp.TargetGroups[0].DBProxyName != rs.Primary.ID {
			return fmt.Errorf("DB Proxy Target Group not found")
		}

		*v = *resp.TargetGroups[0]

		return nil
	}
}

func testAccProxyDefaultTargetGroupBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }

  tags = {
    Name = "%[1]s"
  }
}

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

func testAccProxyDefaultTargetGroupConfig_basic(rName string) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + `
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name
}
`
}

func testAccProxyDefaultTargetGroupConfig_emptyConnectionPoolConfig(rName string) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + `
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
  }
}
`
}

func testAccProxyDefaultTargetGroupConfig_connectionBorrowTimeout(rName string, connectionBorrowTimeout int) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    connection_borrow_timeout = %[2]d
  }
}
`, rName, connectionBorrowTimeout)
}

func testAccProxyDefaultTargetGroupConfig_initQuery(rName, initQuery string) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    init_query = "%[2]s"
  }
}
`, rName, initQuery)
}

func testAccProxyDefaultTargetGroupConfig_maxConnectionsPercent(rName string, maxConnectionsPercent int) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_connections_percent = %[2]d
  }
}
`, rName, maxConnectionsPercent)
}

func testAccProxyDefaultTargetGroupConfig_maxIdleConnectionsPercent(rName string, maxIdleConnectionsPercent int) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_idle_connections_percent = %[2]d
  }
}
`, rName, maxIdleConnectionsPercent)
}

func testAccProxyDefaultTargetGroupConfig_sessionPinningFilters(rName, sessionPinningFilters string) string {
	return testAccProxyDefaultTargetGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    session_pinning_filters = ["%[2]s"]
  }
}
`, rName, sessionPinningFilters)
}
