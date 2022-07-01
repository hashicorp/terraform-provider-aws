package cognitoidentity_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidentity "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidentity"
)

func TestAccCognitoIdentityPoolProviderPrincipalTags_basic(t *testing.T) {
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolProviderPrincipalTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", "value"),
				),
			},
		},
	})
}
func TestAccCognitoIdentityPoolProviderPrincipalTags_updated(t *testing.T) {
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolProviderPrincipalTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolProviderPrincipalTagsConfig_tagsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_pool_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.test", "value"),
					resource.TestCheckResourceAttr(resourceName, "principal_tags.new", "map"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolProviderPrincipalTags_disappears(t *testing.T) {
	resourceName := "aws_cognito_identity_pool_provider_principal_tag.test"
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolProviderPrincipalTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolProviderPrincipalTagsConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolProviderPrincipalTagsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidentity.ResourcePoolProviderPrincipalTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPoolProviderPrincipalTagsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Princpal Tags is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityConn

		_, err := conn.GetPrincipalTagAttributeMap(&cognitoidentity.GetPrincipalTagAttributeMapInput{
			IdentityPoolId:       aws.String(rs.Primary.Attributes["identity_pool_id"]),
			IdentityProviderName: aws.String(rs.Primary.Attributes["identity_provider_name"]),
		})

		return err
	}
}

func testAccCheckPoolProviderPrincipalTagsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_pool_provider_principal_tag" {
			continue
		}

		_, err := conn.GetPrincipalTagAttributeMap(&cognitoidentity.GetPrincipalTagAttributeMapInput{
			IdentityPoolId:       aws.String(rs.Primary.Attributes["identity_pool_id"]),
			IdentityProviderName: aws.String(rs.Primary.Attributes["identity_provider_name"]),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, cognitoidentity.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccPoolProviderPrincipalTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
  supported_identity_providers = compact([
    "COGNITO",
  ])
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false
  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.test.id
    provider_name           = aws_cognito_user_pool.test.endpoint
    server_side_token_check = false
  }
  supported_login_providers = {
    "accounts.google.com" = "new-client-id-url.apps.googleusercontent.com"
  }
}
`, name)
}

func testAccPoolProviderPrincipalTagsConfig_basic(name string) string {
	return acctest.ConfigCompose(testAccPoolProviderPrincipalTagsConfig(name), `
resource "aws_cognito_identity_pool_provider_principal_tag" "test" {
  identity_pool_id       = aws_cognito_identity_pool.test.id
  identity_provider_name = aws_cognito_user_pool.test.endpoint
  use_defaults           = false
  principal_tags = {
    test = "value"
  }
}
`)
}

func testAccPoolProviderPrincipalTagsConfig_tagsUpdated(name string) string {
	return acctest.ConfigCompose(testAccPoolProviderPrincipalTagsConfig(name), `
resource "aws_cognito_identity_pool_provider_principal_tag" "test" {
  identity_pool_id       = aws_cognito_identity_pool.test.id
  identity_provider_name = aws_cognito_user_pool.test.endpoint
  use_defaults           = false
  principal_tags = {
    test = "value"
    new  = "map"
  }
}
`)
}
