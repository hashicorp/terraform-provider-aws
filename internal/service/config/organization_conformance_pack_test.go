package config_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func testAccConfigOrganizationConformancePack_basic(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
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

func testAccConfigOrganizationConformancePack_disappears(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceOrganizationConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigOrganizationConformancePack_excludedAccounts(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackExcludedAccounts1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccConfigOrganizationConformancePackExcludedAccounts2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
				),
			},
		},
	})
}

func testAccConfigOrganizationConformancePack_forceNew(t *testing.T) {
	var before, after configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &before),
				),
			},
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &after),
					testAccCheckConfigOrganizationConformancePackRecreated(&before, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rNameUpdated))),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
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

func testAccConfigOrganizationConformancePack_inputParameters(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	pKey := "ParamKey"
	pValue := "ParamValue"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackInputParameterConfig(rName, pKey, pValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  pKey,
						"parameter_value": pValue,
					}),
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

func testAccConfigOrganizationConformancePack_S3Delivery(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	bucketName := sdkacctest.RandomWithPrefix("awsconfigconforms")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackS3DeliveryConfig(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
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

func testAccConfigOrganizationConformancePack_S3Template(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackS3TemplateConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
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

func testAccConfigOrganizationConformancePack_updateInputParameters(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackInputParameterConfig(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigOrganizationConformancePackUpdateInputParameterConfig(rName, "TestKey1", "TestKey2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey1",
						"parameter_value": "TestValue1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey2",
						"parameter_value": "TestValue2",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
		},
	})
}

func testAccConfigOrganizationConformancePack_updateS3Delivery(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	bucketName := sdkacctest.RandomWithPrefix("awsconfigconforms")
	updatedBucketName := fmt.Sprintf("%s-update", bucketName)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackS3DeliveryConfig(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
				),
			},
			{
				Config: testAccConfigOrganizationConformancePackS3DeliveryConfig(rName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", updatedBucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
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

func testAccConfigOrganizationConformancePack_updateS3Template(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	bucketName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackS3TemplateConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigOrganizationConformancePackS3TemplateConfig(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
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

func testAccConfigOrganizationConformancePack_updateTemplateBody(t *testing.T) {
	var pack configservice.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigOrganizationConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigOrganizationConformancePackUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigOrganizationConformancePackExists(resourceName, &pack),
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_conformance_pack" {
			continue
		}

		pack, err := configDescribeOrganizationConformancePack(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConformancePackException) {
			continue
		}

		// In the event the Organizations Organization is deleted first, its Conformance Packs
		// are deleted and we can continue through the loop
		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeOrganizationAccessDeniedException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Organization Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack != nil {
			return fmt.Errorf("Config Organization Conformance Pack (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConfigOrganizationConformancePackExists(resourceName string, ocp *configservice.OrganizationConformancePack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigConn

		pack, err := configDescribeOrganizationConformancePack(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing Config Organization Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack == nil {
			return fmt.Errorf("Config Organization Conformance Pack (%s) not found", rs.Primary.ID)
		}

		*ocp = *pack

		return nil
	}
}

func testAccCheckConfigOrganizationConformancePackRecreated(before, after *configservice.OrganizationConformancePack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.OrganizationConformancePackArn) == aws.StringValue(after.OrganizationConformancePackArn) {
			return fmt.Errorf("AWS Config Organization Conformance Pack was not recreated")
		}
		return nil
	}
}

func testAccConfigOrganizationConformancePackBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccConfigOrganizationConformancePackBasicConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name          = %[1]q
  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}

func testAccConfigOrganizationConformancePackInputParameterConfig(rName, pKey, pValue string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name       = %q

  input_parameter {
    parameter_name  = %[2]q
    parameter_value = %q
  }

  template_body = <<EOT
Parameters:
  %[2]s:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName, pKey, pValue))
}

func testAccConfigOrganizationConformancePackUpdateInputParameterConfig(rName, pName1, pName2 string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name       = %[1]q

  input_parameter {
    parameter_name  = %[2]q
    parameter_value = "TestValue1"
  }

  input_parameter {
    parameter_name  = %[3]q
    parameter_value = "TestValue2"
  }

  template_body = <<EOT
Parameters:
  %[2]s:
    Type: String
  %[3]s:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName, pName1, pName2))
}

func testAccConfigOrganizationConformancePackS3DeliveryConfig(rName, bName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on             = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name                   = %[1]q
  delivery_s3_bucket     = aws_s3_bucket.test.id
  delivery_s3_key_prefix = %[1]q
  template_body          = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}

resource "aws_s3_bucket" "test" {
  bucket        = %q
  acl           = "private"
  force_destroy = true
}
`, rName, bName))
}

func testAccConfigOrganizationConformancePackS3TemplateConfig(rName, bName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name            = %q
  template_s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.id}"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[2]q
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName, bName))
}

func testAccConfigOrganizationConformancePackUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name          = %q
  template_body = <<EOT
Resources:
  IAMGroupHasUsersCheck:
    Properties:
      ConfigRuleName: IAMGroupHasUsersCheck
      Source:
        Owner: AWS
        SourceIdentifier: IAM_GROUP_HAS_USERS_CHECK
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}

func testAccConfigOrganizationConformancePackExcludedAccounts1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111"]
  name              = %q

  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}

func testAccConfigOrganizationConformancePackExcludedAccounts2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111", "222222222222"]
  name              = %q

  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}
