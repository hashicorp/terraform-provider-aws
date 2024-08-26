// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQLDBLedgerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qldb_ledger.test"
	datasourceName := "data.aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.QLDBEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDeletionProtection, resourceName, names.AttrDeletionProtection),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrKMSKey, resourceName, names.AttrKMSKey),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions_mode", resourceName, "permissions_mode"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccLedgerDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  permissions_mode    = "STANDARD"
  deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

data "aws_qldb_ledger" "test" {
  name = aws_qldb_ledger.test.id
}
`, rName)
}
