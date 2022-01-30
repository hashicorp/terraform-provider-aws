package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"basic":                  testAccClientVPNEndpoint_basic,
			"disappears":             testAccClientVPNEndpoint_disappears,
			"msADAuth":               testAccClientVPNEndpoint_msADAuth,
			"mutualAuthAndMsADAuth":  testAccClientVPNEndpoint_mutualAuthAndMsADAuth,
			"federatedAuth":          testAccClientVPNEndpoint_federatedAuth,
			"withLogGroup":           testAccClientVPNEndpoint_withLogGroup,
			"withDNSServers":         testAccClientVPNEndpoint_withDNSServers,
			"tags":                   testAccClientVPNEndpoint_tags,
			"simpleAttributesUpdate": testAccClientVPNEndpoint_simpleAttributesUpdate,
			"selfServicePortal":      testAccClientVPNEndpoint_selfServicePortal,
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
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`client-vpn-endpoint/cvpn-endpoint-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "certificate-authentication"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.active_directory_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "authentication_options.0.root_certificate_chain_arn"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.saml_provider_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.self_service_saml_provider_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "client_cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_group", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "24"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", ec2.ClientVpnEndpointStatusCodePendingAssociate),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "udp"),
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
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceClientVPNEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNEndpoint_tags(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
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
				Config: testAccEc2ClientVpnEndpointConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_msADAuth(t *testing.T) {
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

func testAccClientVPNEndpoint_mutualAuthAndMsADAuth(t *testing.T) {
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

func testAccClientVPNEndpoint_federatedAuth(t *testing.T) {
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
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	logStream1ResourceName := "aws_cloudwatch_log_stream.test1"
	logStream2ResourceName := "aws_cloudwatch_log_stream.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStream1ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStream2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_group", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "false"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withDNSServers(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_servers.*", "8.8.8.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_servers.*", "8.8.4.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_servers.*", "4.4.4.4"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "0"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_simpleAttributesUpdate(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	serverCertificate1ResourceName := "aws_acm_certificate.test1"
	serverCertificate2ResourceName := "aws_acm_certificate.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSimpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Description1"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "12"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigSimpleAttributesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Description2"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "10"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_selfServicePortal(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, "enabled", idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, "disabled", idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
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

		_, err := tfec2.FindClientVPNEndpointByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Client VPN Endpoint %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckClientVPNEndpointExists(name string, v *ec2.ClientVpnEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Client VPN Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindClientVPNEndpointByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEc2ClientVpnEndpointConfigAcmCertificateBase(n string) string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" %[1]q {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, n, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
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

func testAccEc2ClientVpnEndpointConfigBasic() string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), `
resource "aws_ec2_client_vpn_endpoint" "test" {
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
`)
}

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
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
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string, logStreamIndex int) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test1" {
  name           = "%[1]s-1"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_cloudwatch_log_stream" "test2" {
  name           = "%[1]s-2"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = aws_cloudwatch_log_group.test.name
    cloudwatch_log_stream = %[2]d == 1 ? aws_cloudwatch_log_stream.test1.name : aws_cloudwatch_log_stream.test2.name
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, logStreamIndex))
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigWithDNSServersUpdated(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  dns_servers = ["4.4.4.4"]

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"),
		testAccEc2ClientVpnEndpointMsADBase(domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = aws_directory_service_directory.test.id
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigWithMutualAuthAndMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"),
		testAccEc2ClientVpnEndpointMsADBase(domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigWithFederatedAuth(rName, idpEntityId string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"),
		fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type              = "federated-authentication"
    saml_provider_arn = aws_iam_saml_provider.default.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, idpEntityId))
}

func testAccEc2ClientVpnEndpointConfigWithFederatedAuthSelfServiceSamlProviderArn(rName, idpEntityID string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_iam_saml_provider" "self_service" {
  name                   = "myprovider-selfservice-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
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

  tags = {
    Name = %[1]q
  }
}
`, rName, idpEntityID))
}

func testAccEc2ClientVpnEndpointConfigTags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
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
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEc2ClientVpnEndpointConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
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
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEc2ClientVpnEndpointConfigSimpleAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test1"),
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test2"),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = "Description1"
  server_certificate_arn = aws_acm_certificate.test1.arn
  split_tunnel           = true
  session_timeout_hours  = 12

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test1.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigSimpleAttributesUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test1"),
		testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test2"),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = "Description2"
  server_certificate_arn = aws_acm_certificate.test2.arn
  split_tunnel           = false
  session_timeout_hours  = 10

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test1.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEc2ClientVpnEndpointConfigSelfServicePortal(rName, selfServicePortal, idpEntityID string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfigAcmCertificateBase("test"), fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[3]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"
  self_service_portal    = %[2]q

  authentication_options {
    type              = "federated-authentication"
    saml_provider_arn = aws_iam_saml_provider.test.arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, selfServicePortal, idpEntityID))
}
