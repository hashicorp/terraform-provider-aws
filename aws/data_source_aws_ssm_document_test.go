package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSsmDocumentDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_document.test"
	name := "test_document"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmDocumentDataSourceConfig(name, "JSON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:document/%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "document_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "document_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Command"),
					resource.TestCheckResourceAttrPair(resourceName, "content", "aws_ssm_document.test", "content"),
				),
			},
			{
				Config: testAccCheckAwsSsmDocumentDataSourceConfig(name, "YAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:document/%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
					resource.TestCheckResourceAttr(resourceName, "document_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Command"),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
				),
			},
		},
	})
}

func testAccCheckAwsSsmDocumentDataSourceConfig(name string, documentFormat string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "%s"
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

data "aws_ssm_document" "test" {
  name = "${aws_ssm_document.test.name}"
  document_format = "%s"
}
`, name, documentFormat)
}
