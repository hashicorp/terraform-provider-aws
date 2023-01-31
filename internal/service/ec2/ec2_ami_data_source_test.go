package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2AMIDataSource_natInstance(t *testing.T) {
	datasourceName := "data.aws_ami.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check attributes. Some attributes are tough to test - any not contained here should not be considered
					// stable and should not be used in interpolation. Exception to block_device_mappings which should both
					// show up consistently and break if certain references are not available. However modification of the
					// snapshot ID which is bound to happen on the NAT AMIs will cause testing to break consistently, so
					// deep inspection is not included, simply the count is checked.
					// Tags and product codes may need more testing, but I'm having a hard time finding images with
					// these attributes set.
					resource.TestCheckResourceAttr(datasourceName, "architecture", "x86_64"),
					acctest.MatchResourceAttrRegionalARNNoAccount(datasourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(datasourceName, "block_device_mappings.#", "1"),
					resource.TestMatchResourceAttr(datasourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(datasourceName, "deprecation_time", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(datasourceName, "description", regexp.MustCompile("^Amazon Linux AMI")),
					resource.TestCheckResourceAttr(datasourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(datasourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(datasourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(datasourceName, "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr(datasourceName, "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(datasourceName, "imds_support", ""),
					resource.TestCheckResourceAttr(datasourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(datasourceName, "name", regexp.MustCompile("^amzn-ami-vpc-nat")),
					acctest.MatchResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttr(datasourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(datasourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "public", "true"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_type", "ebs"),
					resource.TestMatchResourceAttr(datasourceName, "root_snapshot_id", regexp.MustCompile("^snap-")),
					resource.TestCheckResourceAttr(datasourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(datasourceName, "state", "available"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(datasourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(datasourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccEC2AMIDataSource_windowsInstance(t *testing.T) {
	datasourceName := "data.aws_ami.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIDataSourceConfig_windows,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(datasourceName, "block_device_mappings.#", "27"),
					resource.TestMatchResourceAttr(datasourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(datasourceName, "deprecation_time", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(datasourceName, "description", regexp.MustCompile("^Microsoft Windows Server")),
					resource.TestCheckResourceAttr(datasourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(datasourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(datasourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(datasourceName, "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr(datasourceName, "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(datasourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(datasourceName, "name", regexp.MustCompile("^Windows_Server-2012-R2")),
					acctest.MatchResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttr(datasourceName, "platform", "windows"),
					resource.TestMatchResourceAttr(datasourceName, "platform_details", regexp.MustCompile(`Windows`)),
					resource.TestCheckResourceAttr(datasourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "public", "true"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_type", "ebs"),
					resource.TestMatchResourceAttr(datasourceName, "root_snapshot_id", regexp.MustCompile("^snap-")),
					resource.TestCheckResourceAttr(datasourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(datasourceName, "state", "available"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(datasourceName, "tpm_support", ""),
					resource.TestMatchResourceAttr(datasourceName, "usage_operation", regexp.MustCompile(`^RunInstances`)),
					resource.TestCheckResourceAttr(datasourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccEC2AMIDataSource_instanceStore(t *testing.T) {
	datasourceName := "data.aws_ami.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(datasourceName, "block_device_mappings.#", "0"),
					resource.TestMatchResourceAttr(datasourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(datasourceName, "deprecation_time", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestCheckResourceAttr(datasourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(datasourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(datasourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(datasourceName, "image_location", regexp.MustCompile("amzn-ami-minimal-hvm")),
					resource.TestCheckResourceAttr(datasourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(datasourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(datasourceName, "name", regexp.MustCompile("amzn-ami-minimal-hvm")),
					acctest.MatchResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttr(datasourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(datasourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "public", "true"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_type", "instance-store"),
					resource.TestCheckResourceAttr(datasourceName, "root_snapshot_id", ""),
					resource.TestCheckResourceAttr(datasourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(datasourceName, "state", "available"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(datasourceName, "tpm_support", ""),
					resource.TestCheckResourceAttr(datasourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(datasourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccEC2AMIDataSource_localNameFilter(t *testing.T) {
	datasourceName := "data.aws_ami.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIDataSourceConfig_nameRegex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(datasourceName, "image_id", regexp.MustCompile("^ami-")),
				),
			},
		},
	})
}

func TestAccEC2AMIDataSource_gp3BlockDevice(t *testing.T) {
	resourceName := "aws_ami.test"
	datasourceName := "data.aws_ami.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAMIDataSourceConfig_gp3BlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "architecture", resourceName, "architecture"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "block_device_mappings.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "image_id", resourceName, "id"),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_device_name", resourceName, "root_device_name"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_type", "ebs"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_snapshot_id", resourceName, "root_snapshot_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "sriov_net_support", resourceName, "sriov_net_support"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "virtualization_type", resourceName, "virtualization_type"),
				),
			},
		},
	})
}

// testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore returns the configuration for a data source that
// describes the latest Amazon Linux AMI using HVM virtualization and an instance store root device.
// The data source is named 'amzn-ami-minimal-hvm-instance-store'.
func testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore() string {
	return `
data "aws_ami" "test" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["instance-store"]
  }
}
`
}

// Using NAT AMIs for testing - I would expect with NAT gateways now a thing,
// that this will possibly be deprecated at some point in time. Other candidates
// for testing this after that may be Ubuntu's AMI's, or Amazon's regular
// Amazon Linux AMIs.
const testAccAMIDataSourceConfig_basic = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "block-device-mapping.volume-type"
    values = ["standard"]
  }
}
`

// Windows image test.
const testAccAMIDataSourceConfig_windows = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2012-R2*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "block-device-mapping.volume-type"
    values = ["gp2"]
  }
}
`

// Testing name_regex parameter
const testAccAMIDataSourceConfig_nameRegex = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
  }

  name_regex = "^amzn-ami-min[a-z]{4}-hvm"
}
`

func testAccAMIDataSourceConfig_gp3BlockDevice(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIConfig_gp3BlockDevice(rName),
		`
data "aws_caller_identity" "current" {}

data "aws_ami" "test" {
  owners = [data.aws_caller_identity.current.account_id]

  filter {
    name   = "image-id"
    values = [aws_ami.test.id]
  }
}
`)
}
