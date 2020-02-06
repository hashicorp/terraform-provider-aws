package aws

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAmplifyBackendEnvironment_basic(t *testing.T) {
	var env1, env2 amplify.BackendEnvironment
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_backend_environment.test"

	envName := "backend"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfig_Required(rName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env1),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/backendenvironments/"+envName)),
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
					testAccCheckAWSAmplifyBackendEnvironmentRecreated(&env1, &env2),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/backendenvironments/"+envName)),
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

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		id := strings.Split(rs.Primary.ID, "/")
		app_id := id[0]
		environment_name := id[2]

		output, err := conn.GetBackendEnvironment(&amplify.GetBackendEnvironmentInput{
			AppId:           aws.String(app_id),
			EnvironmentName: aws.String(environment_name),
		})
		if err != nil {
			return err
		}

		if output == nil || output.BackendEnvironment == nil {
			return fmt.Errorf("Amplify BackendEnvironment (%s) not found", rs.Primary.ID)
		}

		*v = *output.BackendEnvironment

		return nil
	}
}

func testAccCheckAWSAmplifyBackendEnvironmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_backend_environment" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		id := strings.Split(rs.Primary.ID, "/")
		app_id := id[0]
		environment_name := id[2]

		_, err := conn.GetBackendEnvironment(&amplify.GetBackendEnvironmentInput{
			AppId:           aws.String(app_id),
			EnvironmentName: aws.String(environment_name),
		})

		if isAWSErr(err, amplify.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSAmplifyBackendEnvironmentRecreated(i, j *amplify.BackendEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) == aws.TimeValue(j.CreateTime) {
			return errors.New("Amplify BackendEnvironment was not recreated")
		}

		return nil
	}
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
