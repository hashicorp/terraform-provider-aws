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
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TODO update test once that is implemented

func testAccCheckAWSEcsCapacityProviderDestroy(s *terraform.State) error {
	// TODO implement this once delete is implemented
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
			Include:           []*string{aws.String(ecs.ClusterFieldTags)},
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

// TODO pull managed scaling out of _basic and make it a separate test

func testAccAWSEcsCapacityProviderConfig(rName string) string {
	return testAccAWSAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn
		managed_termination_protection = "DISABLED"

		managed_scaling {
			maximum_scaling_step_size = 10
			minimum_scaling_step_size = 1
			status = "DISABLED"
			target_capacity = 1
		}
	}
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSAutoScalingGroupConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
	name = %q

	tags = {
		%q = %q,
	}

	auto_scaling_group_provider {
		auto_scaling_group_arn = aws_autoscaling_group.bar.arn
		managed_termination_protection = "DISABLED"

		managed_scaling {
			maximum_scaling_step_size = 10
			minimum_scaling_step_size = 1
			status = "DISABLED"
			target_capacity = 1
		}
	}
}
`, rName, tag1Key, tag1Value)
}
