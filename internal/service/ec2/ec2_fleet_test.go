package ec2_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
				Config: testAccFleetConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`fleet/.+`)),
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
				Config: testAccFleetConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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
				Config: testAccFleetConfigTags1(rName, "key1", "value1"),
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
				Config: testAccFleetConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFleetConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_memoryMiBAndVCPUCount(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.vcpu_count.0.min", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_mib {
                       min = 500
                       max = 24000
                     }
                     vcpu_count {
                       min = 1
                       max = 8
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_mib.0.max", "24000"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.vcpu_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.vcpu_count.0.max", "8"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_acceleratorCount(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_count {
                       min = 1
                     }
                     memory_mib {
                      min = 500
                     }
                     vcpu_count {
                      min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.0.min", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_count {
                       min = 1
                       max = 4
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.0.max", "4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_count {
                       max = 0
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_count.0.max", "0"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_acceleratorManufacturers(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_manufacturers = ["amd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amd"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_manufacturers = ["amazon-web-services", "amd", "nvidia", "xilinx"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "nvidia"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "xilinx"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_acceleratorNames(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_names = ["a100"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_names = ["a100", "v100", "k80", "t4", "m60", "radeon-pro-v520", "vu9p"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "v100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "k80"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "t4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "m60"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "radeon-pro-v520"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_names.*", "vu9p"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_acceleratorTotalMemoryMiB(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_total_memory_mib {
                       min = 1000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "1000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_total_memory_mib {
                       max = 24000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "24000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_total_memory_mib {
                       min = 1000
                       max = 24000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "1000"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "24000"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_acceleratorTypes(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_types = ["fpga"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_types = ["fpga", "gpu", "inference"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.*", "gpu"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.accelerator_types.*", "inference"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_bareMetal(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.bare_metal", "excluded"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.bare_metal", "included"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.bare_metal", "required"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_baselineEBSBandwidthMbps(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`baseline_ebs_bandwidth_mbps {
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_burstablePerformance(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.burstable_performance", "excluded"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "included"
                     memory_mib {
                       min = 1000
                     }
                     vcpu_count {
                       min = 2
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.burstable_performance", "included"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "required"
                     memory_mib {
                       min = 1000
                     }
                     vcpu_count {
                       min = 2
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.burstable_performance", "required"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_cpuManufacturers(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`cpu_manufacturers = ["amd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amd"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`cpu_manufacturers = ["amazon-web-services", "amd", "intel"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.cpu_manufacturers.*", "intel"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_excludedInstanceTypes(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`excluded_instance_types = ["t2.nano"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`excluded_instance_types = ["t2.nano", "t3*", "t4g.*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.*", "t3*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.excluded_instance_types.*", "t4g.*"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_instanceGenerations(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`instance_generations = ["current"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.instance_generations.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.instance_generations.*", "current"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`instance_generations = ["current", "previous"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.instance_generations.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.instance_generations.*", "current"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.instance_generations.*", "previous"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_localStorage(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage", "excluded"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage", "included"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage", "required"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_localStorageTypes(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage_types = ["hdd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage_types = ["hdd", "ssd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.local_storage_types.*", "ssd"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_memoryGiBPerVCPU(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_gib_per_vcpu {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_gib_per_vcpu {
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_gib_per_vcpu {
                       min = 0.5
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_networkInterfaceCount(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_interface_count {
                       min = 1
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.0.min", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_interface_count {
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.0.max", "10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_interface_count {
                       min = 1
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.network_interface_count.0.max", "10"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_onDemandMaxPricePercentageOverLowestPrice(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`on_demand_max_price_percentage_over_lowest_price = 50
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.on_demand_max_price_percentage_over_lowest_price", "50"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_requireHibernateSupport(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`require_hibernate_support = false
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.require_hibernate_support", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`require_hibernate_support = true
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.require_hibernate_support", "true"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_spotMaxPricePercentageOverLowestPrice(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`spot_max_price_percentage_over_lowest_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.spot_max_price_percentage_over_lowest_price", "75"),
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

func TestAccEC2Fleet_LaunchTemplateOverride_instanceRequirements_totalLocalStorageGB(t *testing.T) {
	var fleet ec2.FleetData
	resourceName := "aws_ec2_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`total_local_storage_gb {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`total_local_storage_gb {
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`total_local_storage_gb {
                       min = 0.5
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &fleet),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
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

func TestAccEC2Fleet_capacityRebalanceInvalidType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckFleet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccFleetConfigWithInvalidTypeForCapacityRebalance(rName),
				ExpectError: regexp.MustCompile(`Capacity Rebalance maintenance strategies can only be specified for fleets of type maintain`),
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

func testAccCheckFleetExists(n string, v *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindFleetByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFleetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_fleet" {
			continue
		}

		_, err := tfec2.FindFleetByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Fleet %s still exists", rs.Primary.ID)
	}

	return nil
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

func testAccFleetConfig_BaseLaunchTemplate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.micro"
  name          = %[1]q
}
`, rName))
}

func testAccFleetConfig_Basic(rName string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), `
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
`)
}

func testAccFleetConfig_multipleNetworkInterfaces(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

resource "aws_subnet" "test" {
  cidr_block = "10.1.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFleetConfig_ExcessCapacityTerminationPolicy(rName, excessCapacityTerminationPolicy string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  excess_capacity_termination_policy = %[2]q

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

  tags = {
    Name = %[1]q
  }
}
`, rName, excessCapacityTerminationPolicy))
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateID(rName, launchTemplateResourceName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_launch_template" "test1" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.micro"
  name          = "%[1]s1"
}

resource "aws_launch_template" "test2" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

  tags = {
    Name = %[1]q
  }
}
`, rName, launchTemplateResourceName))
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_LaunchTemplateName(rName, launchTemplateResourceName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_launch_template" "test1" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.micro"
  name          = "%[1]s1"
}

resource "aws_launch_template" "test2" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

  tags = {
    Name = %[1]q
  }
}
`, rName, launchTemplateResourceName))
}

func testAccFleetConfig_LaunchTemplateConfig_LaunchTemplateSpecification_Version(rName, instanceType string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = %[2]q
  name          = %[1]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_AvailabilityZone(rName string, availabilityZoneIndex int) string {
	return acctest.ConfigCompose(
		testAccFleetConfig_BaseLaunchTemplate(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      availability_zone = data.aws_availability_zones.available.names[%[2]d]
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, availabilityZoneIndex))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_InstanceRequirements(rName, instanceRequirements string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_requirements {
        %[2]s
      }
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "on-demand"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceRequirements))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_InstanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type = %[2]q
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_MaxPrice(rName, maxPrice string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      max_price = %[2]q
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, maxPrice))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_Priority(rName string, priority int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      priority = %[2]d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, priority))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_Priority_Multiple(rName string, priority1, priority2 int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type = aws_launch_template.test.instance_type
      priority      = %[2]d
    }

    override {
      instance_type = "t3.small"
      priority      = %[3]d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, priority1, priority2))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_SubnetID(rName string, subnetIndex int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block = "10.1.${count.index}.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      subnet_id = aws_subnet.test.*.id[%[2]d]
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetIndex))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity(rName string, weightedCapacity int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      weighted_capacity = %[2]d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, weightedCapacity))
}

func testAccFleetConfig_LaunchTemplateConfig_Override_WeightedCapacity_Multiple(rName string, weightedCapacity1, weightedCapacity2 int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }

    override {
      instance_type     = aws_launch_template.test.instance_type
      weighted_capacity = %[2]d
    }

    override {
      instance_type     = "t3.small"
      weighted_capacity = %[3]d
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, weightedCapacity1, weightedCapacity2))
}

func testAccFleetConfig_OnDemandOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  on_demand_options {
    allocation_strategy = %[2]q
  }

  target_capacity_specification {
    default_target_capacity_type = "on-demand"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, allocationStrategy))
}

func testAccFleetConfig_ReplaceUnhealthyInstances(rName string, replaceUnhealthyInstances bool) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  replace_unhealthy_instances = %[2]t

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

  tags = {
    Name = %[1]q
  }
}
`, rName, replaceUnhealthyInstances))
}

func testAccFleetConfig_SpotOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    allocation_strategy = %[2]q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, allocationStrategy))
}

func testAccFleetConfig_SpotOptions_CapacityRebalance(rName, allocationStrategy string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    allocation_strategy = %[2]q
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

  tags = {
    Name = %[1]q
  }
}
`, rName, allocationStrategy))
}

func testAccFleetConfigWithInvalidTypeForCapacityRebalance(rName string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  type = "request"

  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    allocation_strategy = "diversified"
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, instanceInterruptionBehavior string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    instance_interruption_behavior = %[2]q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceInterruptionBehavior))
}

func testAccFleetConfig_SpotOptions_InstancePoolsToUseCount(rName string, instancePoolsToUseCount int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  spot_options {
    instance_pools_to_use_count = %[2]d
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, instancePoolsToUseCount))
}

func testAccFleetConfigTags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  tags = {
    %[1]q = %[2]q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, key1, value1))
}

func testAccFleetConfigTags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, key1, value1, key2, value2))
}

func testAccFleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, defaultTargetCapacityType string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.test.id
      version            = aws_launch_template.test.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = %[2]q
    total_target_capacity        = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, defaultTargetCapacityType))
}

func testAccFleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName string, totalTargetCapacity int) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
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
    total_target_capacity        = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, totalTargetCapacity))
}

func testAccFleetConfig_TerminateInstancesWithExpiration(rName string, terminateInstancesWithExpiration bool) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  terminate_instances_with_expiration = %[2]t

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

  tags = {
    Name = %[1]q
  }
}
`, rName, terminateInstancesWithExpiration))
}

func testAccFleetConfig_Type(rName, fleetType string) string {
	return acctest.ConfigCompose(testAccFleetConfig_BaseLaunchTemplate(rName), fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  type = %[2]q

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

  tags = {
    Name = %[1]q
  }
}
`, rName, fleetType))
}
