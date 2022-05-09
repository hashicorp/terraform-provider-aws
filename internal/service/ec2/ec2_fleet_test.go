package ec2_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2Fleet_basic(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "context", ""),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "termination"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_id"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_template_config.0.launch_template_specification.0.version"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "false"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances", "false"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "maintain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccEC2Fleet_disappears(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					testAccCheckFleetDisappears(&fleet1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Fleet_excessCapacityTerminationPolicy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_ExcessCapacityTerminationPolicy(rName, "no-termination"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "no-termination"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_ExcessCapacityTerminationPolicy(rName, "termination"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "termination"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateLaunchTemplateSpecification_launchTemplateID(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	launchTemplateResourceName1 := "aws_launch_template.test1"
	launchTemplateResourceName2 := "aws_launch_template.test2"
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateID(rName, launchTemplateResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_id", launchTemplateResourceName1, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName1, "latest_version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateID(rName, launchTemplateResourceName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_id", launchTemplateResourceName2, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName2, "latest_version"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateLaunchTemplateSpecification_launchTemplateName(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	launchTemplateResourceName1 := "aws_launch_template.test1"
	launchTemplateResourceName2 := "aws_launch_template.test2"
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateName(rName, launchTemplateResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_name", launchTemplateResourceName1, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName1, "latest_version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateName(rName, launchTemplateResourceName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_name", launchTemplateResourceName2, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName2, "latest_version"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateLaunchTemplateSpecification_version(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	var launchTemplate ec2.LaunchTemplate
	launchTemplateResourceName := "aws_launch_template.test"
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_Version(rName, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(launchTemplateResourceName, &launchTemplate),
					resource.TestCheckResourceAttr(launchTemplateResourceName, "instance_type", "t3.micro"),
					resource.TestCheckResourceAttr(launchTemplateResourceName, "latest_version", "1"),
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName, "latest_version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_Version(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(launchTemplateResourceName, &launchTemplate),
					resource.TestCheckResourceAttr(launchTemplateResourceName, "instance_type", "t3.small"),
					resource.TestCheckResourceAttr(launchTemplateResourceName, "latest_version", "2"),
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.launch_template_specification.0.version", launchTemplateResourceName, "latest_version"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_availabilityZone(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_AvailabilityZone(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.override.0.availability_zone", availabilityZonesDataSourceName, "names.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_AvailabilityZone(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.override.0.availability_zone", availabilityZonesDataSourceName, "names.1"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_instanceType(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceType(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_type", "t3.small"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceType(rName, "t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_type", "t3.medium"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_maxPrice(t *testing.T) {
	acctest.Skip(t, "EC2 API is not correctly returning MaxPrice override")

	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_MaxPrice(rName, "1.01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.max_price", "1.01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_MaxPrice(rName, "1.02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.max_price", "1.02"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_priority(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_Priority(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.priority", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_Priority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.priority", "2"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverridePriority_multiple(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_Priority_Multiple(rName, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.1.priority", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_Priority_Multiple(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.1.priority", "1"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_subnetID(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	subnetResourceName1 := "aws_subnet.test.0"
	subnetResourceName2 := "aws_subnet.test.1"
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_SubnetID(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.override.0.subnet_id", subnetResourceName1, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_SubnetID(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template_config.0.override.0.subnet_id", subnetResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverride_weightedCapacity(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.weighted_capacity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.weighted_capacity", "2"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_LaunchTemplateOverrideWeightedCapacity_multiple(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity_Multiple(rName, 1, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.1.weighted_capacity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity_Multiple(rName, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.1.weighted_capacity", "2"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_OnDemandOptions_allocationStrategy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_OnDemandOptions_AllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "prioritized"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_OnDemandOptions_AllocationStrategy(rName, "lowestPrice"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "lowestPrice"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_replaceUnhealthyInstances(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_ReplaceUnhealthyInstances(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_ReplaceUnhealthyInstances(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "false"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_SpotOptions_allocationStrategy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_SpotOptions_AllocationStrategy(rName, "diversified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "diversified"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_SpotOptions_AllocationStrategy(rName, "lowestPrice"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "lowestPrice"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_SpotOptions_capacityRebalance(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_SpotOptions_CapacityRebalance(rName, "diversified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "diversified"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.maintenance_strategies.0.capacity_rebalance.0.replacement_strategy", "launch"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccEC2Fleet_SpotOptions_instanceInterruptionBehavior(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, "stop"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, "terminate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "terminate"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_SpotOptions_instancePoolsToUseCount(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_SpotOptions_InstancePoolsToUseCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_SpotOptions_InstancePoolsToUseCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "3"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_tags(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_Tags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_Tags(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_TargetCapacitySpecification_defaultTargetCapacityType(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "on-demand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "on-demand"),
				),
			},
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_TargetCapacitySpecificationDefaultTargetCapacityType_onDemand(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "on-demand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "on-demand"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccEC2Fleet_TargetCapacitySpecificationDefaultTargetCapacityType_spot(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccEC2Fleet_TargetCapacitySpecification_totalTargetCapacity(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "2"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_terminateInstancesWithExpiration(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_TerminateInstancesWithExpiration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_TerminateInstancesWithExpiration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					testAccCheckFleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "false"),
				),
			},
		},
	})
}

func TestAccEC2Fleet_type(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_Type(rName, "maintain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "type", "maintain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			// This configuration will fulfill immediately, skip until ValidFrom is implemented
			// {
			// 	Config: testAccFleetConfig_Type(rName, "request"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckFleetExists(resourceName, &fleet2),
			// 		testAccCheckFleetRecreated(&fleet1, &fleet2),
			// 		resource.TestCheckResourceAttr(resourceName, "type", "request"),
			// 	),
			// },
		},
	})
}

// Test for the bug described in https://github.com/hashicorp/terraform-provider-aws/issues/6777
func TestAccEC2Fleet_templateMultipleNetworkInterfaces(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_multipleNetworkInterfaces(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "type", "maintain"),
					testAccCheckFleetHistory(resourceName, "The associatePublicIPAddress parameter cannot be specified when launching with multiple network interfaces"),
				),
			},
		},
	})
}

func testAccCheckFleetHistory(resourceName string, errorMsg string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(time.Minute * 2) // We have to wait a bit for the history to get populated.

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DescribeFleetHistoryInput{
			FleetId:   aws.String(rs.Primary.ID),
			StartTime: aws.Time(time.Now().Add(time.Hour * -2)),
		}

		output, err := conn.DescribeFleetHistory(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EC2 Fleet history not found")
		}

		if output.HistoryRecords == nil {
			return fmt.Errorf("No fleet history records found for fleet %s", rs.Primary.ID)
		}

		for _, record := range output.HistoryRecords {
			if record == nil {
				continue
			}
			if strings.Contains(aws.StringValue(record.EventInformation.EventDescription), errorMsg) {
				return fmt.Errorf("Error %s found in fleet history event", errorMsg)
			}
		}

		return nil
	}
}

func testAccCheckFleetExists(resourceName string, fleet *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DescribeFleetsInput{
			FleetIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFleets(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EC2 Fleet not found")
		}

		for _, fleetData := range output.Fleets {
			if fleetData == nil {
				continue
			}
			if aws.StringValue(fleetData.FleetId) != rs.Primary.ID {
				continue
			}
			*fleet = *fleetData
			break
		}

		if fleet == nil {
			return fmt.Errorf("EC2 Fleet not found")
		}

		return nil
	}
}

func testAccCheckFleetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_fleet" {
			continue
		}

		input := &ec2.DescribeFleetsInput{
			FleetIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFleets(input)

		if tfawserr.ErrCodeEquals(err, "InvalidFleetId.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		for _, fleetData := range output.Fleets {
			if fleetData == nil {
				continue
			}
			if aws.StringValue(fleetData.FleetId) != rs.Primary.ID {
				continue
			}
			if aws.StringValue(fleetData.FleetState) == ec2.FleetStateCodeDeleted {
				break
			}
			terminateInstances, err := strconv.ParseBool(rs.Primary.Attributes["terminate_instances"])
			if err != nil {
				return fmt.Errorf("error converting terminate_instances (%s) to bool: %s", rs.Primary.Attributes["terminate_instances"], err)
			}
			if !terminateInstances && aws.StringValue(fleetData.FleetState) == ec2.FleetStateCodeDeletedRunning {
				break
			}
			// AWS SDK constant is incorrect
			if !terminateInstances && aws.StringValue(fleetData.FleetState) == "deleted_running" {
				break
			}
			return fmt.Errorf("EC2 Fleet (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(fleetData.FleetState))
		}
	}

	return nil
}

func testAccCheckFleetDisappears(fleet *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DeleteFleetsInput{
			FleetIds:           []*string{fleet.FleetId},
			TerminateInstances: aws.Bool(false),
		}

		_, err := conn.DeleteFleets(input)

		return err
	}
}

func testAccCheckFleetNotRecreated(i, j *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreateTime).Equal(aws.TimeValue(j.CreateTime)) {
			return errors.New("EC2 Fleet was recreated")
		}

		return nil
	}
}

func testAccCheckFleetRecreated(i, j *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime).Equal(aws.TimeValue(j.CreateTime)) {
			return errors.New("EC2 Fleet was not recreated")
		}

		return nil
	}
}

func testAccPreCheckFleet(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeFleetsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.DescribeFleets(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFleetConfig_multipleNetworkInterfaces(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.0.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Testacc SSH security group"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  security_groups = [aws_security_group.test.id]
}

resource "aws_launch_template" "test" {
  name     = %[1]q
  image_id = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

  instance_market_options {
    spot_options {
      spot_instance_type = "persistent"
    }
    market_type = "spot"
  }

  network_interfaces {
    device_index          = 0
    delete_on_termination = true
    network_interface_id  = aws_network_interface.test.id
  }
  network_interfaces {
    device_index          = 1
    delete_on_termination = true
    subnet_id             = aws_subnet.test.id
  }
}

resource "aws_ec2_fleet" "test" {
  terminate_instances = true

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    # allow to choose from several instance types if there is no spot capacity for some type
    override {
      instance_type = "t2.micro"
    }
    override {
      instance_type = "t3.micro"
    }
    override {
      instance_type = "t3.small"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 1
  }
}
`, rName))
}

func testAccFleetConfig_BaseLaunchTemplate(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %q
}
`, rName)
}

func testAccFleetConfig_ExcessCapacityTerminationPolicy(rName, excessCapacityTerminationPolicy string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  excess_capacity_termination_policy = %q

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, excessCapacityTerminationPolicy)
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateID(rName, launchTemplateResourceName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test1" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = "%[1]s1"
}

resource "aws_launch_template" "test2" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = "%[1]s2"
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = %[2]s.id
      version            = %[2]s.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, rName, launchTemplateResourceName)
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateName(rName, launchTemplateResourceName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test1" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = "%[1]s1"
}

resource "aws_launch_template" "test2" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = "%[1]s2"
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_name = %[2]s.name
      version              = %[2]s.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, rName, launchTemplateResourceName)
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_Version(rName, instanceType string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = %q
  name          = %q
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instanceType, rName)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_AvailabilityZone(rName string, availabilityZoneIndex int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      availability_zone = data.aws_availability_zones.available.names[%d]
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, availabilityZoneIndex)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_InstanceType(rName, instanceType string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type = %q
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instanceType)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_MaxPrice(rName, maxPrice string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      max_price = %q
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, maxPrice)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_Priority(rName string, priority int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      priority = %d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, priority)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_Priority_Multiple(rName string, priority1, priority2 int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type = aws_launch_template.test.instance_type
      priority      = %d
    }

    override {
      instance_type = "t3.small"
      priority      = %d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, priority1, priority2)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_SubnetID(rName string, subnetIndex int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
variable "TestAccNameTag" {
  default = "tf-acc-test-ec2-fleet-launchtemplateconfig-override-subnetid"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = var.TestAccNameTag
  }
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block = "10.1.${count.index}.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = var.TestAccNameTag
  }
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      subnet_id = aws_subnet.test.*.id[%d]
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, subnetIndex)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity(rName string, weightedCapacity int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      weighted_capacity = %d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, weightedCapacity)
}

func testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity_Multiple(rName string, weightedCapacity1, weightedCapacity2 int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type     = aws_launch_template.test.instance_type
      weighted_capacity = %d
    }

    override {
      instance_type     = "t3.small"
      weighted_capacity = %d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, weightedCapacity1, weightedCapacity2)
}

func testAccFleetConfig_OnDemandOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  on_demand_options {
    allocation_strategy = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "on-demand"
    total_target_capacity        = 0
  }
}
`, allocationStrategy)
}

func testAccFleetConfig_ReplaceUnhealthyInstances(rName string, replaceUnhealthyInstances bool) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  replace_unhealthy_instances = %t

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, replaceUnhealthyInstances)
}

func testAccFleetConfig_SpotOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    allocation_strategy = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, allocationStrategy)
}

func testAccFleetConfig_SpotOptions_CapacityRebalance(rName, allocationStrategy string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    allocation_strategy = %[1]q
    maintenance_strategies {
      capacity_rebalance {
        replacement_strategy = "launch"
      }
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, allocationStrategy)
}

func testAccFleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, instanceInterruptionBehavior string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    instance_interruption_behavior = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instanceInterruptionBehavior)
}

func testAccFleetConfig_SpotOptions_InstancePoolsToUseCount(rName string, instancePoolsToUseCount int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    instance_pools_to_use_count = %d
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instancePoolsToUseCount)
}

func testAccFleetConfig_Tags(rName, key1, value1 string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  tags = {
    %q = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, key1, value1)
}

func testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, defaultTargetCapacityType string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = %q
    total_target_capacity        = 0
  }
}
`, defaultTargetCapacityType)
}

func testAccFleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName string, totalTargetCapacity int) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  terminate_instances = true

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = %d
  }
}
`, totalTargetCapacity)
}

func testAccFleetConfig_TerminateInstancesWithExpiration(rName string, terminateInstancesWithExpiration bool) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  terminate_instances_with_expiration = %t

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, terminateInstancesWithExpiration)
}

func testAccFleetConfig_Type(rName, fleetType string) string {
	return testAccFleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  type = %q

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, fleetType)
}
