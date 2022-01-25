package ec2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2LaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`launch-template/.+`)),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "placement.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", "5"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_asg_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccLaunchTemplateConfig_asg_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "2"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservation_preference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservation_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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

func TestAccEC2LaunchTemplate_associatePublicIPAddress(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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

func TestAccEC2LaunchTemplate_NetworkInterface_ipv6Addresses(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
	var group autoscaling.Group
	groupName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceMarketOptions_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					testAccCheckAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix", "instance_market_options"},
			},
			{
				Config: testAccLaunchTemplateConfig_instanceMarketOptions_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(resourceName, &template),
					testAccCheckAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_licenseSpecification(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
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

func testAccCheckLaunchTemplateExists(n string, t *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.LaunchTemplates) != 1 || *resp.LaunchTemplates[0].LaunchTemplateId != rs.Primary.ID {
			return fmt.Errorf("Launch Template not found")
		}

		*t = *resp.LaunchTemplates[0]

		return nil
	}
}

func testAccCheckLaunchTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(resp.LaunchTemplates) != 0 && *resp.LaunchTemplates[0].LaunchTemplateId == rs.Primary.ID {
				return fmt.Errorf("Launch Template still exists")
			}
		}

		if tfawserr.ErrMessageContains(err, "InvalidLaunchTemplateId.NotFound", "") {
			log.Printf("[WARN] launch template (%s) not found.", rs.Primary.ID)
			continue
		}
		return err
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
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"
}

resource "aws_launch_template" "test" {
  name = %[1]q

  capacity_reservation_specification {
    capacity_reservation_target {
      capacity_reservation_id = aws_ec2_capacity_reservation.test.id
    }
  }
}
`, rName)
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

  tags = {
    Name = %[1]q
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

const testAccLaunchTemplateConfig_asg_basic = `
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "testbar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccLaunchTemplateConfig_asg_update = `
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "testbar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.nano"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccLaunchTemplateConfig_instanceMarketOptions_basic = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "instance_market_options"
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"

  instance_market_options {
    market_type = "spot"

    spot_options {
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccLaunchTemplateConfig_instanceMarketOptions_update = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "instance_market_options"
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"

  instance_market_options {
    market_type = "spot"

    spot_options {
      max_price          = "0.5"
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

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

func testAccCheckAutoScalingGroupExists(n string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		describeGroups, err := conn.DescribeAutoScalingGroups(
			&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			return err
		}

		if len(describeGroups.AutoScalingGroups) != 1 ||
			*describeGroups.AutoScalingGroups[0].AutoScalingGroupName != rs.Primary.ID {
			return fmt.Errorf("Auto Scaling Group not found")
		}

		*group = *describeGroups.AutoScalingGroups[0]

		return nil
	}
}
