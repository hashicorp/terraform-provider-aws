// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolSigningCertificateDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_cognito_user_pool_signing_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolSigningCertificateDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrCertificate),
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

  lifecycle {
    ignore_changes = [
      provider_details["ActiveEncryptionCertificate"],
    ]
  }
}

data "aws_cognito_user_pool_signing_certificate" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName)
}
