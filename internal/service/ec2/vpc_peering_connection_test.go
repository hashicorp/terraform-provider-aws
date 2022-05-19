package ec2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCPeeringConnection_basic(t *testing.T) {
	var v ec2.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v ec2.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPCPeeringConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCPeeringConnection_tags(t *testing.T) {
	var v ec2.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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
				Config: testAccVPCPeeringConnectionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCPeeringConnectionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnection_options(t *testing.T) {
	var v ec2.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	testAccepterChange := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		log.Printf("[DEBUG] Test change to the VPC Peering Connection Options.")

		_, err := conn.ModifyVpcPeeringConnectionOptions(
			&ec2.ModifyVpcPeeringConnectionOptionsInput{
				VpcPeeringConnectionId: v.VpcPeeringConnectionId,
				AccepterPeeringConnectionOptions: &ec2.PeeringConnectionOptionsRequest{
					AllowDnsResolutionFromRemoteVpc: aws.Bool(false),
				},
			})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAccepterRequesterOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&v,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_vpc_to_remote_classic_link",
						"true",
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_vpc_to_remote_classic_link",
						"false",
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
					"auto_accept",
				},
			},
			{
				Config: testAccVPCPeeringConnectionAccepterRequesterOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&v,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_vpc_to_remote_classic_link",
						"true",
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_vpc_to_remote_classic_link",
						"false",
					),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnection_failedState(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionFailedStateConfig(rName),
				ExpectError: regexp.MustCompile(`unexpected state 'failed'`),
			},
		},
	})
}

func TestAccVPCPeeringConnection_peerRegionAutoAccept(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionAlternateRegionAutoAcceptConfig(rName, true),
				ExpectError: regexp.MustCompile("`peer_region` cannot be set whilst `auto_accept` is `true` when creating an EC2 VPC Peering Connection"),
			},
		},
	})
}

func TestAccVPCPeeringConnection_region(t *testing.T) {
	var v ec2.VpcPeeringConnection
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAlternateRegionAutoAcceptConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
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
	var v ec2.VpcPeeringConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAutoAcceptConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
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
				Config: testAccVPCPeeringConnectionAutoAcceptConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
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
				Config: testAccVPCPeeringConnectionAutoAcceptConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCPeeringConnectionAccepterRequesterOptionsNoAutoAcceptConfig(rName),
				ExpectError: regexp.MustCompile(`is not active`),
			},
		},
	})
}

func testAccCheckVPCPeeringConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_peering_connection" {
			continue
		}

		_, err := tfec2.FindVPCPeeringConnectionByID(conn, rs.Primary.ID)

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

func testAccCheckVPCPeeringConnectionExists(n string, v *ec2.VpcPeeringConnection) resource.TestCheckFunc {
	return testAccCheckVPCPeeringConnectionExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckVPCPeeringConnectionExistsWithProvider(n string, v *ec2.VpcPeeringConnection, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPC Peering Connection ID is set.")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindVPCPeeringConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCPeeringConnectionConfig(rName string) string {
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

func testAccVPCPeeringConnectionConfigTags1(rName, tagKey1, tagValue1 string) string {
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

func testAccVPCPeeringConnectionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccVPCPeeringConnectionAccepterRequesterOptionsConfig(rName string) string {
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

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`, rName)
}

func testAccVPCPeeringConnectionFailedStateConfig(rName string) string {
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

func testAccVPCPeeringConnectionAlternateRegionAutoAcceptConfig(rName string, autoAccept bool) string {
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

func testAccVPCPeeringConnectionAutoAcceptConfig(rName string, autoAccept bool) string {
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

func testAccVPCPeeringConnectionAccepterRequesterOptionsNoAutoAcceptConfig(rName string) string {
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

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`, rName)
}
