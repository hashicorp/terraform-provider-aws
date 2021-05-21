package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func testAccAWSAmplifyBackendEnvironment_basic(t *testing.T) {
	var env1, env2 amplify.BackendEnvironment
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_backend_environment.test"

	envName := "backend"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfig_Required(rName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
					resource.TestCheckResourceAttr(resourceName, "environment_name", envName),
					resource.TestMatchResourceAttr(resourceName, "deployment_artifacts", regexp.MustCompile(fmt.Sprintf("^tf-acc-test-.*-%s-.*-deployment$", envName))),
					resource.TestMatchResourceAttr(resourceName, "stack_name", regexp.MustCompile(fmt.Sprintf("^amplify-tf-acc-test-.*-%s-.*$", envName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfigAll(rName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env2),
					resource.TestCheckResourceAttr(resourceName, "environment_name", envName),
					resource.TestCheckResourceAttr(resourceName, "deployment_artifacts", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
				),
			},
		},
	})
}

func testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName string, v *amplify.BackendEnvironment) resource.TestCheckFunc {
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

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		backendEnvironment, err := finder.BackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

		if err != nil {
			return err
		}

		*v = *backendEnvironment

		return nil
	}
}

func testAccCheckAWSAmplifyBackendEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_backend_environment" {
			continue
		}

		appID, environmentName, err := tfamplify.BackendEnvironmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.BackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify BackendEnvironment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyBackendEnvironmentConfig_Required(rName string, envName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%s"
}
`, rName, envName)
}

func testAccAWSAmplifyBackendEnvironmentConfigAll(rName string, envName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%s"

  deployment_artifacts = "%s"
  stack_name           = "%s"
}
`, rName, envName, rName, rName)
}
