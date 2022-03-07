package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2LaunchTemplateDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "hibernation_options", dataSourceName, "hibernation_options"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_ID_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_BasicID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "hibernation_options", dataSourceName, "hibernation_options"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_Filter_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateBasicFilterDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
					resource.TestCheckResourceAttrPair(resourceName, "placement", dataSourceName, "placement"),
					resource.TestCheckResourceAttrPair(resourceName, "license_specification", dataSourceName, "license_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring", dataSourceName, "monitoring"),
					resource.TestCheckResourceAttrPair(resourceName, "network_interfaces", dataSourceName, "network_interfaces"),
					resource.TestCheckResourceAttrPair(resourceName, "ram_disk_id", dataSourceName, "ram_disk_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_security_group_ids", dataSourceName, "vpc_security_group_ids"),
					resource.TestCheckResourceAttrPair(resourceName, "tag_specifications", dataSourceName, "tag_specifications"),
					resource.TestCheckResourceAttrPair(resourceName, "user_data", dataSourceName, "user_data"),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", dataSourceName, "key_name"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", dataSourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_market_options", dataSourceName, "instance_market_options"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_initiated_shutdown_behavior", dataSourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttrPair(resourceName, "image_id", dataSourceName, "image_id"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_instance_profile", dataSourceName, "iam_instance_profile"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_inference_accelerator", dataSourceName, "elastic_inference_accelerator"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_gpu_specifications", dataSourceName, "elastic_gpu_specifications"),
					resource.TestCheckResourceAttrPair(resourceName, "ebs_optimized", dataSourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(resourceName, "disable_api_termination", dataSourceName, "disable_api_termination"),
					resource.TestCheckResourceAttrPair(resourceName, "credit_specification", dataSourceName, "credit_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_reservation_specification", dataSourceName, "capacity_reservation_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "block_device_mappings", dataSourceName, "block_device_mappings")),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_Filter_tags(t *testing.T) {
	rInt := sdkacctest.RandInt()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateFilterTagsDataSourceConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_metadataOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_protocol_ipv6", resourceName, "metadata_options.0.http_protocol_ipv6"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.instance_metadata_tags", resourceName, "metadata_options.0.instance_metadata_tags"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_enclaveOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_enclaveOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "enclave_options.#", resourceName, "enclave_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enclave_options.0.enabled", resourceName, "enclave_options.0.enabled"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_associatePublicIPAddress(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_associatePublicIPAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
			{
				Config: testAccLaunchTemplateDataSourceConfig_associatePublicIPAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
			{
				Config: testAccLaunchTemplateDataSourceConfig_associatePublicIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_associateCarrierIPAddress(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceConfig_associateCarrierIPAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_carrier_ip_address", resourceName, "network_interfaces.0.associate_carrier_ip_address"),
				),
			},
			{
				Config: testAccLaunchTemplateDataSourceConfig_associateCarrierIPAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_carrier_ip_address", resourceName, "network_interfaces.0.associate_carrier_ip_address"),
				),
			},
			{
				Config: testAccLaunchTemplateDataSourceConfig_associateCarrierIPAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_carrier_ip_address", resourceName, "network_interfaces.0.associate_carrier_ip_address"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_NetworkInterfaces_deleteOnTermination(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateNetworkInterfacesDeleteOnTerminationDataSourceConfig(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
			{
				Config: testAccLaunchTemplateNetworkInterfacesDeleteOnTerminationDataSourceConfig(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
			{
				Config: testAccLaunchTemplateNetworkInterfacesDeleteOnTerminationDataSourceConfig(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_nonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccLaunchTemplateDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func testAccLaunchTemplateDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName)
}

func testAccLaunchTemplateDataSourceConfig_BasicID(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  id = aws_launch_template.test.id
}
`, rName)
}

func testAccLaunchTemplateBasicFilterDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  filter {
    name   = "launch-template-name"
    values = [aws_launch_template.test.name]
  }
}
`, rName)
}

func testAccLaunchTemplateFilterTagsDataSourceConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    Name     = "key1"
    TestSeed = "%[2]d"
  }
}

data "aws_launch_template" "test" {
  tags = {
    Name     = aws_launch_template.test.tags["Name"]
    TestSeed = "%[2]d"
  }
}
`, rName, rInt)
}

func testAccLaunchTemplateDataSourceConfig_metadataOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    instance_metadata_tags      = "enabled"
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName)
}

func testAccLaunchTemplateDataSourceConfig_enclaveOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  enclave_options {
    enabled = true
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName)
}

func testAccLaunchTemplateDataSourceConfig_associatePublicIPAddress(rName, associatePublicIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    associate_public_ip_address = %[2]s
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName, associatePublicIPAddress)
}

func testAccLaunchTemplateDataSourceConfig_associateCarrierIPAddress(rName, associateCarrierIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    associate_carrier_ip_address = %[2]s
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName, associateCarrierIPAddress)
}

func testAccLaunchTemplateNetworkInterfacesDeleteOnTerminationDataSourceConfig(rName, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    delete_on_termination = %[2]s
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName, deleteOnTermination)
}

const testAccLaunchTemplateDataSourceConfig_NonExistent = `
data "aws_launch_template" "test" {
  name = "tf-acc-test-nonexistent"
}
`
