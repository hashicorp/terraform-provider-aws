package appautoscaling_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAppAutoScalingTarget_basic(t *testing.T) {
	var target applicationautoscaling.ScalableTarget

	randClusterName := fmt.Sprintf("cluster-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig(randClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.bar", &target),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "service_namespace", "ecs"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "scalable_dimension", "ecs:service:DesiredCount"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "min_capacity", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "max_capacity", "3"),
				),
			},

			{
				Config: testAccTargetUpdateConfig(randClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.bar", &target),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "min_capacity", "2"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "max_capacity", "8"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_target.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc("aws_appautoscaling_target.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_disappears(t *testing.T) {
	var target applicationautoscaling.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(resourceName, &target),
					acctest.CheckResourceDisappears(acctest.Provider, tfappautoscaling.ResourceTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_spotFleetRequest(t *testing.T) {
	var target applicationautoscaling.ScalableTarget
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetSpotFleetRequestConfig(validUntil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.test", &target),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.test", "service_namespace", "ec2"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.test", "scalable_dimension", "ec2:spot-fleet-request:TargetCapacity"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_target.test",
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc("aws_appautoscaling_target.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_emrCluster(t *testing.T) {
	var target applicationautoscaling.ScalableTarget
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetEMRClusterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.bar", &target),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "service_namespace", "elasticmapreduce"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.bar", "scalable_dimension", "elasticmapreduce:instancegroup:InstanceCount"),
				),
			},
			{
				ResourceName:      "aws_appautoscaling_target.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc("aws_appautoscaling_target.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_multipleTargets(t *testing.T) {
	var writeTarget applicationautoscaling.ScalableTarget
	var readTarget applicationautoscaling.ScalableTarget

	rInt := sdkacctest.RandInt()
	tableName := fmt.Sprintf("tf_acc_test_table_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTarget_multipleTargets(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.write", &writeTarget),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.write", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.write", "resource_id", "table/"+tableName),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.write", "scalable_dimension", "dynamodb:table:WriteCapacityUnits"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.write", "min_capacity", "1"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.write", "max_capacity", "10"),

					testAccCheckTargetExists("aws_appautoscaling_target.read", &readTarget),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.read", "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.read", "resource_id", "table/"+tableName),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.read", "scalable_dimension", "dynamodb:table:ReadCapacityUnits"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.read", "min_capacity", "2"),
					resource.TestCheckResourceAttr("aws_appautoscaling_target.read", "max_capacity", "15"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingTarget_optionalRoleARN(t *testing.T) {
	var readTarget applicationautoscaling.ScalableTarget

	rInt := sdkacctest.RandInt()
	tableName := fmt.Sprintf("tf_acc_test_table_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationautoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTarget_optionalRoleARN(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists("aws_appautoscaling_target.read", &readTarget),
					acctest.CheckResourceAttrGlobalARN("aws_appautoscaling_target.read", "role_arn", "iam",
						"role/aws-service-role/dynamodb.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_DynamoDBTable"),
				),
			},
		},
	})
}

func testAccCheckTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appautoscaling_target" {
			continue
		}

		_, err := tfappautoscaling.FindTargetByThreePartKey(conn, rs.Primary.ID, rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes["scalable_dimension"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Application AutoScaling Target %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckTargetExists(n string, v *applicationautoscaling.ScalableTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Application AutoScaling Target ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingConn

		output, err := tfappautoscaling.FindTargetByThreePartKey(conn, rs.Primary.ID, rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes["scalable_dimension"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTargetConfig(
	randClusterName string) string {
	return fmt.Sprintf(`
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
  name            = "foobar"
  cluster         = aws_ecs_cluster.foo.id
  task_definition = aws_ecs_task_definition.task.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "bar" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}
`, randClusterName)
}

func testAccTargetUpdateConfig(
	randClusterName string) string {
	return fmt.Sprintf(`
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
  name            = "foobar"
  cluster         = aws_ecs_cluster.foo.id
  task_definition = aws_ecs_task_definition.task.arn
  desired_count   = 2

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}

resource "aws_appautoscaling_target" "bar" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.foo.name}/${aws_ecs_service.service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 2
  max_capacity       = 8
}
`, randClusterName)
}

func testAccTargetEMRClusterConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # The requested instance type m3.xlarge is not supported in the requested availability zone.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.main.id
    emr_managed_master_security_group = aws_security_group.allow_all.id
    emr_managed_slave_security_group  = aws_security_group.allow_all.id
    instance_profile                  = aws_iam_instance_profile.emr_profile.arn
  }

  master_instance_group {
    instance_type = "m3.xlarge"
  }

  core_instance_group {
    instance_count = 2
    instance_type  = "m3.xlarge"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [aws_main_route_table_association.a]

  service_role     = aws_iam_role.iam_emr_default_role.arn
  autoscaling_role = aws_iam_role.emr-autoscaling-role.arn
}

resource "aws_emr_instance_group" "task" {
  cluster_id     = aws_emr_cluster.tf-test-cluster.id
  instance_count = 1
  instance_type  = "m3.xlarge"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%d"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = [aws_subnet.main]

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-appautoscaling-target-emr-cluster"
  }
}

resource "aws_subnet" "main" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "168.31.0.0/20"
  vpc_id            = aws_vpc.main.id

  tags = {
    Name = "tf-acc-appautoscaling-target-emr-cluster"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = aws_vpc.main.id
  route_table_id = aws_route_table.r.id
}

resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = aws_iam_role.iam_emr_default_role.id
  policy_arn = aws_iam_policy.iam_emr_default_policy.arn
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%d"
  role = aws_iam_role.iam_emr_profile_role.name
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = aws_iam_role.iam_emr_profile_role.id
  policy_arn = aws_iam_policy.iam_emr_profile_policy.arn
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%d"
  assume_role_policy = data.aws_iam_policy_document.emr-autoscaling-role-policy.json
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = aws_iam_role.emr-autoscaling-role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_appautoscaling_target" "bar" {
  service_namespace  = "elasticmapreduce"
  resource_id        = "instancegroup/${aws_emr_cluster.tf-test-cluster.id}/${aws_emr_instance_group.task.id}"
  scalable_dimension = "elasticmapreduce:instancegroup:InstanceCount"
  role_arn           = aws_iam_role.emr-autoscaling-role.arn
  min_capacity       = 1
  max_capacity       = 8
}
`, rInt, rInt, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccTargetSpotFleetRequestConfig(validUntil string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_partition" "current" {}

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
  role       = aws_iam_role.fleet_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.fleet_role.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "m3.medium"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}
`, validUntil)
}

func testAccTarget_multipleTargets(tableName string) string {
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

resource "aws_appautoscaling_target" "write" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_target" "read" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}
`, tableName)
}

func testAccTarget_optionalRoleARN(tableName string) string {
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

resource "aws_appautoscaling_target" "read" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.dynamodb_table_test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}
`, tableName)
}

func testAccTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := fmt.Sprintf("%s/%s/%s",
			rs.Primary.Attributes["service_namespace"],
			rs.Primary.Attributes["resource_id"],
			rs.Primary.Attributes["scalable_dimension"])
		return id, nil
	}
}
