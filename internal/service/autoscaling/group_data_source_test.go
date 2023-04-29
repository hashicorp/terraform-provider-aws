package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAutoScalingGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_autoscaling_group.test"
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "default_cooldown", resourceName, "default_cooldown"),
					resource.TestCheckResourceAttrPair(datasourceName, "desired_capacity", resourceName, "desired_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "desired_capacity_type", resourceName, "desired_capacity_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "enabled_metrics.#", resourceName, "enabled_metrics.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "health_check_grace_period", resourceName, "health_check_grace_period"),
					resource.TestCheckResourceAttrPair(datasourceName, "health_check_type", resourceName, "health_check_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "launch_configuration", resourceName, "launch_configuration"),
					resource.TestCheckResourceAttrPair(datasourceName, "launch_template.#", resourceName, "launch_template.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "load_balancers.#", resourceName, "load_balancers.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "max_instance_lifetime", resourceName, "max_instance_lifetime"),
					resource.TestCheckResourceAttrPair(datasourceName, "max_size", resourceName, "max_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "min_size", resourceName, "min_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.#", resourceName, "mixed_instances_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttr(datasourceName, "new_instances_protected_from_scale_in", "false"),
					resource.TestCheckResourceAttrPair(datasourceName, "placement_group", resourceName, "placement_group"),
					resource.TestCheckResourceAttrPair(datasourceName, "service_linked_role_arn", resourceName, "service_linked_role_arn"),
					resource.TestCheckResourceAttr(datasourceName, "status", ""), // Only set when the DeleteAutoScalingGroup operation is in progress.
					resource.TestCheckResourceAttrPair(datasourceName, "target_group_arns.#", resourceName, "target_group_arns.#"),
					resource.TestCheckResourceAttr(datasourceName, "termination_policies.#", "1"), // Not set in resource.
					resource.TestCheckResourceAttr(datasourceName, "vpc_zone_identifier", ""),     // Not set in resource.
				),
			},
		},
	})
}

func TestAccAutoScalingGroupDataSource_launchTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_autoscaling_group.test"
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_launchTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "launch_template.#", resourceName, "launch_template.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "launch_template.0.id", resourceName, "launch_template.0.id"),
					resource.TestCheckResourceAttrPair(datasourceName, "launch_template.0.name", resourceName, "launch_template.0.name"),
					resource.TestCheckResourceAttrPair(datasourceName, "launch_template.0.version", resourceName, "launch_template.0.version"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroupDataSource_mixedInstancesPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_autoscaling_group.test"
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_mixedInstancesPolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.#", resourceName, "mixed_instances_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.#", resourceName, "mixed_instances_policy.0.instances_distribution.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy", resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy", resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.#", resourceName, "mixed_instances_policy.0.launch_template.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id", resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name", resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.override.#", resourceName, "mixed_instances_policy.0.launch_template.0.override.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_autoscaling_group" "test" {
  name = aws_autoscaling_group.test.name

  depends_on = [aws_autoscaling_group.no_match]
}

resource "aws_autoscaling_group" "test" {
  name                      = %[1]q
  max_size                  = 0
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  enabled_metrics           = ["GroupDesiredCapacity"]
  force_delete              = true
  launch_configuration      = aws_launch_configuration.test.name
  availability_zones        = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
}

resource "aws_autoscaling_group" "no_match" {
  name                      = "%[1]s-1"
  max_size                  = 0
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  enabled_metrics           = ["GroupDesiredCapacity", "GroupStandbyInstances"]
  force_delete              = true
  launch_configuration      = aws_launch_configuration.test.name
  availability_zones        = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
`, rName))
}

func testAccGroupDataSourceConfig_launchTemplate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_autoscaling_group" "test" {
  name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_group" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}

resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
`, rName))
}

func testAccGroupDataSourceConfig_mixedInstancesPolicy(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_autoscaling_group" "test" {
  name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_group" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  mixed_instances_policy {
    instances_distribution {
      on_demand_allocation_strategy            = "prioritized"
      on_demand_base_capacity                  = 1
      on_demand_percentage_above_base_capacity = 1
      spot_allocation_strategy                 = "lowest-price"
      spot_instance_pools                      = 2
      spot_max_price                           = "0.50"
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}

resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
`, rName))
}
