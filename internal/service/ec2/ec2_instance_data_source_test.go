// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

func TestAccEC2InstanceDataSource_instanceTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_instanceTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":     knownvalue.StringExact(rName),
						"TestSeed": knownvalue.NotNull(),
					})),
				},
			},
		},
	})
}

func TestAccEC2InstanceDataSource_azUserData(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_azUser(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

func TestAccEC2InstanceDataSource_gp2IopsDevice(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_gp2IOPSDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_gp3ThroughputDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
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
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_blockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_ebsKMSKeyID(rName),
			},
		},
	})
}

// Test to verify that root_block_device kms_key_id does not elicit a panic
func TestAccEC2InstanceDataSource_RootBlockDevice_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_rootBlockDeviceKMSKeyID(rName),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_rootInstanceStore(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_rootStore(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_optimized", resourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_privateIP(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_privateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name_options.#", resourceName, "private_dns_name_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name_options.0.hostname_type", resourceName, "private_dns_name_options.0.hostname_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ip", resourceName, "private_ip"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_secondaryPrivateIPs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_secondaryPrivateIPs(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "secondary_private_ips", resourceName, "secondary_private_ips"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_ipv6Addresses(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_ipv6Addresses(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "ipv6_addresses.#", resourceName, "ipv6_address_count"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_keyPair(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_keyPair(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_name", resourceName, "key_name"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

func TestAccEC2InstanceDataSource_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "associate_public_ip_address", resourceName, "associate_public_ip_address"),
					resource.TestCheckResourceAttrPair(datasourceName, "tenancy", resourceName, "tenancy"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_placementGroup(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_placementGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "placement_group", resourceName, "placement_group"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_securityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_securityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data", resourceName, "user_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_security_group_ids.#", resourceName, "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_vpcSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_vpcSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_security_group_ids.#", resourceName, "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetPasswordData_trueToFalse(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getPassword(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(datasourceName, "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getPassword(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(datasourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetPasswordData_falseToTrue(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getPassword(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(datasourceName, "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getPassword(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_password_data", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(datasourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_getUserData(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getUser(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtTrue),
					resource.TestCheckResourceAttr(datasourceName, "user_data_base64", "IyEvYmluL2Jhc2gKCmVjaG8gImhlbGxvIHdvcmxkIgo="),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getUser(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getUser(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtTrue),
					resource.TestCheckResourceAttr(datasourceName, "user_data_base64", "IyEvYmluL2Jhc2gKCmVjaG8gImhlbGxvIHdvcmxkIgo="),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_GetUserData_noUserData(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getUserNoUser(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getUserNoUser(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getUserNoUser(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "get_user_data", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(datasourceName, "user_data_base64"),
					resource.TestCheckResourceAttrPair(datasourceName, "user_data_base64", resourceName, "user_data_base64"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_autoRecovery(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_autoRecovery(rName, "default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.0.auto_recovery", "default"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_autoRecovery(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_options.0.auto_recovery", "disabled"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_creditSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{

				Config: testAccInstanceDataSourceConfig_creditSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "credit_specification.#", resourceName, "credit_specification.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "credit_specification.0.cpu_credits", resourceName, "credit_specification.0.cpu_credits"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_metadataOptions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_metaOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_protocol_ipv6", resourceName, "metadata_options.0.http_protocol_ipv6"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.0.instance_metadata_tags", resourceName, "metadata_options.0.instance_metadata_tags"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_enclaveOptions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_enclaveOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "enclave_options.#", resourceName, "enclave_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "enclave_options.0.enabled", resourceName, "enclave_options.0.enabled"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_blockDeviceTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_blockDeviceTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_disableAPIStopTermination(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_disableAPIStopTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "disable_api_stop", acctest.CtTrue),
					resource.TestCheckResourceAttr(datasourceName, "disable_api_termination", acctest.CtTrue),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_disableAPIStopTermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "disable_api_stop", acctest.CtFalse),
					resource.TestCheckResourceAttr(datasourceName, "disable_api_termination", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEC2InstanceDataSource_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_instance.test"
	datasourceName := "data.aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ami", resourceName, "ami"),
					resource.TestCheckResourceAttrSet(datasourceName, "launch_time"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

// Lookup based on InstanceID
func testAccInstanceDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
func testAccInstanceDataSourceConfig_instanceTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
func testAccInstanceDataSourceConfig_azUser(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
func testAccInstanceDataSourceConfig_gp2IOPSDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
func testAccInstanceDataSourceConfig_gp3ThroughputDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
func testAccInstanceDataSourceConfig_blockDevices(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_ebsKMSKeyID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_rootBlockDeviceKMSKeyID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_rootStore(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_privateIP(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_secondaryPrivateIPs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                   = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_ipv6Addresses(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_keyPair(rName, publicKey string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_vpc(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_placementGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_securityGroups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
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
  ami             = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_vpcSecurityGroups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_getPassword(rName, publicKey string, val bool) string {
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

func testAccInstanceDataSourceConfig_getUser(rName string, getUserData bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_getUserNoUser(rName string, getUserData bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_autoRecovery(rName string, val string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_creditSpecification(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_metaOptions(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_enclaveOptions(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("c5a.xlarge", "c5.xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_blockDeviceTags(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccInstanceDataSourceConfig_disableAPIStopTermination(rName string, disableApiStopTermination bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  disable_api_stop        = %[2]t
  disable_api_termination = %[2]t
  instance_type           = "t2.micro"
  subnet_id               = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_instance" "test" {
  instance_id = aws_instance.test.id
}
`, rName, disableApiStopTermination))
}

func testAccInstanceDataSourceConfig_timeout(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

  timeouts {
    read = "60m"
  }
}
`, rName))
}
