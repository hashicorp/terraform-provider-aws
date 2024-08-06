// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudcontrol_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudControlResourceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudcontrolapi_resource.test"
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrProperties, resourceName, names.AttrProperties),
					resource.TestCheckResourceAttrPair(dataSourceName, "type_name", resourceName, "type_name"),
				),
			},
		},
	})
}

func testAccResourceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}

data "aws_cloudcontrolapi_resource" "test" {
  identifier = aws_cloudcontrolapi_resource.test.id
  type_name  = aws_cloudcontrolapi_resource.test.type_name
}
`, rName)
}
