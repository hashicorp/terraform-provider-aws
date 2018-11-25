package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceEcrLifecyclePolicyDocument_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrLifecyclePolicyDocumentConfig,
				Check: resource.TestCheckResourceAttr(
					"data.aws_ecr_lifecycle_policy_document.test", "json",
					testAccAWSEcrLifecyclePolicyDocumentExpectedJSON,
				),
			},
		},
	})
}

var testAccAWSEcrLifecyclePolicyDocumentConfig = `
data "aws_ecr_lifecycle_policy_document" "test" {
    rule {
      priority        = 1
      description     = "This is a test."
      selection       = {
        tag_status      = "tagged"
        tag_prefix_list = ["prod"]
        count_type      = "imageCountMoreThan"
        count_number    = 100
      }
    }
}
`

var testAccAWSEcrLifecyclePolicyDocumentExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "This is a test.",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "prod"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 100
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`
