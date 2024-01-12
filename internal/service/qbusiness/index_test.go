// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccQBusinessIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, qbusiness.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Index name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIndexConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Action": "sts:AssumeRole",
	"Principal": {
		"Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
	},
	"Effect": "Allow",
	"Sid": ""
	}
	]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
  document_attribute_configurations {
    attribute {
        name = "foo1"
        search = "ENABLED"
        type = "STRING"
	}
	attribute {
        name = "foo2"
        search = "ENABLED"
        type = "STRING"
	}
  }
}
`, rName)
}
