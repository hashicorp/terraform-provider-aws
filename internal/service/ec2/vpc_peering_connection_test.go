// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCPeeringConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
		},
	})
}

func TestAccVPCPeeringConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCPeeringConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCPeeringConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
			{
				Config: testAccVPCPeeringConnectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCPeeringConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnection_options(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	testAccepterChange := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		log.Printf("[DEBUG] Test change to the VPC Peering Connection Options.")

		_, err := conn.ModifyVpcPeeringConnectionOptions(ctx, &ec2.ModifyVpcPeeringConnectionOptionsInput{
			VpcPeeringConnectionId: v.VpcPeeringConnectionId,
			AccepterPeeringConnectionOptions: &awstypes.PeeringConnectionOptionsRequest{
				AllowDnsResolutionFromRemoteVpc: aws.Bool(false),
			},
		})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_accepterRequesterOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						acctest.Ct1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						acctest.CtFalse,
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						acctest.Ct1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						acctest.CtTrue,
					),
					testAccepterChange,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accepter.0.allow_remote_vpc_dns_resolution",
					"auto_accept",
				},
			},
			{
				Config: testAccVPCPeeringConnectionConfig_accepterRequesterOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						acctest.Ct1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						acctest.CtFalse,
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						acctest.Ct1,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						acctest.CtTrue,
					),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnection_failedState(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionConfig_failedState(rName),
				ExpectError: regexache.MustCompile(`unexpected state 'failed'`),
			},
		},
	})
}

func TestAccVPCPeeringConnection_peerRegionAutoAccept(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionConfig_alternateRegionAutoAccept(rName, true),
				ExpectError: regexache.MustCompile("`peer_region` cannot be set whilst `auto_accept` is `true` when creating an EC2 VPC Peering Connection"),
			},
		},
	})
}

func TestAccVPCPeeringConnection_region(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_alternateRegionAutoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"pending-acceptance",
					),
				),
			},
		},
	})
}

// Tests the peering connection acceptance functionality for same region, same account.
func TestAccVPCPeeringConnection_accept(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig_autoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"pending-acceptance",
					),
				),
			},
			{
				Config: testAccVPCPeeringConnectionConfig_autoAccept(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"active",
					),
				),
			},
			// Tests that changing 'auto_accept' back to false keeps the connection active.
			{
				Config: testAccVPCPeeringConnectionConfig_autoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(ctx, resourceName,
						&v,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"active",
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
		},
	})
}

// Tests that VPC peering connection options can't be set on non-active connection.
func TestAccVPCPeeringConnection_optionsNoAutoAccept(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionConfig_accepterRequesterOptionsNoAutoAccept(rName),
				ExpectError: regexache.MustCompile(`is not active`),
			},
		},
	})
}

func testAccCheckVPCPeeringConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_peering_connection" {
				continue
			}

			_, err := tfec2.FindVPCPeeringConnectionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC Peering Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCPeeringConnectionExists(ctx context.Context, n string, v *awstypes.VpcPeeringConnection) resource.TestCheckFunc {
	return testAccCheckVPCPeeringConnectionExistsWithProvider(ctx, n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckVPCPeeringConnectionExistsWithProvider(ctx context.Context, n string, v *awstypes.VpcPeeringConnection, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Peering Connection ID is set.")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVPCPeeringConnectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCPeeringConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true
}
`, rName)
}

func testAccVPCPeeringConnectionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCPeeringConnectionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCPeeringConnectionConfig_accepterRequesterOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
`, rName)
}

func testAccVPCPeeringConnectionConfig_failedState(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCPeeringConnectionConfig_alternateRegionAutoAccept(rName string, autoAccept bool) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = %[3]q
  auto_accept = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept, acctest.AlternateRegion()))
}

func testAccVPCPeeringConnectionConfig_autoAccept(rName string, autoAccept bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept)
}

func testAccVPCPeeringConnectionConfig_accepterRequesterOptionsNoAutoAccept(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = false

  tags = {
    Name = %[1]q
  }

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
`, rName)
}
