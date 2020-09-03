package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDBProxyTarget_Instance(t *testing.T) {
	var dbProxyTarget rds.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyTargetConfig_Instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetExists(resourceName, &dbProxyTarget),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_arn", "aws_db_instance.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tracked_cluster_id", ""),
					resource.TestCheckResourceAttr(resourceName, "type", "RDS_INSTANCE"),
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

func TestAccAWSDBProxyTarget_Cluster(t *testing.T) {
	var dbProxyTarget rds.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyTargetConfig_Cluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetExists(resourceName, &dbProxyTarget),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "port", "3306"),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_arn", "aws_rds_cluster.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "tracked_cluster_id", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "TRACKED_CLUSTER"),
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

func TestAccAWSDBProxyTarget_disappears(t *testing.T) {
	var v rds.DBProxy
	resourceName := "aws_db_proxy_target.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyTargetConfig_Instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxyTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDBProxyTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy_target" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBProxyTargets(
			&rds.DescribeDBProxyTargetsInput{
				DBProxyName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.Targets) != 0 &&
				*resp.Targets[0].RdsResourceId == rs.Primary.ID {
				return fmt.Errorf("DB Proxy Target Group still exists")
			}
		}

		if !isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBProxyTargetExists(n string, v *rds.DBProxyTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBProxyTargetsInput{
			DBProxyName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBProxyTargets(&opts)

		if err != nil {
			return err
		}

		if len(resp.Targets) != 1 ||
			*resp.Targets[0].RdsResourceId != rs.Primary.ID {
			return fmt.Errorf("DB Proxy Target not found")
		}

		*v = *resp.Targets[0]

		return nil
	}
}

func testAccAWSDBProxyTargetConfigBase(rName string) string {
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

resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    connection_borrow_timeout    = 120
    init_query                   = "SET x=1, y=2"
    max_connections_percent      = 100
    max_idle_connections_percent = 50
    session_pinning_filters      = []
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test.*.id
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

func testAccAWSDBProxyTargetConfig_Instance(rName string) string {
	return testAccAWSDBProxyTargetConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_target" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  target_group_name      = aws_db_proxy_default_target_group.test.name
  db_instance_identifier = aws_db_instance.test.id
}

resource "aws_db_instance" "test" {
  identifier           = "%[1]s"
  db_subnet_group_name = aws_db_subnet_group.test.id
  allocated_storage    = 20
  engine               = "mysql"
  instance_class       = "db.t2.micro"
  password             = "testtest"
  username             = "test"

  tags = {
    Name = "%[1]s"
  }
}
`, rName)
}

func testAccAWSDBProxyTargetConfig_Cluster(rName string) string {
	return testAccAWSDBProxyTargetConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_target" "test" {
  db_proxy_name         = aws_db_proxy.test.name
  target_group_name     = aws_db_proxy_default_target_group.test.name
  db_cluster_identifier = aws_rds_cluster.test.id
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "%[1]s"
  db_subnet_group_name = aws_db_subnet_group.test.id
  engine               = "aurora-mysql"
  master_username      = "test"
  master_password      = "testtest"

  tags = {
    Name = "%[1]s"
  }
}
`, rName)
}
