// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSAMLProviderDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	dataSourceName := "data.aws_iam_saml_provider.test"
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderDataSourceConfig_basic(rName, idpEntityID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_provider_uuid", resourceName, "saml_provider_uuid"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.#", resourceName, "tags.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "valid_util", resourceName, "valid_util"),
				),
			},
		},
	})
}

func TestAccIAMSAMLProviderDataSource_trustedIdentityPropagation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName(t))
	dataSourceName := "data.aws_iam_saml_provider.test"
	resourceName := "aws_iam_saml_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderDataSourceConfig_trustedIdentityPropagation(rName, idpEntityID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.subject_source_key", "subject"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.subject_type_value", "user"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.0.source_key", names.AttrEmail),
					resource.TestCheckResourceAttr(dataSourceName, "context_provider.0.key_definition.0.target_key", "Email"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_center_instance_arn", resourceName, "identity_center_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_store_arn", resourceName, "identity_store_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_key_hashes.#", resourceName, "saml_key_hashes.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_key_hashes.0.issuer_url", resourceName, "saml_key_hashes.0.issuer_url"),
				),
			},
		},
	})
}

func testAccSAMLProviderDataSourceConfig_basic(rName, idpEntityID string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    Name = %[1]q
  }
}

data "aws_iam_saml_provider" "test" {
  arn = aws_iam_saml_provider.test.arn
}
`, rName, idpEntityID)
}

func testAccSAMLProviderDataSourceConfig_trustedIdentityPropagation(rName, idpEntityID string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  context_provider {
    subject_source_key = "subject"
    subject_type_value = "user"

    key_definition {
      source_key = "email"
      target_key = "Email"
    }
  }
}

data "aws_iam_saml_provider" "test" {
  arn = aws_iam_saml_provider.test.arn
}
`, rName, idpEntityID)
}
