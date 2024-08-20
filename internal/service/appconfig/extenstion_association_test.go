// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigExtensionAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionAssociationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "extension_arn", "appconfig", regexache.MustCompile(`extension/*`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrResourceARN, "appconfig", regexache.MustCompile(`application/*`)),
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

func TestAccAppConfigExtensionAssociation_Parameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension_association.test"
	pName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pDescription1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pValue1 := "ParameterValue1"
	pName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pDescription2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pValue2 := "ParameterValue2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionAssociationConfig_parameters1(rName, pName1, pDescription1, acctest.CtTrue, pValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("parameters.%s", pName1), pValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExtensionAssociationConfig_parameters2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.parameter1", pValue1),
					resource.TestCheckResourceAttr(resourceName, "parameters.parameter2", pValue2),
				),
			},
			{
				Config: testAccExtensionAssociationConfig_parameters1(rName, pName2, pDescription2, acctest.CtFalse, pValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("parameters.%s", pName2), pValue2),
				),
			},
			{
				Config: testAccExtensionAssociationConfig_parametersNotRequired(rName, pName2, pDescription2, acctest.CtFalse, pValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAppConfigExtensionAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionAssociationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceExtensionAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckExtensionAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_environment" {
				continue
			}

			input := &appconfig.GetExtensionAssociationInput{
				ExtensionAssociationId: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetExtensionAssociation(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading AppConfig ExtensionAssociation (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("AppConfig ExtensionAssociation (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckExtensionAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		in := &appconfig.GetExtensionAssociationInput{
			ExtensionAssociationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetExtensionAssociation(ctx, in)

		if err != nil {
			return fmt.Errorf("error reading AppConfig ExtensionAssociation (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig ExtensionAssociation (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccExtensionAssociationConfigBase(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}
`, rName))
}

func testAccExtensionAssociationConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_extension" "test" {
  name        = %[1]q
  description = "test description"
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
}
resource "aws_appconfig_extension_association" "test" {
  extension_arn = aws_appconfig_extension.test.arn
  resource_arn  = aws_appconfig_application.test.arn
}
`, rName))
}

func testAccExtensionAssociationConfig_parameters1(rName string, pName string, pDescription string, pRequired string, pValue string) string {
	return acctest.ConfigCompose(
		testAccExtensionAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_extension" "test" {
  name = %[1]q
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
  parameter {
    name        = %[2]q
    description = %[3]q
    required    = %[4]s
  }
}
resource "aws_appconfig_extension_association" "test" {
  extension_arn = aws_appconfig_extension.test.arn
  resource_arn  = aws_appconfig_application.test.arn
  parameters = {
    %[2]s = %[5]q
  }
}
`, rName, pName, pDescription, pRequired, pValue))
}

func testAccExtensionAssociationConfig_parameters2(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_extension" "test" {
  name = %[1]q
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
  parameter {
    name        = "parameter1"
    description = "description1"
    required    = true
  }
  parameter {
    name        = "parameter2"
    description = "description2"
    required    = false
  }
}
resource "aws_appconfig_extension_association" "test" {
  extension_arn = aws_appconfig_extension.test.arn
  resource_arn  = aws_appconfig_application.test.arn
  parameters = {
    parameter1 = "ParameterValue1"
    parameter2 = "ParameterValue2"
  }
}
`, rName))
}

func testAccExtensionAssociationConfig_parametersNotRequired(rName string, pName string, pDescription string, pRequired string, pValue string) string {
	return acctest.ConfigCompose(
		testAccExtensionAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_extension" "test" {
  name = %[1]q
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
  parameter {
    name        = %[2]q
    description = %[3]q
    required    = %[4]s
  }
}
resource "aws_appconfig_extension_association" "test" {
  extension_arn = aws_appconfig_extension.test.arn
  resource_arn  = aws_appconfig_application.test.arn
}
`, rName, pName, pDescription, pRequired, pValue))
}
