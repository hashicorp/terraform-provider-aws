// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSIntegration_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var integration awstypes.Integration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_baseClusterWithInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					waitUntilDBInstanceRebooted(ctx, rName),
				),
			},
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &integration),
					resource.TestCheckResourceAttr(resourceName, "integration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "data_filter", "include: *.*"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_rds_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, "aws_redshiftserverless_namespace.test", names.AttrARN),
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

func TestAccRDSIntegration_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var integration awstypes.Integration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_baseClusterWithInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					waitUntilDBInstanceRebooted(ctx, rName),
				),
			},
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &integration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrds.ResourceIntegration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSIntegration_optional(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var integration awstypes.Integration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_baseClusterWithInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					waitUntilDBInstanceRebooted(ctx, rName),
				),
			},
			{
				Config: testAccIntegrationConfig_optional(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &integration),
					resource.TestCheckResourceAttr(resourceName, "integration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "data_filter", "include: test.mytable"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_rds_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, "aws_redshiftserverless_namespace.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.department", "test"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func testAccCheckIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_integration" {
				continue
			}

			_, err := tfrds.FindIntegrationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Integration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIntegrationExists(ctx context.Context, n string, v *awstypes.Integration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindIntegrationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func waitUntilDBInstanceRebooted(ctx context.Context, instanceIdentifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Wait for rebooting.
		time.Sleep(60 * time.Second)

		_, err := tfrds.WaitDBInstanceAvailable(ctx, acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx), instanceIdentifier, 30*time.Minute)

		return err
	}
}

func testAccIntegrationConfig_baseClusterWithInstance(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
locals {
  cluster_parameters = {
    "binlog_replication_globaldb" = {
      value        = "0"
      apply_method = "pending-reboot"
    },
    "binlog_format" = {
      value        = "ROW"
      apply_method = "pending-reboot"
    },
    "binlog_row_metadata" = {
      value        = "full"
      apply_method = "immediate"
    },
    "binlog_row_image" = {
      value        = "full"
      apply_method = "immediate"
    },
    "aurora_enhanced_binlog" = {
      value        = "1"
      apply_method = "pending-reboot"
    },
    "binlog_backup" = {
      value        = "0"
      apply_method = "pending-reboot"
    },
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
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

resource "aws_rds_cluster_parameter_group" "test" {
  name   = %[1]q
  family = "aurora-mysql8.0"

  dynamic "parameter" {
    for_each = local.cluster_parameters
    content {
      name         = parameter.key
      value        = parameter.value["value"]
      apply_method = parameter.value["apply_method"]
    }
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-mysql"
  engine_version      = "8.0.mysql_aurora.3.05.2"
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true

  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name

  apply_immediately = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.r6g.large"
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zone   = data.aws_availability_zones.available.names[0]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "Mustbe8characters"
  node_type           = "dc2.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true
}
`, rName))
}

func testAccIntegrationConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_baseClusterWithInstance(rName), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
  base_capacity  = 8

  publicly_accessible = false
  subnet_ids          = aws_subnet.test[*].id

  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "auto_mv"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "datestyle"
    parameter_value = "ISO, MDY"
  }
  config_parameter {
    parameter_key   = "enable_user_activity_logging"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "max_query_execution_time"
    parameter_value = "14400"
  }
  config_parameter {
    parameter_key   = "query_group"
    parameter_value = "default"
  }
  config_parameter {
    parameter_key   = "require_ssl"
    parameter_value = "false"
  }
  config_parameter {
    parameter_key   = "search_path"
    parameter_value = "$user, public"
  }
  config_parameter {
    parameter_key   = "use_fips_ssl"
    parameter_value = "false"
  }
}

# The "aws_redshiftserverless_resource_policy" resource doesn't support the following action types.
# Therefore we need to use the "aws_redshift_resource_policy" resource for RedShift-serverless instead.
resource "aws_redshift_resource_policy" "test" {
  resource_arn = aws_redshiftserverless_namespace.test.arn
  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action   = "redshift:CreateInboundIntegration"
      Resource = aws_redshiftserverless_namespace.test.arn
      }, {
      Effect = "Allow"
      Principal = {
        Service = "redshift.amazonaws.com"
      }
      Action   = "redshift:AuthorizeInboundIntegration"
      Resource = aws_redshiftserverless_namespace.test.arn
      Condition = {
        StringEquals = {
          "aws:SourceArn" = aws_rds_cluster.test.arn
        }
      }
    }]
  })
}
`, rName))
}

func testAccIntegrationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_integration" "test" {
  integration_name = %[1]q
  source_arn       = aws_rds_cluster.test.arn
  target_arn       = aws_redshiftserverless_namespace.test.arn

  depends_on = [
    aws_rds_cluster.test,
    aws_redshiftserverless_namespace.test,
    aws_redshiftserverless_workgroup.test,
    aws_redshift_resource_policy.test,
  ]
}
`, rName))
}

func testAccIntegrationConfig_optional(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
  policy                  = data.aws_iam_policy_document.key_policy.json
}

data "aws_iam_policy_document" "key_policy" {
  statement {
    actions   = ["kms:*"]
    resources = ["*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    actions   = ["kms:CreateGrant"]
    resources = ["*"]
    principals {
      type        = "Service"
      identifiers = ["redshift.amazonaws.com"]
    }
  }
}

resource "aws_rds_integration" "test" {
  integration_name = %[1]q
  source_arn       = aws_rds_cluster.test.arn
  target_arn       = aws_redshiftserverless_namespace.test.arn
  kms_key_id       = aws_kms_key.test.arn

  additional_encryption_context = {
    "department" : "test",
  }

  data_filter = "include: test.mytable"

  tags = {
    Name = %[1]q
  }

  depends_on = [
    aws_rds_cluster.test,
    aws_redshiftserverless_namespace.test,
    aws_redshiftserverless_workgroup.test,
    aws_redshift_resource_policy.test,
  ]
}
`, rName))
}
