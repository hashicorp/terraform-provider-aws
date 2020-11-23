package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSLambdaCodeSigningConfig_basic(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaCodeSigningConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.*", resourceName, "allowed_publishers.0.signing_profile_version_arns.*"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaCodeSigningConfig_PolicyConfigId(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaCodeSigningConfigConfigurePolicy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.*", resourceName, "allowed_publishers.0.signing_profile_version_arns.*"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies", resourceName, "policies"),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_id", resourceName, "config_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaCodeSigningConfig_Description(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaCodeSigningConfigConfigureDescription,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.*", resourceName, "allowed_publishers.0.signing_profile_version_arns.*"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

const testAccDataSourceAWSLambdaCodeSigningConfigBasic = `
resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile1/abcde12345",
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile2/abcde12345"
    ]
  }
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`

const testAccDataSourceAWSLambdaCodeSigningConfigConfigurePolicy = `
resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile1/abcde12345",
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile2/abcde12345"
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`

const testAccDataSourceAWSLambdaCodeSigningConfigConfigureDescription = `
resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile1/abcde12345",
      "arn:aws:signer:us-east-1:123456789012:signing-profiles/my_profile2/abcde12345"
    ]
  }

  description = "Code Signing Config for app A"
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`
