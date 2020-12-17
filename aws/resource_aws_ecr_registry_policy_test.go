package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrRegistryPolicy_basic(t *testing.T) {
	resourceName := "aws_ecr_registry_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRegistryPolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRegistryPolicyExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrRegistryPolicyUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRegistryPolicyExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAWSEcrRegistryPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_registry_policy" {
			continue
		}

		_, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSEcrRegistryPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSEcrRegistryPolicy() string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
			"Principal":{
                "AWS":"arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
            },
            "Action": [
                "ecr:ReplicateImage"
			],
			"Resource": [
                "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
            ]
        }
    ]
}
EOF
}
`)
}

func testAccAWSEcrRegistryPolicyUpdated() string {
	return fmt.Sprintf(`
	data "aws_caller_identity" "current" {}

	data "aws_region" "current" {}
	
	data "aws_partition" "current" {}
	
	resource "aws_ecr_registry_policy" "test" {
	  policy = <<EOF
	{
		"Version":"2012-10-17",
		"Statement": [
			{
				"Sid": "testpolicy",
				"Effect": "Allow",
				"Principal":{
					"AWS":"arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
				},
				"Action": [
					"ecr:ReplicateImage",
					"ecr:CreateRepository"
				],
				"Resource": [
					"arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
				]
			}
		]
	}
	EOF
	}
`)
}
