// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSecurityProfileDataSource_securityProfileID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"
	datasourceName := "data.aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileDataSourceConfig_id(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions.#", resourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_id", resourceName, "security_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccSecurityProfileDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"
	datasourceName := "data.aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileDataSourceConfig_name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions.#", resourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_id", resourceName, "security_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccSecurityProfileBaseDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "test security profile data source"

  permissions = [
    "BasicAgentAccess",
    "OutboundCallAccess",
  ]

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName, rName2)
}

func testAccSecurityProfileDataSourceConfig_id(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_security_profile" "test" {
  instance_id         = aws_connect_instance.test.id
  security_profile_id = aws_connect_security_profile.test.security_profile_id
}
`)
}

func testAccSecurityProfileDataSourceConfig_name(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_security_profile.test.name
}
`)
}
