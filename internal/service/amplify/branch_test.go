package amplify_test

import (
	"encoding/base64"
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

func testAccBranch_basic(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchNameConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/branches/.+`)),
					resource.TestCheckResourceAttr(resourceName, "associated_resources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "backend_environment_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
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

func testAccBranch_disappears(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					acctest.CheckResourceDisappears(acctest.Provider, tfamplify.ResourceBranch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBranch_Tags(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
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
				Config: testAccBranchTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBranchTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccBranch_BasicAuthCredentials(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchBasicAuthCredentialsConfig(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
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
				Config: testAccBranchBasicAuthCredentialsConfig(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				Config: testAccBranchNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
				),
			},
		},
	})
}

func testAccBranch_EnvironmentVariables(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_branch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchEnvironmentVariablesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
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
				Config: testAccBranchEnvironmentVariablesUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", "2"),
				),
			},
			{
				Config: testAccBranchNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
				),
			},
		},
	})
}

func testAccBranch_OptionalArguments(t *testing.T) {
	var branch amplify.Branch
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	environmentName := sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlpha)
	resourceName := "aws_amplify_branch.test"
	backendEnvironment1ResourceName := "aws_amplify_backend_environment.test1"
	backendEnvironment2ResourceName := "aws_amplify_backend_environment.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBranchOptionalArgumentsConfig(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "testdisplayname1"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr1"),
					resource.TestCheckResourceAttr(resourceName, "stage", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBranchOptionalArgumentsUpdatedConfig(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription2"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "testdisplayname2"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "true"),
					resource.TestCheckResourceAttr(resourceName, "framework", "Angular"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr2"),
					resource.TestCheckResourceAttr(resourceName, "stage", "EXPERIMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "15"),
				),
			},
		},
	})
}

func testAccCheckBranchExists(resourceName string, v *amplify.Branch) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

		branch, err := tfamplify.FindBranchByAppIDAndBranchName(conn, appID, branchName)

		if err != nil {
			return err
		}

		*v = *branch

		return nil
	}
}

func testAccCheckBranchDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_branch" {
			continue
		}

		appID, branchName, err := tfamplify.BranchParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfamplify.FindBranchByAppIDAndBranchName(conn, appID, branchName)

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

func testAccBranchNameConfig(rName string) string {
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

func testAccBranchTags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccBranchTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccBranchBasicAuthCredentialsConfig(rName, basicAuthCredentials string) string {
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

func testAccBranchEnvironmentVariablesConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccBranchEnvironmentVariablesUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}

func testAccBranchOptionalArgumentsConfig(rName, environmentName string) string {
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

  backend_environment_arn       = aws_amplify_backend_environment.test1.arn
  description                   = "testdescription1"
  display_name                  = "testdisplayname1"
  enable_auto_build             = false
  enable_notification           = true
  enable_performance_mode       = true
  enable_pull_request_preview   = false
  framework                     = "React"
  pull_request_environment_name = "testpr1"
  stage                         = "DEVELOPMENT"
  ttl                           = "10"
}
`, rName, environmentName)
}

func testAccBranchOptionalArgumentsUpdatedConfig(rName, environmentName string) string {
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

  backend_environment_arn       = aws_amplify_backend_environment.test2.arn
  description                   = "testdescription2"
  display_name                  = "testdisplayname2"
  enable_auto_build             = true
  enable_notification           = false
  enable_performance_mode       = true
  enable_pull_request_preview   = true
  framework                     = "Angular"
  pull_request_environment_name = "testpr2"
  stage                         = "EXPERIMENTAL"
  ttl                           = "15"
}
`, rName, environmentName)
}
