// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxONTAPStorageVirtualMachinesDataSource_Filter(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_fsx_ontap_storage_virtual_machines.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPStorageVirtualMachinesDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func testAccONTAPStorageVirtualMachinesDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccONTAPStorageVirtualMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_ontap_storage_virtual_machine" "test" {
  count = 2

  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = "%[1]s-${count.index}"
}

data "aws_fsx_ontap_storage_virtual_machines" "test" {
  filter {
    name   = "file-system-id"
    values = [aws_fsx_ontap_file_system.test.id]
  }

  depends_on = [aws_fsx_ontap_storage_virtual_machine.test]
}
`, rName))
}
