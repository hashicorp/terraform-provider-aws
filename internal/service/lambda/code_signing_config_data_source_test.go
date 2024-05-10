// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaCodeSigningConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_policyID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_policy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies", resourceName, "policies"),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_id", resourceName, "config_id"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_description(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigDataSourceConfig_description,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
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
