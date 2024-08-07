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

func TestAccAppConfigExtension_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`extension/*`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "action_point.0.point", "ON_DEPLOYMENT_COMPLETE"),
					resource.TestCheckResourceAttr(resourceName, "action_point.0.action.0.name", "test"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
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

func TestAccAppConfigExtension_ActionPoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*", map[string]string{
						"point": "ON_DEPLOYMENT_COMPLETE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*.action.*", map[string]string{
						names.AttrName: "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExtensionConfig_actionPoint2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*", map[string]string{
						"point": "ON_DEPLOYMENT_COMPLETE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*", map[string]string{
						"point": "ON_DEPLOYMENT_ROLLED_BACK",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*.action.*", map[string]string{
						names.AttrName: "test",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*.action.*", map[string]string{
						names.AttrName: "test2",
					}),
				),
			},
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*", map[string]string{
						"point": "ON_DEPLOYMENT_COMPLETE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "action_point.*.action.*", map[string]string{
						names.AttrName: "test",
					}),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_Parameter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"
	pName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pDescription1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pDescription2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_parameter1(rName, pName1, pDescription1, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:        pName1,
						names.AttrDescription: pDescription1,
						"required":            acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExtensionConfig_parameter2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:        "parameter1",
						names.AttrDescription: "description1",
						"required":            acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:        "parameter2",
						names.AttrDescription: "description2",
						"required":            acctest.CtFalse,
					}),
				),
			},
			{
				Config: testAccExtensionConfig_parameter1(rName, pName2, pDescription2, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:        pName2,
						names.AttrDescription: pDescription2,
						"required":            acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_Name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExtensionConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_Description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_appconfig_extension.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_description(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExtensionConfig_description(rName, rDescription2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription2),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceExtension(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckExtensionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_environment" {
				continue
			}

			input := &appconfig.GetExtensionInput{
				ExtensionIdentifier: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetExtension(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading AppConfig Extension (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("AppConfig Extension (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckExtensionExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		in := &appconfig.GetExtensionInput{
			ExtensionIdentifier: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetExtension(ctx, in)

		if err != nil {
			return fmt.Errorf("error reading AppConfig Extension (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Extension (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccExtensionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["appconfig.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccExtensionConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
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
`, rName))
}

func testAccExtensionConfig_description(rName string, rDescription string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_appconfig_extension" "test" {
  name        = %[1]q
  description = %[2]q
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
}
`, rName, rDescription))
}

func testAccExtensionConfig_actionPoint2(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
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
  action_point {
    point = "ON_DEPLOYMENT_ROLLED_BACK"
    action {
      name     = "test2"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }
}
`, rName))
}

func testAccExtensionConfig_parameter1(rName string, pName string, pDescription string, pRequired string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
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
`, rName, pName, pDescription, pRequired))
}

func testAccExtensionConfig_parameter2(rName string) string {
	return acctest.ConfigCompose(
		testAccExtensionConfigBase(rName),
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
`, rName))
}
