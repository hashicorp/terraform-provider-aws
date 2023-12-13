// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EIPDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_publicIP(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_publicIP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_networkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_dns", resourceName, "private_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_instance(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_id", resourceName, "instance"),
					resource.TestCheckResourceAttrPair(dataSourceName, "association_id", resourceName, "association_id"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_carrierIP(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWavelengthZoneAvailable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_carrierIP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "carrier_ip", resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_customerOwnedIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDataSourceConfig_customerOwnedIPv4Pool(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ipv4_pool", dataSourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ip", dataSourceName, "customer_owned_ip"),
				),
			},
		},
	})
}

func testAccEIPDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_eip" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_eip.test.tags.Name]
  }
}
`, rName)
}

func testAccEIPDataSourceConfig_id(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`, rName)
}

func testAccEIPDataSourceConfig_publicIP(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_eip" "test" {
  public_ip = aws_eip.test.public_ip
}
`, rName)
}

func testAccEIPDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_eip" "test" {
  tags = {
    Name = aws_eip.test.tags["Name"]
  }
}
`, rName)
}

func testAccEIPDataSourceConfig_networkInterface(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_networkInterface(rName), `
data "aws_eip" "test" {
  filter {
    name   = "network-interface-id"
    values = [aws_eip.test.network_interface]
  }
}
`)
}

func testAccEIPDataSourceConfig_instance(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_instance(rName), `
data "aws_eip" "test" {
  filter {
    name   = "instance-id"
    values = [aws_eip.test.instance]
  }
}
`)
}

func testAccEIPDataSourceConfig_carrierIP(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_carrierIP(rName), `
data "aws_eip" "test" {
  id = aws_eip.test.id
}
`)
}

func testAccEIPDataSourceConfig_customerOwnedIPv4Pool(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_customerOwnedIPv4Pool(rName), `
data "aws_eip" "test" {
  id = aws_eip.test.id
}
`)
}
