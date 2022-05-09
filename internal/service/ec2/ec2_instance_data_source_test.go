package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2InstanceDataSource_basic(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_tags(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_azUserData(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceAzUserDataConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_gp2IopsDevice(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGp2IopsDeviceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_size", resourceName, "root_block_device.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_type", resourceName, "root_block_device.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.device_name", resourceName, "root_block_device.0.device_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.iops", resourceName, "root_block_device.0.iops"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_gp3ThroughputDevice(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGp3ThroughputDeviceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_size", resourceName, "root_block_device.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_type", resourceName, "root_block_device.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.device_name", resourceName, "root_block_device.0.device_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.throughput", resourceName, "root_block_device.0.throughput"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_blockDevices(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceBlockDevicesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_size", resourceName, "root_block_device.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.volume_type", resourceName, "root_block_device.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.0.device_name", resourceName, "root_block_device.0.device_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
					//resource.TestCheckResourceAttrPair(datasourceName, "ephemeral_block_device.#", resourceName, "ephemeral_block_device.#"),
					// ephemeral block devices don't get saved properly due to API limitations, so this can't actually be tested right now
				),
			},
		},
	})
}

// Test to verify that ebs_block_device kms_key_id does not elicit a panic
func TestAccEC2InstanceDataSource_EBSBlockDevice_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceEbsBlockDeviceKmsKeyIdConfig(rName),
			},
		},
	})
}

// Test to verify that root_block_device kms_key_id does not elicit a panic
func TestAccEC2InstanceDataSource_RootBlockDevice_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceRootBlockDeviceKmsKeyIdConfig(rName),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_rootInstanceStore(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceRootInstanceStoreConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_optimized", resourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_privateIP(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourcePrivateIPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ip", resourceName, "private_ip"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_secondaryPrivateIPs(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceSecondaryPrivateIPsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "secondary_private_ips", resourceName, "secondary_private_ips"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_ipv6Addresses(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceIpv6AddressesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ipv6_addresses.#", resourceName, "ipv6_address_count"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_keyPair(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceKeyPairConfig(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_name", resourceName, "key_name"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_vpc(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "associate_public_ip_address", resourceName, "associate_public_ip_address"),
					resource.TestCheckResourceAttrPair(datasourceName, "tenancy", resourceName, "tenancy"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_placementGroup(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourcePlacementGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "placement_group", resourceName, "placement_group"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_securityGroups(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceSecurityGroupsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_security_group_ids.#", resourceName, "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_vpcSecurityGroups(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceVPCSecurityGroupsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_security_group_ids.#", resourceName, "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetPasswordData_trueToFalse(t *testing.T) {
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGetPasswordDataConfig(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(datasourceName, "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceGetPasswordDataConfig(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", "false"),
					resource.TestCheckNoResourceAttr(datasourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetPasswordData_falseToTrue(t *testing.T) {
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGetPasswordDataConfig(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", "false"),
					resource.TestCheckNoResourceAttr(datasourceName, "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceGetPasswordDataConfig(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(datasourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_getUserData(t *testing.T) {
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGetUserDataConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "true"),
					resource.TestCheckResourceAttr(datasourceName, "user_data_base64", "IyEvYmluL2Jhc2gKCmVjaG8gImhlbGxvIHdvcmxkIgo="),
				),
			},
			{
				Config: testAccInstanceDataSourceGetUserDataConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "false"),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceGetUserDataConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "true"),
					resource.TestCheckResourceAttr(datasourceName, "user_data_base64", "IyEvYmluL2Jhc2gKCmVjaG8gImhlbGxvIHdvcmxkIgo="),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetUserData_noUserData(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceGetUserDataNoUserDataConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "true"),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceGetUserDataNoUserDataConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "false"),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceGetUserDataNoUserDataConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", "true"),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_autoRecovery(t *testing.T) {
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceAutoRecoveryConfig(rName, "default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.0.auto_recovery", "default"),
				),
			},
			{
				Config: testAccInstanceDataSourceAutoRecoveryConfig(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.0.auto_recovery", "disabled"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_creditSpecification(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{

				Config: testAccInstanceDataSourceCreditSpecificationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "credit_specification.#", resourceName, "credit_specification.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "credit_specification.0.cpu_credits", resourceName, "credit_specification.0.cpu_credits"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_metadataOptions(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceMetadataOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.instance_metadata_tags", resourceName, "metadata_options.0.instance_metadata_tags"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_enclaveOptions(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceEnclaveOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "enclave_options.#", resourceName, "enclave_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "enclave_options.0.enabled", resourceName, "enclave_options.0.enabled"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_blockDeviceTags(t *testing.T) {
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceBlockDeviceTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
				),
			},
		},
	})
}

// Lookup based on InstanceID
func testAccInstanceDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  filter {
    name   = "instance-id"
    values = [aws_instance.test.id]
  }
}
`, rName))
}

// Use the tags attribute to filter
func testAccInstanceDataSourceTagsConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"

  tags = {
    Name     = %[1]q
    TestSeed = "%[2]d"
  }
}

data "aws_instance" "test" {
  instance_tags = {
    Name     = aws_instance.test.tags["Name"]
    TestSeed = "%[2]d"
  }
}
`, rName, sdkacctest.RandInt()))
}

// filter on tag, populate more attributes
func testAccInstanceDataSourceAzUserDataConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  availability_zone = data.aws_availability_zones.available.names[0]

  instance_type = "t2.micro"
  user_data     = "test:-with-character's"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

// GP2IopsDevice
func testAccInstanceDataSourceGp2IopsDeviceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

// GP3ThroughputDevice
func testAccInstanceDataSourceGp3ThroughputDeviceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

  root_block_device {
    volume_type = "gp3"
    volume_size = 10
    throughput  = 300
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

// Block Device
func testAccInstanceDataSourceBlockDevicesConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

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

  # Encrypted ebs block device
  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted   = true
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  ebs_block_device {
    device_name = "/dev/sdf"
    volume_size = 10
    volume_type = "gp3"
    throughput  = 300
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceEbsBlockDeviceKmsKeyIdConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    encrypted   = true
    kms_key_id  = aws_kms_key.test.arn
    volume_size = 9
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceRootBlockDeviceKmsKeyIdConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

  root_block_device {
    encrypted   = true
    kms_key_id  = aws_kms_key.test.arn
    volume_type = "gp2"
    volume_size = 11
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceRootInstanceStoreConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.medium"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourcePrivateIPConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
  private_ip    = "10.1.1.42"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceSecondaryPrivateIPsConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                   = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type         = "t2.micro"
  subnet_id             = aws_subnet.test.id
  secondary_private_ips = ["10.1.1.42"]

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceIpv6AddressesConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type      = "t2.micro"
  subnet_id          = aws_subnet.test.id
  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceKeyPairConfig(rName, publicKey string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  key_name      = aws_key_pair.test.key_name

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  filter {
    name   = "tag:Name"
    values = [%[1]q]
  }

  filter {
    name   = "key-name"
    values = [aws_instance.test.key_name]
  }
}
`, rName, publicKey))
}

func testAccInstanceDataSourceVPCConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "t2.small"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  #tenancy                     = "dedicated"
  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourcePlacementGroupConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "c5.large"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  placement_group             = aws_placement_group.test.name

  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceSecurityGroupsConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  ingress {
    protocol  = "icmp"
    from_port = -1
    to_port   = -1
    self      = true
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami             = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.small"
  security_groups = [aws_security_group.test.name]
  user_data       = "foo:-with-character's"

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceVPCSecurityGroupsConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id
  depends_on             = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceGetPasswordDataConfig(rName, publicKey string, val bool) string {
	return acctest.ConfigCompose(testAccLatestWindowsServer2016CoreAMIConfig(), fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.win2016core-ami.id
  instance_type = "t2.medium"
  key_name      = aws_key_pair.test.key_name

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id

  get_password_data = %[3]t
}
`, rName, publicKey, val))
}

func testAccInstanceDataSourceGetUserDataConfig(rName string, getUserData bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  user_data = <<EUD
#!/bin/bash

echo "hello world"
EUD

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  get_user_data = %[2]t
  instance_id   = aws_instance.test.id
}
`, rName, getUserData))
}

func testAccInstanceDataSourceGetUserDataNoUserDataConfig(rName string, getUserData bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  get_user_data = %[2]t
  instance_id   = aws_instance.test.id
}
`, rName, getUserData))
}

func testAccInstanceDataSourceAutoRecoveryConfig(rName string, val string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  maintenance_options {
    auto_recovery = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName, val))
}

func testAccInstanceDataSourceCreditSpecificationConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "unlimited"
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceMetadataOptionsConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    instance_metadata_tags      = "enabled"
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceEnclaveOptionsConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("c5a.xlarge", "c5.xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  enclave_options {
    enabled = true
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceDataSourceBlockDeviceTagsConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }

  ebs_block_device {
    device_name = "/dev/xvdc"
    volume_size = 10

    tags = {
      Name   = %[1]q
      Factum = "SapereAude"
    }
  }

  root_block_device {
    tags = {
      Name   = %[1]q
      Factum = "VincitQuiSeVincit"
    }
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName))
}
