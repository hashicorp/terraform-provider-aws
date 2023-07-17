// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2InstanceConnectEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceConnectEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`instance-connect-endpoint/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "fips_dns_name"),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(resourceName, "network_interface_ids.#", 1),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func TestAccEC2InstanceConnectEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceConnectEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceInstanceConnectEndpoint, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2InstanceConnectEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceConnectEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConnectEndpointConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInstanceConnectEndpointConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2InstanceConnectEndpoint_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_connect_endpoint.test"
	securityGroup1ResourceName := "aws_security_group.test.0"
	securityGroup2ResourceName := "aws_security_group.test.1"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceConnectEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConnectEndpointConfig_securityGroupIDs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceConnectEndpointExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`instance-connect-endpoint/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "fips_dns_name"),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(resourceName, "network_interface_ids.#", 1),
					resource.TestCheckResourceAttr(resourceName, "preserve_client_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func testAccCheckInstanceConnectEndpointExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance Connect Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindInstanceConnectEndpointByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckInstanceConnectEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_instance_connect_endpoint" {
				continue
			}

			_, err := tfec2.FindInstanceConnectEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Instance Connect Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceConnectEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), `
resource "aws_ec2_instance_connect_endpoint" "test" {
  subnet_id = aws_subnet.test[0].id
}
`)
}

func testAccInstanceConnectEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_instance_connect_endpoint" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccInstanceConnectEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_instance_connect_endpoint" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInstanceConnectEndpointConfig_securityGroupIDs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_instance_connect_endpoint" "test" {
  preserve_client_ip = false
  subnet_id          = aws_subnet.test[0].id
  security_group_ids = aws_security_group.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
