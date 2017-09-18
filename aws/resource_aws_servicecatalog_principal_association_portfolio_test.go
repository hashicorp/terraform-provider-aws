package aws

import (
	"github.com/hashicorp/terraform/helper/resource"

	"regexp"
	"testing"
)

func TestAccAWSServiceCatalogPrincipalAssociationPortfolio_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckAwsServiceCatalogPrincipalAssociationPortfolioResourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_servicecatalog_principal_association_portfolio.test", "portfolio_id", regexp.MustCompile("^port-.*")),
				),
			},
		},
	})
}

const testAccCheckAwsServiceCatalogPrincipalAssociationPortfolioResourceConfig_basic = `
data "aws_caller_identity" "current" {}
variable region { default = "us-west-2" }

resource "aws_iam_role" "test" {
  name = "test-me-some-role-assoc-for-tf-sc"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_servicecatalog_portfolio" "test" {
  name = "test-1"
  description = "test-2"
  provider_name = "test-3"
}

resource "aws_servicecatalog_principal_association_portfolio" "test" {
	portfolio_id = "${aws_servicecatalog_portfolio.test.id}"
	principal_arn = "${aws_iam_role.test.arn}"
}
`
