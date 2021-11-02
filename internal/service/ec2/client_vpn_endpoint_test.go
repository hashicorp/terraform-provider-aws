package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

const clientVpnEndpointDefaultLimit = 5

var testAccEc2ClientVpnEndpointSemaphore sync.Semaphore

func init() {
	testAccEc2ClientVpnEndpointSemaphore = sync.InitializeSemaphore("AWS_EC2_CLIENT_VPN_LIMIT", clientVpnEndpointDefaultLimit)
}

// This is part of an experimental feature, do not use this as a starting point for tests
//   "This place is not a place of honor... no highly esteemed deed is commemorated here... nothing valued is here.
//   What is here was dangerous and repulsive to us. This message is a warning about danger."
//   --  https://hyperallergic.com/312318/a-nuclear-warning-designed-to-last-10000-years/
func TestAccEC2ClientVPNEndpoint_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Endpoint": {
			"basic":             testAccClientVPNEndpoint_basic,
			"disappears":        testAccClientVPNEndpoint_disappears,
			"msAD":              testAccClientVPNEndpoint_msAD,
			"mutualAuthAndMsAD": testAccClientVPNEndpoint_mutualAuthAndMsAD,
			"federated":         testAccClientVPNEndpoint_federated,
			"withLogGroup":      testAccClientVPNEndpoint_withLogGroup,
			"withDNSServers":    testAccClientVPNEndpoint_withDNSServers,
			"tags":              testAccClientVPNEndpoint_tags,
			"splitTunnel":       testAccClientVPNEndpoint_splitTunnel,
			"selfServicePortal": testAccClientVPNEndpoint_selfServicePortal,
		},
		"AuthorizationRule": {
			"basic":      testAccClientVPNAuthorizationRule_basic,
			"groups":     testAccClientVPNAuthorizationRule_groups,
			"Subnets":    testAccClientVPNAuthorizationRule_Subnets,
			"disappears": testAccClientVPNAuthorizationRule_disappears,
		},
		"NetworkAssociation": {
			"basic":           testAccClientVPNNetworkAssociation_basic,
			"multipleSubnets": testAccClientVPNNetworkAssociation_multipleSubnets,
			"disappears":      testAccClientVPNNetworkAssociation_disappears,
			"securityGroups":  testAccClientVPNNetworkAssociation_securityGroups,
		},
		"Route": {
			"basic":       testAccClientVPNRoute_basic,
			"description": testAccClientVPNRoute_description,
			"disappears":  testAccClientVPNRoute_disappears,
		},
	}

	t.Parallel()
	for group, m := range testCases {
		m := m
		for name, tc := range m {
			tc := tc
			t.Run(fmt.Sprintf("%s_%s", group, name), func(t *testing.T) {
				t.Cleanup(func() {
					if os.Getenv(resource.TestEnvVar) != "" {
						testAccEc2ClientVpnEndpointSemaphore.Notify()
					}
				})
				tc(t)
			})
		}
	}
}

func testAccClientVPNEndpoint_basic(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`client-vpn-endpoint/cvpn-endpoint-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "certificate-authentication"),
					resource.TestCheckResourceAttr(resourceName, "status", ec2.ClientVpnEndpointStatusCodePendingAssociate),
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

func testAccClientVPNEndpoint_disappears(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceClientVPNEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNEndpoint_msAD(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rStr, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
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

func testAccClientVPNEndpoint_mutualAuthAndMsAD(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMutualAuthAndMicrosoftAD(rStr, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.1.type", "certificate-authentication"),
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

func testAccClientVPNEndpoint_federated(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithFederatedAuth(rStr, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "federated-authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "authentication_options.0.saml_provider_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithFederatedAuthSelfServiceSamlProviderArn(rStr, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "federated-authentication"),
					resource.TestCheckResourceAttrSet(resourceName, "authentication_options.0.saml_provider_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "authentication_options.0.self_service_saml_provider_arn"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withLogGroup(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	logGroupResourceName := "aws_cloudwatch_log_group.lg"
	logStreamResourceName := "aws_cloudwatch_log_stream.ls"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v1),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStreamResourceName, "name"),
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

func testAccClientVPNEndpoint_withDNSServers(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rStr := sdkacctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v1),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v2),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_tags(t *testing.T) {
	var v1, v2, v3 ec2.ClientVpnEndpoint
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	rStr := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig_tags(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig_tagsChanged(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_splitTunnel(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSplitTunnel(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v1),
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
					testAccCheckClientVPNEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_selfServicePortal(t *testing.T) {
	var v1, v2 ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, "enabled", idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, "disabled", idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
				),
			},
		},
	})
}

func testAccPreCheckClientVPNSyncronize(t *testing.T) {
	sync.TestAccPreCheckSyncronize(t, testAccEc2ClientVpnEndpointSemaphore, "Client VPN")
}

func testAccCheckClientVPNEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_endpoint" {
			continue
		}

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{rs.Primary.ID}),
		}

		resp, _ := conn.DescribeClientVpnEndpoints(input)
		for _, v := range resp.ClientVpnEndpoints {
			if aws.StringValue(v.ClientVpnEndpointId) == rs.Primary.ID {
				return fmt.Errorf("Client VPN endpoint (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckClientVPNEndpointExists(name string, endpoint *ec2.ClientVpnEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{rs.Primary.ID}),
		}
		result, err := conn.DescribeClientVpnEndpoints(input)
		if err != nil {
			return err
		}

		if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
			return fmt.Errorf("EC2 Client VPN Endpoint (%s) not found", rs.Primary.ID)
		}

		*endpoint = *result.ClientVpnEndpoints[0]

		return nil
	}
}

func testAccEc2ClientVpnEndpointConfigAcmCertificateBase() string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccEc2ClientVpnEndpointMsADBase(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}
`, domain),
	)
}

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
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
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "lg" {
  name = "terraform-testacc-clientvpn-loggroup-%s"
}

resource "aws_cloudwatch_log_stream" "ls" {
  name           = "${aws_cloudwatch_log_group.lg.name}-stream"
  log_group_name = aws_cloudwatch_log_group.lg.name
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = aws_cloudwatch_log_group.lg.name
    cloudwatch_log_stream = aws_cloudwatch_log_stream.ls.name
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  dns_servers = ["8.8.8.8", "8.8.4.4"]

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase(),
		testAccEc2ClientVpnEndpointMsADBase(domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = aws_directory_service_directory.test.id
  }

  connection_log_options {
    enabled = false
  }
}
`, rName),
	)
}

func testAccEc2ClientVpnEndpointConfigWithMutualAuthAndMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase(),
		testAccEc2ClientVpnEndpointMsADBase(domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = aws_directory_service_directory.test.id
  }

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName),
	)
}

func testAccEc2ClientVpnEndpointConfigWithFederatedAuth(rName, idpEntityId string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type              = "federated-authentication"
    saml_provider_arn = aws_iam_saml_provider.default.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, idpEntityId))
}

func testAccEc2ClientVpnEndpointConfigWithFederatedAuthSelfServiceSamlProviderArn(rName, idpEntityId string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_iam_saml_provider" "self_service" {
  name                   = "myprovider-selfservice-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                           = "federated-authentication"
    saml_provider_arn              = aws_iam_saml_provider.default.arn
    self_service_saml_provider_arn = aws_iam_saml_provider.self_service.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, idpEntityId)
}

func testAccEc2ClientVpnEndpointConfig_tags(rName string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
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
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
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
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = %[1]q
  server_certificate_arn = aws_acm_certificate.test.arn
  split_tunnel           = %[2]t

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, splitTunnel)
}

func testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, selfServicePortal, idpEntityId string) string {
	return testAccEc2ClientVpnEndpointConfigAcmCertificateBase() + fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%[1]s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"
  self_service_portal    = %[3]q

  authentication_options {
    type              = "federated-authentication"
    saml_provider_arn = aws_iam_saml_provider.default.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName, idpEntityId, selfServicePortal)
}
