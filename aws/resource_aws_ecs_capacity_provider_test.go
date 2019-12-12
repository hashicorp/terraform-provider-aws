package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TODO sweepers once Delete is implemented

func TestAccAWSEcsCapacityProvider_basic(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "id", "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.bar", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScaling(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigManagedScaling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.bar", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScalingPartial(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigManagedScalingPartial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.bar", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_Tags(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsCapacityProviderConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsCapacityProviderConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// TODO add an update test config - Reference: https://github.com/aws/containers-roadmap/issues/633

func testAccCheckAWSEcsCapacityProviderDestroy(s *terraform.State) error {
	// Reference: https://github.com/aws/containers-roadmap/issues/632
	return nil
}

func testAccCheckAWSEcsCapacityProviderExists(resourceName string, provider *ecs.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DescribeCapacityProvidersInput{
			CapacityProviders: []*string{aws.String(rs.Primary.ID)},
			Include:           []*string{aws.String(ecs.CapacityProviderFieldTags)},
		}

		output, err := conn.DescribeCapacityProviders(input)

		if err != nil {
			return fmt.Errorf("error reading ECS Capacity Provider (%s): %s", rs.Primary.ID, err)
		}

		for _, cp := range output.CapacityProviders {
			if aws.StringValue(cp.CapacityProviderArn) == rs.Primary.ID {
				*provider = *cp
				return nil
			}
		}

		return fmt.Errorf("ECS Capacity Provider (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
	most_recent = true
	owners      = ["amazon"]

	filter {
		name   = "name"
		values = ["amzn-ami-hvm-*-x86_64-gp2"]
	}
}

resource "aws_launch_configuration" "foobar" {
	image_id      = "${data.aws_ami.test_ami.id}"
	instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
	availability_zones   = ["us-west-2a"]
	name                 = "%[1]s"
	max_size             = 5
	min_size             = 2
	health_check_type    = "ELB"
	desired_capacity     = 4
	force_delete         = true
	termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

	launch_configuration = "${aws_launch_configuration.foobar.name}"

	tags = [
		{
			key                 = "FromTags1"
			value               = "value1"
			propagate_at_launch = true
		},
	]
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfig(rName string) string {
	return testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn
	}
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigManagedScaling(rName string) string {
	return testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn

		managed_scaling {
			maximum_scaling_step_size = 10
			minimum_scaling_step_size = 2
			status = "ENABLED"
			target_capacity = 50
		}
	}
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigManagedScalingPartial(rName string) string {
	return testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn

		managed_scaling {
			minimum_scaling_step_size = 2
			status = "ENABLED"
		}
	}
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	tags = {
		%q = %q,
	}

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn
	}
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSEcsCapacityProviderConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAWSECSCapacityProviderAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	tags = {
		%q = %q,
		%q = %q,
	}

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn
	}
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
