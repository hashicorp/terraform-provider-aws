// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.EMRServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Managed scaling is not available",
		"SSO is not enabled",
		"Account is not whitelisted to use this feature",
		"IAM Identity Center is not enabled",
	)
}

func TestAccEMRManagedScalingPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_managed_scaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedScalingPolicyDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccManagedScalingPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedScalingPolicyExists(ctx, resourceName),
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

func TestAccEMRManagedScalingPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_managed_scaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedScalingPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedScalingPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedScalingPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceManagedScalingPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRManagedScalingPolicy_ComputeLimits_maximumCoreCapacityUnits(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_managed_scaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedScalingPolicyDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccManagedScalingPolicyConfig_computeLimitsMaximumCoreCapacityUnits(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedScalingPolicyExists(ctx, resourceName),
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

func TestAccEMRManagedScalingPolicy_ComputeLimits_maximumOnDemandCapacityUnits(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_managed_scaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedScalingPolicyDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccManagedScalingPolicyConfig_computeLimitsMaximumOnDemandCapacityUnits(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedScalingPolicyExists(ctx, resourceName),
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

func TestAccEMRManagedScalingPolicy_ComputeLimits_maximumOnDemandCapacityUnitsSpotOnly(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_managed_scaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedScalingPolicyDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccManagedScalingPolicyConfig_computeLimitsMaximumOnDemandCapacityUnits(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedScalingPolicyExists(ctx, resourceName),
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

func testAccCheckManagedScalingPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Managed Scaling Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)
		resp, err := conn.GetManagedScalingPolicyWithContext(ctx, &emr.GetManagedScalingPolicyInput{
			ClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if resp.ManagedScalingPolicy == nil {
			return fmt.Errorf("EMR Managed Scaling Policy is empty which shouldn't happen")
		}
		return nil
	}
}

func testAccCheckManagedScalingPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_managed_scaling_policy" {
				continue
			}

			resp, err := conn.GetManagedScalingPolicyWithContext(ctx, &emr.GetManagedScalingPolicyInput{
				ClusterId: aws.String(rs.Primary.ID),
			})

			if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
				continue
			}

			if tfawserr.ErrMessageContains(err, "ValidationException", "A job flow that is shutting down, terminated, or finished may not be modified") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading EMR Managed Scaling Policy (%s): %w", rs.Primary.ID, err)
			}

			if resp != nil && resp.ManagedScalingPolicy != nil {
				return fmt.Errorf("EMR Managed Scaling Policy (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccManagedScalingPolicyConfig_base(rName string) string {
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
    Name = %[1]q
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = false
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = %[1]q
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
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = aws_iam_policy.emr_service.arn
}

resource "aws_iam_policy" "emr_service" {
  name = "%[1]s_emr"

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
resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
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
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
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

resource "aws_emr_cluster" "test" {
  applications                      = ["Hadoop", "Hive"]
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3n://terraform/testlog/"
  name                              = %[1]q
  release_label                     = "emr-5.30.1"
  service_role                      = aws_iam_role.emr_service.arn

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

}
`, rName))
}

func testAccManagedScalingPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccManagedScalingPolicyConfig_base(rName), `
resource "aws_emr_managed_scaling_policy" "test" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type              = "Instances"
    minimum_capacity_units = 1
    maximum_capacity_units = 2
  }
}
`)
}

func testAccManagedScalingPolicyConfig_computeLimitsMaximumCoreCapacityUnits(rName string, maximumCoreCapacityUnits int) string {
	return acctest.ConfigCompose(testAccManagedScalingPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_managed_scaling_policy" "test" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type                   = "Instances"
    minimum_capacity_units      = 1
    maximum_capacity_units      = 2
    maximum_core_capacity_units = %[1]d
  }
}
`, maximumCoreCapacityUnits))
}

func testAccManagedScalingPolicyConfig_computeLimitsMaximumOnDemandCapacityUnits(rName string, maximumOnDemandCapacityUnits int) string {
	return acctest.ConfigCompose(testAccManagedScalingPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_managed_scaling_policy" "test" {
  cluster_id = aws_emr_cluster.test.id
  compute_limits {
    unit_type                       = "Instances"
    minimum_capacity_units          = 1
    maximum_capacity_units          = 2
    maximum_ondemand_capacity_units = %[1]d
  }
}
`, maximumOnDemandCapacityUnits))
}
