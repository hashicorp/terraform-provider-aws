// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineDataSourceConfig_Id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "active_directory_configuration.#", resourceName, "active_directory_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "subtype", resourceName, "subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachineDataSourceConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "active_directory_configuration.#", resourceName, "active_directory_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrFileSystemID, resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "subtype", resourceName, "subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "uuid", resourceName, "uuid"),
				),
			},
		},
	})
}

func testAccONTAPStorageVirtualMachineDataSourceConfig_Id(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_fsx_ontap_storage_virtual_machine" "test" {
  id = aws_fsx_ontap_storage_virtual_machine.test.id
}
`, rName))
}

func testAccONTAPStorageVirtualMachineDataSourceConfig_Filter(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_fsx_ontap_storage_virtual_machine" "test" {
  filter {
    name   = "file-system-id"
    values = [aws_fsx_ontap_file_system.test.id]
  }

  depends_on = [aws_fsx_ontap_storage_virtual_machine.test]
}
`, rName))
}
