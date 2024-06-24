// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppAutoScalingTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var target awstypes.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &target),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_dimension", "ecs:service:DesiredCount"),
					resource.TestCheckResourceAttr(resourceName, "service_namespace", "ecs"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},

			{
				Config: testAccTargetConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &target),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "8"),
					resource.TestCheckResourceAttr(resourceName, "min_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var target awstypes.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &target),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappautoscaling.ResourceTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_spotFleetRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var target awstypes.ScalableTarget
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_spotFleetRequest(rName, validUntil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &target),
					resource.TestCheckResourceAttr(resourceName, "service_namespace", "ec2"),
					resource.TestCheckResourceAttr(resourceName, "scalable_dimension", "ec2:spot-fleet-request:TargetCapacity"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_emrCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var target awstypes.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_emrCluster(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &target),
					resource.TestCheckResourceAttr(resourceName, "service_namespace", "elasticmapreduce"),
					resource.TestCheckResourceAttr(resourceName, "scalable_dimension", "elasticmapreduce:instancegroup:InstanceCount"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTargetImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppAutoScalingTarget_multipleTargets(t *testing.T) {
	ctx := acctest.Context(t)
	var writeTarget awstypes.ScalableTarget
	var readTarget awstypes.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	readResourceName := "aws_appautoscaling_target.read"
	writeResourceName := "aws_appautoscaling_target.write"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, writeResourceName, &writeTarget),
					resource.TestCheckResourceAttr(writeResourceName, "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr(writeResourceName, names.AttrResourceID, "table/"+rName),
					resource.TestCheckResourceAttr(writeResourceName, "scalable_dimension", "dynamodb:table:WriteCapacityUnits"),
					resource.TestCheckResourceAttr(writeResourceName, "min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(writeResourceName, names.AttrMaxCapacity, acctest.Ct10),

					testAccCheckTargetExists(ctx, readResourceName, &readTarget),
					resource.TestCheckResourceAttr(readResourceName, "service_namespace", "dynamodb"),
					resource.TestCheckResourceAttr(readResourceName, names.AttrResourceID, "table/"+rName),
					resource.TestCheckResourceAttr(readResourceName, "scalable_dimension", "dynamodb:table:ReadCapacityUnits"),
					resource.TestCheckResourceAttr(readResourceName, "min_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(readResourceName, names.AttrMaxCapacity, "15"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingTarget_optionalRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var readTarget awstypes.ScalableTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetConfig_optionalRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &readTarget),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", "role/aws-service-role/dynamodb.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_DynamoDBTable"),
				),
			},
		},
	})
}

func testAccCheckTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appautoscaling_target" {
				continue
			}

			_, err := tfappautoscaling.FindTargetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes["scalable_dimension"])

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
}

func testAccCheckTargetExists(ctx context.Context, n string, v *awstypes.ScalableTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Application AutoScaling Target ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		output, err := tfappautoscaling.FindTargetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes["scalable_dimension"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTargetConfig_baseECS(rName string, serviceDesiredCount int) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = "testing"

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

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = %[2]d

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}
`, rName, serviceDesiredCount)
}

func testAccTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTargetConfig_baseECS(rName, 1), `
resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}
`)
}

func testAccTargetConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccTargetConfig_baseECS(rName, 2), `
resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 2
  max_capacity       = 8
}
`)
}

func testAccTargetConfig_spotFleetRequest(rName, validUntil string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}
`, rName, validUntil))
}

func testAccTargetConfig_emrCluster(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}

data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 2
    instance_type  = "c4.large"
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

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling.arn
}

resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 1
  instance_type  = "c4.large"
}

resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

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

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceRole"
}

resource "aws_iam_role" "emr_instance_profile" {
  name = "%[1]s_profile_role"

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

resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role_policy_attachment" "emr_instance_profile" {
  role       = aws_iam_role.emr_instance_profile.id
  policy_arn = aws_iam_policy.emr_instance_profile.arn
}

resource "aws_iam_policy" "emr_instance_profile" {
  name = "%[1]s_profile"

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

resource "aws_iam_role" "emr_autoscaling" {
  name               = "%[1]s_autoscaling_role"
  assume_role_policy = data.aws_iam_policy_document.emr_autoscaling_role_policy.json
}

data "aws_iam_policy_document" "emr_autoscaling_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr_autoscaling" {
  role       = aws_iam_role.emr_autoscaling.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "elasticmapreduce"
  resource_id        = "instancegroup/${aws_emr_cluster.test.id}/${aws_emr_instance_group.test.id}"
  scalable_dimension = "elasticmapreduce:instancegroup:InstanceCount"
  role_arn           = aws_iam_role.emr_autoscaling.arn
  min_capacity       = 1
  max_capacity       = 8

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTargetConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TestKey"

  attribute {
    name = "TestKey"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "write" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_target" "read" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}
`, rName)
}

func testAccTargetConfig_optionalRoleARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TestKey"

  attribute {
    name = "TestKey"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}
`, rName)
}

func testAccTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := fmt.Sprintf("%s/%s/%s",
			rs.Primary.Attributes["service_namespace"],
			rs.Primary.Attributes[names.AttrResourceID],
			rs.Primary.Attributes["scalable_dimension"])
		return id, nil
	}
}
