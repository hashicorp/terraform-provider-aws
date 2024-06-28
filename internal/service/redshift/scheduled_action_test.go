// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftScheduledAction_basicPauseCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduledActionConfig_pauseCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_pauseClusterWithOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseClusterFullOptions(rName, "cron(00 * * * ? *)", "This is test action", true, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is test action"),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(00 * * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.0.cluster_identifier", "tf-test-identifier"),
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

func TestAccRedshiftScheduledAction_basicResumeCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resumeCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduledActionConfig_resumeCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_basicResizeCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resizeClusterBasic(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduledActionConfig_resizeClusterBasic(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_resizeClusterWithOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resizeClusterFullOptions(rName, "cron(00 23 * * ? *)", true, "multi-node", "dc2.large", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.classic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_identifier", "tf-test-identifier"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.node_type", "dc2.large"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.number_of_nodes", acctest.Ct2),
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

func TestAccRedshiftScheduledAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceScheduledAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScheduledActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_scheduled_action" {
				continue
			}

			_, err := tfredshift.FindScheduledActionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Scheduled Action %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScheduledActionExists(ctx context.Context, n string, v *redshift.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Scheduled Action ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		output, err := tfredshift.FindScheduledActionByName(ctx, conn, rs.Primary.ID)

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
          "redshift.amazonaws.com"
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
              "redshift:PauseCluster",
              "redshift:ResumeCluster",
              "redshift:ResizeCluster"
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

func testAccScheduledActionConfig_pauseCluster(rName, schedule string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshift_scheduled_action" "test" {
  name     = %[1]q
  schedule = %[2]q
  iam_role = aws_iam_role.test.arn

  target_action {
    pause_cluster {
      cluster_identifier = "tf-test-identifier"
    }
  }
}
`, rName, schedule))
}

func testAccScheduledActionConfig_pauseClusterFullOptions(rName, schedule, description string, enable bool, startTime, endTime string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshift_scheduled_action" "test" {
  name        = %[1]q
  description = %[2]q
  enable      = %[3]t
  start_time  = %[4]q
  end_time    = %[5]q
  schedule    = %[6]q
  iam_role    = aws_iam_role.test.arn

  target_action {
    pause_cluster {
      cluster_identifier = "tf-test-identifier"
    }
  }
}
`, rName, description, enable, startTime, endTime, schedule))
}

func testAccScheduledActionConfig_resumeCluster(rName, schedule string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshift_scheduled_action" "test" {
  name     = %[1]q
  schedule = %[2]q
  iam_role = aws_iam_role.test.arn

  target_action {
    resume_cluster {
      cluster_identifier = "tf-test-identifier"
    }
  }
}
`, rName, schedule))
}

func testAccScheduledActionConfig_resizeClusterBasic(rName, schedule string) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshift_scheduled_action" "test" {
  name     = %[1]q
  schedule = %[2]q
  iam_role = aws_iam_role.test.arn

  target_action {
    resize_cluster {
      cluster_identifier = "tf-test-identifier"
    }
  }
}
`, rName, schedule))
}

func testAccScheduledActionConfig_resizeClusterFullOptions(rName, schedule string, classic bool, clusterType, nodeType string, numberOfNodes int) string {
	return acctest.ConfigCompose(testAccScheduledActionBaseConfig(rName), fmt.Sprintf(`
resource "aws_redshift_scheduled_action" "test" {
  name     = %[1]q
  schedule = %[2]q
  iam_role = aws_iam_role.test.arn

  target_action {
    resize_cluster {
      cluster_identifier = "tf-test-identifier"
      classic            = %[3]t
      cluster_type       = %[4]q
      node_type          = %[5]q
      number_of_nodes    = %[6]d
    }
  }
}
`, rName, schedule, classic, clusterType, nodeType, numberOfNodes))
}

func TestAccRedshiftScheduledAction_validScheduleName(t *testing.T) {
	t.Parallel()

	var f = validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]{1,63}$`), "")

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
