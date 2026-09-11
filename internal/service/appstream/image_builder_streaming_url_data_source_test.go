// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamImageBuilderStreamingURLDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_appstream_image_builder_streaming_url.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppStreamEndpointID)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderStreamingURLDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttrSet(dataSourceName, "streaming_url"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires"),
				),
			},
		},
	})
}

func testAccImageBuilderStreamingURLDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_image_builder" "test" {
  name          = %[1]q
  instance_type = "stream.standard.small"
  image_name    = "AppStream-WinServer2019-10-05-2022"
}

data "aws_appstream_image_builder_streaming_url" "test" {
  name = aws_appstream_image_builder.test.name
}
`, rName)
}
