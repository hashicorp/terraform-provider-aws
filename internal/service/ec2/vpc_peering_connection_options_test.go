package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-cmp/cmp"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCPeeringConnectionOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection_options.test"
	pcxResourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionOptionsConfig_sameRegionSameAccount(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(resourceName, "requester.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_remote_vpc_dns_resolution", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_vpc_to_remote_classic_link", "false"),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(resourceName, "accepter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "accepter.0.allow_remote_vpc_dns_resolution", "true"),
					resource.TestCheckResourceAttr(resourceName, "accepter.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "accepter.0.allow_vpc_to_remote_classic_link", "false"),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCPeeringConnectionOptionsConfig_sameRegionSameAccount(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionOptions_differentRegionSameAccount(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection_options.test"         // Requester
	resourceNamePeer := "aws_vpc_peering_connection_options.peer"     // Accepter
	pcxResourceName := "aws_vpc_peering_connection.test"              // Requester
	pcxResourceNamePeer := "aws_vpc_peering_connection_accepter.peer" // Accepter

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(t, &providers),
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionOptionsConfig_differentRegionSameAccount(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(resourceName, "requester.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_remote_vpc_dns_resolution", "true"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_vpc_to_remote_classic_link", "false"),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.#", "1"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_remote_vpc_dns_resolution", "true"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_vpc_to_remote_classic_link", "false"),
					testAccCheckVPCPeeringConnectionOptionsWithProvider(ctx, pcxResourceNamePeer,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
						acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers),
					),
				),
			},
			{
				Config:            testAccVPCPeeringConnectionOptionsConfig_differentRegionSameAccount(rName, true, true),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCPeeringConnectionOptionsConfig_differentRegionSameAccount(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.#",
						"1",
					),
					testAccCheckVPCPeeringConnectionOptionsWithProvider(ctx, pcxResourceNamePeer,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
						acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers),
					),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionOptions_sameRegionDifferentAccount(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection_options.test"     // Requester
	resourceNamePeer := "aws_vpc_peering_connection_options.peer" // Accepter
	pcxResourceName := "aws_vpc_peering_connection.test"          // Requester

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckVPCPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionOptionsConfig_sameRegionDifferentAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(resourceName, "requester.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_remote_vpc_dns_resolution", "true"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester.0.allow_vpc_to_remote_classic_link", "false"),
					testAccCheckVPCPeeringConnectionOptions(ctx, pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.#", "1"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_remote_vpc_dns_resolution", "true"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_classic_link_to_remote_vpc", "false"),
					resource.TestCheckResourceAttr(resourceNamePeer, "accepter.0.allow_vpc_to_remote_classic_link", "false"),
				),
			},
			{
				Config:            testAccVPCPeeringConnectionOptionsConfig_sameRegionDifferentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCPeeringConnectionOptions(ctx context.Context, n, block string, options *ec2.VpcPeeringConnectionOptionsDescription) resource.TestCheckFunc {
	return testAccCheckVPCPeeringConnectionOptionsWithProvider(ctx, n, block, options, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckVPCPeeringConnectionOptionsWithProvider(ctx context.Context, n, block string, options *ec2.VpcPeeringConnectionOptionsDescription, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Peering Connection ID is set.")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Conn()

		output, err := tfec2.FindVPCPeeringConnectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		o := output.AccepterVpcInfo
		if block == "requester" {
			o = output.RequesterVpcInfo
		}

		if diff := cmp.Diff(options, o.PeeringOptions); diff != "" {
			return fmt.Errorf("VPC Peering Connection Options mismatch (-want +got):\n%s", diff)
		}

		return nil
	}
}

func testAccVPCPeeringConnectionOptionsConfig_sameRegionSameAccount(rName string, accepterDnsResolution bool) string {
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
}

resource "aws_vpc_peering_connection_options" "test" {
  vpc_peering_connection_id = aws_vpc_peering_connection.test.id

  accepter {
    allow_remote_vpc_dns_resolution = %[2]t
  }

  requester {
    allow_vpc_to_remote_classic_link = false
    allow_classic_link_to_remote_vpc = false
  }
}
`, rName, accepterDnsResolution)
}

func testAccVPCPeeringConnectionOptionsConfig_differentRegionSameAccount(rName string, dnsResolution, dnsResolutionPeer bool) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = false
  peer_region = %[2]q

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection_options" "test" {
  # As options can't be set until the connection has been accepted
  # create an explicit dependency on the accepter.
  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  requester {
    allow_remote_vpc_dns_resolution = %[3]t
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_options" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  accepter {
    allow_remote_vpc_dns_resolution = %[4]t
  }
}
`, rName, acctest.AlternateRegion(), dnsResolution, dnsResolutionPeer))
}

func testAccVPCPeeringConnectionOptionsConfig_sameRegionDifferentAccount(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id        = aws_vpc.test.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection_options" "test" {
  # As options can't be set until the connection has been accepted
  # create an explicit dependency on the accepter.
  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_options" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection_accepter.peer.id

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
`, rName))
}
