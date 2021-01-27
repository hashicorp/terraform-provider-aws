package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccConfigConformancePack_basic(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackConfigRuleIdentifier(rName, rId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					testAccCheckConfigConformancePackSuccessful(resourceName),
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

func testAccConfigConformancePack_disappears(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackConfigRuleIdentifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigConformancePack_InputParameters(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := "IAM_PASSWORD_POLICY"
	pKey := "ParamKey"
	pValue := "ParamValue"
	resourceName := "aws_config_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackConfigRuleIdentifierParameter(rName, rId, pKey, pValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameters."+pKey, pValue),
					testAccCheckConfigConformancePackSuccessful(resourceName),
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

func testAccConfigConformancePack_S3Delivery(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := "awsconfigconforms" + rName
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackConfigRuleIdentifierS3Delivery(rName, rId, bName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rId),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					testAccCheckConfigConformancePackSuccessful(resourceName),
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

func testAccConfigConformancePack_S3Template(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := rName
	kName := rName + ".yaml"
	rId := "IAM_PASSWORD_POLICY"
	resourceName := "aws_config_conformance_pack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackConfigRuleIdentifierS3Template(rName, rId, bName, kName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckNoResourceAttr(resourceName, "input_parameters"),
					testAccCheckConfigConformancePackSuccessful(resourceName),
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

func testAccCheckConfigConformancePackDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_conformance_pack" {
			continue
		}

		rule, err := configDescribeConformancePack(conn, rs.Primary.ID)

		if isAWSErr(err, configservice.ErrCodeNoSuchConformancePackException, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config  Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule != nil {
			return fmt.Errorf("Config  Managed Rule (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConfigConformancePackExists(resourceName string, ocr *configservice.ConformancePackDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		pack, err := configDescribeConformancePack(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing  conformance pack (%s): %s", rs.Primary.ID, err)
		}

		if pack == nil {
			return fmt.Errorf(" conformance pack (%s) not found", rs.Primary.ID)
		}

		*ocr = *pack

		return nil
	}
}

func testAccCheckConfigConformancePackSuccessful(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		packStatus, err := configDescribeConformancePackStatus(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error describing conformance pack status (%s): %s", rs.Primary.ID, err)
		}
		if packStatus == nil {
			return fmt.Errorf("conformance pack status (%s) not found", rs.Primary.ID)
		}
		if *packStatus.ConformancePackState != configservice.ConformancePackStateCreateComplete {
			return fmt.Errorf("conformance pack (%s) returned %s status:  %s", rs.Primary.ID, *packStatus.ConformancePackState, *packStatus.ConformancePackStatusReason)
		}

		return nil
	}
}
func testAccConfigConformancePackConfigRuleIdentifier(rName, ruleIdentifier string) string {
	return fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
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

func testAccConfigConformancePackConfigRuleIdentifierParameter(rName, ruleIdentifier, pKey, pValue string) string {
	return fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  name = %[1]q
  input_parameters = {
    %[3]s = %[4]q
  }
  template_body = <<EOT
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

func testAccConfigConformancePackConfigRuleIdentifierS3Delivery(rName, ruleIdentifier, bName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[3]q
  acl           = "private"
  force_destroy = true
}
resource "aws_config_conformance_pack" "test" {
  name                   = %[1]q
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

func testAccConfigConformancePackConfigRuleIdentifierS3Template(rName, ruleIdentifier, bName, kName string) string {
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
resource "aws_config_conformance_pack" "test" {
  name            = "%[1]s"
  template_s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.id}"
}
`, rName, ruleIdentifier, bName, kName)
}
