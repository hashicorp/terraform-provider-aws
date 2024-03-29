package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcrLifecyclePolicyValue("data.aws_ecr_lifecycle_policy_document.test", "json",
						testAccAWSEcrLifecyclePolicyDocumentExpectedJSON,
					),
				),
			},
		},
	})
}

func testAccCheckEcrLifecyclePolicyValue(id, name, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not found: %s", id)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		v := rs.Primary.Attributes[name]
		if v != value {
			return fmt.Errorf(
				"Value for %s is %s, not %s", name, v, value)
		}

		return nil
	}
}

var testAccAWSEcrLifecyclePolicyDocumentConfig = `
data "aws_ecr_lifecycle_policy_document" "test" {
    rule {
      priority        = 1
      description     = "This is a test."
      tag_status      = "tagged"
      tag_prefix_list = ["prod"]
      count_type      = "imageCountMoreThan"
      count_number    = 100
    }
}
`

var testAccAWSEcrLifecyclePolicyDocumentExpectedJSON = `{
  "rules": [
    {
      "rulePriority": 1,
      "description": "This is a test.",
      "selection": [
        "tagStatus": "tagged",
        "tagPrefixList": ["prod"],
        "countType": "imageCountMoreThan",
        "countNumber": 100
      ],
      "action": [
        "type": "expire"
      ]
    }
  ]
}`
