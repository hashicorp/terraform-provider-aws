package configservice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
)

func testAccConformancePack_basic(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConformancePack_forceNew(t *testing.T) {
	var before, after configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &before),
				),
			},
			{
				Config: testAccConformancePackConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &after),
					testAccCheckConformancePackRecreated(&before, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rNameUpdated))),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConformancePack_disappears(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.CheckResourceDisappears(acctest.Provider, tfconfig.ResourceConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConformancePack_inputParameters(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_inputParameter(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey",
						"parameter_value": "TestValue",
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

func testAccConformancePack_S3Delivery(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_s3Delivery(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
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

func testAccConformancePack_S3Template(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_s3Template(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
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

func testAccConformancePack_updateInputParameters(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_inputParameter(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConformancePackConfig_updateInputParameter(rName, "TestKey1", "TestKey2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
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
				Config: testAccConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
		},
	})
}

func testAccConformancePack_updateS3Delivery(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_s3Delivery(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConformancePackConfig_s3Delivery(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", bucketName),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
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

func testAccConformancePack_updateS3Template(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_s3Template(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConformancePackConfig_s3Template(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
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

func testAccConformancePack_updateTemplateBody(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConformancePackConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConformancePack_S3TemplateAndTemplateBody(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConformancePackConfig_s3TemplateAndTemplateBody(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConformancePackExists(resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
					"template_s3_uri",
				},
			},
		},
	})
}

func testAccCheckConformancePackDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_conformance_pack" {
			continue
		}

		pack, err := tfconfig.DescribeConformancePack(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack != nil {
			return fmt.Errorf("Config Conformance Pack (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConformancePackExists(resourceName string, detail *configservice.ConformancePackDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

		pack, err := tfconfig.DescribeConformancePack(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing Config Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack == nil {
			return fmt.Errorf("Config Conformance Pack (%s) not found", rs.Primary.ID)
		}

		*detail = *pack

		return nil
	}
}

func testAccCheckConformancePackRecreated(before, after *configservice.ConformancePackDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.ConformancePackArn) == aws.StringValue(after.ConformancePackArn) {
			return fmt.Errorf("AWS Config Conformance Pack was not recreated")
		}
		return nil
	}
}

func testAccConformancePackConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

func testAccConformancePackConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test]
  name          = %q
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

func testAccConformancePackConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test]
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

func testAccConformancePackConfig_inputParameter(rName, pName, pValue string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test]
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
`, rName, pName, pValue))
}

func testAccConformancePackConfig_updateInputParameter(rName, pName1, pName2 string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on = [
  aws_config_configuration_recorder.test]
  name = %[1]q

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

func testAccConformancePackConfig_s3Delivery(rName, bucketName string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_config_conformance_pack" "test" {
  depends_on             = [aws_config_configuration_recorder.test]
  name                   = %[2]q
  delivery_s3_bucket     = aws_s3_bucket.test.bucket
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
`, bucketName, rName))
}

func testAccConformancePackConfig_s3Template(rName, bucketName string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
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

resource "aws_config_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test]
  name            = %q
  template_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.id}"
}
`, bucketName, rName))
}

func testAccConformancePackConfig_s3TemplateAndTemplateBody(rName string) string {
	return acctest.ConfigCompose(testAccConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
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

resource "aws_config_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test]
  name            = %[1]q
  template_body   = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
  template_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.id}"
}
`, rName))
}
