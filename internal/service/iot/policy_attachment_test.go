package iot_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iot_policy_attachment", &resource.Sweeper{
		Name: "aws_iot_policy_attachment",
		F:    sweepPolicyAttachments,
	})
}

func sweepPolicyAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListPoliciesInput{}

	err = conn.ListPoliciesPages(input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			input := &iot.ListTargetsForPolicyInput{
				PolicyName: policy.PolicyName,
			}

			err := conn.ListTargetsForPolicyPages(input, func(page *iot.ListTargetsForPolicyOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, target := range page.Targets {
					r := tfiot.ResourcePolicyAttachment()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s|%s", aws.StringValue(policy.PolicyName), aws.StringValue(target)))
					d.Set("policy", policy.PolicyName)
					d.Set("target", target)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy Attachment for %s: %w", region, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy Attachment for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Policy Attachment for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Policy Attachment sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIotPolicyAttachment_basic(t *testing.T) {
	policyName := sdkacctest.RandomWithPrefix("PolicyName-")
	policyName2 := sdkacctest.RandomWithPrefix("PolicyName2-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyAttchmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att"),
					testAccCheckPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName}),
				),
			},
			{
				Config: testAccPolicyAttachmentUpdate1Config(policyName, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att"),
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName, policyName2}),
				),
			},
			{
				Config: testAccPolicyAttachmentUpdate2Config(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName2}),
				),
			},
			{
				Config: testAccPolicyAttachmentUpdate3Config(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att2"),
					testAccCheckPolicyAttachmentExists("aws_iot_policy_attachment.att3"),
					testAccCheckPolicyAttachmentCertStatus("aws_iot_certificate.cert", []string{policyName2}),
					testAccCheckPolicyAttachmentCertStatus("aws_iot_certificate.cert2", []string{policyName2}),
				),
			},
		},
	})

}

func testAccCheckPolicyAttchmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy_attachment" {
			continue
		}

		target := rs.Primary.Attributes["target"]
		policyName := rs.Primary.Attributes["policy"]

		input := &iot.ListAttachedPoliciesInput{
			PageSize:  aws.Int64(250),
			Recursive: aws.Bool(false),
			Target:    aws.String(target),
		}

		var policy *iot.Policy
		err := tfiot.ListPolicyAttachmentPages(conn, input, func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool {
			for _, att := range out.Policies {
				if policyName == aws.StringValue(att.PolicyName) {
					policy = att
					return false
				}
			}
			return true
		})

		if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "The certificate given in the principal does not exist.") {
			continue
		} else if err != nil {
			return err
		}

		if policy == nil {
			continue
		}

		return fmt.Errorf("IOT Policy Attachment (%s) still exists", rs.Primary.Attributes["id"])
	}
	return nil
}

func testAccCheckPolicyAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
		target := rs.Primary.Attributes["target"]
		policyName := rs.Primary.Attributes["policy"]

		policy, err := tfiot.GetPolicyAttachment(conn, target, policyName)

		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for target %s (%s): %s", target, n, err)
		}

		if policy == nil {
			return fmt.Errorf("Error: Policy %s is not attached to target (%s)", policyName, target)
		}

		return nil
	}
}

func testAccCheckPolicyAttachmentCertStatus(n string, policies []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		certARN := rs.Primary.Attributes["arn"]

		out, err := conn.ListAttachedPolicies(&iot.ListAttachedPoliciesInput{
			Target:   aws.String(certARN),
			PageSize: aws.Int64(250),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list attached policies for target %s: %s", certARN, err)
		}

		if len(out.Policies) != len(policies) {
			return fmt.Errorf("Error: Invalid attached policies count for target %s, expected %d, got %d",
				certARN,
				len(policies),
				len(out.Policies))
		}

		for _, p1 := range policies {
			found := false
			for _, p2 := range out.Policies {
				if p1 == aws.StringValue(p2.PolicyName) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("Error: Policy %s is not attached to target %s", p1, certARN)
			}
		}

		return nil
	}
}

func testAccPolicyAttachmentConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "policy" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "att" {
  policy = aws_iot_policy.policy.name
  target = aws_iot_certificate.cert.arn
}
`, policyName)
}

func testAccPolicyAttachmentUpdate1Config(policyName, policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "policy" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy" "policy2" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "att" {
  policy = aws_iot_policy.policy.name
  target = aws_iot_certificate.cert.arn
}

resource "aws_iot_policy_attachment" "att2" {
  policy = aws_iot_policy.policy2.name
  target = aws_iot_certificate.cert.arn
}
`, policyName, policyName2)
}

func testAccPolicyAttachmentUpdate2Config(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "policy2" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "att2" {
  policy = aws_iot_policy.policy2.name
  target = aws_iot_certificate.cert.arn
}
`, policyName2)
}

func testAccPolicyAttachmentUpdate3Config(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_certificate" "cert2" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "policy2" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "att2" {
  policy = aws_iot_policy.policy2.name
  target = aws_iot_certificate.cert.arn
}

resource "aws_iot_policy_attachment" "att3" {
  policy = aws_iot_policy.policy2.name
  target = aws_iot_certificate.cert2.arn
}
`, policyName2)
}
