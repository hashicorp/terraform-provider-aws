// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/acctest/config"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkFunctionEndpointARN(functionName, endpointName string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("lambda", regexache.MustCompile(`web-function/`+functionName+`/endpoint/`+endpointName))
}

func TestAccLambdaWebFunctionEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkFunctionEndpointARN(rName+"-fn", rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auth_type"), tfknownvalue.StringExact(awstypes.AuthTypeApplicationManaged)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_deployment_mode"), tfknownvalue.StringExact(awstypes.AutoDeploymentModeLatestRevision)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDomainName), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("endpoint_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpointType), tfknownvalue.StringExact(awstypes.EndpointTypeHomeRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("function_name"), knownvalue.StringExact(rName+"-fn")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regional_domain_names"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(acctest.Region()),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("revision_weight"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_config"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("throttle_config"), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccFunctionEndpointImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "endpoint_name",
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceFunctionEndpoint, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_description(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("test description 1"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("test description 1")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("test description 2"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("test description 2")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_scalingConfigPlusThrottleConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/scaling_config+throttle_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:    config.StringVariable(rName),
					"max_environments": config.IntegerVariable(4),
					"rate_limit":       config.IntegerVariable(100),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"max_environments": knownvalue.Int64Exact(4),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("throttle_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"rate_limit": knownvalue.Int64Exact(100),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/scaling_config+throttle_config/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:    config.StringVariable(rName),
					"max_environments": config.IntegerVariable(6),
					"rate_limit":       config.IntegerVariable(200),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"max_environments": knownvalue.Int64Exact(6),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("throttle_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"rate_limit": knownvalue.Int64Exact(200),
					})})),
				},
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/regions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"endpoint_type": tfconfig.StringVariable(awstypes.EndpointTypeMultiRegion),
					"regions":       acctest.SetOfStringsVariable(acctest.Region(), acctest.AlternateRegion()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpointType), tfknownvalue.StringExact(awstypes.EndpointTypeMultiRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regional_domain_names"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(acctest.Region()),
						knownvalue.StringExact(acctest.AlternateRegion()),
					})),
				},
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_perRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionEndpoint/regions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"endpoint_type": tfconfig.StringVariable(awstypes.EndpointTypePerRegion),
					"regions":       acctest.SetOfStringsVariable(acctest.Region(), acctest.AlternateRegion()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpointType), tfknownvalue.StringExact(awstypes.EndpointTypePerRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regional_domain_names"), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.Region():          knownvalue.NotNull(),
						acctest.AlternateRegion(): knownvalue.NotNull(),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("regions"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(acctest.Region()),
						knownvalue.StringExact(acctest.AlternateRegion()),
					})),
				},
			},
		},
	})
}

func testAccCheckFunctionEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdaweb_function_endpoint" {
				continue
			}

			_, err := tflambdaweb.FindWebFunctionEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["endpoint_name"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Web Function Endpoint %s still exists", rs.Primary.Attributes["endpoint_name"])
		}

		return nil
	}
}

func testAccCheckFunctionEndpointExists(ctx context.Context, t *testing.T, n string, v *lambdaweb.GetWebFunctionEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		out, err := tflambdaweb.FindWebFunctionEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["endpoint_name"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFunctionEndpointImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(n, ",", "function_name", "endpoint_name")
}
