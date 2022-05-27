package amplify_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccBackendEnvironment_basic(t *testing.T) {
	var env amplify.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentBasicConfig(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(resourceName, &env),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
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
	var env amplify.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentBasicConfig(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(resourceName, &env),
					acctest.CheckResourceDisappears(acctest.Provider, tfamplify.ResourceBackendEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBackendEnvironment_DeploymentArtifacts_StackName(t *testing.T) {
	var env amplify.BackendEnvironment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackendEnvironmentDeploymentArtifactsAndStackNameConfig(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendEnvironmentExists(resourceName, &env),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
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

func testAccCheckBackendEnvironmentExists(resourceName string, v *amplify.BackendEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Backend Environment ID is set")
		}

		appID, environmentName, err := tfamplify.BackendEnvironmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

		backendEnvironment, err := tfamplify.FindBackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

		if err != nil {
			return err
		}

		*v = *backendEnvironment

		return nil
	}
}

func testAccCheckBackendEnvironmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_backend_environment" {
			continue
		}

		appID, environmentName, err := tfamplify.BackendEnvironmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfamplify.FindBackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

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

func testAccBackendEnvironmentBasicConfig(rName string, environmentName string) string {
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

func testAccBackendEnvironmentDeploymentArtifactsAndStackNameConfig(rName string, environmentName string) string {
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
