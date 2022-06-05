package ec2_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2AMI_basic(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tpm_support", ""),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccEC2AMI_deprecateAt(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	deprecateAt := "2027-10-15T13:17:00.000Z"
	deprecateAtUpdated := "2028-10-15T13:17:00.000Z"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_deprecateAt(rName, deprecateAt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "deprecation_time", deprecateAt),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
			{
				Config: testAccAMIConfig_deprecateAt(rName, deprecateAtUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "deprecation_time", deprecateAtUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccEC2AMI_description(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	desc := sdkacctest.RandomWithPrefix("desc")
	descUpdated := sdkacctest.RandomWithPrefix("desc-updated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_desc(rName, desc),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
			{
				Config: testAccAMIConfig_desc(rName, descUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descUpdated),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccEC2AMI_disappears(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2AMI_ephemeralBlockDevices(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_ephemeralBlockDevices(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sdb",
						"virtual_name": "ephemeral0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sdc",
						"virtual_name": "ephemeral1",
					}),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccEC2AMI_gp3BlockDevice(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_gp3BlockDevice(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda1",
						"encrypted":             "false",
						"iops":                  "0",
						"throughput":            "0",
						"volume_size":           "8",
						"outpost_arn":           "",
						"volume_type":           "standard",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "false",
						"device_name":           "/dev/sdb",
						"encrypted":             "true",
						"iops":                  "100",
						"throughput":            "500",
						"volume_size":           "10",
						"outpost_arn":           "",
						"volume_type":           "gp3",
					}),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kernel_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "ramdisk_id", ""),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttrPair(resourceName, "root_snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccEC2AMI_tags(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
			{
				Config: testAccAMIConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAMIConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2AMI_outpost(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.outpost_arn", " data.aws_outposts_outpost.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccEC2AMI_boot(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_boot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "boot_mode", "uefi"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func TestAccEC2AMI_tpmSupport(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIConfig_tpmSupport(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "tpm_support", "v2.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"manage_ebs_snapshots",
				},
			},
		},
	})
}

func testAccCheckAMIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for n, rs := range s.RootModule().Resources {
		// The configuration may contain aws_ami data sources.
		// Ignore them.
		if strings.HasPrefix(n, "data.") {
			continue
		}

		if rs.Type != "aws_ami" {
			continue
		}

		_, err := tfec2.FindImageByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 AMI %s still exists", rs.Primary.ID)
	}

	// Check for managed EBS snapshots.
	return testAccCheckEBSSnapshotDestroy(s)
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

func testAccAMIConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAMIConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName))
}

func testAccAMIConfig_deprecateAt(rName, deprecateAt string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"
  deprecation_time    = %[2]q

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName, deprecateAt))
}

func testAccAMIConfig_desc(rName, desc string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"
  description         = %[2]q

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName, desc))
}

func testAccAMIConfig_ephemeralBlockDevices(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  ephemeral_block_device {
    device_name  = "/dev/sdb"
    virtual_name = "ephemeral0"
  }

  ephemeral_block_device {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral1"
  }
}
`, rName))
}

func testAccAMIConfig_gp3BlockDevice(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = false
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  ebs_block_device {
    delete_on_termination = false
    device_name           = "/dev/sdb"
    encrypted             = true
    iops                  = 100
    throughput            = 500
    volume_size           = 10
    volume_type           = "gp3"
  }
}
`, rName))
}

func testAccAMIConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAMIConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAMIConfig_outpost(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
    outpost_arn = data.aws_outposts_outpost.test.arn
  }
}
`, rName))
}

func testAccAMIConfig_boot(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"
  boot_mode           = "uefi"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName))
}

func testAccAMIConfig_tpmSupport(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = %[1]q
  root_device_name    = "/dev/xvda"
  virtualization_type = "hvm"
  boot_mode           = "uefi"
  tpm_support         = "v2.0"

  ebs_block_device {
    device_name = "/dev/xvda"
    snapshot_id = aws_ebs_snapshot.test.id
  }
}
`, rName))
}
