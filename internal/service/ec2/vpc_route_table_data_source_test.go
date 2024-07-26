// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRouteTableDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rtResourceName := "aws_route_table.test"
	snResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	igwResourceName := "aws_internet_gateway.test"
	datasource1Name := "data.aws_route_table.by_tag"
	datasource2Name := "data.aws_route_table.by_filter"
	datasource3Name := "data.aws_route_table.by_subnet"
	datasource4Name := "data.aws_route_table.by_id"
	datasource5Name := "data.aws_route_table.by_gateway"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					// By tags.
					acctest.MatchResourceAttrRegionalARN(datasource1Name, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrID, rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource1Name, "route_table_id", rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrOwnerID, rtResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(datasource1Name, names.AttrSubnetID),
					resource.TestCheckNoResourceAttr(datasource1Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource1Name, "associations.#", acctest.Ct2),
					testAccCheckListHasSomeElementAttrPair(datasource1Name, "associations", names.AttrSubnetID, snResourceName, names.AttrID),
					testAccCheckListHasSomeElementAttrPair(datasource1Name, "associations", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource1Name, "tags.Name", rName),
					// By filter.
					acctest.MatchResourceAttrRegionalARN(datasource2Name, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrID, rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource2Name, "route_table_id", rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrOwnerID, rtResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(datasource2Name, names.AttrSubnetID),
					resource.TestCheckNoResourceAttr(datasource2Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource2Name, "associations.#", acctest.Ct2),
					testAccCheckListHasSomeElementAttrPair(datasource2Name, "associations", names.AttrSubnetID, snResourceName, names.AttrID),
					testAccCheckListHasSomeElementAttrPair(datasource2Name, "associations", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource2Name, "tags.Name", rName),
					// By subnet ID.
					acctest.MatchResourceAttrRegionalARN(datasource3Name, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrID, rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource3Name, "route_table_id", rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrOwnerID, rtResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrSubnetID, snResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(datasource3Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource3Name, "associations.#", acctest.Ct2),
					testAccCheckListHasSomeElementAttrPair(datasource3Name, "associations", names.AttrSubnetID, snResourceName, names.AttrID),
					testAccCheckListHasSomeElementAttrPair(datasource3Name, "associations", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource3Name, "tags.Name", rName),
					// By route table ID.
					acctest.MatchResourceAttrRegionalARN(datasource4Name, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource4Name, names.AttrID, rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource4Name, "route_table_id", rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource4Name, names.AttrOwnerID, rtResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasource4Name, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(datasource4Name, names.AttrSubnetID),
					resource.TestCheckNoResourceAttr(datasource4Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource4Name, "associations.#", acctest.Ct2),
					testAccCheckListHasSomeElementAttrPair(datasource4Name, "associations", names.AttrSubnetID, snResourceName, names.AttrID),
					testAccCheckListHasSomeElementAttrPair(datasource4Name, "associations", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource4Name, "tags.Name", rName),
					// By gateway ID.
					acctest.MatchResourceAttrRegionalARN(datasource5Name, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource5Name, names.AttrID, rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource5Name, "route_table_id", rtResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource5Name, names.AttrOwnerID, rtResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasource5Name, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(datasource5Name, names.AttrSubnetID),
					resource.TestCheckResourceAttrPair(datasource5Name, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource5Name, "associations.#", acctest.Ct2),
					testAccCheckListHasSomeElementAttrPair(datasource5Name, "associations", names.AttrSubnetID, snResourceName, names.AttrID),
					testAccCheckListHasSomeElementAttrPair(datasource5Name, "associations", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasource5Name, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccVPCRouteTableDataSource_main(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableDataSourceConfig_main(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceName, "associations.0.main", acctest.CtTrue),
				),
			},
		},
	})
}

// testAccCheckListHasSomeElementAttrPair is a TestCheckFunc which validates that the collection on the left has an element with an attribute value
// matching the value on the left
// Based on TestCheckResourceAttrPair from the Terraform SDK testing framework
func testAccCheckListHasSomeElementAttrPair(nameFirst string, resourceAttr string, elementAttr string, nameSecond string, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := acctest.PrimaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := acctest.PrimaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, ok := isSecond.Attributes[keySecond]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", nameSecond, keySecond)
		} else if vSecond == "" {
			return fmt.Errorf("%s: No value was set on attribute %q", nameSecond, keySecond)
		}

		attrsFirst := make([]string, 0, len(isFirst.Attributes))
		attrMatcher := regexache.MustCompile(fmt.Sprintf("%s\\.\\d+\\.%s", resourceAttr, elementAttr))
		for k := range isFirst.Attributes {
			if attrMatcher.MatchString(k) {
				attrsFirst = append(attrsFirst, k)
			}
		}

		found := false
		for _, attrName := range attrsFirst {
			vFirst := isFirst.Attributes[attrName]
			if vFirst == vSecond {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: No element of %q found with attribute %q matching value %q set on %q of %s", nameFirst, resourceAttr, elementAttr, vSecond, keySecond, nameSecond)
		}

		return nil
	}
}

func testAccVPCRouteTableDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "b" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id
}

data "aws_route_table" "by_filter" {
  filter {
    name   = "association.route-table-association-id"
    values = [aws_route_table_association.a.id]
  }

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_tag" {
  tags = {
    Name = aws_route_table.test.tags["Name"]
  }

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_subnet" {
  subnet_id = aws_subnet.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_gateway" {
  gateway_id = aws_internet_gateway.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_id" {
  route_table_id = aws_route_table.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}
`, rName)
}

func testAccVPCRouteTableDataSourceConfig_main(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_route_table" "test" {
  filter {
    name   = "association.main"
    values = ["true"]
  }

  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}
`, rName)
}
