package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDlmLifecyclePolicyBasic(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.basic"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "basic"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDlmLifecyclePolicyFull(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.full"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyFullConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "21:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full"),
				),
			},
			{
				Config: dlmLifecyclePolicyFullUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "24"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "09:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "100"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full-updated"),
				),
			},
		},
	})
}

func dlmLifecyclePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dlmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dlm_lifecycle_policy" {
			continue
		}

		input := dlm.GetLifecyclePolicyInput{
			PolicyId: aws.String(rs.Primary.Attributes["name"]),
		}

		out, err := conn.GetLifecyclePolicy(&input)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}

		if out.Policy != nil {
			return fmt.Errorf("DLM lifecycle policy still exists: %#v", out)
		}
	}

	return nil
}

func checkDlmLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func dlmLifecyclePolicyBasicConfig() string {
	return fmt.Sprint(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = "tf-acc-basic-dlm-lifecycle-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "basic" {
  description        = "tf-acc-basic"
  execution_role_arn = "${aws_iam_role.dlm_lifecycle_role.arn}"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval      = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func dlmLifecyclePolicyFullConfig() string {
	return fmt.Sprint(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = "tf-acc-full-dlm-lifecycle-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "full" {
  description        = "tf-acc-full"
  execution_role_arn = "${aws_iam_role.dlm_lifecycle_role.arn}"
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full"

      create_rule {
        interval      = 12
        interval_unit = "HOURS"
        times         = ["21:42"]
      }

      retain_rule {
        count = 10
      }

      tags_to_add {
        tf-acc-test-added = "full"
      }
    }

    target_tags {
      tf-acc-test = "full"
    }
  }
}
`)
}

func dlmLifecyclePolicyFullUpdateConfig() string {
	return fmt.Sprint(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = "tf-acc-full-dlm-lifecycle-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "full" {
  description        = "tf-acc-full-updated"
  execution_role_arn = "${aws_iam_role.dlm_lifecycle_role.arn}-doesnt-exist"
  state              = "DISABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full-updated"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["09:42"]
      }

      retain_rule {
        count = 100
      }

      tags_to_add {
        tf-acc-test-added = "full-updated"
      }
    }

    target_tags {
      tf-acc-test = "full-updated"
    }
  }
}
`)
}
