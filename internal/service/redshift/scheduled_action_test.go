package redshift_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftScheduledAction_basicPauseCluster(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
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
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_pauseClusterWithOptions(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseClusterFullOptions(rName, "cron(00 * * * ? *)", "This is test action", true, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "This is test action"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 * * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
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
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resumeCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "1"),
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
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_basicResizeCluster(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resizeClusterBasic(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
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
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_identifier", "tf-test-identifier"),
				),
			},
		},
	})
}

func TestAccRedshiftScheduledAction_resizeClusterWithOptions(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_resizeClusterFullOptions(rName, "cron(00 23 * * ? *)", true, "multi-node", "dc2.large", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.pause_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resume_cluster.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.classic", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_identifier", "tf-test-identifier"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.node_type", "dc2.large"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.number_of_nodes", "2"),
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
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_pauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduledActionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceScheduledAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScheduledActionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_scheduled_action" {
			continue
		}

		_, err := tfredshift.FindScheduledActionByName(conn, rs.Primary.ID)

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

func testAccCheckScheduledActionExists(n string, v *redshift.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Scheduled Action ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		output, err := tfredshift.FindScheduledActionByName(conn, rs.Primary.ID)

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
