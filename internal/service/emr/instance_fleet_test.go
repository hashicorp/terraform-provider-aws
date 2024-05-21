// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRInstanceFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var fleet emr.InstanceFleet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_fleet.task"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckInstanceFleetExists(ctx, resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEMRInstanceFleet_Zero_count(t *testing.T) {
	ctx := acctest.Context(t)
	var fleet emr.InstanceFleet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_fleet.task"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckInstanceFleetExists(ctx, resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceFleetConfig_zeroCount(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckInstanceFleetExists(ctx, resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEMRInstanceFleet_ebsBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var fleet emr.InstanceFleet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_fleet.task"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceFleetConfig_ebsBasic(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckInstanceFleetExists(ctx, resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type_configs"},
			},
		},
	})
}

func TestAccEMRInstanceFleet_full(t *testing.T) {
	ctx := acctest.Context(t)
	var fleet emr.InstanceFleet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_fleet.task"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceFleetConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(testAccCheckInstanceFleetExists(ctx, resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "instance_type_configs.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "target_on_demand_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "target_spot_capacity", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceFleetResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type_configs"},
			},
		},
	})
}

func testAccCheckInstanceFleetExists(ctx context.Context, n string, v *emr.InstanceFleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Instance Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		output, err := tfemr.FindInstanceFleetByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceFleetResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["cluster_id"], rs.Primary.ID), nil
	}
}

func testAccInstanceFleetConfig_base(rName string) string {
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
  vpc_id = aws_vpc.test.id
  name   = %[1]q

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

data "aws_partition" "current" {}

resource "aws_iam_role" "emr_service" {
  name               = "%[1]s_default_role"
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

resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role" "emr_instance_profile" {
  name               = "%[1]s_profile_role"
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

resource "aws_iam_role_policy_attachment" "emr_instance_profile" {
  role       = aws_iam_role.emr_instance_profile.id
  policy_arn = aws_iam_policy.emr_instance_profile.arn
}

resource "aws_iam_policy" "emr_instance_profile" {
  name   = "%[1]s_profile"
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
  name          = "%[1]s"
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }

  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100

      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }

      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    name                      = "core fleet"
    target_on_demand_capacity = 1
    target_spot_capacity      = 0
  }

  service_role = aws_iam_role.emr_service.arn
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

func testAccInstanceFleetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceFleetConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    instance_type     = "m3.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    on_demand_specification {
      allocation_strategy = "lowest-price"
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 1
  target_spot_capacity      = 0
}
`, rName))
}

func testAccInstanceFleetConfig_zeroCount(rName string) string {
	return acctest.ConfigCompose(testAccInstanceFleetConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    instance_type     = "m3.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    on_demand_specification {
      allocation_strategy = "lowest-price"
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 0
  target_spot_capacity      = 0
}
`, rName))
}

func testAccInstanceFleetConfig_ebsBasic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceFleetConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }

  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "SWITCH_TO_ON_DEMAND"
      timeout_duration_minutes = 10
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 0
  target_spot_capacity      = 1
}
`, rName))
}

func testAccInstanceFleetConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccInstanceFleetConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.test.id

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }

    ebs_config {
      size                 = 20
      type                 = "gp2"
      volumes_per_instance = 2
    }

    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }

  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 80

    ebs_config {
      size                 = 10
      type                 = "gp2"
      volumes_per_instance = 1
    }

    instance_type     = "m4.2xlarge"
    weighted_capacity = 2
  }

  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "SWITCH_TO_ON_DEMAND"
      timeout_duration_minutes = 10
    }
  }

  name                      = "emr_instance_fleet_%[1]s"
  target_on_demand_capacity = 2
  target_spot_capacity      = 2
}
`, rName))
}
