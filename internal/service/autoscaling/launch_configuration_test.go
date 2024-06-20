// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	ec2awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingLaunchConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(`launchConfiguration:.+`)),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_monitoring", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile", ""),
					resource.TestCheckResourceAttrSet(resourceName, "image_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "key_name", ""),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "placement_tenancy", ""),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "spot_price", ""),
					resource.TestCheckNoResourceAttr(resourceName, "user_data"),
					resource.TestCheckNoResourceAttr(resourceName, "user_data_base64"),
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

func TestAccAutoScalingLaunchConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceLaunchConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_Name_generated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
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

func TestAccAutoScalingLaunchConfiguration_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
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

func TestAccAutoScalingLaunchConfiguration_withBlockDevices(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_blockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdc",
						names.AttrIOPS:       "100",
						names.AttrVolumeSize: acctest.Ct10,
						names.AttrVolumeType: "io1",
					}),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						names.AttrDeviceName:  "/dev/sde",
						names.AttrVirtualName: "ephemeral0",
					}),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						names.AttrVolumeSize: "11",
						names.AttrVolumeType: "gp2",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_RootBlockDevice_amiDisappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ami ec2awstypes.Image
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	amiCopyResourceName := "aws_ami_copy.test"
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_cofingRootBlockDeviceCopiedAMI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckAMIExists(ctx, amiCopyResourceName, &ami),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceAMI(), amiCopyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_rootBlockDeviceVolumeSize(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_RootBlockDevice_volumeSize(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_rootBlockDeviceVolumeSize(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "11"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_rootBlockDeviceVolumeSize(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "20"),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_encryptedRootBlockDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_encryptedRootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						names.AttrEncrypted:  acctest.CtTrue,
						names.AttrVolumeSize: "11",
						names.AttrVolumeType: "gp2",
					}),
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

func TestAccAutoScalingLaunchConfiguration_withSpotPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_spotPrice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
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

func TestAccAutoScalingLaunchConfiguration_withIAMProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_iamProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "iam_instance_profile"),
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

func TestAccAutoScalingLaunchConfiguration_withGP3(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_gp3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						names.AttrEncrypted:  acctest.CtTrue,
						names.AttrThroughput: "150",
						names.AttrVolumeSize: "9",
						names.AttrVolumeType: "gp3",
					}),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						names.AttrVolumeSize: "11",
						names.AttrVolumeType: "gp3",
					}),
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

func TestAccAutoScalingLaunchConfiguration_encryptedEBSBlockDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_encryptedEBSBlockDevice(rName, 9),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						names.AttrEncrypted:  acctest.CtTrue,
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						names.AttrVolumeSize: "11",
						names.AttrVolumeType: "gp2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_encryptedEBSBlockDevice(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						names.AttrEncrypted:  acctest.CtTrue,
						names.AttrVolumeSize: acctest.Ct10,
					}),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						names.AttrVolumeSize: "11",
						names.AttrVolumeType: "gp2",
					}),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_metadataOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
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

func TestAccAutoScalingLaunchConfiguration_EBS_noDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_ebsNoDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sda2",
						"no_device":          acctest.CtTrue,
					}),
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

func TestAccAutoScalingLaunchConfiguration_userData(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_userData(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_userDataBase64(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetFalseConfigNull(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, false),
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

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetFalseConfigFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, false, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, false),
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

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetFalseConfigTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, false, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, true),
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

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetTrueConfigNull(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, true, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, true),
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

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetTrueConfigFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, true, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, false),
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

func TestAccAutoScalingLaunchConfiguration_AssociatePublicIPAddress_subnetTrueConfigTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.LaunchConfiguration
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"
	groupResourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_associatePublicIPAddress(rName, true, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(ctx, resourceName, &conf),
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 1),
					testAccCheckInstanceHasPublicIPAddress(ctx, &group, 0, true),
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

func testAccCheckLaunchConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_launch_configuration" {
				continue
			}

			_, err := tfautoscaling.FindLaunchConfigurationByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Launch Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLaunchConfigurationExists(ctx context.Context, n string, v *awstypes.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindLaunchConfigurationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAMIExists(ctx context.Context, n string, v *ec2awstypes.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindImageByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckInstanceHasPublicIPAddress(ctx context.Context, group *awstypes.AutoScalingGroup, idx int, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instanceID := aws.ToString(group.Instances[idx].InstanceId)
		instance, err := tfec2.FindInstanceByID(ctx, conn, instanceID)

		if err != nil {
			return err
		}

		hasPublicIPAddress := aws.ToString(instance.PublicIpAddress) != ""

		if hasPublicIPAddress != expected {
			return fmt.Errorf("%s has public IP address; got %t, expected %t", instanceID, hasPublicIPAddress, expected)
		}

		return nil
	}
}

func testAccLaunchConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}
`, rName))
}

func testAccLaunchConfigurationConfig_nameGenerated() string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), `
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}
`)
}

func testAccLaunchConfigurationConfig_namePrefix(namePrefix string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}
`, namePrefix))
}

func testAccLaunchConfigurationConfig_blockDevices(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m1.small"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_cofingRootBlockDeviceCopiedAMI(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = aws_ami_copy.test.id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = 10
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_rootBlockDeviceVolumeSize(rName string, volumeSize int) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = %[2]d
  }
}
`, rName, volumeSize))
}

func testAccLaunchConfigurationConfig_encryptedRootBlockDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.nano"

  root_block_device {
    encrypted   = true
    volume_type = "gp2"
    volume_size = 11
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_spotPrice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  spot_price    = "0.05"
}
`, rName))
}

func testAccLaunchConfigurationConfig_iamProfile(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_launch_configuration" "test" {
  name                 = %[1]q
  image_id             = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = "t2.nano"
  iam_instance_profile = aws_iam_instance_profile.test.name
}
`, rName))
}

func testAccLaunchConfigurationConfig_gp3(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  root_block_device {
    volume_type = "gp3"
    volume_size = 11
  }

  ebs_block_device {
    volume_type = "gp3"
    device_name = "/dev/sdb"
    volume_size = 9
    encrypted   = true
    throughput  = 150
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_encryptedEBSBlockDevice(rName string, size int) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = %[2]d
    encrypted   = true
  }
}
`, rName, size))
}

func testAccLaunchConfigurationConfig_metadataOptions(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.nano"
  name          = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_ebsNoDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m1.small"

  ebs_block_device {
    device_name = "/dev/sda2"
    no_device   = true
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_userData(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  user_data     = "foo:-with-character's"
}
`, rName))
}

func testAccLaunchConfigurationConfig_userDataBase64(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name             = %[1]q
  image_id         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type    = "t2.micro"
  user_data_base64 = base64encode("hello world")
}
`, rName))
}

func testAccLaunchConfigurationConfig_associatePublicIPAddress(rName string, subnetMapPublicIPOnLaunch bool, associatePublicIPAddress string) string {
	if associatePublicIPAddress == "" {
		associatePublicIPAddress = "null"
	}

	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[1]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = %[2]t

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  associate_public_ip_address = %[3]s
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = [aws_subnet.test.id]
  max_size             = 1
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, subnetMapPublicIPOnLaunch, associatePublicIPAddress))
}
