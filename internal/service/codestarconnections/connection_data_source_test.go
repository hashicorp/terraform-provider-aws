// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codestarconnections_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeStarConnectionsConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codestarconnections_connection.test_arn"
	dataSourceName2 := "data.aws_codestarconnections_connection.test_name"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "provider_type", dataSourceName, "provider_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "connection_status", dataSourceName, "connection_status"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName2, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName2, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "provider_type", dataSourceName2, "provider_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName2, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "connection_status", dataSourceName2, "connection_status"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName2, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccCodeStarConnectionsConnectionDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codestarconnections_connection.test"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeStarConnectionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarConnectionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccConnectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}

data "aws_codestarconnections_connection" "test_arn" {
  arn = aws_codestarconnections_connection.test.arn
}

data "aws_codestarconnections_connection" "test_name" {
  name = aws_codestarconnections_connection.test.name
}
`, rName)
}

func testAccConnectionDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    "key1" = "value1"
    "key2" = "value2"
  }
}

data "aws_codestarconnections_connection" "test" {
  arn = aws_codestarconnections_connection.test.arn
}
`, rName)
}
