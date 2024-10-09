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

func TestAccRDSProxyTarget_instance(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget types.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(ctx, resourceName, &dbProxyTarget),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEndpoint, "aws_db_instance.test", names.AttrAddress),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPort, "aws_db_instance.test", names.AttrPort),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTargetARN, ""),
					resource.TestCheckResourceAttr(resourceName, "tracked_cluster_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "RDS_INSTANCE"),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget types.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_cluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(ctx, resourceName, &dbProxyTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpoint, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPort, "aws_rds_cluster.test", names.AttrPort),
					resource.TestCheckResourceAttr(resourceName, "rds_resource_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTargetARN, ""),
					resource.TestCheckResourceAttr(resourceName, "tracked_cluster_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "TRACKED_CLUSTER"),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxyTarget types.DBProxyTarget
	resourceName := "aws_db_proxy_target.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyTargetConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyTargetExists(ctx, resourceName, &dbProxyTarget),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceProxyTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProxyTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_target" {
				continue
			}

			dbProxyName, targetGroupName, targetType, rdsResourceID, err := tfrds.ProxyTargetParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfrds.FindDBProxyTargetByFourPartKey(ctx, conn, dbProxyName, targetGroupName, targetType, rdsResourceID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Proxy Target %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyTargetExists(ctx context.Context, n string, v *types.DBProxyTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		dbProxyName, targetGroupName, targetType, rdsResourceID, err := tfrds.ProxyTargetParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfrds.FindDBProxyTargetByFourPartKey(ctx, conn, dbProxyName, targetGroupName, targetType, rdsResourceID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyTargetConfig_base(rName string) string {
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

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id

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

  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccProxyTargetConfig_instance(rName string) string {
	return acctest.ConfigCompose(testAccProxyTargetConfig_base(rName), fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["8.0.33", "8.0.32", "8.0.31"]
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
  identifier             = %[1]q
  instance_class         = data.aws_rds_orderable_db_instance.test.instance_class
  password               = "testtest"
  skip_final_snapshot    = true
  username               = "test"
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_proxy_target" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_proxy_name          = aws_db_proxy.test.name
  target_group_name      = "default"
}
`, rName))
}

func testAccProxyTargetConfig_cluster(rName string) string {
	return acctest.ConfigCompose(testAccProxyTargetConfig_base(rName), fmt.Sprintf(`
data "aws_rds_engine_version" "test" {
  engine = "aurora-mysql"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = %[1]q
  db_subnet_group_name   = aws_db_subnet_group.test.id
  engine                 = data.aws_rds_engine_version.test.engine
  engine_version         = data.aws_rds_engine_version.test.version
  master_username        = "test"
  master_password        = "testtest"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_proxy_target" "test" {
  db_cluster_identifier = aws_rds_cluster.test.cluster_identifier
  db_proxy_name         = aws_db_proxy.test.name
  target_group_name     = "default"
}
`, rName))
}
