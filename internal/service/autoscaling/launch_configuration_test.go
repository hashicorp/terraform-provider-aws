package autoscaling_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingLaunchConfiguration_basic(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(`launchConfiguration:.+`)),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_monitoring", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile", ""),
					resource.TestCheckResourceAttrSet(resourceName, "image_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "key_name", ""),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "placement_tenancy", ""),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", ""),
					resource.TestCheckNoResourceAttr(resourceName, "user_data"),
					resource.TestCheckNoResourceAttr(resourceName, "user_data_base64"),
					resource.TestCheckResourceAttr(resourceName, "vpc_classic_link_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_classic_link_security_groups.#", "0"),
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
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscaling.ResourceLaunchConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_Name_generated(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
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

func TestAccAutoScalingLaunchConfiguration_namePrefix(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
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

func TestAccAutoScalingLaunchConfiguration_withBlockDevices(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withBlockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdb",
						"volume_size": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdc",
						"iops":        "100",
						"volume_size": "10",
						"volume_type": "io1",
					}),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sde",
						"virtual_name": "ephemeral0",
					}),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						"volume_size": "11",
						"volume_type": "gp2",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_withInstanceStoreAMI(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withInstanceStoreAMI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
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

func TestAccAutoScalingLaunchConfiguration_RootBlockDevice_amiDisappears(t *testing.T) {
	var ami ec2.Image
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	amiCopyResourceName := "aws_ami_copy.test"
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationCofing_withRootBlockDeviceCopiedAMI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					testAccCheckAMIExists(amiCopyResourceName, &ami),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMI(), amiCopyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_withRootBlockDeviceVolumeSize(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_RootBlockDevice_volumeSize(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withRootBlockDeviceVolumeSize(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "11"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfigurationConfig_withRootBlockDeviceVolumeSize(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "20"),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_encryptedRootBlockDevice(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withEncryptedRootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "root_block_device.*", map[string]string{
						"encrypted":   "true",
						"volume_size": "11",
						"volume_type": "gp2",
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
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationWithSpotPriceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_withVPCClassicLink(t *testing.T) {
	var vpc ec2.Vpc
	var group ec2.SecurityGroup
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withVPCClassicLink(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					acctest.CheckVPCExists("aws_vpc.test", &vpc),
					testAccCheckSecurityGroupExists("aws_security_group.test", &group),
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
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_withIAMProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_withEncryption(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationWithEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists("aws_launch_configuration.test", &conf),
					testAccCheckLaunchConfigurationWithEncryption(&conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_withGP3(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationWithGP3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists("aws_launch_configuration.test", &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_type": "gp3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"throughput": "150",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_updateEBSBlockDevices(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationWithEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "9",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
			{
				Config: testAccLaunchConfigurationWithEncryptionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "10",
					}),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_metadataOptions(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationMetadataOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
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
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationEBSNoDeviceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sda2",
						"no_device":   "true",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAutoScalingLaunchConfiguration_userData(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationConfig_userData(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
			{
				Config: testAccLaunchConfigurationConfig_userDataBase64(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
		},
	})
}

func testAccCheckLaunchConfigurationWithEncryption(conf *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Map out the block devices by name, which should be unique.
		blockDevices := make(map[string]*autoscaling.BlockDeviceMapping)
		for _, blockDevice := range conf.BlockDeviceMappings {
			blockDevices[*blockDevice.DeviceName] = blockDevice
		}

		// Check if the root block device exists.
		if _, ok := blockDevices["/dev/xvda"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/xvda")
		} else if blockDevices["/dev/xvda"].Ebs.Encrypted != nil {
			return fmt.Errorf("root device should not include value for Encrypted")
		}

		// Check if the secondary block device exists.
		if _, ok := blockDevices["/dev/sdb"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sdb")
		} else if !*blockDevices["/dev/sdb"].Ebs.Encrypted {
			return fmt.Errorf("block device isn't encrypted as expected: /dev/sdb")
		}

		return nil
	}
}

func testAccCheckLaunchConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_configuration" {
			continue
		}

		_, err := tfautoscaling.FindLaunchConfigurationByName(conn, rs.Primary.ID)

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

func testAccCheckLaunchConfigurationExists(n string, v *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Launch Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindLaunchConfigurationByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAMIExists(n string, v *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 AMI ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindImageByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSecurityGroupExists(n string, v *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Security Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindSecurityGroupByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLaunchConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}
`, rName))
}

func testAccLaunchConfigurationConfig_nameGenerated() string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), `
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}
`)
}

func testAccLaunchConfigurationConfig_namePrefix(namePrefix string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}
`, namePrefix))
}

func testAccLaunchConfigurationConfig_withBlockDevices(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

// testAccLatestAmazonLinuxPVInstanceStoreAMIConfig returns the configuration for a data source that
// describes the latest Amazon Linux AMI using PV virtualization and an instance store root device.
// The data source is named 'amzn-ami-minimal-pv-ebs'.
func testAccLatestAmazonLinuxPVInstanceStoreAMIConfig() string {
	return `
data "aws_ami" "amzn-ami-minimal-pv-instance-store" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["instance-store"]
  }
}
`
}

func testAccLaunchConfigurationConfig_withInstanceStoreAMI(rName string) string {
	return acctest.ConfigCompose(testAccLatestAmazonLinuxPVInstanceStoreAMIConfig(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name     = %[1]q
  image_id = data.aws_ami.amzn-ami-minimal-pv-instance-store.id

  # When the instance type is updated, the new type must support ephemeral storage.
  instance_type = "m1.small"
}
`, rName))
}

func testAccLaunchConfigurationCofing_withRootBlockDeviceCopiedAMI(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccLaunchConfigurationConfig_withRootBlockDeviceVolumeSize(rName string, volumeSize int) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = %[2]d
  }
}
`, rName, volumeSize))
}

func testAccLaunchConfigurationConfig_withEncryptedRootBlockDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix                 = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t3.nano"
  associate_public_ip_address = false

  root_block_device {
    encrypted   = true
    volume_type = "gp2"
    volume_size = 11
  }
}
`, rName))
}

func testAccLaunchConfigurationWithSpotPriceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  spot_price    = "0.05"
}
`, rName))
}

func testAccLaunchConfigurationMetadataOptionsConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccLaunchConfigurationWithEncryption(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.micro"
  associate_public_ip_address = false

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
    encrypted   = true
  }
}
`, rName))
}

func testAccLaunchConfigurationWithGP3(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.micro"
  associate_public_ip_address = false

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

func testAccLaunchConfigurationWithEncryptionUpdated(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.micro"
  associate_public_ip_address = false

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 10
    encrypted   = true
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_withVPCClassicLink(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block         = "10.0.0.0/16"
  enable_classiclink = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  vpc_classic_link_id              = aws_vpc.test.id
  vpc_classic_link_security_groups = [aws_security_group.test.id]
}
`, rName))
}

func testAccLaunchConfigurationConfig_withIAMProfile(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
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
  image_id             = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = "t2.nano"
  iam_instance_profile = aws_iam_instance_profile.test.name
}
`, rName))
}

func testAccLaunchConfigurationEBSNoDeviceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m1.small"

  ebs_block_device {
    device_name = "/dev/sda2"
    no_device   = true
  }
}
`, rName))
}

func testAccLaunchConfigurationConfig_userData(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix                 = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.micro"
  user_data                   = "foo:-with-character's"
  associate_public_ip_address = false
}
`, rName))
}

func testAccLaunchConfigurationConfig_userDataBase64(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix                 = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.micro"
  user_data_base64            = base64encode("hello world")
  associate_public_ip_address = false
}
`, rName))
}
