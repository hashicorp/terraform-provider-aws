package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ec2_client_vpn_endpoint", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_endpoint",
		F:    testSweepEc2ClientVpnEndpoints,
		Dependencies: []string{
			"aws_directory_service_directory",
		},
	})
}

func testSweepEc2ClientVpnEndpoints(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeClientVpnEndpointsInput{}

	for {
		output, err := conn.DescribeClientVpnEndpoints(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Client VPN Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Client VPN Endpoints: %s", err)
		}

		for _, clientVpnEndpoint := range output.ClientVpnEndpoints {
			if clientVpnEndpoint.Status != nil && aws.StringValue(clientVpnEndpoint.Status.Code) == ec2.ClientVpnEndpointStatusCodeDeleted {
				continue
			}

			id := aws.StringValue(clientVpnEndpoint.ClientVpnEndpointId)
			input := &ec2.DeleteClientVpnEndpointInput{
				ClientVpnEndpointId: clientVpnEndpoint.ClientVpnEndpointId,
			}

			log.Printf("[INFO] Deleting Client VPN Endpoint: %s", id)
			_, err := conn.DeleteClientVpnEndpoint(input)

			if err != nil {
				return fmt.Errorf("error deleting Client VPN Endpoint (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAwsEc2ClientVpnEndpoint_basic(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
					resource.TestCheckResourceAttr("aws_ec2_client_vpn_endpoint.test", "authentication_options.#", "1"),
					resource.TestCheckResourceAttr("aws_ec2_client_vpn_endpoint.test", "authentication_options.0.type", "certificate-authentication"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_msAD(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
					resource.TestCheckResourceAttr("aws_ec2_client_vpn_endpoint.test", "authentication_options.#", "1"),
					resource.TestCheckResourceAttr("aws_ec2_client_vpn_endpoint.test", "authentication_options.0.type", "directory-service-authentication"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withLogGroup(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				ResourceName:      "aws_ec2_client_vpn_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withDNSServers(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists("aws_ec2_client_vpn_endpoint.test"),
				),
			},

			{
				ResourceName:            "aws_ec2_client_vpn_endpoint.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_servers"},
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_tags(t *testing.T) {
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig_tags(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig_tagsChanged(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_splitTunnel(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSplitTunnel(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigSplitTunnel(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_endpoint" {
			continue
		}

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, _ := conn.DescribeClientVpnEndpoints(input)
		for _, v := range resp.ClientVpnEndpoints {
			if *v.ClientVpnEndpointId == rs.Primary.ID && !(*v.Status.Code == ec2.ClientVpnEndpointStatusCodeDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN endpoint (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsEc2ClientVpnEndpointExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_cloudwatch_log_group" "lg" {
  name = "terraform-testacc-clientvpn-loggroup-%s"
}

resource "aws_cloudwatch_log_stream" "ls" {
  name           = "${aws_cloudwatch_log_group.lg.name}-stream"
  log_group_name = "${aws_cloudwatch_log_group.lg.name}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = "${aws_cloudwatch_log_group.lg.name}"
    cloudwatch_log_stream = "${aws_cloudwatch_log_stream.ls.name}"
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  dns_servers = ["8.8.8.8", "8.8.4.4"]

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
  }
}

resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = "${aws_directory_service_directory.test.id}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfig_tagsChanged(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Usage = "changed"
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigSplitTunnel(rName string, splitTunnel bool) string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "server_auth",
  ]

  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  validity_period_hours = 12

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = %[1]q
  server_certificate_arn = "${aws_acm_certificate.test.arn}"
  split_tunnel           = %[2]t

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.test.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, splitTunnel)
}
