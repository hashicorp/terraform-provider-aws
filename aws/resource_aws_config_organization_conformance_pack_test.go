package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"
)

func TestAccConfigOrganizationConformancePack_basic(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsMasterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifier(rName, rId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					resource.TestCheckNoResourceAttr(resourceName, "excluded_accounts"),
					testAccCheckConfigOrganizationConformancePackSuccessful(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func TestAccConfigOrganizationConformancePack_disappears(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsMasterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigOrganizationConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccConfigOrganizationConformancePack_inputParameters(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := "IAM_PASSWORD_POLICY"
	pKey := "ParamKey"
	pValue := "ParamValue"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsMasterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifierParameter(rName, rId, pKey, pValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameters."+pKey, pValue),
					resource.TestCheckNoResourceAttr(resourceName, "excluded_accounts"),
					testAccCheckConfigOrganizationConformancePackSuccessful(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func TestAccConfigOrganizationConformancePack_s3Delivery(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := "awsconfigconforms" + rName
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsMasterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifierS3Delivery(rName, rId, bName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rId),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					resource.TestCheckNoResourceAttr(resourceName, "excluded_accounts"),
					testAccCheckConfigOrganizationConformancePackSuccessful(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func TestAccConfigOrganizationConformancePack_s3Template(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := rName
	kName := rName + ".yaml"
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsMasterPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifierS3Template(rName, rId, bName, kName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					resource.TestCheckNoResourceAttr(resourceName, "excluded_accounts"),
					testAccCheckConfigOrganizationConformancePackSuccessful(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_s3_uri"},
			},
		},
	})
}

func TestAccConfigOrganizationConformancePack_excludedAccounts(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsMasterPreCheck(t)
			testAccOrganizationsMinAccountsPreCheck(t, 2)
			// TODO: All accounts in the organization must also have configuration recorders in the current region,
			//       which is a little complicated for a precheck.  If you get an 'unexpected state' error with this
			//       test, try enabling configuration recorders across the org.
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackConfigRuleIdentifierExcludedAccounts(rName, rId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
					testAccCheckConfigOrganizationConformancePackSuccessful(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func testAccCheckConfigOrganizationConformancePackDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_conformance_pack" {
			continue
		}

		rule, err := configDescribeOrganizationConformancePack(conn, rs.Primary.ID)

		if isAWSErr(err, configservice.ErrCodeNoSuchOrganizationConformancePackException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule != nil {
			return fmt.Errorf("Config Organization Managed Rule (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConfigOrganizationConformancePackExists(resourceName string, ocr *configservice.OrganizationConformancePack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		pack, err := configDescribeOrganizationConformancePack(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing organization conformance pack (%s): %s", rs.Primary.ID, err)
		}

		if pack == nil {
			return fmt.Errorf("organization conformance pack (%s) not found", rs.Primary.ID)
		}

		*ocr = *pack

		return nil
	}
}

func testAccCheckConfigOrganizationConformancePackSuccessful(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		packStatus, err := configDescribeOrganizationConformancePackStatus(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error describing organization conformance pack status (%s): %s", rs.Primary.ID, err)
		}
		if packStatus == nil {
			return fmt.Errorf("organization conformance pack status (%s) not found", rs.Primary.ID)
		}
		if *packStatus.Status != configservice.OrganizationResourceStatusCreateSuccessful {
			return fmt.Errorf("organization conformance pack (%s) returned %s status (%s):  %s", rs.Primary.ID, *packStatus.Status, *packStatus.ErrorCode, *packStatus.ErrorMessage)
		}

		detailedStatus, err := configDescribeOrganizationConformancePackDetailedStatus(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error describing organization conformance pack detailed status (%s): %s", rs.Primary.ID, err)
		}
		if detailedStatus == nil {
			return fmt.Errorf("organization conformance pack detailed status (%s) not found", rs.Primary.ID)
		}
		for _, s := range detailedStatus {
			if *s.Status != configservice.OrganizationResourceDetailedStatusCreateSuccessful {
				return fmt.Errorf("organization conformance pack (%s) on account %s returned %s status (%s):  %s", rs.Primary.ID, *s.Status, *s.AccountId, *s.ErrorCode, *s.ErrorMessage)
			}
		}

		return nil
	}
}

/*func testAccConfigOrganizationConformancePackBase(rName string) string {
	return fmt.Sprintf(`

data "aws_partition" "current" {
}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.test.name
}

`, rName)
}
*/
func testAccConfigOrganizationConformancePackConfigRuleIdentifier(rName, ruleIdentifier string) string {
	return fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  name          = %[1]q
  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: %[2]q
    Type: AWS::Config::ConfigRule
EOT
}

`, rName, ruleIdentifier)
}

func testAccConfigOrganizationConformancePackConfigRuleIdentifierParameter(rName, ruleIdentifier, pKey, pValue string) string {
	return fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  name             = %[1]q
  input_parameters = {
    %[3]s = %[4]q
  }
  template_body    = <<EOT
Parameters:
  %[3]s:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: %[2]q
    Type: AWS::Config::ConfigRule
EOT
}

`, rName, ruleIdentifier, pKey, pValue)
}

func testAccConfigOrganizationConformancePackConfigRuleIdentifierS3Delivery(rName, ruleIdentifier, bName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[3]q
  acl           = "private"
  force_destroy = true
}

resource "aws_config_organization_conformance_pack" "test" {
  name = %[1]q
  delivery_s3_bucket     = aws_s3_bucket.test.id
  delivery_s3_key_prefix = %[2]q
  template_body          = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: %[2]q
    Type: AWS::Config::ConfigRule
EOT
}

`, rName, ruleIdentifier, bName)
}

func testAccConfigOrganizationConformancePackConfigRuleIdentifierS3Template(rName, ruleIdentifier, bName, kName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[3]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[4]q
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: %[2]q
    Type: AWS::Config::ConfigRule
EOT
}

resource "aws_config_organization_conformance_pack" "test" {
  name            = "%[1]s"
  template_s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.id}"
}

`, rName, ruleIdentifier, bName, kName)
}

func testAccConfigOrganizationConformancePackConfigRuleIdentifierExcludedAccounts(rName, ruleIdentifier string) string {
	return fmt.Sprintf(`

data "aws_caller_identity" "current" {}

resource "aws_config_organization_conformance_pack" "test" {
  name              = %[1]q
  excluded_accounts = [data.aws_caller_identity.current.account_id]
  template_body     = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: %[2]q
    Type: AWS::Config::ConfigRule
EOT
}

`, rName, ruleIdentifier)
}
