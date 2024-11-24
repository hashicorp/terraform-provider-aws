// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessScheduledAction_basicCreateSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshiftserverless.ScheduledAction
	resourceName := "aws_redshiftserverless_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_createSnapshot(rName, "(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.retention_period", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.namespace_name", "aws_redshiftserverless_namespace.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.snapshot_name", "aws_redshiftserverless_snapshot.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduledActionConfig_createSnapshot(rName, "(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.retention_period", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.namespace_name", "aws_redshiftserverless_namespace.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.snapshot_name", "aws_redshiftserverless_snapshot.test", "id"),
				),
			},
		},
	})
}

func TestAccRedshiftServerlessScheduledAction_createSnapshotWithOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshiftserverless.ScheduledAction
	resourceName := "aws_redshiftserverless_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_CreateSnapshotFullOptions(rName, "(00 * * * ? *)", "This is test action", true, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "This is test action"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "(00 * * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.retention_period", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.namespace_name", "aws_redshiftserverless_namespace.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.create_snapshot.0.snapshot_name", "aws_redshiftserverless_snapshot.test", "id"),
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

func TestAccRedshiftServerlessScheduledAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshiftserverless.ScheduledAction
	resourceName := "aws_redshiftserverless_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_createSnapshot(rName, "(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceScheduledAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScheduledActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_scheduled_action" {
				continue
			}

			_, err := tfredshiftserverless.FindScheduledActionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Serverless Scheduled Action %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScheduledActionExists(ctx context.Context, n string, v *redshiftserverless.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Scheduled Action is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn(ctx)

		output, err := tfredshiftserverless.FindScheduledActionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccScheduledActionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "scheduler.redshift.amazonaws.com",
          "redshift-serverless.amazonaws.com""
        ]
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Sid": "VisualEditor0",
          "Effect": "Allow",
          "Action": [
              "redshift-serverless:CreateScheduledAction",
              "redshift-serverless:CreateSnapshot"
          ],
          "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = aws_iam_policy.test.arn
  role       = aws_iam_role.test.name
}
`, rName, rName)
}

func testAccScheduledActionConfig_createSnapshot(rName, schedule string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}
	
resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_scheduled_action" "test" {
  name     = %[1]q
  schedule = %[2]q
  role_arn = aws_iam_role.test.arn

  target_action {
    create_snapshot {
      namespace_name = aws_redshiftserverless_workgroup.test.namespace_name
	  snapshot_name  = %[1]q
	  retention_period = 1
    }
  }
}
`, rName, schedule))
}

func testAccScheduledActionConfig_createSnapshotFullOptions(rName, schedule, description string, enabled bool, startTime, endTime string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}
	
resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftserverless_scheduled_action" "test" {
  name        = %[1]q
  description = %[2]q
  enabled     = %[3]t
  start_time  = %[4]q
  end_time    = %[5]q
  schedule    = %[6]q
  role_arn    = aws_iam_role.test.arn

  target_action {
    create_snapshot {
      namespace_name = aws_redshiftserverless_workgroup.test.namespace_name
	  snapshot_name  = %[1]q
	  retention_period = 1
    }
  }
}
`, rName, description, enabled, startTime, endTime, schedule))
}

func TestAccRedshiftScheduledAction_validScheduleName(t *testing.T) {
	t.Parallel()

	var f = validation.StringMatch(regexache.MustCompile(`^[a-z0-9-]+{3,60}$`), "")

	validIds := []string{
		"tf-test-schedule-action-1",
		acctest.ResourcePrefix,
		sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
	}

	for _, s := range validIds {
		_, errors := f(s, "")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication instance id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_schedule-action_1",
		"tfTestScheduleACtion",
		"tf.test.schedule.action.1",
		"tf test schedule action 1",
		"tf-test-schedule-action-1!",
	}

	for _, s := range invalidIds {
		_, errors := f(s, "")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication instance id: %v", s, errors)
		}
	}
}
