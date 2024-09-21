// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMServiceLinkedRoleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	awsServiceName := "inspector.amazonaws.com"
	dataSourceName := "data.aws_iam_service_linked_role.test"
	resourceName := "aws_iam_service_linked_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleDataSourceConfig_basic(awsServiceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "aws_service_name", awsServiceName),
					testAccCheckServiceLinkedRoleExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "unique_id", resourceName, "unique_id"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccServiceLinkedRoleDataSourceConfig_basic(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = %[1]q
}

data "aws_iam_service_linked_role" "test" {
    aws_service_name = %[1]q
}
`, awsServiceName)
}
