// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClientVPNEndpoint_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_basic(t, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`client-vpn-endpoint/cvpn-endpoint-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "certificate-authentication",
					}),
					resource.TestCheckResourceAttr(resourceName, "client_cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.lambda_function_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_group", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal_url", ""),
					resource.TestCheckResourceAttrSet(resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "24"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "udp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCID, ""),
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

func testAccClientVPNEndpoint_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceClientVPNEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNEndpoint_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_tags1(t, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_tags2(t, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_tags1(t, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_msADAuth(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	domainName := acctest.RandomDomainName()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_microsoftAD(t, rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "directory-service-authentication",
					}),
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

func testAccClientVPNEndpoint_msADAuthAndMutualAuth(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	domainName := acctest.RandomDomainName()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_mutualAuthAndMicrosoftAD(t, rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "directory-service-authentication",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "certificate-authentication",
					}),
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

func testAccClientVPNEndpoint_federatedAuth(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_federatedAuth(t, rName, idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "federated-authentication",
					}),
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

func testAccClientVPNEndpoint_federatedAuthWithSelfServiceProvider(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_federatedAuthAndSelfServiceSAMLProvider(t, rName, idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "authentication_options.*", map[string]string{
						names.AttrType: "federated-authentication",
					}),
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

func testAccClientVPNEndpoint_withClientConnectOptions(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	lambdaFunction1ResourceName := "aws_lambda_function.test1"
	lambdaFunction2ResourceName := "aws_lambda_function.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(t, rName, true, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "client_connect_options.0.lambda_function_arn", lambdaFunction1ResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(t, rName, true, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "client_connect_options.0.lambda_function_arn", lambdaFunction2ResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_clientConnectOptions(t, rName, false, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_connect_options.0.lambda_function_arn", ""),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withClientLoginBannerOptions(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(t, rName, true, "Options 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", "Options 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(t, rName, true, "Options 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", "Options 2"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_clientLoginBannerOptions(t, rName, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_login_banner_options.0.banner_text", ""),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withConnectionLogOptions(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	logStream1ResourceName := "aws_cloudwatch_log_stream.test1"
	logStream2ResourceName := "aws_cloudwatch_log_stream.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(t, rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "connection_log_options.0.cloudwatch_log_stream"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(t, rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStream1ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_connectionLogOptions(t, rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_group", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "connection_log_options.0.cloudwatch_log_stream", logStream2ResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_group", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.cloudwatch_log_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_withDNSServers(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_dnsServers(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", acctest.Ct2),
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
				Config: testAccClientVPNEndpointConfig_dnsServersUpdated(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.0", "4.4.4.4"),
				),
			},
			{
				Config: testAccClientVPNEndpointConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "dns_servers.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_simpleAttributesUpdate(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	serverCertificate1ResourceName := "aws_acm_certificate.test1"
	serverCertificate2ResourceName := "aws_acm_certificate.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_simpleAttributes(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description1"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", "12"),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", acctest.CtTrue),
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
				Config: testAccClientVPNEndpointConfig_simpleAttributesUpdated(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Description2"),
					resource.TestCheckResourceAttrPair(resourceName, "server_certificate_arn", serverCertificate2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "session_timeout_hours", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "split_tunnel", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "vpn_port", "443"),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_selfServicePortal(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_selfServicePortal(t, rName, names.AttrEnabled, idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", names.AttrEnabled),
					resource.TestCheckResourceAttrSet(resourceName, "self_service_portal_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_selfServicePortal(t, rName, "disabled", idpEntityID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "self_service_portal_url", ""),
				),
			},
		},
	})
}

func testAccClientVPNEndpoint_vpcNoSecurityGroups(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	defaultSecurityGroupResourceName := "aws_default_security_group.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(t, rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", defaultSecurityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
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

func testAccClientVPNEndpoint_vpcSecurityGroups(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnEndpoint
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	securityGroup1ResourceName := "aws_security_group.test.0"
	securityGroup2ResourceName := "aws_security_group.test.1"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(t, rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1ResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup2ResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNEndpointConfig_securityGroups(t, rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroup1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckClientVPNEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_client_vpn_endpoint" {
				continue
			}

			_, err := tfec2.FindClientVPNEndpointByID(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckClientVPNEndpointExists(ctx context.Context, name string, v *awstypes.ClientVpnEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindClientVPNEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClientVPNEndpointConfig_acmCertificateBase(t *testing.T, n string) string {
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" %[1]q {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, n, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccClientVPNEndpointConfig_msADBase(rName, domain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, domain))
}

func testAccClientVPNEndpointConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccClientVPNEndpointConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_clientConnectOptions(t *testing.T, rName string, enabled bool, lambdaFunctionIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"),
		fmt.Sprintf(`
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "AWSClientVPN-%[1]s-1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "AWSClientVPN-%[1]s-2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
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

func testAccClientVPNEndpointConfig_clientLoginBannerOptions(t *testing.T, rName string, enabled bool, bannerText string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_connectionLogOptions(t *testing.T, rName string, logStreamIndex int) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_dnsServers(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_dnsServersUpdated(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_microsoftAD(t *testing.T, rName, domain string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"),
		testAccClientVPNEndpointConfig_msADBase(rName, domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.1.0.0/20"

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

func testAccClientVPNEndpointConfig_mutualAuthAndMicrosoftAD(t *testing.T, rName, domain string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"),
		testAccClientVPNEndpointConfig_msADBase(rName, domain),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.1.0.0/20"

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

func testAccClientVPNEndpointConfig_federatedAuth(t *testing.T, rName, idpEntityID string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"),
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

func testAccClientVPNEndpointConfig_federatedAuthAndSelfServiceSAMLProvider(t *testing.T, rName, idpEntityID string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_tags1(t *testing.T, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_tags2(t *testing.T, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_simpleAttributes(t *testing.T, rName string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test1"),
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test2"),
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

func testAccClientVPNEndpointConfig_simpleAttributesUpdated(t *testing.T, rName string) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test1"),
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test2"),
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

func testAccClientVPNEndpointConfig_selfServicePortal(t *testing.T, rName, selfServicePortal, idpEntityID string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"), fmt.Sprintf(`
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

func testAccClientVPNEndpointConfig_securityGroups(t *testing.T, rName string, nSecurityGroups int) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_acmCertificateBase(t, "test"),
		testAccClientVPNEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
locals {
  security_group_count = %[2]d
  security_group_ids   = local.security_group_count == 0 ? null : (local.security_group_count == 1 ? [aws_security_group.test[0].id] : aws_security_group.test[*].id)
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.1.0.0/22"

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
