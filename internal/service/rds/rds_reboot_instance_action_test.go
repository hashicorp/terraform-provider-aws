// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSRebootInstanceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RDS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRebootInstanceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRebootInstanceAction(ctx, rName),
				),
			},
		},
	})
}

func TestAccRDSRebootInstanceAction_forceFailover(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RDS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckDBInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRebootInstanceActionConfig_forceFailover(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRebootInstanceAction(ctx, rName),
				),
			},
		},
	})
}

// Test helper functions

func testAccCheckRebootInstanceAction(ctx context.Context, dbInstanceId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		// Verify the instance exists and is available
		input := &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: &dbInstanceId,
		}

		output, err := conn.DescribeDBInstances(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe DB instance %s: %w", dbInstanceId, err)
		}

		if len(output.DBInstances) == 0 {
			return fmt.Errorf("DB instance %s not found", dbInstanceId)
		}

		instance := output.DBInstances[0]
		if *instance.DBInstanceStatus != "available" {
			return fmt.Errorf("expected instance %s to be available after reboot, got status: %s", dbInstanceId, *instance.DBInstanceStatus)
		}

		return nil
	}
}

// Configuration functions

func testAccRebootInstanceActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRebootInstanceActionConfig_base(rName),
		`
action "aws_rds_reboot_instance" "test" {
  config {
    db_instance_identifier = aws_db_instance.test.identifier
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_rds_reboot_instance.test]
    }
  }
}
`)
}

func testAccRebootInstanceActionConfig_forceFailover(rName string) string {
	return acctest.ConfigCompose(
		testAccRebootInstanceActionConfig_baseMultiAZ(rName),
		`
action "aws_rds_reboot_instance" "test" {
  config {
    db_instance_identifier = aws_db_instance.test.identifier
    force_failover          = true
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_rds_reboot_instance.test]
    }
  }
}
`)
}

func testAccRebootInstanceActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier     = %[1]q
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.micro"
  
  allocated_storage = 20
  storage_type      = "gp2"
  
  db_name  = "testdb"
  username = "testuser"
  password = "testpass123"
  
  skip_final_snapshot = true
  
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccRebootInstanceActionConfig_baseMultiAZ(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_instance" "test" {
  identifier     = %[1]q
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.small"  # Multi-AZ requires larger instance
  
  allocated_storage = 20
  storage_type      = "gp2"
  multi_az          = true
  
  db_name  = "testdb"
  username = "testuser"
  password = "testpass123"
  
  skip_final_snapshot = true
  
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
