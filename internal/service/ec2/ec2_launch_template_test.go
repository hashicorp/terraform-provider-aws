// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2LaunchTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`launch-template/.+`)),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_api_stop", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_initiated_shutdown_behavior", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, ""),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "key_name", ""),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "license_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "placement.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ram_disk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "security_group_names.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_data", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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
	ctx := acctest.Context(t)
	var launchTemplate awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &launchTemplate),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceLaunchTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2LaunchTemplate_BlockDeviceMappings_ebs(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_blockDeviceMappingsEBS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.encrypted", ""),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.throughput", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_type", ""),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_blockDeviceMappingsEBSDeleteOnTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_blockDeviceMappingsEBSDeleteOnTermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_blockDeviceMappingsEBSGP3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_ebsOptimized(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_ebsOptimized(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_ebsOptimized(rName, "\"true\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtTrue),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_ebsOptimized(rName, "\"false\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_ebsOptimized(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_elasticInferenceAccelerator(t *testing.T) {
	ctx := acctest.Context(t)
	var template1 awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_elasticInferenceAccelerator(rName, "eia1.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.medium"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_elasticInferenceAccelerator(rName, "eia1.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.large"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_NetworkInterfaces_deleteOnTermination(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfacesDeleteOnTermination(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", acctest.CtTrue),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_networkInterfacesDeleteOnTermination(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_networkInterfacesDeleteOnTermination(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_networkInterfacesDeleteOnTermination(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_data(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "disable_api_stop"),
					resource.TestCheckResourceAttrSet(resourceName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "image_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(resourceName, "kernel_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_name"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "placement.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, "Test Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Description 1"),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Description 2"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	asgResourceName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_asgBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_asgUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_CapacityReservation_preference(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservationPreference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_capacityReservationTarget(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalCoreCount := 2
	updatedCoreCount := 3
	originalThreadsPerCore := 2
	updatedThreadsPerCore := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_cpuOptions(rName, string(awstypes.AmdSevSnpSpecificationEnabled), originalCoreCount, originalThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resName, &template),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.core_count", strconv.Itoa(originalCoreCount)),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.threads_per_core", strconv.Itoa(originalThreadsPerCore)),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_cpuOptions(rName, string(awstypes.AmdSevSnpSpecificationDisabled), updatedCoreCount, updatedThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resName, &template),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.core_count", strconv.Itoa(updatedCoreCount)),
					resource.TestCheckResourceAttr(resName, "cpu_options.0.threads_per_core", strconv.Itoa(updatedThreadsPerCore)),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_CreditSpecification_nonBurstable(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "m1.small", "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "t2.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "t3.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
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

func TestAccEC2LaunchTemplate_CreditSpecification_t4g(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_creditSpecification(rName, "t4g.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template1 awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_iamInstanceProfileEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2LaunchTemplate_networkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.device_index", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.interface_type", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefix_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefix_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.network_card_index", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.primary_ipv6", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.private_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceAddresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_addresses.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceTypeEFA(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceCardIndex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.network_card_index", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv4PrefixCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefix_count", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv4Prefixes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_prefixes.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv6PrefixCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefix_count", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv6Prefixes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_prefixes.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_associatePublicIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_associateCarrierIPAddress(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_associateCarrierIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_Placement_hostResourceGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_placementHostResourceGroupARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "placement.0.host_resource_group_arn", "aws_resourcegroups_group.test", names.AttrARN),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_partition(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_partition(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_privateDNSNameOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_privateDNSNameOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_networkInterfaceIPv6Addresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_addresses.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_ipv6Count(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_address_count", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	asgResourceName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceMarketOptionsBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceMarketOptionsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(asgResourceName, "launch_template.0.version", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_primaryIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_primaryIPv6(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.primary_ipv6", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_primaryIPv6(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.primary_ipv6", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_primaryIPv6(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.primary_ipv6", ""),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_instanceRequirements_memoryMiBAndVCPUCount(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.0.min", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_mib.0.max", "4000"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.vcpu_count.0.min", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.min", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.min", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.max", acctest.Ct4),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_count.0.max", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_manufacturers.#", acctest.Ct4),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_names.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_types.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.accelerator_types.#", acctest.Ct3),
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

func TestAccEC2LaunchTemplate_instanceRequirements_allowedInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`allowed_instance_types = ["m4.large"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.allowed_instance_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.allowed_instance_types.*", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`allowed_instance_types = ["m4.large", "m5.*", "m6*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.allowed_instance_types.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.allowed_instance_types.*", "m4.large"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.allowed_instance_types.*", "m5.*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_requirements.0.allowed_instance_types.*", "m6*"),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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

func TestAccEC2LaunchTemplate_instanceRequirements_baselineEBSBandwidthMbps(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", acctest.Ct10),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", acctest.Ct10),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.cpu_manufacturers.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.cpu_manufacturers.#", acctest.Ct3),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.excluded_instance_types.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.excluded_instance_types.#", acctest.Ct3),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.instance_generations.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.instance_generations.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage_types.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.local_storage_types.#", acctest.Ct2),
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

func TestAccEC2LaunchTemplate_instanceRequirements_maxSpotPriceAsPercentageOfOptimalOnDemandPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`max_spot_price_as_percentage_of_optimal_on_demand_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.max_spot_price_as_percentage_of_optimal_on_demand_price", "75"),
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

func TestAccEC2LaunchTemplate_instanceRequirements_memoryGiBPerVCPU(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
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

func TestAccEC2LaunchTemplate_instanceRequirements_networkBandwidthGbps(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_bandwidth_gbps {
                       min = 1.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.0.min", "1.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_bandwidth_gbps {
                       max = 200
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.0.max", "200"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_instanceRequirements(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix),
					`network_bandwidth_gbps {
                       min = 2.5
                       max = 250
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.0.min", "2.5"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_bandwidth_gbps.0.max", "250"),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.min", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.max", acctest.Ct10),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.min", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.network_interface_count.0.max", acctest.Ct10),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.require_hibernate_support", acctest.CtFalse),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.require_hibernate_support", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/license-manager.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_licenseSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "license_specification.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", ""),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", ""),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", ""),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", names.AttrEnabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_metadataOptionsNoHTTPEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled), //Setting any of the values in metadata options will set the http_endpoint to enabled, you will not see it via the Console, but will in the API for any instance made from the template
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", names.AttrEnabled),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_enclaveOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", acctest.CtTrue),
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
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_hibernation(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_hibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchTemplateConfig_hibernation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", acctest.CtFalse),
				),
			},
			{
				Config: testAccLaunchTemplateConfig_hibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplate_defaultVersion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct1),
				),
			},
			// An updated config should cause a new version to be created
			// but keep the default_version unchanged if unset
			{
				Config: testAccLaunchTemplateConfig_description(rName, descriptionNew),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
				),
			},
			// An updated config to set the default_version to an existing version
			// should not cause a new version to be created
			{
				Config: testAccLaunchTemplateConfig_descriptionDefaultVersion(rName, descriptionNew, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
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
	ctx := acctest.Context(t)
	resourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct1),
				),
			},
			// Updating a field should create a new version but not update the default_version
			// if update_default_version is set to false
			{
				Config: testAccLaunchTemplateConfig_configDescriptionUpdateDefaultVersion(rName, descriptionNew, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
				),
			},
			// Only updating the update_default_version to true should not create a new version
			// but update the template version to the latest available
			{
				Config: testAccLaunchTemplateConfig_configDescriptionUpdateDefaultVersion(rName, descriptionNew, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
				),
			},
			// Updating a field should create a new version and update the default_version
			// if update_default_version is set to true
			{
				Config: testAccLaunchTemplateConfig_configDescriptionUpdateDefaultVersion(rName, description, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct3),
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

func testAccCheckLaunchTemplateExists(ctx context.Context, n string, v *awstypes.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Launch Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindLaunchTemplateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLaunchTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_launch_template" {
				continue
			}

			_, err := tfec2.FindLaunchTemplateByID(ctx, conn, rs.Primary.ID)

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
}

func testAccLaunchTemplateConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}
`, rName)
}

func testAccLaunchTemplateConfig_nameGenerated() string {
	return `
resource "aws_launch_template" "test" {}
`
}

func testAccLaunchTemplateConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccLaunchTemplateConfig_ipv6Count(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_address_count = 1
  }
}
`, rName)
}

func testAccLaunchTemplateConfig_blockDeviceMappingsEBS(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_blockDeviceMappingsEBSDeleteOnTermination(rName string, deleteOnTermination bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_blockDeviceMappingsEBSGP3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_networkInterfacesDeleteOnTermination(rName string, deleteOnTermination string) string {
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

func testAccLaunchTemplateConfig_ebsOptimized(rName, ebsOptimized string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  ebs_optimized = %[1]s
  name          = %[2]q
}
`, ebsOptimized, rName)
}

func testAccLaunchTemplateConfig_elasticInferenceAccelerator(rName, elasticInferenceAcceleratorType string) string {
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

  disable_api_stop        = true
  disable_api_termination = true
  ebs_optimized           = false

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
    resource_type = "network-interface"

    tags = {
      Name = "test"
    }
  }
}
`, rName)) //lintignore:AWSAT002
}

func testAccLaunchTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLaunchTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccLaunchTemplateConfig_capacityReservationPreference(rName string, preference string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  capacity_reservation_specification {
    capacity_reservation_preference = %[2]q
  }
}
`, rName, preference)
}

func testAccLaunchTemplateConfig_capacityReservationTarget(rName string) string {
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

func testAccLaunchTemplateConfig_cpuOptions(rName, amdSevSnp string, coreCount, threadsPerCore int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  cpu_options {
    amd_sev_snp      = %[2]q
    core_count       = %[3]d
    threads_per_core = %[4]d
  }
}
`, rName, amdSevSnp, coreCount, threadsPerCore)
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

func testAccLaunchTemplateConfig_iamInstanceProfileEmptyConfigurationBlock(rName string) string {
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

func testAccLaunchTemplateConfig_partition(rName string, partNum int) string {
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

func testAccLaunchTemplateConfig_placementHostResourceGroupARN(rName string) string {
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

func testAccLaunchTemplateConfig_privateDNSNameOptions(rName string) string {
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

func testAccLaunchTemplateConfig_primaryIPv6(rName, primaryIPv6 string) string {
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
    primary_ipv6         = %[2]s
  }
}
`, rName, primaryIPv6)
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

func testAccLaunchTemplateConfig_networkInterfaceIPv6Addresses(rName string) string {
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

func testAccLaunchTemplateConfig_networkInterfaceTypeEFA(rName string) string {
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

func testAccLaunchTemplateConfig_asgBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_asgUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.nano", "t2.nano"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_instanceMarketOptionsBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccLaunchTemplateConfig_instanceMarketOptionsUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name     = %[1]q
  image_id = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_requirements {
    %[2]s
  }
}
`, rName, instanceRequirements))
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
func testAccLaunchTemplateConfig_metadataOptionsNoHTTPEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
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

func testAccLaunchTemplateConfig_hibernation(rName string, enabled bool) string {
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

func testAccLaunchTemplateConfig_configDescriptionUpdateDefaultVersion(rName, description string, update bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name                   = %[1]q
  description            = %[2]q
  update_default_version = %[3]t
}
`, rName, description, update)
}
