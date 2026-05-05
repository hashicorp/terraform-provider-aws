// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigExtension_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`extension/*`)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", "1"),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", "2"),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action_point.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"
	pName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	pDescription1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	pName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	pDescription2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_parameter1(rName, pName1, pDescription1, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_Description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rDescription := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rDescription2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_appconfig_extension.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_description(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
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
					testAccCheckExtensionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rDescription2),
				),
			},
		},
	})
}

func TestAccAppConfigExtension_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_extension.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExtensionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExtensionConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExtensionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappconfig.ResourceExtension(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckExtensionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_extension" {
				continue
			}

			_, err := tfappconfig.FindExtensionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppConfig Extension %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckExtensionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		_, err := tfappconfig.FindExtensionByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccExtensionConfig_base(rName string) string {
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
		testAccExtensionConfig_base(rName),
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
		testAccExtensionConfig_base(rName),
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
		testAccExtensionConfig_base(rName),
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
		testAccExtensionConfig_base(rName),
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
		testAccExtensionConfig_base(rName),
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
