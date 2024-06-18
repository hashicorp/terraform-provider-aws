// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInsightsPath_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_basic(rName, "tcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-insights-path/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestination, "aws_network_interface.test.1", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, "aws_network_interface.test.1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "destination_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_network_interface.test.0", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_network_interface.test.0", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "source_ip", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccVPCNetworkInsightsPath_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_basic(rName, "udp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkInsightsPath(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInsightsPath_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkInsightsPathConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNetworkInsightsPathConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNetworkInsightsPath_sourceAndDestinationARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_sourceAndDestinationARN(rName, "tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestination, "aws_network_interface.test.1", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, "aws_network_interface.test.1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSource, "aws_network_interface.test.0", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", "aws_network_interface.test.0", names.AttrARN),
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

func TestAccVPCNetworkInsightsPath_sourceIP(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_sourceIP(rName, "1.1.1.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_ip", "1.1.1.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkInsightsPathConfig_sourceIP(rName, "8.8.8.8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_ip", "8.8.8.8"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInsightsPath_destinationIP(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_destinationIP(rName, "1.1.1.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_ip", "1.1.1.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkInsightsPathConfig_destinationIP(rName, "8.8.8.8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_ip", "8.8.8.8"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInsightsPath_destinationPort(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInsightsPathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathConfig_destinationPort(rName, 80),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_port", "80"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNetworkInsightsPathConfig_destinationPort(rName, 443),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInsightsPathExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_port", "443"),
				),
			},
		},
	})
}

func testAccCheckNetworkInsightsPathExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindNetworkInsightsPathByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckNetworkInsightsPathDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_network_insights_path" {
				continue
			}

			_, err := tfec2.FindNetworkInsightsPathByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network Insights Path %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCNetworkInsightsPathConfig_basic(rName, protocol string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test[0].id
  destination = aws_network_interface.test[1].id
  protocol    = %[2]q
}
`, rName, protocol))
}

func testAccVPCNetworkInsightsPathConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test[0].id
  destination = aws_network_interface.test[1].id
  protocol    = "tcp"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccVPCNetworkInsightsPathConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test[0].id
  destination = aws_network_interface.test[1].id
  protocol    = "tcp"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCNetworkInsightsPathConfig_sourceAndDestinationARN(rName, protocol string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test[0].arn
  destination = aws_network_interface.test[1].arn
  protocol    = %[2]q
}
`, rName, protocol))
}

func testAccVPCNetworkInsightsPathConfig_sourceIP(rName, sourceIP string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_internet_gateway.test.id
  destination = aws_network_interface.test.id
  protocol    = "tcp"
  source_ip   = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, sourceIP))
}

func testAccVPCNetworkInsightsPathConfig_destinationIP(rName, destinationIP string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source         = aws_network_interface.test.id
  destination    = aws_internet_gateway.test.id
  protocol       = "tcp"
  destination_ip = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationIP))
}

func testAccVPCNetworkInsightsPathConfig_destinationPort(rName string, destinationPort int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source           = aws_network_interface.test[0].id
  destination      = aws_network_interface.test[1].id
  protocol         = "tcp"
  destination_port = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationPort))
}
