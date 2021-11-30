package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccSSMAssociation_basic(t *testing.T) {
	name := fmt.Sprintf("tf-acc-ssm-association-%s", sdkacctest.RandString(10))
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "false"),
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

func TestAccSSMAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMAssociation_applyOnlyAtCronInterval(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithApplyOnlyAtCronIntervalConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithApplyOnlyAtCronIntervalConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "false"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withTargets(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"
	oneTarget := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}
`

	twoTargets := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}

targets {
  key    = "tag:ExtraName"
  values = ["acceptanceTest"]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithTargetsConfig(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithTargetsConfig(name, twoTargets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.1.key", "tag:ExtraName"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.1.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAssociationBasicWithTargetsConfig(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withParameters(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithParametersConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAssociationBasicWithParametersUpdatedConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withAssociationName(t *testing.T) {
	assocName1 := sdkacctest.RandString(10)
	assocName2 := sdkacctest.RandString(10)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithAssociationNameConfig(rName, assocName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithAssociationNameConfig(rName, assocName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withAssociationNameAndScheduleExpression(t *testing.T) {
	assocName := sdkacctest.RandString(10)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_ssm_association.test"
	scheduleExpression1 := "cron(0 16 ? * TUE *)"
	scheduleExpression2 := "cron(0 16 ? * WED *)"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationWithAssociationNameAndScheduleExpressionConfig(rName, assocName, scheduleExpression1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationWithAssociationNameAndScheduleExpressionConfig(rName, assocName, scheduleExpression2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withDocumentVersion(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithDocumentVersionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "document_version", "1"),
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

func TestAccSSMAssociation_withOutputLocation(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithOutPutLocationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithOutPutLocationUpdateBucketNameConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAssociationBasicWithOutPutLocationUpdateKeyPrefixConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "UpdatedAssociation"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withOutputLocation_s3Region(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationWithOutputLocationS3RegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", acctest.Region()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationWithOutputLocationUpdateS3RegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationWithOutputLocationNoS3RegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_region", ""),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withAutomationTargetParamName(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithAutomationTargetParamNameConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAssociationBasicWithParametersUpdatedConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withScheduleExpression(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithScheduleExpressionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithScheduleExpressionUpdatedConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccSSMAssociation_withComplianceSeverity(t *testing.T) {
	assocName := sdkacctest.RandString(10)
	rName := sdkacctest.RandString(10)
	compSeverity1 := "HIGH"
	compSeverity2 := "LOW"
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationBasicWithComplianceSeverityConfig(compSeverity1, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(
						resourceName, "compliance_severity", compSeverity1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationBasicWithComplianceSeverityConfig(compSeverity2, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(
						resourceName, "compliance_severity", compSeverity2),
				),
			},
		},
	})
}

func TestAccSSMAssociation_rateControl(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationRateControlConfig(name, "10%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "max_concurrency", "10%"),
					resource.TestCheckResourceAttr(
						resourceName, "max_errors", "10%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationRateControlConfig(name, "20%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "max_concurrency", "20%"),
					resource.TestCheckResourceAttr(
						resourceName, "max_errors", "20%"),
				),
			},
		},
	})
}

func testAccCheckAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Assosciation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		_, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, ssm.ErrCodeAssociationDoesNotExist, "") {
				return nil
			}
			return err
		}

		return nil
	}
}

func testAccCheckAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_association" {
			continue
		}

		out, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, ssm.ErrCodeAssociationDoesNotExist, "") {
				continue
			}
			return err
		}

		if out != nil {
			return fmt.Errorf("Expected AWS SSM Association to be gone, but was still found")
		}
	}

	return nil
}

func testAccAssociationBasicWithApplyOnlyAtCronIntervalConfig(rName string, applyOnlyAtCronInterval bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                        = aws_ssm_document.test.name
  schedule_expression         = "cron(0 16 ? * TUE *)"
  apply_only_at_cron_interval = %[2]t

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, applyOnlyAtCronInterval)
}

func testAccAssociationBasicWithAutomationTargetParamNameConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
resource "aws_iam_instance_profile" "ssm_profile" {
  name = "ssm_profile-%[1]s"
  role = aws_iam_role.ssm_role.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "ssm_role" {
  name = "ssm_role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_ssm_document" "foo" {
  name          = "test_document-%[1]s"
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC

}

resource "aws_ssm_association" "test" {
  name                             = aws_ssm_document.foo.name
  automation_target_parameter_name = "Directory"

  parameters = {
    AutomationAssumeRole = aws_iam_role.ssm_role.id
    Directory            = "myWorkSpace"
  }

  targets {
    key    = "tag:myTagName"
    values = ["myTagValue"]
  }

  schedule_expression = "rate(60 minutes)"
}
`, rName))
}

func testAccAssociationBasicWithParametersUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpaceUpdated"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationBasicWithParametersConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpace"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationBasicWithTargetsConfig(rName, targetsStr string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name
  %s
}
`, rName, targetsStr)
}

func testAccAssociationBasicConfig(rName string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_ami" "amzn" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.name
  }
}

resource "aws_subnet" "first" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_security_group" "test" {
  name        = var.name
  description = "foo"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn.image_id
  availability_zone      = data.aws_availability_zones.available.names[0]
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.first.id

  tags = {
    Name = var.name
  }
}

resource "aws_ssm_document" "test" {
  name          = var.name
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name        = var.name
  instance_id = aws_instance.test.id
}
`, rName)
}

func testAccAssociationBasicWithDocumentVersionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = "test_document_association-%s"
  document_version = aws_ssm_document.test.latest_version

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rName)
}

func testAccAssociationBasicWithScheduleExpressionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * TUE *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationBasicWithScheduleExpressionUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * WED *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAssociationBasicWithOutPutLocationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName, rName)
}

func testAccAssociationWithOutputLocationS3RegionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}
`, rName)
}

func testAccAssociationWithOutputLocationS3RegionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
    s3_region      = aws_s3_bucket.test.region
  }
}
`)
}

func testAccAssociationWithOutputLocationUpdateS3RegionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
    s3_region      = %[1]q
  }
}
`, acctest.AlternateRegion()))
}

func testAccAssociationWithOutputLocationNoS3RegionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociationWithOutputLocationS3RegionConfigBase(rName),
		`
resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.test.id
  }
}
`)
}

func testAccAssociationBasicWithOutPutLocationUpdateBucketNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName, rName, rName)
}

func testAccAssociationBasicWithOutPutLocationUpdateKeyPrefixConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "UpdatedAssociation"
  }
}
`, rName, rName, rName)
}

func testAccAssociationBasicWithAssociationNameConfig(rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = aws_ssm_document.test.name
  association_name = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName)
}

func testAccAssociationWithAssociationNameAndScheduleExpressionConfig(rName, associationName, scheduleExpression string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  association_name    = %q
  name                = aws_ssm_document.test.name
  schedule_expression = %q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, associationName, scheduleExpression)
}

func testAccAssociationBasicWithComplianceSeverityConfig(compSeverity, rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  association_name    = "%s"
  compliance_severity = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName, compSeverity)
}

func testAccAssociationRateControlConfig(rName, rate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "tf-test-ssm-document-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name            = aws_ssm_document.test.name
  max_concurrency = "%s"
  max_errors      = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rate, rate)
}
