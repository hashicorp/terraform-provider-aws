package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTProvisioningTemplate_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("Fleet-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTProvisioningTemplateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTProvisioningTemplateConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_provisioning_template.fleet", "default_version_id", "1"),
					resource.TestCheckResourceAttr("aws_iot_provisioning_template.fleet", "description", "My provisioning template"),
					resource.TestCheckResourceAttr("aws_iot_provisioning_template.fleet", "enabled", "false"),
					resource.TestCheckResourceAttrSet("aws_iot_provisioning_template.fleet", "provisioning_role_arn"),
					resource.TestCheckResourceAttrSet("aws_iot_provisioning_template.fleet", "template_arn"),
					resource.TestCheckResourceAttr("aws_iot_provisioning_template.fleet", "template_name", rName),
					resource.TestCheckResourceAttrSet("aws_iot_provisioning_template.fleet", "template_body"),
					testAccAWSIoTProvisioningTemplateCheckVersionExists(rName, 1),
				),
			},
			{
				Config: testAccAWSIoTProvisioningTemplateConfigTemplateBodyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_provisioning_template.fleet", "default_version_id", "2"),
					resource.TestCheckResourceAttrSet("aws_iot_provisioning_template.fleet", "template_body"),
					testAccAWSIoTProvisioningTemplateCheckVersionExists(rName, 2),
				),
			},
		},
	})
}

func testAccCheckAWSIoTProvisioningTemplateDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_provisioning_template" {
			continue
		}

		_, err := conn.DescribeProvisioningTemplate(&iot.DescribeProvisioningTemplateInput{
			TemplateName: aws.String(rs.Primary.Attributes["template_name"]),
		})

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			return nil
		} else if err != nil {
			return err
		}

		return fmt.Errorf("IoT Provisioning Template still exists")
	}

	return nil
}

func testAccAWSIoTProvisioningTemplateCheckVersionExists(templateName string, numVersions int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotconn

		resp, err := conn.ListProvisioningTemplateVersions(&iot.ListProvisioningTemplateVersionsInput{
			TemplateName: aws.String(templateName),
		})

		if err != nil {
			return err
		}

		if len(resp.Versions) != numVersions {
			return fmt.Errorf("Expected %d versions for template %s but found %d", numVersions, templateName, len(resp.Versions))
		}

		return nil
	}
}

func testAccAWSIoTProvisioningTemplateConfigInitialState(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "iot_assume_role_policy" {
	statement {
		actions = ["sts:AssumeRole"]

		principals {
			type        = "Service"
			identifiers = ["iot.amazonaws.com"]
		}
	}
}

resource "aws_iam_role" "iot_fleet_provisioning" {
	name = "IoTProvisioningServiceRole"
	path = "/service-role/"
	assume_role_policy = data.aws_iam_policy_document.iot_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "iot_fleet_provisioning_registration" {
	role       = aws_iam_role.iot_fleet_provisioning.name
	policy_arn = "arn:aws:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

resource "aws_iot_provisioning_template" "fleet" {
	template_name         = "%s"
	description           = "My provisioning template"
	provisioning_role_arn = aws_iam_role.iot_fleet_provisioning.arn

  template_body = <<EOF
{
  "Parameters": {
    "SerialNumber": {
      "Type": "String"
    },
    "AWS::IoT::Certificate::Id": {
      "Type": "String"
    }
  },
  "Resources": {
    "certificate": {
      "Properties": {
        "CertificateId": {
          "Ref": "AWS::IoT::Certificate::Id"
        },
        "Status": "Active"
      },
      "Type": "AWS::IoT::Certificate"
    },
    "policy": {
      "Properties": {
        "PolicyDocument": "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Action\": \"iot:*\",\n      \"Resource\": \"*\"\n    }\n  ]\n}"
      },
      "Type": "AWS::IoT::Policy"
    }
  }
}
EOF
}
`, rName)
}

func testAccAWSIoTProvisioningTemplateConfigTemplateBodyUpdate(rName string) string {
	return strings.ReplaceAll(testAccAWSIoTProvisioningTemplateConfigInitialState(rName), "Allow", "Deny")
}
