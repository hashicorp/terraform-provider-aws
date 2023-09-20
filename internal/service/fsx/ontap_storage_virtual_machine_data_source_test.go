// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFSxOntapStorageVirtualMachineDataSource_Id(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	dataSourceName := "data.aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFSxOntapStorageVirtualMachineDataSourceConfig_Id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "active_directory_configuration.#", resourceName, "active_directory_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subtype", resourceName, "subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "uuid", resourceName, "uuid"),
				),
			},
		},
	})
}

func TestAccFSxOntapStorageVirtualMachineDataSource_Filter(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_fsx_ontap_storage_virtual_machine.test"
	dataSourceName := "data.aws_fsx_ontap_storage_virtual_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFSxOntapStorageVirtualMachineDataSourceConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "active_directory_configuration.#", resourceName, "active_directory_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subtype", resourceName, "subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "uuid", resourceName, "uuid"),
				),
			},
		},
	})
}

func testAccOntapStorageVirtualMachineDataSourceBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFSxOntapStorageVirtualMachineDataSourceConfig_Id(rName string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineDataSourceBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
	file_system_id = aws_fsx_ontap_file_system.test.id
	name           = %[1]q
}

data "aws_fsx_ontap_storage_virtual_machine" "test" {
  id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccFSxOntapStorageVirtualMachineDataSourceConfig_Filter(rName string) string {
	return acctest.ConfigCompose(testAccOntapStorageVirtualMachineDataSourceBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
	file_system_id = aws_fsx_ontap_file_system.test.id
	name           = %[1]q
}

data "aws_fsx_ontap_storage_virtual_machine" "test" {
  filter {
		name = "file-system-id"
		values = [aws_fsx_ontap_file_system.test.id]
	}

	depends_on = [
		aws_fsx_ontap_storage_virtual_machine.test
	]
}
`, rName))
}
