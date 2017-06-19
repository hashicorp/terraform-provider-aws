package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAppautoScalingPolicy_basic(t *testing.T) {
	var policy applicationautoscaling.ScalingPolicy

	randClusterName := fmt.Sprintf("cluster%s", acctest.RandString(10))
	randPolicyName := fmt.Sprintf("terraform-test-foobar-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAppautoscalingPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAppautoscalingPolicyConfig(randClusterName, randPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppautoscalingPolicyExists("aws_appautoscaling_policy.foobar_simple", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "cooldown", "60"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_adjustment.#", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_adjustment.2087484785.scaling_adjustment", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_adjustment.2087484785.metric_interval_lower_bound", "0"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "name", randPolicyName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "resource_id", fmt.Sprintf("service/%s/foobar", randClusterName)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "service_namespace", "ecs"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "scalable_dimension", "ecs:service:DesiredCount"),
				),
			},
		},
	})
}

func TestAccAWSAppautoScalingPolicy_nestedSchema(t *testing.T) {
	var policy applicationautoscaling.ScalingPolicy

	randClusterName := fmt.Sprintf("cluster%s", acctest.RandString(10))
	randPolicyName := fmt.Sprintf("terraform-test-foobar-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAppautoscalingPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAppautoscalingPolicyNestedSchemaConfig(randClusterName, randPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppautoscalingPolicyExists("aws_appautoscaling_policy.foobar_simple", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.adjustment_type", "PercentChangeInCapacity"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.cooldown", "60"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.step_adjustment.#", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.step_adjustment.2252990027.scaling_adjustment", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.step_adjustment.2252990027.metric_interval_lower_bound", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "step_scaling_policy_configuration.0.step_adjustment.2252990027.metric_interval_upper_bound", "-1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "name", randPolicyName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "resource_id", fmt.Sprintf("service/%s/foobar", randClusterName)),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "service_namespace", "ecs"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.foobar_simple", "scalable_dimension", "ecs:service:DesiredCount"),
				),
			},
		},
	})
}

func TestAccAWSAppautoScalingPolicy_spotFleetRequest(t *testing.T) {
	var policy applicationautoscaling.ScalingPolicy

	randPolicyName := fmt.Sprintf("test-appautoscaling-policy-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAppautoscalingPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAppautoscalingPolicySpotFleetRequestConfig(randPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppautoscalingPolicyExists("aws_appautoscaling_policy.test", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", "name", randPolicyName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", "service_namespace", "ec2"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.test", "scalable_dimension", "ec2:spot-fleet-request:TargetCapacity"),
				),
			},
		},
	})
}

func TestAccAWSAppautoScalingPolicy_dynamoDb(t *testing.T) {
	var policy applicationautoscaling.ScalingPolicy

	randPolicyName := fmt.Sprintf("test-appautoscaling-policy-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAppautoscalingPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAppautoscalingPolicyDynamoDB(randPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppautoscalingPolicyExists("aws_appautoscaling_policy.dynamo_test", &policy),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.dynamo_test", "name", randPolicyName),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.dynamo_test", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_policy.dynamo_test", "scalable_dimension", "dynamodb:table:WriteCapacityUnits"),
				),
			},
		},
	})
}

func testAccCheckAWSAppautoscalingPolicyExists(n string, policy *applicationautoscaling.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).appautoscalingconn
		params := &applicationautoscaling.DescribeScalingPoliciesInput{
			ServiceNamespace: aws.String(rs.Primary.Attributes["service_namespace"]),
			PolicyNames:      []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeScalingPolicies(params)
		if err != nil {
			return err
		}
		if len(resp.ScalingPolicies) == 0 {
			return fmt.Errorf("ScalingPolicy %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSAppautoscalingPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appautoscalingconn

	for _, rs := range s.RootModule().Resources {
		params := applicationautoscaling.DescribeScalingPoliciesInput{
			ServiceNamespace: aws.String(rs.Primary.Attributes["service_namespace"]),
			PolicyNames:      []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeScalingPolicies(&params)

		if err == nil {
			if len(resp.ScalingPolicies) != 0 &&
				*resp.ScalingPolicies[0].PolicyName == rs.Primary.ID {
				return fmt.Errorf("Application autoscaling policy still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSAppautoscalingPolicyConfig(
	randClusterName string,
	randPolicyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "autoscale_role" {
	name = "%s"
	path = "/"

	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_role_policy" "autoscale_role_policy" {
	name = "%s"
	role = "${aws_iam_role.autoscale_role.id}"

	policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:DescribeServices",
                "ecs:UpdateService",
				"cloudwatch:DescribeAlarms"
            ],
            "Resource": ["*"]
        }
    ]
}
EOF
}

resource "aws_ecs_cluster" "foo" {
	name = "%s"
}

resource "aws_ecs_task_definition" "task" {
	family = "foobar"
	container_definitions = <<EOF
[
	{
		"name": "busybox",
		"image": "busybox:latest",
		"cpu": 10,
		"memory": 128,
		"essential": true
	}
]
EOF
}

resource "aws_ecs_service" "service" {
	name = "foobar"
	cluster = "${aws_ecs_cluster.foo.id}"
	task_definition = "${aws_ecs_task_definition.task.arn}"
	desired_count = 1
	deployment_maximum_percent = 200
	deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "tgt" {
	service_namespace = "ecs"
	resource_id = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
	scalable_dimension = "ecs:service:DesiredCount"
	role_arn = "${aws_iam_role.autoscale_role.arn}"
	min_capacity = 1
	max_capacity = 4
}

resource "aws_appautoscaling_policy" "foobar_simple" {
	name = "%s"
	service_namespace = "ecs"
	resource_id = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
	scalable_dimension = "ecs:service:DesiredCount"
	adjustment_type = "ChangeInCapacity"
	cooldown = 60
	metric_aggregation_type = "Average"
	step_adjustment {
		metric_interval_lower_bound = 0
		scaling_adjustment = 1
	}
	depends_on = ["aws_appautoscaling_target.tgt"]
}
`, randClusterName, randClusterName, randClusterName, randPolicyName)
}

func testAccAWSAppautoscalingPolicySpotFleetRequestConfig(
	randPolicyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "fleet_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "fleet_role_policy" {
  role = "${aws_iam_role.fleet_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role = "${aws_iam_role.fleet_role.arn}"
  spot_price = "0.005"
  target_capacity = 2
  valid_until = "2019-11-04T20:44:20Z"
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "m3.medium"
    ami = "ami-d06a90b0"
  }
}

resource "aws_iam_role" "autoscale_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "application-autoscaling.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "autoscale_role_policy_a" {
  role = "${aws_iam_role.autoscale_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetRole"
}

resource "aws_iam_role_policy_attachment" "autoscale_role_policy_b" {
  role = "${aws_iam_role.autoscale_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetAutoscaleRole"
}

resource "aws_appautoscaling_target" "test" {
  service_namespace = "ec2"
  resource_id = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  role_arn = "${aws_iam_role.autoscale_role.arn}"
  min_capacity = 1
  max_capacity = 3
}

resource "aws_appautoscaling_policy" "test" {
  name = "%s"
  service_namespace = "ec2"
  resource_id = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  adjustment_type = "ChangeInCapacity"
  cooldown = 60
  metric_aggregation_type = "Average"

  step_adjustment {
    metric_interval_lower_bound = 0
    scaling_adjustment = 1
  }

  depends_on = ["aws_appautoscaling_target.test"]
}
`, randPolicyName)
}

func testAccAWSAppautoscalingPolicyNestedSchemaConfig(
	randClusterName string,
	randPolicyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "autoscale_role" {
	name = "%s"
	path = "/"

	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_role_policy" "autoscale_role_policy" {
	name = "%s"
	role = "${aws_iam_role.autoscale_role.id}"

	policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:DescribeServices",
                "ecs:UpdateService",
				"cloudwatch:DescribeAlarms"
            ],
            "Resource": ["*"]
        }
    ]
}
EOF
}

resource "aws_ecs_cluster" "foo" {
	name = "%s"
}

resource "aws_ecs_task_definition" "task" {
	family = "foobar"
	container_definitions = <<EOF
[
	{
		"name": "busybox",
		"image": "busybox:latest",
		"cpu": 10,
		"memory": 128,
		"essential": true
	}
]
EOF
}

resource "aws_ecs_service" "service" {
	name = "foobar"
	cluster = "${aws_ecs_cluster.foo.id}"
	task_definition = "${aws_ecs_task_definition.task.arn}"
	desired_count = 1
	deployment_maximum_percent = 200
	deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "tgt" {
	service_namespace = "ecs"
	resource_id = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
	scalable_dimension = "ecs:service:DesiredCount"
	role_arn = "${aws_iam_role.autoscale_role.arn}"
	min_capacity = 1
	max_capacity = 4
}

resource "aws_appautoscaling_policy" "foobar_simple" {
	name = "%s"
	service_namespace = "ecs"
	resource_id = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
	scalable_dimension = "ecs:service:DesiredCount"

	step_scaling_policy_configuration {
		adjustment_type = "PercentChangeInCapacity"
		cooldown = 60
		metric_aggregation_type = "Average"
		step_adjustment {
			metric_interval_lower_bound = 1
			scaling_adjustment = 1
		}
	}
	depends_on = ["aws_appautoscaling_target.tgt"]
}
`, randClusterName, randClusterName, randClusterName, randPolicyName)
}

func testAccAWSAppautoscalingPolicyDynamoDB(
	randPolicyName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "dynamodb_table_test" {
  name           = "%s"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "FooKey"
  attribute {
    name = "FooKey"
    type = "S"
  }
}
resource "aws_iam_role" "dynamodb_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "dynamodb.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
resource "aws_iam_role_policy_attachment" "dynamodb_role_policy" {
  role = "${aws_iam_role.dynamodb_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDynamoDBFullAccess"
}
resource "aws_iam_role" "autoscale_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "application-autoscaling.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
resource "aws_iam_role_policy_attachment" "autoscale_role_policy" {
  role = "${aws_iam_role.autoscale_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AutoScalingFullAccess"
}
resource "aws_appautoscaling_target" "dynamo_test" {
  service_namespace = "dynamodb"
  resource_id = "${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  role_arn = "${aws_iam_role.autoscale_role.arn}"
  min_capacity = 1
  max_capacity = 10
}
resource "aws_appautoscaling_policy" "dynamo_test" {
  name = "%s"
  service_namespace = "dynamodb"
  resource_id = "${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  adjustment_type = "ChangeInCapacity"
  cooldown = 60
  metric_aggregation_type = "Maximum"
  step_adjustment {
    metric_interval_lower_bound = 0
    scaling_adjustment = 1
  }
  
  customized_metric_specification = {
    metric_name = "foo"
    namespace = "dynamodb"
    statistic = "Sum"
  }
  predefined_metric_specification = {
    predefined_metric_type = "DynamoDBWriteCapacityUtilization"
  }
  
  scale_in_cooldown = 10
  scale_out_cooldown = 10
  target_value = 70
  depends_on = ["aws_appautoscaling_target.dynamo_test"]
}
`, randPolicyName, randPolicyName)
}
