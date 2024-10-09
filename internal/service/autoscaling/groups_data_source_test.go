// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasource1Name := "data.aws_autoscaling_groups.group_list"
	datasource2Name := "data.aws_autoscaling_groups.group_list_tag_lookup"
	datasource3Name := "data.aws_autoscaling_groups.group_list_by_name"
	datasource4Name := "data.aws_autoscaling_groups.group_list_multiple_values"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasource1Name, "names.#", acctest.Ct3),
					resource.TestCheckResourceAttr(datasource1Name, "arns.#", acctest.Ct3),
					resource.TestCheckResourceAttr(datasource2Name, "names.#", acctest.Ct3),
					resource.TestCheckResourceAttr(datasource2Name, "arns.#", acctest.Ct3),
					resource.TestCheckResourceAttr(datasource3Name, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasource3Name, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasource4Name, "names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(datasource4Name, "arns.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccGroupsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_autoscaling_group" "test1" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  name               = "%[1]s-1"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "MetaGroup"
    value               = %[1]q
    propagate_at_launch = true
  }

  tag {
    key                 = "Name"
    value               = "%[1]s-1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "test2" {
  availability_zones = [data.aws_availability_zones.available.names[1]]
  name               = "%[1]s-2"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "MetaGroup"
    value               = %[1]q
    propagate_at_launch = true
  }

  tag {
    key                 = "Name"
    value               = "%[1]s-2"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "test3" {
  availability_zones = [data.aws_availability_zones.available.names[2]]
  name               = "%[1]s-3"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "MetaGroup"
    value               = %[1]q
    propagate_at_launch = true
  }

  tag {
    key                 = "Name"
    value               = "%[1]s-3"
    propagate_at_launch = true
  }
}

data "aws_autoscaling_groups" "group_list" {
  filter {
    name   = "key"
    values = ["MetaGroup"]
  }

  filter {
    name   = "value"
    values = [%[1]q]
  }

  depends_on = [aws_autoscaling_group.test1, aws_autoscaling_group.test2, aws_autoscaling_group.test3]
}

data "aws_autoscaling_groups" "group_list_tag_lookup" {
  filter {
    name   = "tag:MetaGroup"
    values = [%[1]q]
  }

  depends_on = [aws_autoscaling_group.test1, aws_autoscaling_group.test2, aws_autoscaling_group.test3]
}

data "aws_autoscaling_groups" "group_list_by_name" {
  names = [aws_autoscaling_group.test1.name]

  depends_on = [aws_autoscaling_group.test1, aws_autoscaling_group.test2, aws_autoscaling_group.test3]
}

data "aws_autoscaling_groups" "group_list_multiple_values" {
  filter {
    name   = "tag:Name"
    values = [aws_autoscaling_group.test2.name, aws_autoscaling_group.test3.name]
  }

  depends_on = [aws_autoscaling_group.test1, aws_autoscaling_group.test2, aws_autoscaling_group.test3]
}
`, rName))
}
