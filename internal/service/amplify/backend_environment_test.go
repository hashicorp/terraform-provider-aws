// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBackendEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var env types.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentConfig_basic(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(ctx, resourceName, &env),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_artifacts"),
					resource.TestCheckResourceAttr(resourceName, "environment_name", environmentName),
					resource.TestCheckResourceAttrSet(resourceName, "stack_name"),
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

func testAccBackendEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var env types.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentConfig_basic(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(ctx, resourceName, &env),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamplify.ResourceBackendEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBackendEnvironment_DeploymentArtifacts_StackName(t *testing.T) {
	ctx := acctest.Context(t)
	var env types.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentConfig_deploymentArtifactsAndStackName(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(ctx, resourceName, &env),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "amplify", regexache.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_artifacts", rName),
					resource.TestCheckResourceAttr(resourceName, "environment_name", environmentName),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
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

func testAccCheckBackendEnvironmentExists(ctx context.Context, resourceName string, v *types.BackendEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		output, err := tfamplify.FindBackendEnvironmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes["environment_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBackendEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_amplify_backend_environment" {
				continue
			}

			_, err := tfamplify.FindBackendEnvironmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes["environment_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amplify Backend Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBackendEnvironmentConfig_basic(rName string, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = %[2]q
}
`, rName, environmentName)
}

func testAccBackendEnvironmentConfig_deploymentArtifactsAndStackName(rName string, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = %[2]q

  deployment_artifacts = %[1]q
  stack_name           = %[1]q
}
`, rName, environmentName)
}
