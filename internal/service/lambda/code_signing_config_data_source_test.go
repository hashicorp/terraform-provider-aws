package lambda_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaCodeSigningConfigDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_policyID(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_policy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies", resourceName, "policies"),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_id", resourceName, "config_id"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_description(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_description,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

const testAccCodeSigningConfigDataSourceConfig_basic = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
    ]
  }
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`

const testAccCodeSigningConfigDataSourceConfig_policy = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
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

const testAccCodeSigningConfigDataSourceConfig_description = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
    ]
  }

  description = "Code Signing Config for app A"
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`
