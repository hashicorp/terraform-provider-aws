// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2Listener_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb_listener.test[0]"
	resourceName2 := "aws_lb_listener.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownValueApplicationListenerARN(rName)),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownValueApplicationListenerARN(rName)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_listener.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownValueApplicationListenerARN(rName)),
					tfquerycheck.ExpectNoResourceObject("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_lb_listener.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownValueApplicationListenerARN(rName)),
					tfquerycheck.ExpectNoResourceObject("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccELBV2Listener_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb_listener.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownValueApplicationListenerARN(rName)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_listener.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownValueApplicationListenerARN(rName)),
					querycheck.ExpectResourceKnownValues("aws_lb_listener.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("alpn_policy"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), knownValueApplicationListenerARN(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCertificateARN), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDefaultAction), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownValueApplicationListenerARN(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("load_balancer_arn"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("mutual_authentication"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPort), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrProtocol), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_issuer_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_leaf_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_subject_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_mtls_clientcert_validity_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_tls_cipher_suite_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_request_x_amzn_tls_version_header_name"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_allow_credentials_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_allow_headers_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_allow_methods_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_allow_origin_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_expose_headers_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_access_control_max_age_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_content_security_policy_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_server_enabled"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_strict_transport_security_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_x_content_type_options_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("routing_http_response_x_frame_options_header_value"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ssl_policy"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("tcp_idle_timeout_seconds"), knownvalue.Null()),
					}),
				},
			},
		},
	})
}

func TestAccELBV2Listener_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_lb_listener.test[0]"
	resourceName2 := "aws_lb_listener.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy: testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownValueApplicationListenerAlternateRegionARN(rName)),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownValueApplicationListenerAlternateRegionARN(rName)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Listener/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_lb_listener.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_lb_listener.test", identity2.Checks()),
				},
			},
		},
	})
}

func knownValueApplicationListenerARN(lbName string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("elasticloadbalancing", regexache.MustCompile(applicationListenerARNPattern(lbName)))
}

func knownValueApplicationListenerAlternateRegionARN(lbName string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("elasticloadbalancing", regexache.MustCompile(applicationListenerARNPattern(lbName)))
}
