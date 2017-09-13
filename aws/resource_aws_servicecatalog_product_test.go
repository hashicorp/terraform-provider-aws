package aws

import (
	"github.com/hashicorp/terraform/helper/resource"

	"testing"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogProductResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_description", "ad"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "artifact_name", "an"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "description", "test"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "distributor", "disco"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", "test1234"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "owner", "brett"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "product_type", "CLOUD_FORMATION_TEMPLATE"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_description", "a test"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_email", "mailid@mail.com"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "support_url", "https://url/support.html"),
				),
			},
		},
	})
}

const testAccCheckAwsServiceCatalogProductResourceConfig_basic = `
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
  description = "test"
  distributor = "disco"
  name = "test1234"
  owner = "brett"
  product_type = "CLOUD_FORMATION_TEMPLATE"
  support_description = "a test"
  support_email = "mailid@mail.com"
  support_url = "https://url/support.html"
}
`
