package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

func TestAccRDSProxyTarget_instance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget rds.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_Instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(resourceName, &dbProxyTarget),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", "aws_db_instance.test", "address"),
					resource.TestCheckResourceAttrPair(resourceName, "port", "aws_db_instance.test", "port"),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttr(resourceName, "target_arn", ""),
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

func TestAccRDSProxyTarget_cluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget rds.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_Cluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(resourceName, &dbProxyTarget),
					resource.TestCheckResourceAttr(resourceName, "endpoint", ""),
					resource.TestCheckResourceAttrPair(resourceName, "port", "aws_rds_cluster.test", "port"),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttr(resourceName, "target_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tracked_cluster_id", ""),
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

func TestAccRDSProxyTarget_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget rds.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccDBProxyPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProxyTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_Instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(resourceName, &dbProxyTarget),
					acctest.CheckResourceDisappears(acctest.Provider, tfrds.ResourceProxyTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProxyTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy_target" {
			continue
		}

		dbProxyName, targetGroupName, targetType, rdsResourceId, err := tfrds.ProxyTargetParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		dbProxyTarget, err := tfrds.FindDBProxyTarget(conn, dbProxyName, targetGroupName, targetType, rdsResourceId)

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if dbProxyTarget != nil {
			return fmt.Errorf("RDS DB Proxy Target (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckProxyTargetExists(n string, v *rds.DBProxyTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		dbProxyName, targetGroupName, targetType, rdsResourceId, err := tfrds.ProxyTargetParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		dbProxyTarget, err := tfrds.FindDBProxyTarget(conn, dbProxyName, targetGroupName, targetType, rdsResourceId)

		if err != nil {
			return err
		}

		if dbProxyTarget == nil {
			return fmt.Errorf("RDS DB Proxy Target (%s) not found", rs.Primary.ID)
		}

		*v = *dbProxyTarget

		return nil
	}
}

func testAccProxyTargetBaseConfig(rName string) string {
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

  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }
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

func testAccProxyTargetConfig_Instance(rName string) string {
	return testAccProxyTargetBaseConfig(rName) + fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["5.7.31", "5.7.30"]
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  allocated_storage      = 20
  db_subnet_group_name   = aws_db_subnet_group.test.id
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  engine_version         = data.aws_rds_orderable_db_instance.test.engine_version
  identifier             = "%[1]s"
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "testtest"
  skip_final_snapshot    = true
  username               = "test"
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    Name = "%[1]s"
  }
}

resource "aws_db_proxy_target" "test" {
  db_instance_identifier = aws_db_instance.test.id
  db_proxy_name          = aws_db_proxy.test.name
  target_group_name      = "default"
}
`, rName)
}

func testAccProxyTargetConfig_Cluster(rName string) string {
	return testAccProxyTargetBaseConfig(rName) + fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine = "aurora-mysql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = "%[1]s"
  db_subnet_group_name   = aws_db_subnet_group.test.id
  engine                 = data.aws_rds_engine_version.test.engine
  engine_version         = data.aws_rds_engine_version.test.version
  master_username        = "test"
  master_password        = "testtest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    Name = "%[1]s"
  }
}

resource "aws_db_proxy_target" "test" {
  db_cluster_identifier = aws_rds_cluster.test.cluster_identifier
  db_proxy_name         = aws_db_proxy.test.name
  target_group_name     = "default"
}
`, rName)
}
