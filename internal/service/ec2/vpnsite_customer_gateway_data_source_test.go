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

func TestAccSiteVPNCustomerGatewayDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	asn := sdkacctest.RandIntRange(64512, 65534)
	hostOctet := sdkacctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayDataSourceConfig_filter(rName, asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDeviceName, dataSourceName, names.AttrDeviceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIPAddress, dataSourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
				),
			},
		},
	})
}

func TestAccSiteVPNCustomerGatewayDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	asn := sdkacctest.RandIntRange(64512, 65534)
	hostOctet := sdkacctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayDataSourceConfig_id(rName, asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDeviceName, dataSourceName, names.AttrDeviceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIPAddress, dataSourceName, names.AttrIPAddress),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
				),
			},
		},
	})
}

func testAccSiteVPNCustomerGatewayDataSourceConfig_filter(rName string, asn, hostOctet int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "50.0.0.%[3]d"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

data "aws_customer_gateway" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_customer_gateway.test.tags.Name]
  }
}
`, rName, asn, hostOctet)
}

func testAccSiteVPNCustomerGatewayDataSourceConfig_id(rName string, asn, hostOctet int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %[2]d
  ip_address  = "50.0.0.%[3]d"
  device_name = "test"
  type        = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

data "aws_customer_gateway" "test" {
  id = aws_customer_gateway.test.id
}
`, rName, asn, hostOctet)
}
