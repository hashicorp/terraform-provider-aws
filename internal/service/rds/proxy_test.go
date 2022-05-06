package rds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSProxy_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_family", "MYSQL"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db-proxy:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auth.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"auth_scheme": "SECRETS",
						"description": "test",
						"iam_auth":    "DISABLED",
					}),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "1800"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`))),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyNameConfig(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "name", nName),
				),
			},
		},
	})
}

func TestAccRDSProxy_debugLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDebugLoggingConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyDebugLoggingConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "false"),
				),
			},
		},
	})
}

func TestAccRDSProxy_idleClientTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyIdleClientTimeoutConfig(rName, 900),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "900"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyIdleClientTimeoutConfig(rName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "3600"),
				),
			},
		},
	})
}

func TestAccRDSProxy_requireTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRequireTLSConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyRequireTLSConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "false"),
				),
			},
		},
	})
}

func TestAccRDSProxy_roleARN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyRoleARNConfig(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
				),
			},
		},
	})
}

func TestAccRDSProxy_vpcSecurityGroupIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyVPCSecurityGroupIDsConfig(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test2", "id"),
				),
			},
		},
	})
}

func TestAccRDSProxy_authDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.description", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyAuthDescriptionConfig(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.description", description),
				),
			},
		},
	})
}

func TestAccRDSProxy_authIAMAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamAuth := "REQUIRED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.iam_auth", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyAuthIAMAuthConfig(rName, iamAuth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.iam_auth", iamAuth),
				),
			},
		},
	})
}

func TestAccRDSProxy_authSecretARN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "auth.0.secret_arn", "aws_secretsmanager_secret.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyAuthSecretARNConfig(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "auth.0.secret_arn", "aws_secretsmanager_secret.test2", "arn"),
				),
			},
		},
	})
}

func TestAccRDSProxy_authUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.username", ""),
				),
			},
		},
	})
}

func TestAccRDSProxy_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := "foo"
	value := "bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyNameConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyTagsConfig(rName, key, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", value),
				),
			},
		},
	})
}

func TestAccRDSProxy_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceProxy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccDBProxyPreCheck checks if a call to describe db proxies errors out meaning feature not supported
func testAccDBProxyPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBProxiesInput{}
	_, err := conn.DescribeDBProxies(input)

	if tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance test, RDS Proxy not supported: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckProxyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy" {
			continue
		}

		_, err := tfrds.FindDBProxyByName(conn, rs.Primary.ID)

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

func testAccCheckProxyExists(n string, v *rds.DBProxy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS DB Proxy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		output, err := tfrds.FindDBProxyByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyBaseConfig(rName string) string {
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

func testAccProxyConfig(rName string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccProxyNameConfig(rName, nName string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[2]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, nName)
}

func testAccProxyDebugLoggingConfig(rName string, debugLogging bool) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = "%[1]s"
  debug_logging  = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, debugLogging)
}

func testAccProxyIdleClientTimeoutConfig(rName string, idleClientTimeout int) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                = "%[1]s"
  idle_client_timeout = %[2]d
  engine_family       = "MYSQL"
  role_arn            = aws_iam_role.test.arn
  vpc_subnet_ids      = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, idleClientTimeout)
}

func testAccProxyRequireTLSConfig(rName string, requireTls bool) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = "%[1]s"
  require_tls    = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, requireTls)
}

func testAccProxyRoleARNConfig(rName, nName string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test2
  ]

  name           = "%[1]s"
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test2.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

# IAM setup

resource "aws_iam_role" "test2" {
  name               = "%[2]s"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName, nName)
}

func testAccProxyVPCSecurityGroupIDsConfig(rName, nName string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test2.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

resource "aws_security_group" "test2" {
  name   = "%[2]s"
  vpc_id = aws_vpc.test.id
}
`, rName, nName)
}

func testAccProxyAuthDescriptionConfig(rName, description string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "%[2]s"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, description)
}

func testAccProxyAuthIAMAuthConfig(rName, iamAuth string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  require_tls            = true
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "%[2]s"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, iamAuth)
}

func testAccProxyAuthSecretARNConfig(rName, nName string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test2.arn
  }
}

resource "aws_secretsmanager_secret" "test2" {
  name                    = "%[2]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}
`, rName, nName)
}

func testAccProxyTagsConfig(rName, key, value string) string {
	return testAccProxyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
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
    %[2]s = "%[3]s"
  }
}
`, rName, key, value)
}
