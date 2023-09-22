// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFSxONTAPStorageVirtualMachineDataSource_Id(t *testing.T) {
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
				Config: testAccONTAPStorageVirtualMachineDataSourceConfig_Id(rName),
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

func TestAccFSxONTAPStorageVirtualMachineDataSource_Filter(t *testing.T) {
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
				Config: testAccONTAPStorageVirtualMachineDataSourceConfig_Filter(rName),
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

func testAccONTAPStorageVirtualMachineDataSourceConfig_Id(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_basic(rName), `
data "aws_fsx_ontap_storage_virtual_machine" "test" {
  id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`)
}

func testAccONTAPStorageVirtualMachineDataSourceConfig_Filter(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_basic(rName), `
data "aws_fsx_ontap_storage_virtual_machine" "test" {
  filter {
    name   = "file-system-id"
    values = [aws_fsx_ontap_file_system.test.id]
  }

  depends_on = [aws_fsx_ontap_storage_virtual_machine.test]
}
`)
}
