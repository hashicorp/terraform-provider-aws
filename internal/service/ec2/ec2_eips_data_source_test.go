// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EIPsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue("data.aws_eips.all", "allocation_ids.#", 1),
					resource.TestCheckResourceAttr("data.aws_eips.by_tags", "allocation_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_eips.by_tags", "public_ips.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_eips.none", "allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr("data.aws_eips.none", "public_ips.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccEIPsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test1" {
  domain = "vpc"

  tags = {
    Name = "%[1]s-1"
  }
}

resource "aws_eip" "test2" {
  domain = "vpc"

  tags = {
    Name = "%[1]s-2"
  }
}

data "aws_eips" "all" {
  depends_on = [aws_eip.test1, aws_eip.test2]
}

data "aws_eips" "by_tags" {
  tags = {
    Name = "%[1]s-1"
  }

  depends_on = [aws_eip.test1, aws_eip.test2]
}

data "aws_eips" "none" {
  filter {
    name   = "tag-key"
    values = ["%[1]s-3"]
  }

  depends_on = [aws_eip.test1, aws_eip.test2]
}
`, rName)
}
