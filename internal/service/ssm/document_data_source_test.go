package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMDocumentDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_document.test"
	name := fmt.Sprintf("test_document-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDocumentDataSourceConfig(name, "JSON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ssm_document.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", "aws_ssm_document.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "document_format", "aws_ssm_document.test", "document_format"),
					resource.TestCheckResourceAttr(resourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "document_type", "aws_ssm_document.test", "document_type"),
					resource.TestCheckResourceAttrPair(resourceName, "content", "aws_ssm_document.test", "content"),
				),
			},
			{
				Config: testAccCheckDocumentDataSourceConfig(name, "YAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ssm_document.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", "aws_ssm_document.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
					resource.TestCheckResourceAttr(resourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "document_type", "aws_ssm_document.test", "document_type"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
				),
			},
		},
	})
}

func TestAccSSMDocumentDataSource_awsManaged(t *testing.T) {
	resourceName := "data.aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDocumentDataSourceAWSManagedDocumentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "AWS-StartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "arn", "AWS-StartEC2Instance"),
				),
			},
		},
	})
}

func testAccCheckDocumentDataSourceConfig(name, documentFormat string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}

data "aws_ssm_document" "test" {
  name            = aws_ssm_document.test.name
  document_format = "%s"
}
`, name, documentFormat)
}

func testAccCheckDocumentDataSourceAWSManagedDocumentConfig() string {
	return `
data "aws_ssm_document" "test" {
  name = "AWS-StartEC2Instance"
}
`
}
