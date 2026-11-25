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

func TestAccVPCEndpointsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_vpc_endpoints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "ids.#", 0),
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "vpc_endpoints.#", 0),
				),
			},
		},
	})
}

func TestAccVPCEndpointsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointsDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoints.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.0", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.service_name", resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.vpc_id", resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.state", resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.vpc_endpoint_type", resourceName, "vpc_endpoint_type"),
				),
			},
		},
	})
}

func TestAccVPCEndpointsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointsDataSourceConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoints.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.0", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.tags.%", resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccVPCEndpointsDataSource_serviceName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	datasourceName := "data.aws_vpc_endpoints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointsDataSourceConfig_serviceName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoints.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "ids.0", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_endpoints.0.service_name", resourceName, names.AttrServiceName),
				),
			},
		},
	})
}

func testAccVPCEndpointsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoints" "test" {
  depends_on = [aws_vpc_endpoint.test]
}
`, rName)
}

func testAccVPCEndpointsDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoints" "test" {
  vpc_id = aws_vpc.test.id

  depends_on = [aws_vpc_endpoint.test]
}
`, rName)
}

func testAccVPCEndpointsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
    TestTag = "test-value"
  }
}

data "aws_vpc_endpoints" "test" {
  tags = {
    TestTag = "test-value"
  }

  depends_on = [aws_vpc_endpoint.test]
}
`, rName)
}

func testAccVPCEndpointsDataSourceConfig_serviceName(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "kms" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.kms"
  vpc_endpoint_type = "Interface"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoints" "test" {
  vpc_id = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  depends_on = [aws_vpc_endpoint.test, aws_vpc_endpoint.kms]
}
`, rName)
}
