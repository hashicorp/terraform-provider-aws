package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfscheduler "github.com/hashicorp/terraform-provider-aws/internal/service/scheduler"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceScheduleIDFromARN(t *testing.T) {
	testCases := []struct {
		ARN   string
		ID    string
		Fails bool
	}{
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test",
			ID:    "default/test",
			Fails: false,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test/test",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//test",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "",
			ID:    "",
			Fails: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ARN, func(t *testing.T) {
			id, err := tfscheduler.ResourceScheduleIDFromARN(tc.ARN)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if id != tc.ID {
				t.Errorf("expected id %s, got: %s", tc.ID, id)
			}
		})
	}
}

func TestResourceScheduleParseID(t *testing.T) {
	testCases := []struct {
		ID           string
		GroupName    string
		ScheduleName string
		Fails        bool
	}{
		{
			ID:           "default/test",
			GroupName:    "default",
			ScheduleName: "test",
			Fails:        false,
		},
		{
			ID:           "default/test/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "//",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			groupName, scheduleName, err := tfscheduler.ResourceScheduleParseID(tc.ID)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if groupName != tc.GroupName {
				t.Errorf("expected group name %s, got: %s", tc.GroupName, groupName)
			}

			if scheduleName != tc.ScheduleName {
				t.Errorf("expected schedule name %s, got: %s", tc.ScheduleName, scheduleName)
			}
		})
	}
}

func TestAccSchedulerSchedule_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "scheduler", regexp.MustCompile(regexp.QuoteMeta(`schedule/default/`+name))),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("default/%s", name)),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", "arn"),
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

func TestAccSchedulerSchedule_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.CheckResourceDisappears(acctest.Provider, tfscheduler.ResourceSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_scheduleExpression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 hour)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 day)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 day)"),
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
func TestAccSchedulerSchedule_targetArn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetArn(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.0", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetArn(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.1", "arn"),
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

func TestAccSchedulerSchedule_targetRoleArn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetRoleArn(name, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetRoleArn(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test1", "arn"),
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

func testAccCheckScheduleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SchedulerClient
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_scheduler_schedule" {
			continue
		}

		parts := strings.Split(rs.Primary.ID, "/")

		input := &scheduler.GetScheduleInput{
			GroupName: aws.String(parts[0]),
			Name:      aws.String(parts[1]),
		}
		_, err := conn.GetSchedule(ctx, input)
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.Scheduler, create.ErrActionCheckingDestroyed, tfscheduler.ResNameSchedule, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckScheduleExists(name string, schedule *scheduler.GetScheduleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not set"))
		}

		parts := strings.Split(rs.Primary.ID, "/")

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchedulerClient
		ctx := context.Background()
		resp, err := conn.GetSchedule(ctx, &scheduler.GetScheduleInput{
			Name:      aws.String(parts[1]),
			GroupName: aws.String(parts[0]),
		})

		if err != nil {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, rs.Primary.ID, err)
		}

		*schedule = *resp

		return nil
	}
}

const testAccScheduleConfig_base = `
data "aws_partition" "main" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
    }
  })
}
`

func testAccScheduleConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name),
	)
}

func testAccScheduleConfig_scheduleExpression(name, expression string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, expression),
	)
}

func testAccScheduleConfig_targetArn(name string, i int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test[%[2]d].arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, i),
	)
}

func testAccScheduleConfig_targetRoleArn(name, resourceName string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_iam_role" "test1" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
    }
  })
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.%[2]s.arn
  }
}
`, name, resourceName),
	)
}
