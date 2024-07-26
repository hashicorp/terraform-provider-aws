// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConformancePack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct0),
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

func testAccOrganizationConformancePack_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceOrganizationConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationConformancePack_excludedAccounts(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_excludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccOrganizationConformancePackConfig_excludedAccounts2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccOrganizationConformancePack_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccOrganizationConformancePackConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &after),
					testAccCheckOrganizationConformancePackRecreated(&before, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rNameUpdated))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct0),
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

func testAccOrganizationConformancePack_inputParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pKey := "ParamKey"
	pValue := "ParamValue"
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_inputParameter(rName, pKey, pValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct1),
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

func testAccOrganizationConformancePack_S3Delivery(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix("awsconfigconforms")
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_s3Delivery(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
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

func testAccOrganizationConformancePack_S3Template(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_s3Template(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile(fmt.Sprintf("organization-conformance-pack/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct0),
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

func testAccOrganizationConformancePack_updateInputParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_inputParameter(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
				),
			},
			{
				Config: testAccOrganizationConformancePackConfig_updateInputParameter(rName, "TestKey1", "TestKey2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct2),
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
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccOrganizationConformancePack_updateS3Delivery(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix("awsconfigconforms")
	updatedBucketName := fmt.Sprintf("%s-update", bucketName)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_s3Delivery(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
				),
			},
			{
				Config: testAccOrganizationConformancePackConfig_s3Delivery(rName, updatedBucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
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

func testAccOrganizationConformancePack_updateS3Template(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_s3Template(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
				),
			},
			{
				Config: testAccOrganizationConformancePackConfig_s3Template(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
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

func testAccOrganizationConformancePack_updateTemplateBody(t *testing.T) {
	ctx := acctest.Context(t)
	var pack types.OrganizationConformancePack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConformancePackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConformancePackConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
				),
			},
			{
				Config: testAccOrganizationConformancePackConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConformancePackExists(ctx, resourceName, &pack),
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

func testAccCheckOrganizationConformancePackDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_organization_conformance_pack" {
				continue
			}

			_, err := tfconfig.FindOrganizationConformancePackByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Organization Conformance Pack %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationConformancePackExists(ctx context.Context, n string, v *types.OrganizationConformancePack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindOrganizationConformancePackByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOrganizationConformancePackRecreated(before, after *types.OrganizationConformancePack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.OrganizationConformancePackArn) == aws.ToString(after.OrganizationConformancePackArn) {
			return errors.New("ConfigService Organization Conformance Pack was not recreated")
		}
		return nil
	}
}

func testAccOrganizationConformancePackBase(rName string) string {
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
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccOrganizationConformancePackConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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

func testAccOrganizationConformancePackConfig_inputParameter(rName, pKey, pValue string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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

func testAccOrganizationConformancePackConfig_updateInputParameter(rName, pName1, pName2 string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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

func testAccOrganizationConformancePackConfig_s3Delivery(rName, bName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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
  force_destroy = true
}
`, rName, bName))
}

func testAccOrganizationConformancePackConfig_s3Template(rName, bName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
		fmt.Sprintf(`
resource "aws_config_organization_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test, aws_organizations_organization.test]
  name            = %q
  template_s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.test.id}"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
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

func testAccOrganizationConformancePackConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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

func testAccOrganizationConformancePackConfig_excludedAccounts1(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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

func testAccOrganizationConformancePackConfig_excludedAccounts2(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationConformancePackBase(rName),
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
