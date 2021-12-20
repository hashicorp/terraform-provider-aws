package cognitoidp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIDPUserPoolSigningCertDataSource_basic(t *testing.T) {
	testName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", sdkacctest.RandString(7))
	resourceName := "aws_cognito_user_pool.saml"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSigningCertDataSourceConfig_basic(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolExists(resourceName, nil),
					resource.TestCheckResourceAttrSet("data.aws_cognito_user_pool_signing_certificate.saml", "certificate"),
				),
			},
		},
	})
}

func testAccUserPoolSigningCertDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "saml" {
	name                     = "%s"
	auto_verified_attributes = ["email"]
}
resource "aws_cognito_identity_provider" "saml" {
	user_pool_id  = aws_cognito_user_pool.saml.id
	provider_name = "SAML"
	provider_type = "SAML"
  
	provider_details = {
	  MetadataFile = file("./test-fixtures/saml-metadata.xml")
	  // if we don't specify below, terraform always thinks this resource has
	  // changed: https://github.com/terraform-providers/terraform-provider-aws/issues/4831
	  SSORedirectBindingURI = "https://terraform-dev-ed.my.salesforce.com/idp/endpoint/HttpRedirect"
	}
  
	attribute_mapping = {
	  email = "email"
	}
}
  
data "aws_cognito_user_pool_signing_certificate" "saml" {
	user_pool_id = aws_cognito_user_pool.saml.id
}
`, rName)
}
