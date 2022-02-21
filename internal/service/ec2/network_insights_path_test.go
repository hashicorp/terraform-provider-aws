package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2NetworkInsightsPath_basic(t *testing.T) {
	resourceName := "aws_ec2_network_insights_path.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsPath("tcp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "source", "aws_network_interface.test_source", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination", "aws_network_interface.test_destination", "id"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-insights-path/.+$`)),
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

func TestAccEC2NetworkInsightsPath_SourceIP(t *testing.T) {
	resourceName := "aws_ec2_network_insights_path.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsPath_source_ip("1.1.1.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_ip", "1.1.1.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEC2NetworkInsightsPath_source_ip("8.8.8.8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_ip", "8.8.8.8"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInsightsPath_DestinationIP(t *testing.T) {
	resourceName := "aws_ec2_network_insights_path.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsPath_destination_ip("1.1.1.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_ip", "1.1.1.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEC2NetworkInsightsPath_destination_ip("8.8.8.8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_ip", "8.8.8.8"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInsightsPath_DestinationPort(t *testing.T) {
	resourceName := "aws_ec2_network_insights_path.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEC2NetworkInsightsPath_destination_port("80"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_port", "80"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEC2NetworkInsightsPath_destination_port("443"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEC2NetworkInsightsPathExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_port", "443"),
				),
			},
		},
	})
}

func testAccCheckEC2NetworkInsightsPathExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Insights Path is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		id := rs.Primary.ID

		response, err := tfec2.FindNetworkInsightsPathByID(conn, id)
		if err != nil {
			return err
		}

		if response == nil || *response.NetworkInsightsPathId != id {
			return fmt.Errorf("Network Insights Path (%s) not found", id)
		}

		return nil
	}
}

func testAccCheckEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_network_insights_path" {
			continue
		}

		id := rs.Primary.ID

		_, err := tfec2.FindNetworkInsightsPathByID(conn, id)
		if err != nil && !tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidNetworkInsightsPathIDNotFound) {
			return err
		}
	}

	return nil
}

func testAccEC2NetworkInsightsPath(protocol string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_network_interface.test_source.id
  destination = aws_network_interface.test_destination.id
  protocol    = "%s"
}
`, protocol)
}

func testAccEC2NetworkInsightsPath_source_ip(source_ip string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_ec2_network_insights_path" "test" {
  source      = aws_internet_gateway.test.id
  destination = aws_network_interface.test.id
  protocol    = "tcp"
  source_ip   = "%s"
}
`, source_ip)
}

func testAccEC2NetworkInsightsPath_destination_ip(destination_ip string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_ec2_network_insights_path" "test" {
  source         = aws_network_interface.test.id
  destination    = aws_internet_gateway.test.id
  protocol       = "tcp"
  destination_ip = "%s"
}
`, destination_ip)
}

func testAccEC2NetworkInsightsPath_destination_port(destination_port string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/16"
}

resource "aws_network_interface" "test_source" {
  subnet_id = aws_subnet.test.id
}

resource "aws_network_interface" "test_destination" {
  subnet_id = aws_subnet.test.id
}

resource "aws_ec2_network_insights_path" "test" {
  source           = aws_network_interface.test_source.id
  destination      = aws_network_interface.test_destination.id
  protocol         = "tcp"
  destination_port = %s
}
`, destination_port)
}
