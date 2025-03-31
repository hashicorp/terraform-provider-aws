// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2Listener_Application_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Application_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "application"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_issuer_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_leaf_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_subject_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_validity_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_tls_cipher_suite_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_tls_version_header_name"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_credentials_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_headers_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_methods_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_origin_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_expose_headers_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_max_age_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_content_security_policy_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_server_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_strict_transport_security_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_content_type_options_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_frame_options_header_value", ""),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tcp_idle_timeout_seconds"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
					"tcp_idle_timeout_seconds",
					"routing_http_response_server_enabled",
					"routing_http_response_strict_transport_security_header_value",
					"routing_http_response_access_control_allow_origin_header_value",
					"routing_http_response_access_control_allow_methods_header_value",
					"routing_http_response_access_control_allow_headers_header_value",
					"routing_http_response_access_control_allow_credentials_header_value",
					"routing_http_response_access_control_expose_headers_header_value",
					"routing_http_response_access_control_max_age_header_value",
					"routing_http_response_content_security_policy_header_value",
					"routing_http_response_x_content_type_options_header_value",
					"routing_http_response_x_frame_options_header_value",
					"routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_issuer_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_subject_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_validity_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_leaf_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_header_name",
					"routing_http_request_x_amzn_tls_version_header_name",
					"routing_http_request_x_amzn_tls_cipher_suite_header_name",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Network_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Network_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "network"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TCP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_issuer_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_leaf_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_subject_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_validity_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_tls_cipher_suite_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_request_x_amzn_tls_version_header_name"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_allow_credentials_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_allow_headers_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_allow_methods_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_allow_origin_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_expose_headers_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_access_control_max_age_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_content_security_policy_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_server_enabled"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_strict_transport_security_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_x_content_type_options_header_value"),
					resource.TestCheckNoResourceAttr(resourceName, "routing_http_response_x_frame_options_header_value"),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", "350"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
					"tcp_idle_timeout_seconds",
					"routing_http_response_server_enabled",
					"routing_http_response_strict_transport_security_header_value",
					"routing_http_response_access_control_allow_origin_header_value",
					"routing_http_response_access_control_allow_methods_header_value",
					"routing_http_response_access_control_allow_headers_header_value",
					"routing_http_response_access_control_allow_credentials_header_value",
					"routing_http_response_access_control_expose_headers_header_value",
					"routing_http_response_access_control_max_age_header_value",
					"routing_http_response_content_security_policy_header_value",
					"routing_http_response_x_content_type_options_header_value",
					"routing_http_response_x_frame_options_header_value",
					"routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_issuer_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_subject_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_validity_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_leaf_header_name",
					"routing_http_request_x_amzn_mtls_clientcert_header_name",
					"routing_http_request_x_amzn_tls_version_header_name",
					"routing_http_request_x_amzn_tls_cipher_suite_header_name",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Gateway_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr("aws_lb.test", "load_balancer_type", "gateway"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckNoResourceAttr(resourceName, "alpn_policy"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "0"), // A Gateway Listener can only have one action, so the API never returns a value
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "0"),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", "350"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Application_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_update(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_basic(rName, "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_Forward_basic(rName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

// providerlint doesn't allow 'import' in the name of the test although that's
// exactly the point of this test.
func TestAccELBV2Listener_Forward_ingest(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_targetGroup(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			// The config just applied does not include default_action.0.target_group_arn as verified above.
			// This cannot be imported without changes because default_action.0.target_group_arn will be set and
			// will show as a diff.
			// See: https://github.com/hashicorp/terraform-provider-aws/issues/37211
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttrSet("default_action.0.target_group_arn", true), // this will cause a change on import
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.stickiness.0.duration", "3600"),
					acctest.ImportCheckResourceAttrSet("default_action.0.forward.0.target_group.0.arn", true),
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.target_group.0.weight", "1"),
				),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_targetGroup(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.stickiness.0.duration", "3600"),
					acctest.ImportCheckResourceAttrSet("default_action.0.forward.0.target_group.0.arn", true),
					acctest.ImportCheckResourceAttr("default_action.0.forward.0.target_group.0.weight", "1"),
				),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_weighted(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_weighted(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerConfig_Forward_changeWeightedStickiness(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerConfig_Forward_changeWeightedToBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_tgARNAndForward(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_tgARNAndForward(rName, true), // no errors expected
			},
			{
				Config:      testAccListenerConfig_Forward_tgARNAndForward(rName, false),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(`You can specify both a top-level target group ARN ("default_action[0].target_group_arn") and, with "default_action[0].forward", a target group list with ARNs, only if the ARNs match.`)),
			},
		},
	})
}

func TestAccELBV2Listener_Forward_TGARNToForward_noChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_tgARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_addStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_stickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_removeStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_stickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_TGARNToForward_weightAndStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_tgARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_Forward_weightAndStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_ToTGARN_noChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_tgARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_ToTGARN_weightStickiness(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_weightAndStickiness(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "3600"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_tgARN(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_addAction(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_addAction(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_removeAction(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_addAction(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.target_group_arn",
				},
			},
			{
				Config: testAccListenerConfig_Forward_cert(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_group.0.arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.0.weight", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.target_group_arn",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Forward_ignoreFields(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_Forward_multiTargetWithIgnore(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "440"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.stickiness.0.duration", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
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

func TestAccELBV2Listener_attributes_gwlb_TCPIdleTimeoutSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbResourceName := "aws_lb.test"
	resourceName := "aws_lb_listener.test"
	tcpTimeout1 := 60
	tcpTimeout2 := 6000

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_attributes_gwlbTCPIdleTimeoutSeconds(rName, tcpTimeout1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", lbResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "0"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", strconv.Itoa(tcpTimeout1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_attributes_gwlbTCPIdleTimeoutSeconds(rName, tcpTimeout2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", lbResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "0"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", strconv.Itoa(tcpTimeout2)),
				),
			},
		},
	})
}

func TestAccELBV2Listener_attributes_nlb_TCPIdleTimeoutSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tcpTimeout1 := 60
	tcpTimeout2 := 6000

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_attributes_nlbTCPIdleTimeoutSeconds(rName, tcpTimeout1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TCP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", strconv.Itoa(tcpTimeout1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_attributes_nlbTCPIdleTimeoutSeconds(rName, tcpTimeout2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TCP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "tcp_idle_timeout_seconds", strconv.Itoa(tcpTimeout2)),
				),
			},
		},
	})
}

func TestAccELBV2Listener_attributes_alb_HTTPRequestHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_attributes_albHTTPRequestHeaders(rName, "https://example.com", "DENY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_server_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_strict_transport_security_header_value", "max-age=31536000; includeSubDomains"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_origin_header_value", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_methods_header_value", "GET,POST,OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_headers_header_value", "Content-Type,X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_credentials_header_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_expose_headers_header_value", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_max_age_header_value", "3600"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_content_security_policy_header_value", "default-src 'self'"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_content_type_options_header_value", "nosniff"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_frame_options_header_value", "DENY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
			{
				Config: testAccListenerConfig_attributes_albHTTPRequestHeaders(rName, "https://www.example.com", "SAMEORIGIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_server_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_strict_transport_security_header_value", "max-age=31536000; includeSubDomains"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_origin_header_value", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_methods_header_value", "GET,POST,OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_headers_header_value", "Content-Type,X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_credentials_header_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_expose_headers_header_value", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_max_age_header_value", "3600"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_content_security_policy_header_value", "default-src 'self'"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_content_type_options_header_value", "nosniff"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_frame_options_header_value", "SAMEORIGIN"),
				),
			},
		},
	})
}

func TestAccELBV2Listener_attributes_alb_HTTPSRequestHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = rName[:min(len(rName), 30)]
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_attributes_albHTTPSRequestHeaders(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_server_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_strict_transport_security_header_value", "max-age=31536000; includeSubDomains"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_origin_header_value", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_methods_header_value", "GET,POST,OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_headers_header_value", "Content-Type,X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_allow_credentials_header_value", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_expose_headers_header_value", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_access_control_max_age_header_value", "3600"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_content_security_policy_header_value", "default-src 'self'"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_content_type_options_header_value", "nosniff"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_response_x_frame_options_header_value", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name", "X-Custom-Serial-Number"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_issuer_header_name", "X-Custom-Issuer"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_subject_header_name", "X-Custom-Subject"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_validity_header_name", "X-Custom-Validity"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_leaf_header_name", "X-Custom-Leaf"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_mtls_clientcert_header_name", "X-Custom-Mtls-Cert"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_tls_version_header_name", "X-Custom-TLS-Version"),
					resource.TestCheckResourceAttr(resourceName, "routing_http_request_x_amzn_tls_cipher_suite_header_name", "X-Custom-Cipher-Suite"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_upd(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basicUdp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "UDP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "514"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

// TestAccELBV2Listener_backwardsCompatibility confirms that the resource type `aws_alb_listener` works
func TestAccELBV2Listener_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_alb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_alb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_alb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_https(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_https(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.advertise_trust_store_ca_names", ""),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationOff),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.trust_store_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_mutualAuthentication(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_mutualAuthentication(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.advertise_trust_store_ca_names", "off"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationVerify),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_authentication.0.trust_store_arn", "aws_lb_trust_store.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_mutualAuthenticationPassthrough(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_mutualAuthenticationPassthrough(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationPassthrough),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.trust_store_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_mutualAuthenticationAdvertiseCASubject(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_mutualAuthenticationAdvertiseCASubject(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.advertise_trust_store_ca_names", "on"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.ignore_client_certificate_expiry", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.0.mode", tfelbv2.MutualAuthenticationVerify),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_authentication.0.trust_store_arn", "aws_lb_trust_store.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_Gateway_lbARN(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbResourceName := "aws_lb.test"
	resourceName := "aws_lb_listener.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_arnGateway(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", lbResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccELBV2Listener_Protocol_tls(t *testing.T) {
	ctx := acctest.Context(t)
	var listener1 awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_protocolTLS(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener1),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TLS"),
					resource.TestCheckResourceAttr(resourceName, "alpn_policy", tfelbv2.AlpnPolicyHTTP2Preferred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
					resource.TestCheckResourceAttr(resourceName, "mutual_authentication.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_redirect(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_redirect(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.host", "#{host}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.path", "/#{path}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.query", "#{query}"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.0.status_code", "HTTP_301"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "0"),
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

func TestAccELBV2Listener_fixedResponse(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_fixedResponse(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "fixed-response"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.message_body", "Fixed response content"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "200"),
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

func TestAccELBV2Listener_cognito(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_cognito(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-cognito"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_action.0.authenticate_cognito.0.user_pool_domain"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_cognito.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_oidc(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_oidc(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authorization_endpoint", "https://example.com/authorization_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.client_id", "s6BhdRkqt3"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.client_secret", "7Fjfp0ZBr1KtDRbnfVdmIw"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.token_endpoint", "https://example.com/token_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.user_info_endpoint", "https://example.com/user_info_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.authenticate_oidc.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.type", "forward"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.1.target_group_arn", "aws_lb_target_group.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, "aws_iam_server_certificate.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_DefaultAction_defaultOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_DefaultAction_defaultOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

func TestAccELBV2Listener_DefaultAction_specifyOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_DefaultAction_specifyOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", "4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"default_action.0.authenticate_oidc.0.client_secret",
					"default_action.1.forward",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6171
func TestAccELBV2Listener_DefaultAction_actionDisappears(t *testing.T) {
	ctx := acctest.Context(t)
	var listener awstypes.Listener
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_DefaultAction_defaultOrder(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", "2"),
					testAccCheckListenerDefaultActionOrderDisappears(ctx, &listener, 1),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						// TODO: change `default_action[0]`
						// TODO: add `default_action[1]`
					},
				},
			},
		},
	})
}

func TestAccELBV2Listener_DefaultAction_empty(t *testing.T) {
	t.Parallel()

	testcases := map[awstypes.ActionTypeEnum]struct {
		actionType    awstypes.ActionTypeEnum
		expectedError *regexp.Regexp
	}{
		awstypes.ActionTypeEnumForward: {
			actionType: awstypes.ActionTypeEnumForward,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Either %q or %q must be specified when %q is %q.",
				"default_action[0].target_group_arn", "default_action[0].forward",
				"default_action[0].type",
				awstypes.ActionTypeEnumForward,
			))),
		},

		awstypes.ActionTypeEnumAuthenticateOidc: {
			actionType: awstypes.ActionTypeEnumAuthenticateOidc,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].authenticate_oidc",
				"default_action[0].type",
				awstypes.ActionTypeEnumAuthenticateOidc,
			))),
		},

		awstypes.ActionTypeEnumAuthenticateCognito: {
			actionType: awstypes.ActionTypeEnumAuthenticateCognito,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].authenticate_cognito",
				"default_action[0].type",
				awstypes.ActionTypeEnumAuthenticateCognito,
			))),
		},

		awstypes.ActionTypeEnumRedirect: {
			actionType: awstypes.ActionTypeEnumRedirect,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].redirect",
				"default_action[0].type",
				awstypes.ActionTypeEnumRedirect,
			))),
		},

		awstypes.ActionTypeEnumFixedResponse: {
			actionType: awstypes.ActionTypeEnumFixedResponse,
			expectedError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Attribute %q must be specified when %q is %q.",
				"default_action[0].fixed_response",
				"default_action[0].type",
				awstypes.ActionTypeEnumFixedResponse,
			))),
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // uses t.Setenv
		t.Run(string(name), func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckListenerDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config:      testAccListenerConfig_DefaultAction_empty(rName, testcase.actionType),
						ExpectError: testcase.expectedError,
					},
				},
			})
		})
	}
}

// https://github.com/hashicorp/terraform-provider-aws/issues/35668.
func TestAccELBV2Listener_redirectWithTargetGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Listener
	resourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.34.0",
					},
				},
				Config: testAccListenerConfig_redirectWithTargetGroupARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_arn", "aws_lb.test", names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "elasticloadbalancing", regexache.MustCompile("listener/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.redirect.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.target_group_arn", ""),
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.36.0",
					},
				},
				Config:             testAccListenerConfig_redirectWithTargetGroupARN(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckListenerDefaultActionOrderDisappears(ctx context.Context, listener *awstypes.Listener, actionOrderToDelete int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var newDefaultActions []awstypes.Action

		for i, action := range listener.DefaultActions {
			if int(aws.ToInt32(action.Order)) == actionOrderToDelete {
				newDefaultActions = slices.Delete(listener.DefaultActions, i, i+1)
				break
			}
		}

		if len(newDefaultActions) == 0 {
			return fmt.Errorf("Unable to find default action order %d from default actions: %#v", actionOrderToDelete, listener.DefaultActions)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		input := &elasticloadbalancingv2.ModifyListenerInput{
			DefaultActions: newDefaultActions,
			ListenerArn:    listener.ListenerArn,
		}

		_, err := conn.ModifyListener(ctx, input)

		return err
	}
}

func testAccCheckListenerExists(ctx context.Context, n string, v *awstypes.Listener) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := tfelbv2.FindListenerByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_listener" && rs.Type != "aws_alb_listener" {
				continue
			}

			_, err := tfelbv2.FindListenerByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Listener %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccListenerConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Application_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "application"
  internal           = true
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Network_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "TCP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "network"
  internal           = true
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Gateway_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "gateway"
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_Forward_basic(rName, targetName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "440"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.%[2]s.arn
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test1" {
  name     = "%[1]s-1"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

resource "aws_lb_target_group" "test2" {
  name     = "%[1]s-2"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}
`, rName, targetName))
}

func testAccListenerConfig_Forward_weighted(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test1.arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 1
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[2]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_Forward_changeWeightedStickiness(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test1.arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 1
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_Forward_tgARNAndForward(rName string, sameTG bool) string {
	tg := "test"
	if !sameTG {
		tg = "test2"
	}
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "440"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn

    forward {
      target_group {
        arn    = aws_lb_target_group.%[2]s.arn
        weight = 1
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = "%[1]s2"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8082
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = "%[1]s2"
  }
}
`, rName, tg))
}

func testAccListenerConfig_Forward_targetGroup(rName string, useDATG bool) string {
	daTG := "target_group_arn = aws_lb_target_group.test.arn"
	if !useDATG {
		daTG = ""
	}
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "440"

  default_action {
    order = 1
    type  = "forward"
    %[1]s

    forward {
      stickiness {
        enabled  = true
        duration = 3600
      }

      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 1
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[2]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[2]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[2]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[2]q
  }
}
`, daTG, rName))
}

func testAccListenerConfig_Forward_tgARN(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
  `, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_cert(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_stickiness(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_weightAndStickiness(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 2
      }

      stickiness {
        enabled  = true
        duration = 3600
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_addAction(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  default_action {
    type = "forward"

    forward {
      target_group {
        arn = aws_lb_target_group.test.arn
      }
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_multiTargetWithIgnore(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "440"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test.arn
        weight = 100
      }

      target_group {
        arn    = aws_lb_target_group.test2.arn
        weight = 0
      }
    }
  }

  lifecycle {
    ignore_changes = [
      default_action[0].forward,
      default_action[0].target_group_arn,
    ]
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = "%[1]s-2"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_attributes_gwlbTCPIdleTimeoutSeconds(rName string, seconds int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 80
    protocol = "HTTP"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }

  tcp_idle_timeout_seconds = %[2]d
}
`, rName, seconds))
}

func testAccListenerConfig_attributes_nlbTCPIdleTimeoutSeconds(rName string, seconds int) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "TCP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  tcp_idle_timeout_seconds = %[2]d
}

resource "aws_lb" "test" {
  name               = %[1]q
  load_balancer_type = "network"
  internal           = true
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "TCP"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = 8081
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, seconds))
}

func testAccListenerConfig_attributes_albHTTPRequestHeaders(rName, allowOriginHeaderValue, frameOptionsHeaderValue string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  protocol          = "HTTP"
  port              = 80

  routing_http_response_server_enabled                                = true
  routing_http_response_strict_transport_security_header_value        = "max-age=31536000; includeSubDomains"
  routing_http_response_access_control_allow_origin_header_value      = %[2]q
  routing_http_response_access_control_allow_methods_header_value     = "GET,POST,OPTIONS"
  routing_http_response_access_control_allow_headers_header_value     = "Content-Type,X-Custom-Header"
  routing_http_response_access_control_allow_credentials_header_value = "true"
  routing_http_response_access_control_expose_headers_header_value    = "X-Custom-Header"
  routing_http_response_access_control_max_age_header_value           = "3600"
  routing_http_response_content_security_policy_header_value          = "default-src 'self'"
  routing_http_response_x_content_type_options_header_value           = "nosniff"
  routing_http_response_x_frame_options_header_value                  = %[3]q

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, allowOriginHeaderValue, frameOptionsHeaderValue))
}

func testAccListenerConfig_attributes_albHTTPSRequestHeaders(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn                                                     = aws_lb.test.arn
  protocol                                                              = "HTTPS"
  port                                                                  = 443
  ssl_policy                                                            = "ELBSecurityPolicy-2016-08"
  certificate_arn                                                       = aws_iam_server_certificate.test.arn
  routing_http_response_server_enabled                                  = true
  routing_http_response_strict_transport_security_header_value          = "max-age=31536000; includeSubDomains"
  routing_http_response_access_control_allow_origin_header_value        = "https://example.com"
  routing_http_response_access_control_allow_methods_header_value       = "GET,POST,OPTIONS"
  routing_http_response_access_control_allow_headers_header_value       = "Content-Type,X-Custom-Header"
  routing_http_response_access_control_allow_credentials_header_value   = "true"
  routing_http_response_access_control_expose_headers_header_value      = "X-Custom-Header"
  routing_http_response_access_control_max_age_header_value             = "3600"
  routing_http_response_content_security_policy_header_value            = "default-src 'self'"
  routing_http_response_x_content_type_options_header_value             = "nosniff"
  routing_http_response_x_frame_options_header_value                    = "DENY"
  routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name = "X-Custom-Serial-Number"
  routing_http_request_x_amzn_mtls_clientcert_issuer_header_name        = "X-Custom-Issuer"
  routing_http_request_x_amzn_mtls_clientcert_subject_header_name       = "X-Custom-Subject"
  routing_http_request_x_amzn_mtls_clientcert_validity_header_name      = "X-Custom-Validity"
  routing_http_request_x_amzn_mtls_clientcert_leaf_header_name          = "X-Custom-Leaf"
  routing_http_request_x_amzn_mtls_clientcert_header_name               = "X-Custom-Mtls-Cert"
  routing_http_request_x_amzn_tls_version_header_name                   = "X-Custom-TLS-Version"
  routing_http_request_x_amzn_tls_cipher_suite_header_name              = "X-Custom-Cipher-Suite"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.test.id]
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_Forward_changeWeightedToBasic(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test1.arn
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test1" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name     = %[2]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[2]q
  }
}
`, rName, rName2))
}

func testAccListenerConfig_basicUdp(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "UDP"
  port              = "514"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 514
  protocol = "UDP"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 514
    protocol = "TCP"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

// testAccListenerConfig_backwardsCompatibility should be the equivalent of `testAccListenerConfig_basic`
// but using the legacy `aws_alb*` resource types.
func testAccListenerConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_alb_listener" "test" {
  load_balancer_arn = aws_alb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_alb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_alb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_https(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = false
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_mutualAuthentication(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		testAccTrustStoreConfig_baseS3BucketCA(rName),
		fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  mutual_authentication {
    mode            = "verify"
    trust_store_arn = aws_lb_trust_store.test.arn
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_mutualAuthenticationAdvertiseCASubject(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		testAccTrustStoreConfig_baseS3BucketCA(rName),
		fmt.Sprintf(`
resource "aws_lb_trust_store" "test" {
  name                             = %[1]q
  ca_certificates_bundle_s3_bucket = aws_s3_bucket.test.bucket
  ca_certificates_bundle_s3_key    = aws_s3_object.test.key
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  mutual_authentication {
    mode                           = "verify"
    trust_store_arn                = aws_lb_trust_store.test.arn
    advertise_trust_store_ca_names = "on"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_mutualAuthenticationPassthrough(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  mutual_authentication {
    mode = "passthrough"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_arnGateway(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.test.id

  health_check {
    port     = 80
    protocol = "HTTP"
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_protocolTLS(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  certificate_arn   = aws_acm_certificate.test.arn
  load_balancer_arn = aws_lb.test.arn
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  alpn_policy       = "HTTP2Preferred"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TLS"
  vpc_id   = aws_vpc.test.id

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_redirect(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_fixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code  = "200"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccListenerConfig_cognito(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = %[1]q
  internal                   = false
  security_groups            = [aws_security_group.test.id]
  subnets                    = aws_subnet.test[*].id
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_cognito_user_pool_client" "test" {
  name                                 = %[1]q
  user_pool_id                         = aws_cognito_user_pool.test.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls                        = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri                 = "https://www.example.com/redirect"
  logout_urls                          = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "authenticate-cognito"

    authenticate_cognito {
      user_pool_arn       = aws_cognito_user_pool.test.arn
      user_pool_client_id = aws_cognito_user_pool_client.test.id
      user_pool_domain    = aws_cognito_user_pool_domain.test.domain

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_oidc(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb" "test" {
  name                       = %[1]q
  internal                   = false
  security_groups            = [aws_security_group.test.id]
  subnets                    = aws_subnet.test[*].id
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_DefaultAction_defaultOrder(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb" "test" {
  internal        = true
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_DefaultAction_specifyOrder(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    order = 2
    type  = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param = "test"
      }
    }
  }

  default_action {
    order            = 4
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb" "test" {
  internal        = true
  name            = %[1]q
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccListenerConfig_DefaultAction_empty(rName string, action awstypes.ActionTypeEnum) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    type = %[2]q
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, action))
}

func testAccListenerConfig_redirectWithTargetGroupARN(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
