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

// Note: updating 'build_spec' does not work on AWS side
func TestAccAWSAmplifyBranch_basic(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	branchName := "master"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/branches/[^/]+$")),
					resource.TestCheckResourceAttr(resourceName, "branch_name", branchName),
					resource.TestCheckResourceAttr(resourceName, "build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "display_name", branchName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "framework", ""),
					resource.TestCheckResourceAttr(resourceName, "stage", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "associated_resources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "destination_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "source_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
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
		},
	})
}

func TestAccAWSAmplifyBranch_rename(t *testing.T) {
	var branch1, branch2 amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	branchName1 := "master"
	branchName2 := "development"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBranch(rName, branchName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch1),
					resource.TestCheckResourceAttr(resourceName, "branch_name", branchName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigBranch(rName, branchName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch2),
					testAccCheckAWSAmplifyBranchRecreated(&branch1, &branch2),
					resource.TestCheckResourceAttr(resourceName, "branch_name", branchName2),
				),
			},
		},
	})
}

func TestAccAWSAmplifyBranch_simple(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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

func TestAccAWSAmplifyBranch_basicAuthConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	username1 := "username1"
	password1 := "password1"
	username2 := "username2"
	password2 := "password2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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

func TestAccAWSAmplifyBranch_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigTags1(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TAG1", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigTags2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TAG1", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TAG2", "2"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSAmplifyBranchExists(resourceName string, v *amplify.Branch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		id := strings.Split(rs.Primary.ID, "/")
		app_id := id[0]
		branch := id[2]

		output, err := conn.GetBranch(&amplify.GetBranchInput{
			AppId:      aws.String(app_id),
			BranchName: aws.String(branch),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Branch == nil {
			return fmt.Errorf("Amplify Branch (%s) not found", rs.Primary.ID)
		}

		*v = *output.Branch

		return nil
	}
}

func testAccCheckAWSAmplifyBranchDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_branch" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		id := strings.Split(rs.Primary.ID, "/")
		app_id := id[0]
		branch := id[2]

		_, err := conn.GetBranch(&amplify.GetBranchInput{
			AppId:      aws.String(app_id),
			BranchName: aws.String(branch),
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

func testAccCheckAWSAmplifyBranchRecreated(i, j *amplify.Branch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) == aws.TimeValue(j.CreateTime) {
			return errors.New("Amplify Branch was not recreated")
		}

		return nil
	}
}

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

func testAccAWSAmplifyBranchConfigTags1(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  tags = {
    TAG1 = "1",
  }
}
`, rName)
}

func testAccAWSAmplifyBranchConfigTags2(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"

  tags = {
    TAG1 = "2",
    TAG2 = "2",
  }
}
`, rName)
}
