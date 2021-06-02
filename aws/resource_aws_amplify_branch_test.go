package aws

import (
	"encoding/base64"
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

func testAccAWSAmplifyBranch_basic(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/branches/.+`)),
					resource.TestCheckResourceAttr(resourceName, "associated_resources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "backend_environment_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
					resource.TestCheckResourceAttr(resourceName, "build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "framework", ""),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "source_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "stage", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccAWSAmplifyBranch_disappears(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAmplifyBranch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAmplifyBranch_Tags(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSAmplifyBranch_BackendEnvironmentArn(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	environmentName := acctest.RandStringFromCharSet(9, acctest.CharSetAlpha)
	resourceName := "aws_amplify_branch.test"
	backendEnvironment1ResourceName := "aws_amplify_backend_environment.test1"
	backendEnvironment2ResourceName := "aws_amplify_backend_environment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBackendEnvironmentARN(rName, environmentName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment1ResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigBackendEnvironmentARN(rName, environmentName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment2ResourceName, "arn"),
				),
			},
		},
	})
}

func testAccAWSAmplifyBranch_BasicAuthCredentials(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials1),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
				),
			},
		},
	})
}

/*
func TestAccAWSAmplifyBranch_simple(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigSimple(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayname"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "framework", "WEB"),
					resource.TestCheckResourceAttr(resourceName, "stage", "PRODUCTION"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "10"),
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

func TestAccAWSAmplifyBranch_backendEnvironment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBackendEnvironment(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "backend_environment_arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/backendenvironments/prod")),
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

func TestAccAWSAmplifyBranch_pullRequestPreview(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigPullRequestPreview(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "true"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "prod"),
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

func TestAccAWSAmplifyBranch_BasicAuthCredentials(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	username1 := "username1"
	password1 := "password1"
	username2 := "username2"
	password2 := "password2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthConfig(rName, username1, password1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.enable_basic_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.username", username1),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.password", password1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthConfig(rName, username2, password2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.enable_basic_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.username", username2),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.0.password", password2),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "basic_auth_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSAmplifyBranch_notification(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigNotification(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "true"),
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

func TestAccAWSAmplifyBranch_environmentVariables(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigEnvironmentVariables1(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigEnvironmentVariables2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", "2"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
				),
			},
		},
	})
}
*/

func testAccCheckAWSAmplifyBranchExists(resourceName string, v *amplify.Branch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Branch ID is set")
		}

		appID, branchName, err := tfamplify.BranchParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		branch, err := finder.BranchByAppIDAndBranchName(conn, appID, branchName)

		if err != nil {
			return err
		}

		*v = *branch

		return nil
	}
}

func testAccCheckAWSAmplifyBranchDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_branch" {
			continue
		}

		appID, branchName, err := tfamplify.BranchParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.BranchByAppIDAndBranchName(conn, appID, branchName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Branch %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyBranchConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}
`, rName)
}

func testAccAWSAmplifyBranchConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAmplifyBranchConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAmplifyBranchConfigBackendEnvironmentARN(rName, environmentName string, index int) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test1" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sa"
}

resource "aws_amplify_backend_environment" "test2" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sb"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  backend_environment_arn = aws_amplify_backend_environment.test%[3]d.arn
}
`, rName, environmentName, index)
}

func testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  basic_auth_credentials = %[2]q
  enable_basic_auth      = true
}
`, rName, basicAuthCredentials)
}

/*
func testAccAWSAmplifyBranchConfig_Required(rName string) string {
	return testAccAWSAmplifyBranchConfigBranch(rName, "master")
}

func testAccAWSAmplifyBranchConfigBranch(rName string, branchName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%s"
}
`, rName, branchName)
}

func testAccAWSAmplifyBranchConfigSimple(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  description       = "description"
  display_name      = "displayname"
  enable_auto_build = false
  framework         = "WEB"
  stage             = "PRODUCTION"
  ttl               = "10"
}
`, rName)
}

func testAccAWSAmplifyBranchConfigBackendEnvironment(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = "prod"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  backend_environment_arn = aws_amplify_backend_environment.test.arn
}
`, rName)
}

func testAccAWSAmplifyBranchConfigPullRequestPreview(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  enable_pull_request_preview = true
  pull_request_environment_name = "prod"
}
`, rName)
}

func testAccAWSAmplifyBranchConfigBasicAuthConfig(rName string, username, password string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  basic_auth_config {
    enable_basic_auth = true
    username          = "%s"
    password          = "%s"
  }
}
`, rName, username, password)
}

func testAccAWSAmplifyBranchConfigNotification(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  enable_notification = true
}
`, rName)
}

func testAccAWSAmplifyBranchConfigEnvironmentVariables1(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccAWSAmplifyBranchConfigEnvironmentVariables2(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}
*/
