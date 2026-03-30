// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaCodeSigningConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	signingProfile1 := "aws_signer_signing_profile.test1"
	signingProfile2 := "aws_signer_signing_profile.test2"
	var conf awstypes.CodeSigningConfig

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeSigningConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile2, "version_arn"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Warn"),
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

func TestAccLambdaCodeSigningConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	var conf awstypes.CodeSigningConfig

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeSigningConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tflambda.ResourceCodeSigningConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaCodeSigningConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var conf awstypes.CodeSigningConfig

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeSigningConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeSigningConfigConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCodeSigningConfigConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfig_updatePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	var conf awstypes.CodeSigningConfig

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeSigningConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Warn"),
				),
			},
			{
				Config: testAccCodeSigningConfigConfig_updatePolicy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "policies.0.untrusted_artifact_on_deployment", "Enforce"),
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

func TestAccLambdaCodeSigningConfig_updatePublishers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_code_signing_config.code_signing_config"
	signingProfile1 := "aws_signer_signing_profile.test1"
	signingProfile2 := "aws_signer_signing_profile.test2"
	var conf awstypes.CodeSigningConfig

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeSigningConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Code Signing Config for test account"),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile2, "version_arn"),
				),
			},
			{
				Config: testAccCodeSigningConfigConfig_updatePublishers(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeSigningConfigExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "allowed_publishers.0.signing_profile_version_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_publishers.0.signing_profile_version_arns.*", signingProfile1, "version_arn"),
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

func testAccCheckCodeSigningConfigExists(ctx context.Context, t *testing.T, n string, v *awstypes.CodeSigningConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		output, err := tflambda.FindCodeSigningConfigByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCodeSigningConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_code_signing_config" {
				continue
			}

			_, err := tflambda.FindCodeSigningConfigByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Code Signing Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCodeSigningConfigConfig_basic() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn,
      aws_signer_signing_profile.test2.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  description = "Code Signing Config for test account"
}`
}

func testAccCodeSigningConfigConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn,
    ]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCodeSigningConfigConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn,
    ]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCodeSigningConfigConfig_updatePublishers() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn
    ]
  }
}`
}

func testAccCodeSigningConfigConfig_updatePolicy() string {
	return `
resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn,
      aws_signer_signing_profile.test2.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Enforce"
  }
}`
}
