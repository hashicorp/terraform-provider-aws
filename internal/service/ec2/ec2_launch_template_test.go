package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2LaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNameConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`launch-template/.+`)),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", "false"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
					resource.TestCheckResourceAttr(resourceName, "elastic_gpu_specifications.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_initiated_shutdown_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", ""),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "key_name", ""),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "license_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "placement.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ram_disk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "security_group_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "user_data", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "0"),
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

func TestAccEC2LaunchTemplate_Name_generated(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNameGeneratedConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccEC2LaunchTemplate_Name_prefix(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNamePrefixConfig("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccEC2LaunchTemplate_disappears(t *testing.T) {
	var launchTemplate ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &launchTemplate),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceLaunchTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2LaunchTemplate_BlockDeviceMappings_ebs(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_BlockDeviceMappings_EBS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
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

func TestAccEC2LaunchTemplate_BlockDeviceMappingsEBS_deleteOnTermination(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "false"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
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

func TestAccEC2LaunchTemplate_BlockDeviceMappingsEBS_gp3(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_BlockDeviceMappings_EBS_GP3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.iops", "4000"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_type", "gp3"),
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

func TestAccEC2LaunchTemplate_ebsOptimized(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_EBSOptimized(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_EBSOptimized(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_EBSOptimized(rName, "\"true\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_EBSOptimized(rName, "\"false\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_EBSOptimized(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_elasticInferenceAccelerator(t *testing.T) {
	var template1 ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateElasticInferenceAcceleratorConfig(rName, "eia1.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.medium"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateElasticInferenceAcceleratorConfig(rName, "eia1.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.large"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_NetworkInterfaces_deleteOnTermination(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", "true"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", "false"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
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

func TestAccEC2LaunchTemplate_data(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_data(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "elastic_gpu_specifications.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "image_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttrSet(resourceName, "kernel_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_name"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "placement.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
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

func TestAccEC2LaunchTemplate_description(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, "Test Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_description(rName, "Test Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Description 2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_update(t *testing.T) {
	var template ec2.LaunchTemplate
	asgResourceName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_ASG_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_ASG_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_tags(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLaunchTemplateTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_CapacityReservation_preference(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservation_preference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
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

func TestAccEC2LaunchTemplate_CapacityReservation_target(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservation_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", ""),
					resource.TestCheckResourceAttrSet(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn", ""),
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

func TestAccEC2LaunchTemplate_cpuOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_cpuOptions(rName, 4, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resName, &template),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_CreditSpecification_nonBurstable(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "m1.small", "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification"},
			},
		},
	})
}

func TestAccEC2LaunchTemplate_CreditSpecification_t2(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "t2.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

func TestAccEC2LaunchTemplate_CreditSpecification_t3(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "t3.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6757
func TestAccEC2LaunchTemplate_IAMInstanceProfile_emptyBlock(t *testing.T) {
	var template1 ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateIAMInstanceProfileEmptyConfigurationBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2LaunchTemplate_networkInterface(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.device_index", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.interface_type", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_addresses.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefix_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefixes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefix_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefixes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.network_card_index", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.private_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.subnet_id", ""),
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

func TestAccEC2LaunchTemplate_networkInterfaceAddresses(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceAddresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_addresses.#", "2"),
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

func TestAccEC2LaunchTemplate_networkInterfaceType(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceType_efa(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.interface_type", "efa"),
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

func TestAccEC2LaunchTemplate_networkInterfaceCardIndex(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceCardIndex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.network_card_index", "1"),
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

func TestAccEC2LaunchTemplate_networkInterfaceIPv4PrefixCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv4PrefixCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefix_count", "1"),
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

func TestAccEC2LaunchTemplate_networkInterfaceIPv4Prefixes(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv4Prefixes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefixes.#", "2"),
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

func TestAccEC2LaunchTemplate_networkInterfaceIPv6PrefixCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv6PrefixCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefix_count", "2"),
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

func TestAccEC2LaunchTemplate_networkInterfaceIPv6Prefixes(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv6Prefixes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefixes.#", "1"),
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

func TestAccEC2LaunchTemplate_associatePublicIPAddress(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_associateCarrierIPAddress(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_Placement_hostResourceGroupARN(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplatePlacementHostResourceGroupARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "placement.0.host_resource_group_arn", "aws_resourcegroups_group.test", "arn"),
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

func TestAccEC2LaunchTemplate_Placement_partitionNum(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplatePartitionConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplatePartitionConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_privateDNSNameOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplatePrivateDNSNameOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", "true"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", "false"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "resource-name"),
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

func TestAccEC2LaunchTemplate_NetworkInterface_ipv6Addresses(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterface_ipv6Addresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_addresses.#", "2"),
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

func TestAccEC2LaunchTemplate_NetworkInterface_ipv6AddressCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_ipv6_count(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_address_count", "1"),
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

func TestAccEC2LaunchTemplate_instanceMarketOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	asgResourceName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_InstanceMarketOptions_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_InstanceMarketOptions_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_instanceRequirements_memoryMiBAndVCpuCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.0.min", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`memory_mib {
                       min = 500
                       max = 4000
                     }
                     vcpu_count {
                       min = 1
                       max = 8
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.max", "4000"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.0.max", "8"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_acceleratorCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.min", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.max", "4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.max", "0"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_acceleratorManufacturers(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_manufacturers = ["amd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.*", "amd"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_manufacturers = ["amazon-web-services", "amd", "nvidia", "xilinx"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.*", "nvidia"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.*", "xilinx"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_acceleratorNames(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_names = ["a100"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "a100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_names = ["a100", "v100", "k80", "t4", "m60", "radeon-pro-v520", "vu9p"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_names.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "a100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "v100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "k80"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "t4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "m60"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "radeon-pro-v520"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_names.*", "vu9p"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_acceleratorTotalMemoryMiB(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.0.min", "1000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.0.max", "24000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.0.min", "1000"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.0.max", "24000"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_acceleratorTypes(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_types = ["fpga"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_types.*", "fpga"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`accelerator_types = ["fpga", "gpu", "inference"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_types.*", "fpga"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_types.*", "gpu"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.accelerator_types.*", "inference"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_bareMetal(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.bare_metal", "excluded"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.bare_metal", "included"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`bare_metal = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.bare_metal", "required"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_baselineEbsBandwidthMbps(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", "10"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_burstablePerformance(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.burstable_performance", "excluded"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "included"
                     memory_mib {
                       min = 1000
                     }
                     vcpu_count {
                       min = 2
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.burstable_performance", "included"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`burstable_performance = "required"
                     memory_mib {
                       min = 1000
                     }
                     vcpu_count {
                       min = 2
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.burstable_performance", "required"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_cpuManufacturers(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`cpu_manufacturers = ["amd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.cpu_manufacturers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.cpu_manufacturers.*", "amd"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`cpu_manufacturers = ["amazon-web-services", "amd", "intel"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.cpu_manufacturers.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.cpu_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.cpu_manufacturers.*", "intel"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_excludedInstanceTypes(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`excluded_instance_types = ["t2.nano"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.excluded_instance_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.excluded_instance_types.*", "t2.nano"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`excluded_instance_types = ["t2.nano", "t3*", "t4g.*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.excluded_instance_types.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.excluded_instance_types.*", "t2.nano"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.excluded_instance_types.*", "t3*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.excluded_instance_types.*", "t4g.*"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_instanceGenerations(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`instance_generations = ["current"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.instance_generations.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.instance_generations.*", "current"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`instance_generations = ["current", "previous"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.instance_generations.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.instance_generations.*", "current"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.instance_generations.*", "previous"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_localStorage(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage", "excluded"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage", "included"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage", "required"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_localStorageTypes(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage_types = ["hdd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.local_storage_types.*", "hdd"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`local_storage_types = ["hdd", "ssd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.local_storage_types.*", "hdd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.local_storage_types.*", "ssd"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_memoryGiBPerVCpu(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_networkInterfaceCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.min", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.max", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.max", "10"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_onDemandMaxPricePercentageOverLowestPrice(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`on_demand_max_price_percentage_over_lowest_price = 50
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.on_demand_max_price_percentage_over_lowest_price", "50"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_requireHibernateSupport(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`require_hibernate_support = false
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.require_hibernate_support", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`require_hibernate_support = true
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.require_hibernate_support", "true"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_spotMaxPricePercentageOverLowestPrice(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`spot_max_price_percentage_over_lowest_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.spot_max_price_percentage_over_lowest_price", "75"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_totalLocalStorageGB(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
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
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
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

func TestAccEC2LaunchTemplate_licenseSpecification(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/license-manager.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_licenseSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "license_specification.#", "1"),
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

func TestAccEC2LaunchTemplate_metadataOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_metadataOptionsIPv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_metadataOptionsInstanceTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "enabled"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_enclaveOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_enclaveOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "false"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_hibernation(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateHibernationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateHibernationConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "false"),
				),
			},
			{
				Config: testAccLaunchTemplateHibernationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "true"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_defaultVersion(t *testing.T) {
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
				),
			},
			// An updated config should cause a new version to be created
			// but keep the default_version unchanged if unset
			{
				Config: testAccLaunchTemplateConfig_description(rName, descriptionNew),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// An updated config to set the default_version to an existing version
			// should not cause a new version to be created
			{
				Config: testAccLaunchTemplateConfig_descriptionDefaultVersion(rName, descriptionNew, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
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

func TestAccEC2LaunchTemplate_updateDefaultVersion(t *testing.T) {
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
				),
			},
			// Updating a field should create a new version but not update the default_version
			// if update_default_version is set to false
			{
				Config: testAccLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, descriptionNew, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// Only updating the update_default_version to true should not create a new version
			// but update the template version to the latest available
			{
				Config: testAccLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, descriptionNew, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// Updating a field should create a new version and update the default_version
			// if update_default_version is set to true
			{
				Config: testAccLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, description, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "3"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"update_default_version",
				},
			},
		},
	})
}

func testAccCheckLaunchTemplateExists(n string, v *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Launch Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindLaunchTemplateByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLaunchTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		_, err := tfec2.FindLaunchTemplateByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Launch Template %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccLaunchTemplateNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}
`, rName)
}

func testAccLaunchTemplateNameGeneratedConfig() string {
	return `
resource "aws_launch_template" "test" {}
`
}

func testAccLaunchTemplateNamePrefixConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccLaunchTemplateConfig_ipv6_count(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_address_count = 1
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_BlockDeviceMappings_EBS(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName string, deleteOnTermination bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      delete_on_termination = %[2]t
      volume_size           = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName, deleteOnTermination))
}

func testAccLaunchTemplateConfig_BlockDeviceMappings_EBS_GP3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      iops        = 4000
      throughput  = 500
      volume_size = 15
      volume_type = "gp3"
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName string, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id  = "eni-123456ab"
    security_groups       = ["sg-1a23bc45"]
    delete_on_termination = %[2]s
  }
}
`, rName, deleteOnTermination)
}

func testAccLaunchTemplateConfig_EBSOptimized(rName, ebsOptimized string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  ebs_optimized = %[1]s
  name          = %[2]q
}
`, ebsOptimized, rName)
}

func testAccLaunchTemplateElasticInferenceAcceleratorConfig(rName, elasticInferenceAcceleratorType string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  elastic_inference_accelerator {
    type = %[2]q
  }
}
`, rName, elasticInferenceAcceleratorType)
}

func testAccLaunchTemplateConfig_data(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  block_device_mappings {
    device_name = "test"
  }

  maintenance_options {
    auto_recovery = "disabled"
  }

  disable_api_termination = true
  ebs_optimized           = false

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id                             = "ami-12a3b456"
  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

  instance_type = "t2.micro"
  kernel_id     = "aki-a12bc3de"
  key_name      = "test"

  monitoring {
    enabled = true
  }

  network_interfaces {
    network_interface_id = "eni-123456ab"
    security_groups      = ["sg-1a23bc45"]
  }

  placement {
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  ram_disk_id            = "ari-a12bc3de"
  vpc_security_group_ids = ["sg-12a3b45c"]

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "spot-instances-request"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "elastic-gpu"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "network-interface"

    tags = {
      Name = "test"
    }
  }
}
`, rName)) //lintignore:AWSAT002
}

func testAccLaunchTemplateTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLaunchTemplateTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccLaunchTemplateConfig_capacityReservation_preference(rName string, preference string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  capacity_reservation_specification {
    capacity_reservation_preference = %[2]q
  }
}
`, rName, preference)
}

func testAccLaunchTemplateConfig_capacityReservation_target(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  capacity_reservation_specification {
    capacity_reservation_target {
      capacity_reservation_id = aws_ec2_capacity_reservation.test.id
    }
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_cpuOptions(rName string, coreCount, threadsPerCore int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  cpu_options {
    core_count       = %[2]d
    threads_per_core = %[3]d
  }
}
`, rName, coreCount, threadsPerCore)
}

func testAccLaunchTemplateConfig_creditSpecification(rName, instanceType, cpuCredits string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  instance_type = %[1]q
  name          = %[2]q

  credit_specification {
    cpu_credits = %[3]q
  }
}
`, instanceType, rName, cpuCredits)
}

func testAccLaunchTemplateIAMInstanceProfileEmptyConfigurationBlockConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  iam_instance_profile {}
}
`, rName)
}

func testAccLaunchTemplateConfig_licenseSpecification(rName string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                  = "Test"
  license_counting_type = "vCPU"
}

resource "aws_launch_template" "test" {
  name = %[1]q

  license_specification {
    license_configuration_arn = aws_licensemanager_license_configuration.test.id
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccLaunchTemplateConfig_networkInterface(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id = aws_network_interface.test.id
    ipv4_address_count   = 2
  }
}
`, rName)
}

func testAccLaunchTemplatePartitionConfig(rName string, partNum int) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "partition"
}

resource "aws_launch_template" "test" {
  name = %[1]q

  placement {
    group_name       = aws_placement_group.test.name
    partition_number = %[2]d
  }
}
`, rName, partNum)
}

func testAccLaunchTemplatePlacementHostResourceGroupARNConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name = %[1]q

  resource_query {
    query = jsonencode({
      ResourceTypeFilters = ["AWS::EC2::Instance"]
      TagFilters = [
        {
          Key    = "Stage"
          Values = ["Test"]
        },
      ]
    })
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  placement {
    host_resource_group_arn = aws_resourcegroups_group.test.arn
  }
}
`, rName)
}

func testAccLaunchTemplatePrivateDNSNameOptionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  private_dns_name_options {
    enable_resource_name_dns_aaaa_record = true
    enable_resource_name_dns_a_record    = false
    hostname_type                        = "resource-name"
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceAddresses(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id = aws_network_interface.test.id
    ipv4_addresses       = ["10.1.0.10", "10.1.0.11"]
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_associatePublicIPAddress(rName, associatePublicIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id        = aws_network_interface.test.id
    associate_public_ip_address = %[2]s
    ipv4_address_count          = 2
  }
}
`, rName, associatePublicIPAddress)
}

func testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, associateCarrierIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id         = aws_network_interface.test.id
    associate_carrier_ip_address = %[2]s
    ipv4_address_count           = 2
  }
}
`, rName, associateCarrierIPAddress)
}

func testAccLaunchTemplateConfig_networkInterface_ipv6Addresses(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_addresses = [
      "0:0:0:0:0:ffff:a01:5",
      "0:0:0:0:0:ffff:a01:6",
    ]
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceType_efa(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    interface_type = "efa"
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceCardIndex(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  instance_type = "p4d.24xlarge"

  network_interfaces {
    network_card_index = 1
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceIPv4PrefixCount(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv4_prefix_count = 1
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceIPv4Prefixes(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv4_prefixes = ["172.16.10.16/28", "172.16.10.32/28"]
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceIPv6PrefixCount(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_prefix_count = 2
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_networkInterfaceIPv6Prefixes(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_prefixes = ["2001:db8::/80"]
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_ASG_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_ASG_update(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.nano", "t2.nano"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_InstanceMarketOptions_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  instance_market_options {
    market_type = "spot"

    spot_options {
      spot_instance_type = "one-time"
    }
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_InstanceMarketOptions_update(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  instance_market_options {
    market_type = "spot"

    spot_options {
      max_price          = "0.5"
      spot_instance_type = "one-time"
    }
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`, rName))
}

func testAccLaunchTemplateConfig_instanceRequirements(rName, instanceRequirements string) string {
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
  name     = %[1]q
  image_id = data.aws_ami.test.id

  instance_requirements {
    %[2]s
  }
}
`, rName, instanceRequirements)
}

func testAccLaunchTemplateConfig_metadataOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_metadataOptionsIPv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    http_protocol_ipv6          = "enabled"
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_metadataOptionsInstanceTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    http_protocol_ipv6          = "enabled"
    instance_metadata_tags      = "enabled"
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_enclaveOptions(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  enclave_options {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccLaunchTemplateHibernationConfig(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  hibernation_options {
    configured = %[2]t
  }
}
`, rName, enabled)
}

func testAccLaunchTemplateConfig_descriptionDefaultVersion(rName, description string, version int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name            = %[1]q
  description     = %[2]q
  default_version = %[3]d
}
`, rName, description, version)
}

func testAccLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, description string, update bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name                   = %[1]q
  description            = %[2]q
  update_default_version = %[3]t
}
`, rName, description, update)
}
