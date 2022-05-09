package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMSAMLProviderDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityID := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	dataSourceName := "data.aws_iam_saml_provider.test"
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderDataSourceConfig(rName, idpEntityID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "create_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.#", resourceName, "tags.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "valid_util", resourceName, "valid_util"),
				),
			},
		},
	})
}

func testAccSAMLProviderDataSourceConfig(rName, idpEntityID string) string {
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
