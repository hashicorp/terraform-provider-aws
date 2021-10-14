package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/redshift/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_redshift_scheduled_action", &resource.Sweeper{
		Name: "aws_redshift_scheduled_action",
		F:    testSweepRedshiftScheduledActions,
	})
}

func testSweepRedshiftScheduledActions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).redshiftconn
	input := &redshift.DescribeScheduledActionsInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeScheduledActionsPages(input, func(page *redshift.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, scheduledAction := range page.ScheduledActions {
			r := resourceAwsRedshiftScheduledAction()
			d := r.Data(nil)
			d.SetId(aws.StringValue(scheduledAction.ScheduledActionName))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Redshift Scheduled Action sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Redshift Scheduled Actions (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Redshift Scheduled Actions (%s): %w", region, err)
	}

	return nil
}

func TestAccAWSRedshiftScheduledAction_basicPauseCluster(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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
				Config: testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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

func TestAccAWSRedshiftScheduledAction_PauseClusterWithOptions(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	startTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseClusterWithFullOptions(rName, "cron(00 * * * ? *)", "This is test action", true, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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

func TestAccAWSRedshiftScheduledAction_basicResumeCluster(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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
				Config: testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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

func TestAccAWSRedshiftScheduledAction_basicResizeCluster(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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

func TestAccAWSRedshiftScheduledAction_ResizeClusterWithOptions(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterWithFullOptions(rName, "cron(00 23 * * ? *)", true, "multi-node", "dc1.large", 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
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
					resource.TestCheckResourceAttr(resourceName, "target_action.0.resize_cluster.0.node_type", "dc1.large"),
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

func TestAccAWSRedshiftScheduledAction_disappears(t *testing.T) {
	var v redshift.ScheduledAction
	resourceName := "aws_redshift_scheduled_action.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, redshift.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRedshiftScheduledAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSRedshiftScheduledActionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_scheduled_action" {
			continue
		}

		_, err := finder.ScheduledActionByName(conn, rs.Primary.ID)

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

func testAccCheckAWSRedshiftScheduledActionExists(n string, v *redshift.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Scheduled Action ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn

		output, err := finder.ScheduledActionByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSRedshiftScheduledActionConfigBase(rName string) string {
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

func testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, schedule string) string {
	return composeConfig(testAccAWSRedshiftScheduledActionConfigBase(rName), fmt.Sprintf(`
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

func testAccAWSRedshiftScheduledActionConfigPauseClusterWithFullOptions(rName, schedule, description string, enable bool, startTime, endTime string) string {
	return composeConfig(testAccAWSRedshiftScheduledActionConfigBase(rName), fmt.Sprintf(`
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

func testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, schedule string) string {
	return composeConfig(testAccAWSRedshiftScheduledActionConfigBase(rName), fmt.Sprintf(`
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

func testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, schedule string) string {
	return composeConfig(testAccAWSRedshiftScheduledActionConfigBase(rName), fmt.Sprintf(`
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

func testAccAWSRedshiftScheduledActionConfigResizeClusterWithFullOptions(rName, schedule string, classic bool, clusterType, nodeType string, numberOfNodes int) string {
	return composeConfig(testAccAWSRedshiftScheduledActionConfigBase(rName), fmt.Sprintf(`
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
