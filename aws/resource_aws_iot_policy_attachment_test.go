package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTPolicyAttachment_basic(t *testing.T) {
	policyName := acctest.RandomWithPrefix("PolicyName-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTPolicyAttchmentDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTPolicyAttachmentConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTPolicyAttachmentExists("aws_iot_policy_attachment.att"),
				),
			},
		},
	})

}

func testAccCheckAWSIoTPolicyAttchmentDestroy_basic(s *terraform.State) error {
	return nil
}

func testAccCheckAWSIoTPolicyAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		target := rs.Primary.Attributes["target"]
		policyName := rs.Primary.Attributes["policy"]

		out, err := conn.ListAttachedPolicies(&iot.ListAttachedPoliciesInput{
			Target:   aws.String(target),
			PageSize: aws.Int64(250),
		})

		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for target %s (%s)", target, n)
		}
		if len(out.Policies) != 1 {
			return fmt.Errorf("Error: Target (%s) has wrong number of policies attached on initial creation", target)
		}

		attPolicyName := aws.StringValue(out.Policies[0].PolicyName)

		if policyName != attPolicyName {
			return fmt.Errorf("Error: Target (%s) has wrong policy attached, expected %s, got %s", target, policyName, attPolicyName)
		}

		return nil
	}
}

func testAccAWSIoTPolicyAttachmentConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_policy" "policy" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iot_policy_attachment" "att" {
  policy = "${aws_iot_policy.policy.name}"
  target = "${aws_iot_certificate.cert.arn}"
}
`, policyName)
}
