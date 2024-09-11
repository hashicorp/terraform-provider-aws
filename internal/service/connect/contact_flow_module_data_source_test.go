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

func testAccContactFlowModuleDataSource_contactFlowModuleID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"
	datasourceName := "data.aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleDataSourceConfig_id(rName, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_module_id", resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrContent, resourceName, names.AttrContent),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccContactFlowModuleDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow_module.test"
	datasourceName := "data.aws_connect_contact_flow_module.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowModuleDataSourceConfig_name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_module_id", resourceName, "contact_flow_module_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrContent, resourceName, names.AttrContent),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccContactFlowModuleBaseDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_contact_flow_module" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Contact Flow Module Description"
  content     = file("./test-fixtures/connect_contact_flow_module.json")

  tags = {
    "Name"        = "Test Contact Flow Module",
    "Application" = "Terraform",
    "Method"      = "Create"
  }
}
    `, rName, rName2)
}

func testAccContactFlowModuleDataSourceConfig_id(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_contact_flow_module" "test" {
  instance_id            = aws_connect_instance.test.id
  contact_flow_module_id = aws_connect_contact_flow_module.test.contact_flow_module_id
}
`)
}

func testAccContactFlowModuleDataSourceConfig_name(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_contact_flow_module" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_contact_flow_module.test.name
}
`)
}
