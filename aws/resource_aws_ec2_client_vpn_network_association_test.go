package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_ec2_client_vpn_network_association", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_network_association",
		F:    testSweepEc2ClientVpnNetworkAssociations,
		Dependencies: []string{
			"aws_directory_service_directory",
		},
	})
}

func testSweepEc2ClientVpnNetworkAssociations(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeClientVpnEndpointsInput{}
	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, clientVpnEndpoint := range page.ClientVpnEndpoints {

			input := &ec2.DescribeClientVpnTargetNetworksInput{
				ClientVpnEndpointId: clientVpnEndpoint.ClientVpnEndpointId,
			}
			err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, isLast bool) bool {
				if page == nil {
					return !isLast
				}

				for _, networkAssociation := range page.ClientVpnTargetNetworks {
					networkAssociationID := aws.StringValue(networkAssociation.AssociationId)
					clientVpnEndpointID := aws.StringValue(networkAssociation.ClientVpnEndpointId)

					log.Printf("[INFO] Deleting Client VPN network association (%s,%s)", clientVpnEndpointID, networkAssociationID)
					err := deleteClientVpnNetworkAssociation(conn, networkAssociationID, clientVpnEndpointID)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting Client VPN network association (%s,%s): %w", clientVpnEndpointID, networkAssociationID, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}

				return !isLast
			})

			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Client VPN network association sweeper for %q: %s", region, err)
				return false
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Client VPN network associations: %w", err))
				return false
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Client VPN network association sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Client VPN network associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsEc2ClientVpnNetworkAssociation_basic(t *testing.T) {
	var assoc ec2.TargetNetwork
	var group ec2.SecurityGroup
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc),
					testAccCheckAWSDefaultSecurityGroupExists(defaultSecurityGroupResourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(resourceName, "security_groups.*", &group),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func testAccAwsEc2ClientVpnNetworkAssociation_disappears(t *testing.T) {
	var assoc ec2.TargetNetwork
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnNetworkAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsEc2ClientVpnNetworkAssociation_securityGroups(t *testing.T) {
	var assoc1, assoc2 ec2.TargetNetwork
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_network_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc1),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
				),
			},
			{
				Config: testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists(resourceName, &assoc2),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	defer testAccEc2ClientVpnEndpointSemaphore.Notify()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_network_association" {
			continue
		}

		resp, _ := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		for _, v := range resp.ClientVpnTargetNetworks {
			if *v.AssociationId == rs.Primary.ID && !(*v.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN network association (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationExists(name string, assoc *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error reading Client VPN network association (%s): %w", rs.Primary.ID, err)
		}

		for _, a := range resp.ClientVpnTargetNetworks {
			if *a.AssociationId == rs.Primary.ID && !(*a.Status.Code == ec2.AssociationStatusCodeDisassociated) {
				*assoc = *a
				return nil
			}
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationSecurityGroupID(name, key string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return tfawsresource.TestCheckTypeSetElemAttr(name, key, aws.StringValue(group.GroupId))(s)
	}
}

func testAccEc2ClientVpnNetworkAssociationConfigBasic(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName, 1),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationVpcBase(rName string, subnetCount int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidParameterValue: AZ us-west-2d is not currently supported. Please choose another az in this region
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_subnet" "test" {
  count                   = %[2]d
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}
`, rName, subnetCount)
}

func testAccEc2ClientVpnNetworkAssociationTwoSecurityGroups(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName, 1),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
		protocol = "tcp"
    from_port = 22
    to_port = 22
    cidr_blocks = ["10.1.1.0/24"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 80
    cidr_blocks = ["10.1.1.0/24"]
  }

  egress {
    protocol = "-1"
    from_port = 0
    to_port = 0
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.1.2.0/24"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.1.2.0/24"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = "${aws_acm_certificate.test.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.test.arn}"
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  subnet_id              = "${aws_subnet.test[0].id}"
  security_groups        = ["${aws_security_group.test1.id}", "${aws_security_group.test2.id}"]
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationOneSecurityGroup(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnNetworkAssociationVpcBase(rName, 1),
		testAccEc2ClientVpnNetworkAssociationAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name        = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
		protocol = "tcp"
    from_port = 22
    to_port = 22
    cidr_blocks = ["10.1.1.0/24"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 80
    cidr_blocks = ["10.1.1.0/24"]
  }

  egress {
    protocol = "-1"
    from_port = 0
    to_port = 0
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_security_group" "test2" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.1.2.0/24"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.1.2.0/24"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["10.1.0.0/16"]
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = "${aws_acm_certificate.test.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.test.arn}"
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  subnet_id              = "${aws_subnet.test[0].id}"
  security_groups        = ["${aws_security_group.test1.id}"]
}
`, rName))
}

func testAccEc2ClientVpnNetworkAssociationAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
