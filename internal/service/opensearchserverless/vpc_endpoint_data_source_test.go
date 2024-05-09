// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessVPCEndpointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_vpc_endpoint.test"
	dataSourceName := "data.aws_opensearchserverless_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccVPCEndpointDataSourceConfig_networkingBase(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func testAccVPCEndpointDataSourceConfig_securityGroupBase(rName string, sgCount int) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count  = %[2]d
  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, sgCount),
	)
}

func testAccVPCEndpointDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointDataSourceConfig_networkingBase(rName, 2),
		testAccVPCEndpointDataSourceConfig_securityGroupBase(rName, 2),
		fmt.Sprintf(`
resource "aws_opensearchserverless_vpc_endpoint" "test" {
  name               = %[1]q
  security_group_ids = aws_security_group.test[*].id
  subnet_ids         = aws_subnet.test[*].id
  vpc_id             = aws_vpc.test.id
}

data "aws_opensearchserverless_vpc_endpoint" "test" {
  vpc_endpoint_id = aws_opensearchserverless_vpc_endpoint.test.id
}
`, rName))
}
