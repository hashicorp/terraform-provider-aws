// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaWebEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`web-function/.+/endpoint/.+`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", "extra"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypeHomeRegion)),
					resource.TestCheckResourceAttr(resourceName, "auth_type", string(awstypes.AuthTypeApplicationManaged)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEndpointImportStateIDFunc(resourceName),
			},
			{
				Config: testAccEndpointConfig_authType(rName, string(awstypes.AuthTypeIamAuth)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "auth_type", string(awstypes.AuthTypeIamAuth)),
				),
			},
		},
	})
}

func TestAccLambdaWebEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceEndpoint, resourceName),
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

func TestAccLambdaWebEndpoint_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypeMultiRegion)),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment_mode", string(awstypes.AutoDeploymentModeDisabled)),
					resource.TestCheckResourceAttr(resourceName, "revision_weights.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revision_weights.0.weight", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEndpointImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccLambdaWebEndpoint_perRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint lambdaweb.GetWebFunctionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_typeWithRegions(rName, string(awstypes.EndpointTypePerRegion)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, string(awstypes.EndpointTypePerRegion)),
					// PerRegion endpoints expose one independent domain per region.
					resource.TestCheckResourceAttr(resourceName, "regional_domain_names.%", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEndpointImportStateIDFunc(resourceName),
			},
		},
	})
}

func testAccCheckEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdaweb_endpoint" {
				continue
			}

			_, err := tflambdaweb.FindEndpointByName(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["endpoint_name"])
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return smarterr.NewError(err)
			}

			return fmt.Errorf("Web Functions Endpoint %s still exists", rs.Primary.Attributes["endpoint_name"])
		}

		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, t *testing.T, n string, v *lambdaweb.GetWebFunctionEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		out, err := tflambdaweb.FindEndpointByName(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["endpoint_name"])
		if err != nil {
			return smarterr.NewError(err)
		}

		*v = *out

		return nil
	}
}

func testAccEndpointImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes["endpoint_name"]), nil
	}
}

func testAccEndpointConfig_basic(rName string) string {
	return testAccEndpointConfig_authType(rName, "ApplicationManaged")
}

func testAccEndpointConfig_authType(rName, authType string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), fmt.Sprintf(`
resource "aws_lambdaweb_endpoint" "test" {
  function_name = aws_lambdaweb_function.test.function_name
  endpoint_name = "extra"
  endpoint_type = "HomeRegion"
  auth_type     = %[1]q
}
`, authType))
}

func testAccEndpointConfig_multiRegion(rName string) string {
	return testAccEndpointConfig_typeWithRegions(rName, "MultiRegion")
}

func testAccEndpointConfig_typeWithRegions(rName, endpointType string) string {
	// Pick a second region distinct from the test region. Regions with the
	// service active as of July 2026: us-east-1 and eu-west-1 (us-west-2
	// rollout pending; PreCheckRegion admits it for when it lands).
	secondRegion := endpoints.EuWest1RegionID
	if acctest.Region() == endpoints.EuWest1RegionID {
		secondRegion = endpoints.UsEast1RegionID
	}
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), fmt.Sprintf(`
resource "aws_lambdaweb_endpoint" "test" {
  function_name        = aws_lambdaweb_function.test.function_name
  endpoint_name        = "multi"
  endpoint_type        = %[3]q
  regions              = [%[1]q, %[2]q]
  auth_type            = "ApplicationManaged"
  auto_deployment_mode = "Disabled"

  revision_weights {
    revision_id = aws_lambdaweb_function.test.latest_revision_id
    weight      = 100
  }
}
`, secondRegion, acctest.Region(), endpointType))
}
