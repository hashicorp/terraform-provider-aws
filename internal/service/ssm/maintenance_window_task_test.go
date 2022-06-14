package ssm_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccSSMMaintenanceWindowTask_basic(t *testing.T) {
	var before, after ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssm", regexp.MustCompile(`windowtask/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "window_task_id"),
					resource.TestCheckResourceAttrPair(resourceName, "window_id", "aws_ssm_maintenance_window.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "1"),
				),
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_basicUpdate(rName, "test description", "RUN_COMMAND", "AWS-InstallPowerShellModule", 3, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("maintenance-window-task-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "task_type", "RUN_COMMAND"),
					resource.TestCheckResourceAttr(resourceName, "task_arn", "AWS-InstallPowerShellModule"),
					resource.TestCheckResourceAttr(resourceName, "priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "3"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "2"),
					testAccCheckWindowsTaskNotRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_noTarget(t *testing.T) {
	var before ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_noTarget(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_cutoff(t *testing.T) {
	var before ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_cutoff(rName, "CANCEL_TASK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "cutoff_behavior", "CANCEL_TASK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_cutoff(rName, "CONTINUE_TASK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "cutoff_behavior", "CONTINUE_TASK"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_noRole(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_noRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_updateForcesNewResource(t *testing.T) {
	var before, after ssm.MaintenanceWindowTask
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
				),
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_basicUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", "TestMaintenanceWindowTask"),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					testAccCheckWindowsTaskRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_description(t *testing.T) {
	var task1, task2 ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					testAccCheckWindowsTaskNotRecreated(t, &task1, &task2),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_taskInvocationAutomationParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_automation(rName, "$DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$DEFAULT"),
				),
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_automationUpdate(rName, "$LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$LATEST"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_taskInvocationLambdaParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	rString := sdkacctest.RandString(8)
	rInt := sdkacctest.RandInt()

	funcName := fmt.Sprintf("tf_acc_lambda_func_tags_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tags_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tags_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tags_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_lambda(funcName, policyName, roleName, sgName, rString, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_taskInvocationRunCommandParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	serviceRoleResourceName := "aws_iam_role.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_runCommand(rName, "test comment", 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "30"),
				),
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_runCommandUpdate(rName, "test comment update", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment update"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "60"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.output_s3_bucket", s3BucketResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_taskInvocationRunCommandParametersCloudWatch(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	serviceRoleResourceName := "aws_iam_role.test"
	cwResourceName := "aws_cloudwatch_log_group.test"

	name := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_runCommandCloudWatch(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.cloudwatch_config.0.cloudwatch_log_group_name", cwResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.cloudwatch_config.0.cloudwatch_output_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_runCommandCloudWatch(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.cloudwatch_config.0.cloudwatch_output_enabled", "false"),
				),
			},
			{
				Config: testAccMaintenanceWindowTaskConfig_runCommandCloudWatch(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.cloudwatch_config.0.cloudwatch_log_group_name", cwResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.cloudwatch_config.0.cloudwatch_output_enabled", "true"),
				),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_taskInvocationStepFunctionParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"
	rString := sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_stepFunction(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_emptyNotification(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_emptyNotifcation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &task),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTask_disappears(t *testing.T) {
	var before ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.test"

	name := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTaskConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTaskExists(resourceName, &before),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceMaintenanceWindowTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWindowsTaskNotRecreated(t *testing.T,
	before, after *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.WindowTaskId) != aws.StringValue(after.WindowTaskId) {
			t.Fatalf("Unexpected change of Windows Task IDs, but both were %s and %s", aws.StringValue(before.WindowTaskId), aws.StringValue(after.WindowTaskId))
		}
		return nil
	}
}

func testAccCheckWindowsTaskRecreated(t *testing.T,
	before, after *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.WindowTaskId == after.WindowTaskId {
			t.Fatalf("Expected change of Windows Task IDs, but both were %v", before.WindowTaskId)
		}
		return nil
	}
}

func testAccCheckMaintenanceWindowTaskExists(n string, task *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Task Window ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		resp, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})
		if err != nil {
			return err
		}

		for _, i := range resp.Tasks {
			if aws.StringValue(i.WindowTaskId) == rs.Primary.ID {
				*task = *i
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window task found")
	}
}

func testAccCheckMaintenanceWindowTaskDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window_task" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})

		if err != nil {
			// Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
				continue
			}
			return err
		}

		if len(out.Tasks) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Task to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccMaintenanceWindowTaskImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}

func testAccMaintenanceWindowTaskBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  resource_type = "INSTANCE"
  window_id     = aws_ssm_maintenance_window.test.id

  targets {
    key    = "tag:Name"
    values = ["tf-acc-test"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
POLICY
}
`, rName)
}

func testAccMaintenanceWindowTaskConfig_basic(rName string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}
`)
}

func testAccMaintenanceWindowTaskConfig_noTarget(rName string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
}
`)
}

func testAccMaintenanceWindowTaskConfig_cutoff(rName, cutoff string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  cutoff_behavior  = %[1]q
}
`, cutoff)
}

func testAccMaintenanceWindowTaskConfig_basicUpdate(rName, description, taskType, taskArn string, priority, maxConcurrency, maxErrors int) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = %[2]q
  task_arn         = %[3]q
  name             = "maintenance-window-task-%[1]s"
  description      = %[4]q
  priority         = %[5]d
  service_role_arn = aws_iam_role.ssm_role_update.arn
  max_concurrency  = %[6]d
  max_errors       = %[7]d

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}

resource "aws_iam_role" "ssm_role_update" {
  name = "ssm-role-update-%[1]s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_update_%[1]s"
  role = aws_iam_role.ssm_role_update.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}
`, rName, taskType, taskArn, description, priority, maxConcurrency, maxErrors)
}

func testAccMaintenanceWindowTaskConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  name             = "TestMaintenanceWindowTask"
  description      = "This resource is for test purpose only"
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`)
}

func testAccMaintenanceWindowTaskConfig_description(rName string, description string) string {
	return acctest.ConfigCompose(
		testAccMaintenanceWindowTaskBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ssm_maintenance_window_task" "test" {
  description     = %[1]q
  max_concurrency = 2
  max_errors      = 1
  task_arn        = "AWS-RunShellScript"
  task_type       = "RUN_COMMAND"
  window_id       = aws_ssm_maintenance_window.test.id

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}
`, description))
}

func testAccMaintenanceWindowTaskConfig_emptyNotifcation(rName string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName) + `

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      timeout_seconds = 600

      notification_config {}

      parameter {
        name   = "Operation"
        values = ["Install"]
      }
    }
  }
}
`)
}

func testAccMaintenanceWindowTaskConfig_noRole(rName string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName) + `
resource "aws_ssm_maintenance_window_task" "test" {
  description     = "This resource is for test purpose only"
  max_concurrency = 2
  max_errors      = 1
  name            = "TestMaintenanceWindowTask"
  priority        = 1
  task_arn        = "AWS-RunShellScript"
  task_type       = "RUN_COMMAND"
  window_id       = aws_ssm_maintenance_window.test.id

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      parameter {
        name   = "commands"
        values = ["pwd"]
      }
    }
  }
}
`)
}

func testAccMaintenanceWindowTaskConfig_automation(rName, version string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    automation_parameters {
      document_version = %[2]q

      parameter {
        name   = "InstanceId"
        values = ["{{TARGET_ID}}"]
      }

      parameter {
        name   = "NoReboot"
        values = ["false"]
      }
    }
  }
}
`, rName, version)
}

func testAccMaintenanceWindowTaskConfig_automationUpdate(rName, version string) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "AUTOMATION"
  task_arn         = "AWS-CreateImage"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    automation_parameters {
      document_version = %[2]q

      parameter {
        name   = "InstanceId"
        values = ["{{TARGET_ID}}"]
      }

      parameter {
        name   = "NoReboot"
        values = ["false"]
      }
    }
  }
}
`, rName, version)
}

func testAccMaintenanceWindowTaskConfig_lambda(funcName, policyName, roleName, sgName, rName string, rInt int) string {
	return fmt.Sprintf(testAccLambdaBasicConfig(funcName, policyName, roleName, sgName)+
		testAccMaintenanceWindowTaskBaseConfig(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "LAMBDA"
  task_arn         = aws_lambda_function.test.arn
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    lambda_parameters {
      client_context = base64encode(jsonencode({
        key1 = "value1"
        key2 = "value2"
        key3 = "value3"
      }))
      payload = jsonencode({
        number = %[2]d
      })
    }
  }
}
`, rName, rInt)
}

func testAccMaintenanceWindowTaskConfig_runCommand(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      comment            = %[2]q
      document_hash      = sha256("COMMAND")
      document_hash_type = "Sha256"
      service_role_arn   = aws_iam_role.test.arn
      timeout_seconds    = %[3]d

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`, rName, comment, timeoutSeconds)
}

func testAccMaintenanceWindowTaskConfig_runCommandUpdate(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      comment              = %[2]q
      document_hash        = sha256("COMMAND")
      document_hash_type   = "Sha256"
      service_role_arn     = aws_iam_role.test.arn
      timeout_seconds      = %[3]d
      output_s3_bucket     = aws_s3_bucket.test.id
      output_s3_key_prefix = "foo"

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
`, rName, comment, timeoutSeconds)
}

func testAccMaintenanceWindowTaskConfig_runCommandCloudWatch(rName string, enabled bool) string {
	return fmt.Sprintf(testAccMaintenanceWindowTaskBaseConfig(rName)+`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "RUN_COMMAND"
  task_arn         = "AWS-RunShellScript"
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      document_hash      = sha256("COMMAND")
      document_hash_type = "Sha256"
      service_role_arn   = aws_iam_role.test.arn

      parameter {
        name   = "commands"
        values = ["date"]
      }

      cloudwatch_config {
        cloudwatch_log_group_name = aws_cloudwatch_log_group.test.name
        cloudwatch_output_enabled = %[2]t
      }
    }
  }
}
`, rName, enabled)
}

func testAccMaintenanceWindowTaskConfig_stepFunction(rName string) string {
	return testAccMaintenanceWindowTaskBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = %[1]q
}

resource "aws_ssm_maintenance_window_task" "test" {
  window_id        = aws_ssm_maintenance_window.test.id
  task_type        = "STEP_FUNCTIONS"
  task_arn         = aws_sfn_activity.test.id
  priority         = 1
  service_role_arn = aws_iam_role.test.arn
  max_concurrency  = "2"
  max_errors       = "1"

  targets {
    key    = "WindowTargetIds"
    values = [aws_ssm_maintenance_window_target.test.id]
  }

  task_invocation_parameters {
    step_functions_parameters {
      input = jsonencode({
        key1 = "value1"
        key2 = "value2"
        key3 = "value3"
      })
      name = "tf-step-function-%[1]s"
    }
  }
}
`, rName)
}

func testAccLambdaBasicConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}
