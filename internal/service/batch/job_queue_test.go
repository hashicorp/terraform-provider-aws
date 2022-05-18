package batch_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func TestAccBatchJobQueue_basic(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("job-queue/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccBatchJobQueue_disappears(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckBatchJobQueueDisappears(&jobQueue1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8083
func TestAccBatchJobQueue_ComputeEnvironments_externalOrderUpdate(t *testing.T) {
	var jobQueue1 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate(&jobQueue1),
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

func TestAccBatchJobQueue_priority(t *testing.T) {
	var jobQueue1, jobQueue2 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigPriority(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "priority", "1"),
				),
			},
			{
				Config: testAccBatchJobQueueConfigPriority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, "priority", "2"),
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

func TestAccBatchJobQueue_schedulingPolicy(t *testing.T) {
	var jobQueue1, jobQueue2 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedulingPolicyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedulingPolicyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				// last variable selects the scheduling policy's arn. In this case, the first scheduling policy's arn.
				Config: testAccBatchJobQueueConfigSchedulingPolicy(rName, schedulingPolicyName1, schedulingPolicyName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					resource.TestCheckResourceAttrSet(resourceName, "scheduling_policy_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// test switching the scheduling_policy_arn by changing the last variable to select the second scheduling policy's arn.
				Config: testAccBatchJobQueueConfigSchedulingPolicy(rName, schedulingPolicyName1, schedulingPolicyName2, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue2),
					resource.TestCheckResourceAttrSet(resourceName, "scheduling_policy_arn"),
				),
			},
		},
	})
}

func TestAccBatchJobQueue_state(t *testing.T) {
	var jobQueue1, jobQueue2 batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateDisabled),
				),
			},
			{
				Config: testAccBatchJobQueueConfigState(rName, batch.JQStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, "state", batch.JQStateEnabled),
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

func TestAccBatchJobQueue_tags(t *testing.T) {
	var jobQueue batch.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchJobQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchJobQueueConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBatchJobQueueConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBatchJobQueueConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBatchJobQueueExists(resourceName, &jobQueue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckBatchJobQueueExists(n string, jq *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		log.Printf("State: %#v", s.RootModule().Resources)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Job Queue ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn
		name := rs.Primary.Attributes["name"]
		queue, err := tfbatch.GetJobQueue(conn, name)
		if err != nil {
			return err
		}
		if queue == nil {
			return fmt.Errorf("Not found: %s", n)
		}
		*jq = *queue

		return nil
	}
}

func testAccCheckBatchJobQueueDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_job_queue" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn
		jq, err := tfbatch.GetJobQueue(conn, rs.Primary.Attributes["name"])
		if err == nil {
			if jq != nil {
				return fmt.Errorf("Error: Job Queue still exists")
			}
		}
		return nil
	}
	return nil
}

func testAccCheckBatchJobQueueDisappears(jobQueue *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn
		name := aws.StringValue(jobQueue.JobQueueName)

		err := tfbatch.DisableJobQueue(name, conn)
		if err != nil {
			return fmt.Errorf("error disabling Batch Job Queue (%s): %s", name, err)
		}

		return tfbatch.DeleteJobQueue(name, conn)
	}
}

// testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate simulates the change of a Compute Environment Order
// An external update to the Batch Job Queue (e.g. console) may trigger changes to the value of the Order
// parameter that do not affect the operation of the queue itself, but the resource logic needs to handle.
// For example, Terraform may set a single Compute Environment with Order 0, but the console updates it to 1.
func testAccCheckBatchJobQueueComputeEnvironmentOrderUpdate(jobQueue *batch.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn

		input := &batch.UpdateJobQueueInput{
			ComputeEnvironmentOrder: jobQueue.ComputeEnvironmentOrder,
			JobQueue:                jobQueue.JobQueueName,
		}
		name := aws.StringValue(jobQueue.JobQueueName)

		if len(input.ComputeEnvironmentOrder) != 1 {
			return fmt.Errorf("expected one ComputeEnvironmentOrder in Batch Job Queue (%s)", name)
		}

		if aws.Int64Value(input.ComputeEnvironmentOrder[0].Order) != 0 {
			return fmt.Errorf("expected first ComputeEnvironmentOrder in Batch Job Queue (%s) to have existing Order value of 0", name)
		}

		input.ComputeEnvironmentOrder[0].Order = aws.Int64(1)

		_, err := conn.UpdateJobQueue(input)

		if err != nil {
			return fmt.Errorf("error updating Batch Job Queue (%s): %s", name, err)
		}

		return nil
	}
}

func testAccBatchJobQueueConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "batch.${data.aws_partition.current.dns_suffix}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "%[1]s-ecs"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = aws_iam_role.ecs_instance_role.name
  role = aws_iam_role_policy_attachment.ecs_instance_role.role
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-batch-job-queue"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-batch-job-queue"
  }
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q
  service_role             = aws_iam_role.test.arn
  type                     = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccBatchJobQueueConfigPriority(rName string, priority int) string {
	return acctest.ConfigCompose(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = %[2]d
  state                = "ENABLED"
}
`, rName, priority))
}

func testAccBatchJobQueueSchedulingPolicy(rName string, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test1" {
  name = %[1]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }
  }
}

resource "aws_batch_scheduling_policy" "test2" {
  name = %[2]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }
}
`, rName, rName2)
}

func testAccBatchJobQueueConfigSchedulingPolicy(rName string, schedulingPolicyName1 string, schedulingPolicyName2 string, selectSchedulingPolicy string) string {
	return acctest.ConfigCompose(
		testAccBatchJobQueueConfigBase(rName),
		testAccBatchJobQueueSchedulingPolicy(schedulingPolicyName1, schedulingPolicyName2),
		fmt.Sprintf(`
locals {
  select_scheduling_policy = %[2]q
}

resource "aws_batch_job_queue" "test" {
  compute_environments  = [aws_batch_compute_environment.test.arn]
  name                  = %[1]q
  priority              = 1
  scheduling_policy_arn = local.select_scheduling_policy == "first" ? aws_batch_scheduling_policy.test1.arn : aws_batch_scheduling_policy.test2.arn
  state                 = "ENABLED"
}
`, rName, selectSchedulingPolicy))
}

func testAccBatchJobQueueConfigState(rName string, state string) string {
	return acctest.ConfigCompose(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = %[2]q
}
`, rName, state))
}

func testAccBatchJobQueueConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "DISABLED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccBatchJobQueueConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccBatchJobQueueConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "DISABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCheckLaunchTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(resp.LaunchTemplates) != 0 && *resp.LaunchTemplates[0].LaunchTemplateId == rs.Primary.ID {
				return fmt.Errorf("Launch Template still exists")
			}
		}

		if tfawserr.ErrCodeEquals(err, "InvalidLaunchTemplateId.NotFound") {
			log.Printf("[WARN] launch template (%s) not found.", rs.Primary.ID)
			continue
		}
		return err
	}

	return nil
}
