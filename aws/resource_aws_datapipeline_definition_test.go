package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/datapipeline"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDataPipelineDefinition_basic(t *testing.T) {
	var conf datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	bucketName := fmt.Sprintf("tf-test-pipeline-log-bucket-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "cron"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline_pipeline.default", "id",
						resourceName, "pipeline_id"),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.schedule_type", "cron"),
					resource.TestCheckResourceAttr(resourceName, "default.0.failure_and_rerun_mode", "none"),
					resource.TestCheckResourceAttrPair(
						"aws_iam_role.role", "arn",
						resourceName, "default.0.role"),
					resource.TestCheckResourceAttrPair(
						"aws_iam_instance_profile.resource_role", "arn",
						resourceName, "default.0.resource_role"),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "ondemand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.schedule_type", "ondemand"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigDefaultUpdate(rName, "ondemand", bucketName, "cascade"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.failure_and_rerun_mode", "cascade"),
					resource.TestCheckResourceAttr(resourceName, "default.0.pipeline_log_uri", fmt.Sprintf("s3://%s/log_prefix", bucketName)),
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

func TestAccAWSDataPipelineDefinition_schedule(t *testing.T) {
	var conf datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"
	nowTime := time.Now()
	startDateTime1 := nowTime.Format("2006-01-02T15:04:05")
	startDateTime2 := nowTime.AddDate(0, 0, -1).Format("2006-01-02T15:04:05")
	endDateTime1 := nowTime.AddDate(0, 1, 0).Format("2006-01-02T15:04:05")
	endDateTime2 := nowTime.AddDate(1, 0, 0).Format("2006-01-02T15:04:05")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "cron"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithSchedule(rName, "DefaultSchedule", "1 hours", startDateTime1, endDateTime1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.id", "DefaultSchedule"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.name", "DefaultSchedule"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.period", "1 hours"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_date_time", startDateTime1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.end_date_time", endDateTime1),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithSchedule(rName, "DefaultSchedule1", "1 Day", startDateTime2, endDateTime2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.id", "DefaultSchedule1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.name", "DefaultSchedule1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_date_time", startDateTime2),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.end_date_time", endDateTime2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithScheduleUpdate(rName, "DefaultSchedule1", "1 Day", "FIRST_ACTIVATION_DATE_TIME", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.start_at", "FIRST_ACTIVATION_DATE_TIME"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.occurrences", "10"),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "cron"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSDataPipelineDefinition_parameters(t *testing.T) {
	var conf datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithParameter(rName, "myDDBTableName", "test-table"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.optional", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.string_value", "test-table"),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithParameterUpdate(rName, "myDDBTableName", "test-table2", "myItemList", "item"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.default", rName),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.optional", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.1.id", "myItemList"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.1.is_array", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.1.description", "Array value myItemList"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.string_value", "test-table2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.1.id", "myItemList"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.1.string_value", "item-1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.2.id", "myItemList"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.2.string_value", "item-2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigWithParameter(rName, "myDDBTableName", "test-table"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.optional", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.id", "myDDBTableName"),
					resource.TestCheckResourceAttr(resourceName, "parameter_value.0.string_value", "test-table"),
				),
			},
		},
	})
}

func TestAccAWSDataPipelineDefinition_disappears(t *testing.T) {
	var description datapipeline.PipelineDescription
	var definition datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "ondemand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists("aws_datapipeline_pipeline.default", &description),
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &definition),
					testAccCheckAWSDataPipelinePipelineDisappears(&description),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDataPipelineDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_definition" {
			continue
		}
		// Try to find the Pipeline
		pipelineDefinition, err := resourceAwsDataPipelineDefinitionRetrieve(rs.Primary.ID, conn)
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") {
			continue
		} else if isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			continue
		}

		if err != nil {
			return err
		}
		if pipelineDefinition != nil {
			return fmt.Errorf("Pipeline Definition still exists")
		}
	}

	return nil
}

func testAccCheckAWSDataPipelineDefinitionExists(n string, v *datapipeline.GetPipelineDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

		pipelineDefinition, err := resourceAwsDataPipelineDefinitionRetrieve(rs.Primary.ID, conn)

		if err != nil {
			return err
		}
		if pipelineDefinition == nil {
			return fmt.Errorf("DataPipeline Definition not found")
		}

		*v = *pipelineDefinition
		return nil
	}
}

func testAccAWSDataPipelineDefinitionConfigCommon(rName string) string {
	return fmt.Sprintf(testAccAWSDataPipelinePipelineConfig(rName)+`
resource "aws_iam_role" "role" {
  name = "tf-test-datapipeline-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "elasticmapreduce.amazonaws.com",
          "datapipeline.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

data "aws_iam_policy" "role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}

resource "aws_iam_role_policy_attachment" "role" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${data.aws_iam_policy.role.arn}"
}

resource "aws_iam_role" "resource_role" {
  name = "tf-test-datapipeline-resource-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

data "aws_iam_policy" "resource_role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}

resource "aws_iam_role_policy_attachment" "resource_role" {
  role       = "${aws_iam_role.resource_role.name}"
  policy_arn = "${data.aws_iam_policy.resource_role.arn}"
}

resource "aws_iam_instance_profile" "resource_role" {
  name = "tf-test-datapipeline-resource-role-profile-%[1]s"
  role = "${aws_iam_role.resource_role.name}"
}

`, rName)
}

func testAccAWSDataPipelineDefinitionConfig(rName, scheduleType string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`

resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type = %[2]q
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }
}`, rName, scheduleType)

}

func testAccAWSDataPipelineDefinitionConfigDefaultUpdate(rName, scheduleType, bucketName, failureAndRerunMode string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`
resource "aws_s3_bucket" "log_bucket" {
  bucket = %[3]q
  acl    = "log-delivery-write"
		
  force_destroy = true
}


resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type          = %[2]q
	role                   = "${aws_iam_role.role.arn}"
	resource_role          = "${aws_iam_instance_profile.resource_role.arn}"
	pipeline_log_uri       = "s3://${aws_s3_bucket.log_bucket.id}/log_prefix"
	failure_and_rerun_mode = %[4]q
  }
}`, rName, scheduleType, bucketName, failureAndRerunMode)

}

func testAccAWSDataPipelineDefinitionConfigWithSchedule(rName, scheduleID, period, startDateTime, endDateTime string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`
resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type          = "cron"
	role                   = "${aws_iam_role.role.arn}"
	resource_role          = "${aws_iam_instance_profile.resource_role.arn}"

	schedule			   = %[1]q
  }

  schedule {
	  id              = %[1]q
	  name            = %[1]q
	  period          = %[2]q
	  start_date_time = %[3]q
	  end_date_time   = %[4]q
  }
}`, scheduleID, period, startDateTime, endDateTime)

}

func testAccAWSDataPipelineDefinitionConfigWithScheduleUpdate(rName, scheduleID, period, startAt string, occurrences int) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`
resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type          = "cron"
	role                   = "${aws_iam_role.role.arn}"
	resource_role          = "${aws_iam_instance_profile.resource_role.arn}"

	schedule			   = %[1]q
  }

  schedule {
	  id           = %[1]q
	  name         = %[1]q
	  period       = %[2]q
	  start_at     = %[3]q
	  occurrences  = %[4]d
  }
}`, scheduleID, period, startAt, occurrences)

}

func testAccAWSDataPipelineDefinitionConfigWithParameter(rName, parameterObjectID, parameterObjectValue string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`

resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type = "cron"
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }

  parameter_object {
	id = %[2]q
  }

  parameter_value {
	id = %[2]q
	string_value = %[3]q
  }
}`, rName, parameterObjectID, parameterObjectValue)

}

func testAccAWSDataPipelineDefinitionConfigWithParameterUpdate(rName, parameterObjectID1, parameterObjectValue1, parameterObjectID2, parameterObjectValue2 string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`

resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type = "cron"
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }

  parameter_object {
	id       = %[2]q
	optional = true
	default  = %[1]q
  }

  parameter_value {
	id = %[2]q
	string_value = %[3]q
  }

  parameter_object {
	id          = %[4]q
	description = "Array value %[4]s"
	is_array    = true
  }

  parameter_value {
	id           = %[4]q
	string_value = "%[5]s-1"
  }

  parameter_value {
	id           = %[4]q
	string_value = "%[5]s-2"
  }

}`, rName, parameterObjectID1, parameterObjectValue1, parameterObjectID2, parameterObjectValue2)

}
