package aws

import (
	"github.com/hashicorp/terraform/helper/resource"

	"testing"

	"bytes"
	"text/template"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	template := template.Must(template.New("hcl").Parse(testAccCheckAwsServiceCatalogProductResourceConfig_basic_tempate))
	var template1, template2 bytes.Buffer
	template.Execute(&template1, Input{"dsc1", "dst1", "nm1", "own1", "sd1", "a@b.com", "https://url/support1.html"})
	template.Execute(&template2, Input{"dsc2", "dst2", "nm2", "own2", "sd2", "c@d.com", "https://url/support2.html"})

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: template1.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "nm1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "a@b.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support1.html"),
				),
			},
			resource.TestStep{
				Config: template2.String(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "dsc2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "dst2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "nm2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "own2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "sd2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "c@d.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support2.html"),
				),
			},
		},
	})
}

type Input struct {
	Description        string
	Distributor        string
	Name               string
	Owner              string
	SupportDescription string
	SupportEmail       string
	SupportUrl         string
}

const testAccCheckAwsServiceCatalogProductResourceConfig_basic_tempate = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

resource "aws_s3_bucket" "bucket" {
	bucket = "deving-me-some-tf-sc-${data.aws_caller_identity.current.account_id}-${var.region}"
	region = "${var.region}"
	acl    = "private"
	force_destroy = true
}

resource "aws_s3_bucket_object" "template" {
  bucket = "${aws_s3_bucket.bucket.id}"
  key = "test_templates_for_terraform_sc_dev.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_product" "test" {
  artifact_description = "ad"
  artifact_name = "an"
  cloud_formation_template_url = "https://s3-${var.region}.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template.key}"
  description = "{{.Description}}"
  distributor = "{{.Distributor}}"
  name = "{{.Name}}"
  owner = "{{.Owner}}"
  product_type = "CLOUD_FORMATION_TEMPLATE"
  support_description = "{{.SupportDescription}}"
  support_email = "{{.SupportEmail}}"
  support_url = "{{.SupportUrl}}"
}
`
