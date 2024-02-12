// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
)

func TestAccAppConfigConfigurationProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"
	appResourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", appResourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/configurationprofile/[0-9a-z]{4,7}`)),
					resource.TestMatchResourceAttr(resourceName, "configuration_profile_id", regexache.MustCompile(`[0-9a-z]{4,7}`)),
					resource.TestCheckResourceAttr(resourceName, "location_uri", "hosted"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "AWS.Freeform"),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "0"),
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

func TestAccAppConfigConfigurationProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceConfigurationProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigConfigurationProfile_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_identifier", "alias/"+rName),
				),
			},
		},
	})
}

func TestAccAppConfigConfigurationProfile_Validators_json(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_validatorJSON(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "validator.*", map[string]string{
						"type": appconfig.ValidatorTypeJsonSchema,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationProfileConfig_validatorNoJSONContent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "validator.*", map[string]string{
						"content": "",
						"type":    appconfig.ValidatorTypeJsonSchema,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test Validator Removal
				Config: testAccConfigurationProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "0"),
				),
			},
		},
	})
}

func TestAccAppConfigConfigurationProfile_Validators_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_validatorLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "validator.*.content", "aws_lambda_function.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "validator.*", map[string]string{
						"type": appconfig.ValidatorTypeLambda,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test Validator Removal
				Config: testAccConfigurationProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "0"),
				),
			},
		},
	})
}

func TestAccAppConfigConfigurationProfile_Validators_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_validatorMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "validator.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "validator.*", map[string]string{
						"content": "{\"$schema\":\"http://json-schema.org/draft-05/schema#\",\"description\":\"BasicFeatureToggle-1\",\"title\":\"$id$\"}",
						"type":    appconfig.ValidatorTypeJsonSchema,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "validator.*.content", "aws_lambda_function.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "validator.*", map[string]string{
						"type": appconfig.ValidatorTypeLambda,
					}),
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

func TestAccAppConfigConfigurationProfile_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccConfigurationProfileConfig_name(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func TestAccAppConfigConfigurationProfile_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationProfileConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
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

func TestAccAppConfigConfigurationProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationProfileConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccConfigurationProfileConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationProfileExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConfigurationProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_configuration_profile" {
				continue
			}

			confProfID, appID, err := tfappconfig.ConfigurationProfileParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			input := &appconfig.GetConfigurationProfileInput{
				ApplicationId:          aws.String(appID),
				ConfigurationProfileId: aws.String(confProfID),
			}

			output, err := conn.GetConfigurationProfileWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading AppConfig Configuration Profile (%s) for Application (%s): %w", confProfID, appID, err)
			}

			if output != nil {
				return fmt.Errorf("AppConfig Configuration Profile (%s) for Application (%s) still exists", confProfID, appID)
			}
		}

		return nil
	}
}

func testAccCheckConfigurationProfileExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		confProfID, appID, err := tfappconfig.ConfigurationProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		output, err := conn.GetConfigurationProfileWithContext(ctx, &appconfig.GetConfigurationProfileInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
		})

		if err != nil {
			return fmt.Errorf("error reading AppConfig Configuration Profile (%s) for Application (%s): %w", confProfID, appID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Configuration Profile (%s) for Application (%s) not found", confProfID, appID)
		}

		return nil
	}
}

func testAccConfigurationProfileConfig_kms(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id     = aws_appconfig_application.test.id
  name               = %[1]q
  kms_key_identifier = aws_kms_alias.test.name
  location_uri       = "hosted"
}
`, rName))
}

func testAccConfigurationProfileConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %q
  location_uri   = "hosted"
}
`, rName))
}

func testAccConfigurationProfileConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_description(rName, description),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  description    = %[2]q
  location_uri   = "hosted"
}
`, rName, description))
}

func testAccConfigurationProfileConfig_validatorJSON(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %q
  location_uri   = "hosted"

  validator {
    content = jsonencode({
      "$schema"            = "http://json-schema.org/draft-04/schema#"
      title                = "$id$"
      description          = "BasicFeatureToggle-1"
      type                 = "object"
      additionalProperties = false
      patternProperties = {
        "[^\\s]+$" = {
          type = "boolean"
        }
      }
      minProperties = 1
    })

    type = "JSON_SCHEMA"
  }
}
`, rName))
}

func testAccConfigurationProfileConfig_validatorNoJSONContent(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %q
  location_uri   = "hosted"

  validator {
    type = "JSON_SCHEMA"
  }
}
`, rName))
}

func testAccApplicationLambdaBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccConfigurationProfileConfig_validatorLambda(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		testAccApplicationLambdaBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"

  validator {
    content = aws_lambda_function.test.arn
    type    = "LAMBDA"
  }
}
`, rName))
}

func testAccConfigurationProfileConfig_validatorMultiple(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		testAccApplicationLambdaBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"

  validator {
    content = jsonencode({
      "$schema"   = "http://json-schema.org/draft-05/schema#"
      title       = "$id$"
      description = "BasicFeatureToggle-1"
    })

    type = "JSON_SCHEMA"
  }

  validator {
    content = aws_lambda_function.test.arn
    type    = "LAMBDA"
  }
}
`, rName))
}

func testAccConfigurationProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
  name           = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccConfigurationProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
  name           = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
