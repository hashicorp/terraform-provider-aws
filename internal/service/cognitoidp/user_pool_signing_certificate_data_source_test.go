package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIDPUserPoolSigningCertificateDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_cognito_user_pool_signing_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSigningCertificateDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "certificate"),
				),
			},
		},
	})
}

func testAccUserPoolSigningCertificateDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name                     = %[1]q
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "test" {
  user_pool_id  = aws_cognito_user_pool.test.id
  provider_name = "SAML"
  provider_type = "SAML"

  provider_details = {
    MetadataFile          = file("./test-fixtures/saml-metadata.xml")
    SSORedirectBindingURI = "https://terraform-dev-ed.my.salesforce.com/idp/endpoint/HttpRedirect"
  }

  attribute_mapping = {
    email = "email"
  }
}

data "aws_cognito_user_pool_signing_certificate" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName)
}
