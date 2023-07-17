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
		ErrorCheck:               acctest.ErrorCheck(t, names.QLDBEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLedgerDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "deletion_protection", resourceName, "deletion_protection"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key", resourceName, "kms_key"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions_mode", resourceName, "permissions_mode"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
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
