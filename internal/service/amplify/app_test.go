package amplify_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccApp_basic(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckNoResourceAttr(resourceName, "access_token"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "default_domain", regexp.MustCompile(`\.amplifyapp\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_branch_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_branch_auto_deletion", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "iam_service_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "oauth_token"),
					resource.TestCheckResourceAttr(resourceName, "platform", "WEB"),
					resource.TestCheckResourceAttr(resourceName, "production_branch.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository", ""),
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

func testAccApp_disappears(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					acctest.CheckResourceDisappears(acctest.Provider, tfamplify.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccApp_Tags(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
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
				Config: testAccAppConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccApp_AutoBranchCreationConfig(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	credentials := base64.StdEncoding.EncodeToString([]byte("username1:password1"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_autoBranchCreationNoAutoBranchCreation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.1", "*/**"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				Config: testAccAppConfig_autoBranchCreationAutoBranchCreation(rName, credentials),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", credentials),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", "version: 0.1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.ENVVAR1", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "feature/*"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_autoBranchCreationAutoBranchCreationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "1"),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.build_spec", "version: 0.2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.pull_request_environment_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.0.stage", "EXPERIMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.0", "feature/*"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "true"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					// No change is reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_config.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "auto_branch_creation_patterns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_branch_creation", "false"),
				),
			},
		},
	})
}

func testAccApp_BasicAuthCredentials(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basicAuthCredentials(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
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
				Config: testAccAppConfig_basicAuthCredentials(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
				),
			},
		},
	})
}

func testAccApp_BuildSpec(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_buildSpec(rName, "version: 0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_buildSpec(rName, "version: 0.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.2"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					// build_spec is Computed.
					resource.TestCheckResourceAttr(resourceName, "build_spec", "version: 0.2"),
				),
			},
		},
	})
}

func testAccApp_CustomRules(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_customRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.source", "/<*>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.status", "404"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.target", "/index.html"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_customRulesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.condition", "<US>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.source", "/documents"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.status", "302"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.0.target", "/documents/us"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.source", "/<*>"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.status", "200"),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.1.target", "/index.html"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "custom_rule.#", "0"),
				),
			},
		},
	})
}

func testAccApp_Description(t *testing.T) {
	var app1, app2, app3 amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app1),
					resource.TestCheckResourceAttr(resourceName, "description", "description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_description(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app2),
					testAccCheckAppNotRecreated(&app1, &app2),
					resource.TestCheckResourceAttr(resourceName, "description", "description 2"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app3),
					testAccCheckAppRecreated(&app2, &app3),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func testAccApp_EnvironmentVariables(t *testing.T) {
	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_environmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
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
				Config: testAccAppConfig_environmentVariablesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", "2"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
				),
			},
		},
	})
}

func testAccApp_IAMServiceRole(t *testing.T) {
	var app1, app2, app3 amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"
	iamRole1ResourceName := "aws_iam_role.test1"
	iamRole2ResourceName := "aws_iam_role.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_iamServiceRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app1),
					resource.TestCheckResourceAttrPair(resourceName, "iam_service_role_arn", iamRole1ResourceName, "arn")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_iamServiceRoleARNUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app2),
					testAccCheckAppNotRecreated(&app1, &app2),
					resource.TestCheckResourceAttrPair(resourceName, "iam_service_role_arn", iamRole2ResourceName, "arn"),
				),
			},
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app3),
					testAccCheckAppRecreated(&app2, &app3),
					resource.TestCheckResourceAttr(resourceName, "iam_service_role_arn", ""),
				),
			},
		},
	})
}

func testAccApp_Name(t *testing.T) {
	var app amplify.App
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccApp_Repository(t *testing.T) {
	key := "AMPLIFY_GITHUB_ACCESS_TOKEN"
	accessToken := os.Getenv(key)
	if accessToken == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	key = "AMPLIFY_GITHUB_REPOSITORY"
	repository := os.Getenv(key)
	if repository == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var app amplify.App
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, amplify.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_repository(rName, repository, accessToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "access_token", accessToken),
					resource.TestCheckResourceAttr(resourceName, "repository", repository),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// access_token is ignored because AWS does not store access_token and oauth_token
				// See https://docs.aws.amazon.com/sdk-for-go/api/service/amplify/#CreateAppInput
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccCheckAppExists(n string, v *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify App ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

		output, err := tfamplify.FindAppByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAppDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_app" {
			continue
		}

		_, err := tfamplify.FindAppByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify App %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	if acctest.Partition() == "aws-us-gov" {
		t.Skip("AWS Amplify is not supported in GovCloud partition")
	}
}

func testAccCheckAppNotRecreated(before, after *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.AppId), aws.StringValue(after.AppId); before != after {
			return fmt.Errorf("Amplify App (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccCheckAppRecreated(before, after *amplify.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.AppId), aws.StringValue(after.AppId); before == after {
			return fmt.Errorf("Amplify App (%s) not recreated", before)
		}

		return nil
	}
}

func testAccAppConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppConfig_autoBranchCreationNoAutoBranchCreation(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "*",
    "*/**",
  ]
}
`, rName)
}

func testAccAppConfig_autoBranchCreationAutoBranchCreation(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "feature/*",
  ]

  auto_branch_creation_config {
    build_spec = "version: 0.1"
    framework  = "React"
    stage      = "DEVELOPMENT"

    enable_basic_auth      = true
    basic_auth_credentials = %[2]q

    enable_auto_build             = true
    enable_pull_request_preview   = true
    pull_request_environment_name = "test1"

    environment_variables = {
      ENVVAR1 = "1"
    }
  }
}

`, rName, basicAuthCredentials)
}

func testAccAppConfig_autoBranchCreationAutoBranchCreationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  enable_auto_branch_creation = true

  auto_branch_creation_patterns = [
    "feature/*",
  ]

  auto_branch_creation_config {
    build_spec = "version: 0.2"
    framework  = "React"
    stage      = "EXPERIMENTAL"

    enable_basic_auth = false

    enable_auto_build           = false
    enable_pull_request_preview = false

    pull_request_environment_name = "test2"
  }
}
`, rName)
}

func testAccAppConfig_basicAuthCredentials(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  basic_auth_credentials = %[2]q
  enable_basic_auth      = true
}
`, rName, basicAuthCredentials)
}

func testAccAppConfig_buildSpec(rName, buildSpec string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  build_spec = %[2]q
}
`, rName, buildSpec)
}

func testAccAppConfig_customRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  custom_rule {
    source = "/<*>"
    status = "404"
    target = "/index.html"
  }
}
`, rName)
}

func testAccAppConfig_customRulesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  custom_rule {
    condition = "<US>"
    source    = "/documents"
    status    = "302"
    target    = "/documents/us"
  }

  custom_rule {
    source = "/<*>"
    status = "200"
    target = "/index.html"
  }
}
`, rName)
}

func testAccAppConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  description = %[2]q
}
`, rName, description)
}

func testAccAppConfig_environmentVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccAppConfig_environmentVariablesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}

func testAccAppIAMServiceRoleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
  name = "%[1]s-1"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "amplify.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "amplify.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}
`, rName)
}

func testAccAppConfig_iamServiceRoleARN(rName string) string {
	return acctest.ConfigCompose(testAccAppIAMServiceRoleBaseConfig(rName), fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  iam_service_role_arn = aws_iam_role.test1.arn
}
`, rName))
}

func testAccAppConfig_iamServiceRoleARNUpdated(rName string) string {
	return acctest.ConfigCompose(testAccAppIAMServiceRoleBaseConfig(rName), fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  iam_service_role_arn = aws_iam_role.test2.arn
}
`, rName))
}

func testAccAppConfig_repository(rName, repository, accessToken string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q

  repository   = %[2]q
  access_token = %[3]q
}
`, rName, repository, accessToken)
}
