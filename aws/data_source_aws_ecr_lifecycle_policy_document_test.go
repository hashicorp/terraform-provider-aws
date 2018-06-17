package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceECRLifecyclePolicyDocument_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSECRLifecyclePolicyDocumentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStateValue("data.aws_ecr_lifecycle_policy_document.test", "json",
						testAccAWSECRLifecyclePolicyDocumentExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceECRLifecyclePolicyDocument_source(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSECRLifecyclePolicyDocumentSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStateValue("data.aws_ecr_lifecycle_policy_document.test_source", "json",
						testAccAWSECRLifecyclePolicyDocumentSourceExpectedJSON,
					),
				),
			},
			{
				Config: testAccAWSECRLifecyclePolicyDocumentSourceBlankConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStateValue("data.aws_ecr_lifecycle_policy_document.test_source_blank", "json",
						testAccAWSECRLifecyclePolicyDocumentSourceBlankExpectedJSON,
					),
				),
			},
		},
	})
}

var testAccAWSECRLifecyclePolicyDocumentConfig = `
data "aws_ecr_lifecycle_policy_document" "test" {
  rule {
    priority = 1
    description = "Expire images older than 14 days"

    selection {
      tag_status = "untagged"
      count_type = "sinceImagePushed"
      count_unit = "days"
      count_number = 14
    }
  }
}
`

var testAccAWSECRLifecyclePolicyDocumentExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire images older than 14 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 14
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`

var testAccAWSECRLifecyclePolicyDocumentSourceConfig = `
data "aws_ecr_lifecycle_policy_document" "test" {
  rule {
    priority = 1
    description = "Expire images older than 14 days"

    selection {
      tag_status = "untagged"
      count_type = "sinceImagePushed"
      count_unit = "days"
      count_number = 14
    }
  }
}

data "aws_ecr_lifecycle_policy_document" "test_source" {
  source_json = "${data.aws_ecr_lifecycle_policy_document.test.json}"

  rule {
    priority = 2
    description = "Keep last 30 images"

    selection {
      tag_status = "tagged"
      tag_prefixes = ["v"]
      count_type = "imageCountMoreThan"
      count_number = 30
    }
  }
}
`
var testAccAWSECRLifecyclePolicyDocumentSourceExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire images older than 14 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 14
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 2,
      "description": "Keep last 30 images",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "v"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 30
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`

var testAccAWSECRLifecyclePolicyDocumentSourceBlankConfig = `
data "aws_ecr_lifecycle_policy_document" "test_source_blank" {
  source_json = ""

  rule {
    priority = 2
    description = "Keep last 30 images"

    selection {
      tag_status = "tagged"
      tag_prefixes = ["v"]
      count_type = "imageCountMoreThan"
      count_number = 30
    }
  }
}
`
var testAccAWSECRLifecyclePolicyDocumentSourceBlankExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 2,
      "description": "Keep last 30 images",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": [
          "v"
        ],
        "countType": "imageCountMoreThan",
        "countNumber": 30
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}`
