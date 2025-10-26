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

func TestAccRDSCreateSnapshotAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snapshotName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccCreateSnapshotActionConfig_basic(rName, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCreateSnapshotAction(ctx, rName, snapshotName),
				),
			},
		},
	})
}

func TestAccRDSCreateSnapshotAction_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snapshotName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccCreateSnapshotActionConfig_withTags(rName, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCreateSnapshotAction(ctx, rName, snapshotName),
					testAccCheckSnapshotHasTags(ctx, snapshotName, map[string]string{
						"Environment": "test",
						"Purpose":     "backup",
					}),
				),
			},
		},
	})
}

// Test helper functions

func testAccCheckCreateSnapshotAction(ctx context.Context, dbInstanceId, snapshotId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		// Verify the snapshot exists and is available
		input := &rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: &snapshotId,
		}

		output, err := conn.DescribeDBSnapshots(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe DB snapshot %s: %w", snapshotId, err)
		}

		if len(output.DBSnapshots) == 0 {
			return fmt.Errorf("DB snapshot %s not found", snapshotId)
		}

		snapshot := output.DBSnapshots[0]
		if *snapshot.Status != "available" {
			return fmt.Errorf("expected snapshot %s to be available, got status: %s", snapshotId, *snapshot.Status)
		}

		if *snapshot.DBInstanceIdentifier != dbInstanceId {
			return fmt.Errorf("expected snapshot to be from instance %s, got %s", dbInstanceId, *snapshot.DBInstanceIdentifier)
		}

		return nil
	}
}

func testAccCheckSnapshotHasTags(ctx context.Context, snapshotId string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		// Get snapshot ARN first
		describeInput := &rds.DescribeDBSnapshotsInput{
			DBSnapshotIdentifier: &snapshotId,
		}
		describeOutput, err := conn.DescribeDBSnapshots(ctx, describeInput)
		if err != nil {
			return fmt.Errorf("failed to describe DB snapshot %s: %w", snapshotId, err)
		}

		if len(describeOutput.DBSnapshots) == 0 {
			return fmt.Errorf("DB snapshot %s not found", snapshotId)
		}

		snapshotArn := *describeOutput.DBSnapshots[0].DBSnapshotArn

		// List tags
		tagsInput := &rds.ListTagsForResourceInput{
			ResourceName: &snapshotArn,
		}
		tagsOutput, err := conn.ListTagsForResource(ctx, tagsInput)
		if err != nil {
			return fmt.Errorf("failed to list tags for snapshot %s: %w", snapshotId, err)
		}

		actualTags := make(map[string]string)
		for _, tag := range tagsOutput.TagList {
			actualTags[*tag.Key] = *tag.Value
		}

		for expectedKey, expectedValue := range expectedTags {
			if actualValue, exists := actualTags[expectedKey]; !exists {
				return fmt.Errorf("expected tag %s not found on snapshot %s", expectedKey, snapshotId)
			} else if actualValue != expectedValue {
				return fmt.Errorf("expected tag %s to have value %s, got %s", expectedKey, expectedValue, actualValue)
			}
		}

		return nil
	}
}

// Configuration functions

func testAccCreateSnapshotActionConfig_basic(rName, snapshotName string) string {
	return acctest.ConfigCompose(
		testAccCreateSnapshotActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_rds_create_snapshot" "test" {
  config {
    db_instance_identifier = aws_db_instance.test.identifier
    snapshot_identifier    = %[1]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_rds_create_snapshot.test]
    }
  }
}
`, snapshotName))
}

func testAccCreateSnapshotActionConfig_withTags(rName, snapshotName string) string {
	return acctest.ConfigCompose(
		testAccCreateSnapshotActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_rds_create_snapshot" "test" {
  config {
    db_instance_identifier = aws_db_instance.test.identifier
    snapshot_identifier    = %[1]q
    tags = {
      Environment = "test"
      Purpose     = "backup"
    }
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_rds_create_snapshot.test]
    }
  }
}
`, snapshotName))
}

func testAccCreateSnapshotActionConfig_base(rName string) string {
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
