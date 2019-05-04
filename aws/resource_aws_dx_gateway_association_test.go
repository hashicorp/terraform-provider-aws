package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway_association", &resource.Sweeper{
		Name: "aws_dx_gateway_association",
		F:    testSweepDirectConnectGatewayAssociations,
	})
}

func testSweepDirectConnectGatewayAssociations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dxconn
	gatewayInput := &directconnect.DescribeDirectConnectGatewaysInput{}

	for {
		gatewayOutput, err := conn.DescribeDirectConnectGateways(gatewayInput)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Direct Connect Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Direct Connect Gateways: %s", err)
		}

		for _, gateway := range gatewayOutput.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			associationInput := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: gateway.DirectConnectGatewayId,
			}

			for {
				associationOutput, err := conn.DescribeDirectConnectGatewayAssociations(associationInput)

				if err != nil {
					return fmt.Errorf("error retrieving Direct Connect Gateway (%s) Associations: %s", directConnectGatewayID, err)
				}

				for _, association := range associationOutput.DirectConnectGatewayAssociations {
					gatewayID := aws.StringValue(association.AssociatedGateway.Id)

					if aws.StringValue(association.AssociationState) != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, aws.StringValue(association.AssociationState), gatewayID)
						continue
					}

					input := &directconnect.DeleteDirectConnectGatewayAssociationInput{
						AssociationId: association.AssociationId,
					}

					log.Printf("[INFO] Deleting Direct Connect Gateway (%s) Association: %s", directConnectGatewayID, gatewayID)
					_, err := conn.DeleteDirectConnectGatewayAssociation(input)

					if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
						continue
					}

					if err != nil {
						return fmt.Errorf("error deleting Direct Connect Gateway (%s) Association (%s): %s", directConnectGatewayID, gatewayID, err)
					}

					if err := waitForDirectConnectGatewayAssociationDeletion(conn, aws.StringValue(association.AssociationId), 20*time.Minute); err != nil {
						return fmt.Errorf("error waiting for Direct Connect Gateway (%s) Association (%s) to be deleted: %s", directConnectGatewayID, gatewayID, err)
					}
				}

				if aws.StringValue(associationOutput.NextToken) == "" {
					break
				}

				associationInput.NextToken = associationOutput.NextToken
			}
		}

		if aws.StringValue(gatewayOutput.NextToken) == "" {
			break
		}

		gatewayInput.NextToken = gatewayOutput.NextToken
	}

	return nil
}

func TestAccAwsDxGatewayAssociation_deprecated(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_deprecated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckNoResourceAttr(resourceName, "associated_gateway_id"),
					resource.TestCheckNoResourceAttr(resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.1216997074", "10.255.255.0/28"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicVpnGateway(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckNoResourceAttr(resourceName, "vpn_gateway_id"),
					resource.TestCheckNoResourceAttr(resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.1216997074", "10.255.255.0/28"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not Found: %s", resourceName)
					}

					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["dx_gateway_id"], rs.Primary.Attributes["associated_gateway_id"]), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_basicTransitGateway(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_basicTransitGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckNoResourceAttr(resourceName, "vpn_gateway_id"),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2173830893", "10.255.255.0/30"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2984398124", "10.255.255.8/30"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not Found: %s", resourceName)
					}

					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["dx_gateway_id"], rs.Primary.Attributes["associated_gateway_id"]), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_multiVgws(t *testing.T) {
	resourceName1 := "aws_dx_gateway_association.test1"
	resourceName2 := "aws_dx_gateway_association.test2"
	rName1 := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rName2 := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVgws(rName1, rName2, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName1),
					testAccCheckAwsDxGatewayAssociationExists(resourceName2),
					resource.TestCheckResourceAttrSet(resourceName1, "dx_gateway_association_id"),
					resource.TestCheckResourceAttr(resourceName1, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName1, "allowed_prefixes.704201654", "10.255.255.16/28"),
					resource.TestCheckResourceAttrSet(resourceName2, "dx_gateway_association_id"),
					resource.TestCheckResourceAttr(resourceName2, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName2, "allowed_prefixes.2444313725", "10.255.255.32/28"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_allowedPrefixesVpnGateway(t *testing.T) {
	resourceName := "aws_dx_gateway_association.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("terraform-testacc-dxgwassoc-%d", acctest.RandInt())
	rBgpAsn := randIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "dx_gateway_association_id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2173830893", "10.255.255.0/30"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.2984398124", "10.255.255.8/30"),
				),
			},
			{
				Config: testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.1642241106", "10.255.255.8/29"),
				),
			},
		},
	})
}

func testAccCheckAwsDxGatewayAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association" {
			continue
		}

		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			AssociationId: aws.String(rs.Primary.Attributes["dx_gateway_association_id"]),
		})
		if err != nil {
			return err
		}

		if len(resp.DirectConnectGatewayAssociations) > 0 {
			return fmt.Errorf("Direct Connect Gateway (%s) is not dissociated from GW %s",
				aws.StringValue(resp.DirectConnectGatewayAssociations[0].DirectConnectGatewayId),
				aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociatedGateway.Id))
		}
	}
	return nil
}

func testAccCheckAwsDxGatewayAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccDxGatewayAssociationConfig_deprecated(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id  = "${aws_dx_gateway.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway_attachment.test.vpn_gateway_id}"
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_basicVpnGateway(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_vpn_gateway_attachment.test.vpn_gateway_id}"
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_basicTransitGateway(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_ec2_transit_gateway.test.id}"

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_multiVgws(rName1, rName2 string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[3]d"
}

resource "aws_vpc" "test1" {
  cidr_block = "10.255.255.16/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test1" {
  vpc_id         = "${aws_vpc.test1.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test1.id}"
}

resource "aws_dx_gateway_association" "test1" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_vpn_gateway_attachment.test1.vpn_gateway_id}"
}

resource "aws_vpc" "test2" {
  cidr_block = "10.255.255.32/28"

  tags = {
    Name = %[2]q
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = %[2]q
  }
}

resource "aws_vpn_gateway_attachment" "test2" {
  vpc_id         = "${aws_vpc.test2.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test2.id}"
}

resource "aws_dx_gateway_association" "test2" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_vpn_gateway_attachment.test2.vpn_gateway_id}"
}
`, rName1, rName2, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGateway(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_vpn_gateway_attachment.test.vpn_gateway_id}"

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationConfig_allowedPrefixesVpnGatewayUpdated(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = "${aws_dx_gateway.test.id}"
  associated_gateway_id = "${aws_vpn_gateway_attachment.test.vpn_gateway_id}"

  allowed_prefixes = [
    "10.255.255.8/29",
  ]
}
`, rName, rBgpAsn)
}
