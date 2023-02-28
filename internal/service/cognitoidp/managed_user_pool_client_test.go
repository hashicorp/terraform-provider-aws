package cognitoidp_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

const (
	openSearchDomainMaxLen = 28

	openSearchDomainPrefix    = "tf-acc-"
	openSearchDomainPrefixLen = len(openSearchDomainPrefix)

	openSearchDomainRemainderLen = openSearchDomainMaxLen - openSearchDomainPrefixLen
)

func randomOpenSearchDomainName() string {
	return fmt.Sprintf(openSearchDomainPrefix+"%s", sdkacctest.RandString(openSearchDomainRemainderLen))
}

func TestAccCognitoIDPManagedUserPoolClient_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var client cognitoidentityprovider.UserPoolClientType
	rName := randomOpenSearchDomainName()
	resourceName := "aws_cognito_managed_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(fmt.Sprintf(`^AmazonOpenSearchService-%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "0"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.0", "code"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.0", "email"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.1", "openid"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.2", "phone"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.3", "profile"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auth_session_validity", "3"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "callback_urls.0", regexp.MustCompile(fmt.Sprintf(`https://search-%s-\w+.%s.es.amazonaws.com/_dashboards/app/home`, rName, acctest.Region()))),
					resource.TestMatchResourceAttr(resourceName, "client_secret", regexp.MustCompile(`\w+`)),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_propagate_additional_user_context_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "explicit_auth_flows"),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "0"),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "logout_urls.0", regexp.MustCompile(fmt.Sprintf(`https://search-%s-\w+.%s.es.amazonaws.com/_dashboards/app/home`, rName, acctest.Region()))),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", ""),
					resource.TestCheckNoResourceAttr(resourceName, "read_attributes"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "30"),
					resource.TestCheckResourceAttr(resourceName, "supported_identity_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "supported_identity_providers.0", "COGNITO"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", "aws_cognito_user_pool.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func testAccManagedUserPoolClientBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false

  lifecycle {
    ignore_changes = [cognito_identity_providers]
  }
}

resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  engine_version = "OpenSearch_1.1"

  cognito_options {
    enabled          = true
    user_pool_id     = aws_cognito_user_pool.test.id
    identity_pool_id = aws_cognito_identity_pool.test.id
    role_arn         = aws_iam_role.test.arn
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  depends_on = [
    aws_cognito_user_pool_domain.test,
    aws_iam_role_policy_attachment.test,
  ]
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    sid     = ""
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type = "Service"
      identifiers = [
        "es.${data.aws_partition.current.dns_suffix}",
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonESCognitoAccess"
}
`, rName)
}

func testAccManagedUserPoolClientConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccManagedUserPoolClientBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cognito_managed_user_pool_client" "test" {
  name_prefix  = "AmazonOpenSearchService-%[1]s"
  user_pool_id = aws_cognito_user_pool.test.id

  depends_on = [
    aws_opensearch_domain.test,
  ]
}
`, rName))
}
