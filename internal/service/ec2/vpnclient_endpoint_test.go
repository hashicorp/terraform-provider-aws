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

const clientVPNEndpointDefaultLimit = 5

var testAccEc2ClientVpnEndpointSemaphore sync.Semaphore

func init() {
	testAccEc2ClientVpnEndpointSemaphore = sync.InitializeSemaphore("AWS_EC2_CLIENT_VPN_LIMIT", clientVPNEndpointDefaultLimit)
}

// This is part of an experimental feature, do not use this as a starting point for tests
//   "This place is not a place of honor... no highly esteemed deed is commemorated here... nothing valued is here.
//   What is here was dangerous and repulsive to us. This message is a warning about danger."
//   --  https://hyperallergic.com/312318/a-nuclear-warning-designed-to-last-10000-years/
func TestAccClientVPNEndpoint_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Endpoint": {
			"basic":                        testAccClientVPNEndpoint_basic,
			"disappears":                   testAccClientVPNEndpoint_disappears,
			"msADAuth":                     testAccClientVPNEndpoint_msADAuth,
			"msADAuthAndMutualAuth":        testAccClientVPNEndpoint_msADAuthAndMutualAuth,
			"federatedAuth":                testAccClientVPNEndpoint_federatedAuth,
			"federatedAuthWithSelfService": testAccClientVPNEndpoint_federatedAuthWithSelfServiceProvider,
			"withClientConnect":            testAccClientVPNEndpoint_withClientConnectOptions,
			"withClientLoginBanner":        testAccClientVPNEndpoint_withClientLoginBannerOptions,
			"withLogGroup":                 testAccClientVPNEndpoint_withConnectionLogOptions,
			"withDNSServers":               testAccClientVPNEndpoint_withDNSServers,
			"tags":                         testAccClientVPNEndpoint_tags,
			"simpleAttributesUpdate":       testAccClientVPNEndpoint_simpleAttributesUpdate,
			"selfServicePortal":            testAccClientVPNEndpoint_selfServicePortal,
			"vpcNoSecurityGroups":          testAccClientVPNEndpoint_vpcNoSecurityGroups,
			"vpcSecurityGroups":            testAccClientVPNEndpoint_vpcSecurityGroups,
			"basicDataSource":              testAccClientVPNEndpointDataSource_basic,
		},
		"AuthorizationRule": {
			"basic":              testAccClientVPNAuthorizationRule_basic,
			"groups":             testAccClientVPNAuthorizationRule_groups,
			"subnets":            testAccClientVPNAuthorizationRule_subnets,
			"disappears":         testAccClientVPNAuthorizationRule_disappears,
			"disappearsEndpoint": testAccClientVPNAuthorizationRule_Disappears_endpoint,
		},
		"NetworkAssociation": {
			"basic":                    testAccClientVPNNetworkAssociation_basic,
			"multipleSubnets":          testAccClientVPNNetworkAssociation_multipleSubnets,
			"disappears":               testAccClientVPNNetworkAssociation_disappears,
			"securityGroups":           testAccClientVPNNetworkAssociation_securityGroups,
			"securityGroupsOnEndpoint": testAccClientVPNNetworkAssociation_securityGroupsOnEndpoint,
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_basic(rName),
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
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.lambda_function_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_group", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "24"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", ec2.ClientVpnEndpointStatusCodePendingAssociate),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "udp"),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpn_port", "443"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_basic(rName),
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
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_tags1("key1", "value1"),
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
				Config: testAccClientVPNEndpointConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_tags1("key2", "value2"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	dsDirectoryResourceName := "aws_directory_service_directory.test"
	domainName := acctest.RandomDomainName()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_microsoftAD(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.active_directory_id", dsDirectoryResourceName, "id"),
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

func testAccClientVPNEndpoint_msADAuthAndMutualAuth(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	dsDirectoryResourceName := "aws_directory_service_directory.test"
	serverCertificateResourceName := "aws_acm_certificate.test"
	domainName := acctest.RandomDomainName()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_mutualAuthAndMicrosoftAD(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.active_directory_id", dsDirectoryResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.1.type", "certificate-authentication"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.root_certificate_chain_arn", serverCertificateResourceName, "arn"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	samlProviderResourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_federatedAuth(rName, idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "federated-authentication"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.saml_provider_arn", samlProviderResourceName, "arn"),
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

func testAccClientVPNEndpoint_federatedAuthWithSelfServiceProvider(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	samlProvider1ResourceName := "aws_iam_saml_provider.test1"
	samlProvider2ResourceName := "aws_iam_saml_provider.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckClientVPNSyncronize(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_federatedAuthAndSelfServiceSAMLProvider(rName, idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "federated-authentication"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.saml_provider_arn", samlProvider1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authentication_options.0.self_service_saml_provider_arn", samlProvider2ResourceName, "arn"),
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

func testAccClientVPNEndpoint_withClientConnectOptions(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	lambdaFunction1ResourceName := "aws_lambda_function.test1"
	lambdaFunction2ResourceName := "aws_lambda_function.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(rName, true, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "client_connect_options.0.lambda_function_arn", lambdaFunction1ResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(rName, true, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "client_connect_options.0.lambda_function_arn", lambdaFunction2ResourceName, "arn"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(rName, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.lambda_function_arn", ""),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withClientLoginBannerOptions(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(rName, true, "Options 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", "Options 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(rName, true, "Options 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", "Options 2"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(rName, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", ""),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withConnectionLogOptions(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	logStream1ResourceName := "aws_cloudwatch_log_stream.test1"
	logStream2ResourceName := "aws_cloudwatch_log_stream.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_log_options.0.cloudwatch_log_stream"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(rName, 1),
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
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStream2ResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_basic(rName),
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
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_dnsServers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.1", "8.8.4.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_dnsServersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.0", "4.4.4.4"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_basic(rName),
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
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Description1"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "12"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "true"),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "vpn_port", "1194"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_simpleAttributesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Description2"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "10"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", "false"),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "vpn_port", "443"),
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
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_selfServicePortal(rName, "enabled", idpEntityID),
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
				Config: testAccClientVPNEndpointConfig_selfServicePortal(rName, "disabled", idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_vpcNoSecurityGroups(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", defaultSecurityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func testAccClientVPNEndpoint_vpcSecurityGroups(t *testing.T) {
	var v ec2.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	securityGroup1ResourceName := "aws_security_group.test1"
	securityGroup2ResourceName := "aws_security_group.test2"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1ResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func testAccClientVPNEndpointConfig_acmCertificateBase(n string) string {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" %[1]q {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, n, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccClientVPNEndpointConfig_msADBase(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, domain))
}

func testAccClientVPNEndpointConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccClientVPNEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_clientConnectOptions(rName string, enabled bool, lambdaFunctionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		fmt.Sprintf(`
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "AWSClientVPN-%[1]s-1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}

resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "AWSClientVPN-%[1]s-2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}

locals {
  enabled             = %[2]t
  index               = %[3]d
  lambda_function_arn = local.enabled ? (local.index == 1 ? aws_lambda_function.test1.arn : aws_lambda_function.test2.arn) : null
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  client_connect_options {
    enabled             = local.enabled
    lambda_function_arn = local.lambda_function_arn
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, enabled, lambdaFunctionIndex))
}

func testAccClientVPNEndpointConfig_clientLoginBannerOptions(rName string, enabled bool, bannerText string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
locals {
  enabled     = %[2]t
  text        = %[3]q
  banner_text = local.enabled ? local.text : null
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  client_login_banner_options {
    enabled     = local.enabled
    banner_text = local.banner_text
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, enabled, bannerText))
}

func testAccClientVPNEndpointConfig_connectionLogOptions(rName string, logStreamIndex int) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

locals {
  log_stream_index = %[2]d
  log_stream       = local.log_stream_index == 0 ? null : (local.log_stream_index == 1 ? aws_cloudwatch_log_stream.test1.name : aws_cloudwatch_log_stream.test2.name)
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
    cloudwatch_log_stream = local.log_stream
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, logStreamIndex))
}

func testAccClientVPNEndpointConfig_dnsServers(rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_dnsServersUpdated(rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_microsoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		testAccClientVPNEndpointConfig_msADBase(rName, domain),
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

func testAccClientVPNEndpointConfig_mutualAuthAndMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		testAccClientVPNEndpointConfig_msADBase(rName, domain),
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

func testAccClientVPNEndpointConfig_federatedAuth(rName, idpEntityID string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

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
`, rName, idpEntityID))
}

func testAccClientVPNEndpointConfig_federatedAuthAndSelfServiceSAMLProvider(rName, idpEntityID string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
resource "aws_iam_saml_provider" "test1" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_iam_saml_provider" "test2" {
  name                   = "%[1]s-self-service"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                           = "federated-authentication"
    saml_provider_arn              = aws_iam_saml_provider.test1.arn
    self_service_saml_provider_arn = aws_iam_saml_provider.test2.arn
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

func testAccClientVPNEndpointConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_simpleAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test1"),
		testAccClientVPNEndpointConfig_acmCertificateBase("test2"),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = "Description1"
  server_certificate_arn = aws_acm_certificate.test1.arn
  split_tunnel           = true
  session_timeout_hours  = 12
  transport_protocol     = "tcp"
  vpn_port               = 1194

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

func testAccClientVPNEndpointConfig_simpleAttributesUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test1"),
		testAccClientVPNEndpointConfig_acmCertificateBase("test2"),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  client_cidr_block      = "10.0.0.0/16"
  description            = "Description2"
  server_certificate_arn = aws_acm_certificate.test2.arn
  split_tunnel           = false
  session_timeout_hours  = 10
  transport_protocol     = "tcp"
  vpn_port               = 443

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

func testAccClientVPNEndpointConfig_selfServicePortal(rName, selfServicePortal, idpEntityID string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase("test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_securityGroups(rName string, nSecurityGroups int) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase("test"),
		testAccClientVPNEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
locals {
  security_group_count = %[2]d
  security_group_ids   = local.security_group_count == 0 ? null : (local.security_group_count == 1 ? [aws_security_group.test1.id] : [aws_security_group.test1.id, aws_security_group.test2.id])
}

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

  vpc_id             = aws_vpc.test.id
  security_group_ids = local.security_group_ids

  depends_on = [aws_subnet.test[0], aws_subnet.test[1]]
}
`, rName, nSecurityGroups))
}
