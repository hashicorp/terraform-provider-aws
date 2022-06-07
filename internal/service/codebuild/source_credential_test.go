package codebuild_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/codebuild"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodeBuildSourceCredential_basic(t *testing.T) {
	var sourceCredentialsInfo codebuild.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSourceCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/github`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB_ENTERPRISE", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/github_enterprise`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB_ENTERPRISE"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "user_name"},
			},
		},
	})
}

func TestAccCodeBuildSourceCredential_basicAuth(t *testing.T) {
	var sourceCredentialsInfo codebuild.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSourceCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basicAuth(token, "user1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/bitbucket`)),
					resource.TestCheckResourceAttr(resourceName, "user_name", "user1"),
					resource.TestCheckResourceAttr(resourceName, "server_type", "BITBUCKET"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "BASIC_AUTH"),
				),
			},
			{
				Config: testAccSourceCredentialConfig_basicAuth(token, "user2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					resource.TestCheckResourceAttr(resourceName, "user_name", "user2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "user_name"},
			},
		},
	})
}

func TestAccCodeBuildSourceCredential_disappears(t *testing.T) {
	var sourceCredentialsInfo codebuild.SourceCredentialsInfo
	token := sdkacctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSourceCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSourceCredentialConfig_basic("PERSONAL_ACCESS_TOKEN", "GITHUB_ENTERPRISE", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceSourceCredential(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceSourceCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSourceCredentialDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_source_credential" {
			continue
		}

		_, err := tfcodebuild.FindSourceCredentialByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodeBuild Source Credential %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckSourceCredentialExists(name string, sourceCredential *codebuild.SourceCredentialsInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

		output, err := tfcodebuild.FindSourceCredentialByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("CodeBuild Source Credential (%s) not found", rs.Primary.ID)
		}

		*sourceCredential = *output

		return nil
	}
}

func testAccSourceCredentialConfig_basic(authType, serverType, token string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "%s"
  server_type = "%s"
  token       = "%s"
}
`, authType, serverType, token)
}

func testAccSourceCredentialConfig_basicAuth(token, userName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "BASIC_AUTH"
  server_type = "BITBUCKET"
  token       = "%s"
  user_name   = "%s"
}
`, token, userName)
}
