// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldProtectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	dataSourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionDataSource_basicByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, protectionResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "protection_id", protectionResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, protectionResourceName, names.AttrResourceARN),
					acctest.MatchResourceAttrGlobalARN(dataSourceName, "protection_arn", "shield", regexache.MustCompile(`protection/.+$`)),
				),
			},
			{
				Config: testAccProtectionDataSource_basicById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, protectionResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "protection_id", protectionResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, protectionResourceName, names.AttrResourceARN),
					acctest.MatchResourceAttrGlobalARN(dataSourceName, "protection_arn", "shield", regexache.MustCompile(`protection/.+$`)),
				),
			},
		},
	})
}

func testAccProtectionDataSourceConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = true
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_globalaccelerator_accelerator.test.id
}
`, rName)
}

func testAccProtectionDataSource_basicByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfigBase(rName),
		`
data "aws_shield_protection" "test" {
  resource_arn = aws_shield_protection.test.resource_arn
}
`)
}

func testAccProtectionDataSource_basicById(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfigBase(rName),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}
