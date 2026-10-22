// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaWebFunctionEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint, updated lambdaweb.GetWebFunctionEndpointOutput
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
				Config: testAccFunctionEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`web-function/.+/endpoint/.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypeHomeRegion)),
					resource.TestCheckResourceAttr(resourceName, "auth_type", string(awstypes.AuthTypeApplicationManaged)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
				),
			},
			{
				ImportStateIdFunc:                    testAccFunctionEndpointImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "endpoint_name",
			},
			{
				// auth_type is updated in place and the domain name is kept.
				Config: testAccFunctionEndpointConfig_authType(rName, string(awstypes.AuthTypeIamAuth)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &updated),
					testAccCheckFunctionEndpointNotRecreated(&endpoint, &updated),
					resource.TestCheckResourceAttr(resourceName, "auth_type", string(awstypes.AuthTypeIamAuth)),
				),
			},
		},
	})
}

func TestAccLambdaWebFunctionEndpoint_scalingAndThrottle(t *testing.T) {
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
				Config: testAccFunctionEndpointConfig_scalingAndThrottle(rName, 4, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_environments", "4"),
					resource.TestCheckResourceAttr(resourceName, "throttle_config.0.rate_limit", "100"),
				),
			},
			{
				ImportStateIdFunc:                    testAccFunctionEndpointImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "endpoint_name",
			},
			{
				Config: testAccFunctionEndpointConfig_scalingAndThrottle(rName, 6, 200),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "scaling_config.0.max_environments", "6"),
					resource.TestCheckResourceAttr(resourceName, "throttle_config.0.rate_limit", "200"),
				),
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
				Config: testAccFunctionEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceFunctionEndpoint, resourceName),
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
				Config: testAccFunctionEndpointConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypeMultiRegion)),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment_mode", string(awstypes.AutoDeploymentModeDisabled)),
					resource.TestCheckResourceAttr(resourceName, "revision_weight.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revision_weight.0.weight", "100"),
				),
			},
			{
				ImportStateIdFunc:                    testAccFunctionEndpointImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "endpoint_name",
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
				Config: testAccFunctionEndpointConfig_typeWithRegions(rName, string(awstypes.EndpointTypePerRegion)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypePerRegion)),
					// PerRegion endpoints expose one independent domain per region.
					resource.TestCheckResourceAttr(resourceName, "regional_domain_names.%", "2"),
				),
			},
			{
				ImportStateIdFunc:                    testAccFunctionEndpointImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "endpoint_name",
			},
		},
	})
}

func testAccCheckFunctionEndpointNotRecreated(before, after *lambdaweb.GetWebFunctionEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if b, a := aws.ToString(before.DomainName), aws.ToString(after.DomainName); b != a {
			return fmt.Errorf("Lambda Web Endpoint was recreated: domain name %s became %s", b, a)
		}
		return nil
	}
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

func testAccFunctionEndpointConfig_basic(rName string) string {
	return testAccFunctionEndpointConfig_authType(rName, "ApplicationManaged")
}

func testAccFunctionEndpointConfig_authType(rName, authType string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function_endpoint" "test" {
  function_name = aws_lambdaweb_function.test.function_name
  endpoint_name = "extra"
  endpoint_type = "HomeRegion"
  auth_type     = %[1]q
}
`, authType))
}

func testAccFunctionEndpointConfig_multiRegion(rName string) string {
	return testAccFunctionEndpointConfig_typeWithRegions(rName, "MultiRegion")
}

func testAccFunctionEndpointConfig_scalingAndThrottle(rName string, maxEnvironments, rateLimit int) string {
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function_endpoint" "test" {
  function_name = aws_lambdaweb_function.test.function_name
  endpoint_name = "extra"
  endpoint_type = "HomeRegion"
  auth_type     = "ApplicationManaged"

  scaling_config {
    max_environments = %[1]d
  }

  throttle_config {
    rate_limit = %[2]d
  }
}
`, maxEnvironments, rateLimit))
}

func testAccFunctionEndpointConfig_typeWithRegions(rName, endpointType string) string {
	// Pick a second region distinct from the test region. Regions with the
	// service active as of July 2026: us-east-1 and eu-west-1 (us-west-2
	// rollout pending; PreCheckRegion admits it for when it lands).
	secondRegion := endpoints.EuWest1RegionID
	if acctest.Region() == endpoints.EuWest1RegionID {
		secondRegion = endpoints.UsEast1RegionID
	}
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function_endpoint" "test" {
  function_name        = aws_lambdaweb_function.test.function_name
  endpoint_name        = "multi"
  endpoint_type        = %[3]q
  regions              = [%[1]q, %[2]q]
  auth_type            = "ApplicationManaged"
  auto_deployment_mode = "Disabled"

  revision_weight {
    revision_id = aws_lambdaweb_function.test.latest_revision_id
    weight      = 100
  }
}
`, secondRegion, acctest.Region(), endpointType))
}
