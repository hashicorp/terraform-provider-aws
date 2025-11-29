// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AllowedImagesSettings_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		acctest.CtBasic:                    testAccEC2AllowedImagesSettings_basic,
		acctest.CtDisappears:               testAccEC2AllowedImagesSettings_disappears,
		"auditMode":                        testAccEC2AllowedImagesSettings_auditMode,
		"imageCriteria":                    testAccEC2AllowedImagesSettings_imageCriteria,
		"imageCriteriaMultiple":            testAccEC2AllowedImagesSettings_imageCriteriaMultiple,
		"imageCriteriaWithNames":           testAccEC2AllowedImagesSettings_imageCriteriaWithNames,
		"imageCriteriaWithMarketplace":     testAccEC2AllowedImagesSettings_imageCriteriaWithMarketplace,
		"imageCriteriaWithCreationDate":    testAccEC2AllowedImagesSettings_imageCriteriaWithCreationDate,
		"imageCriteriaWithDeprecationTime": testAccEC2AllowedImagesSettings_imageCriteriaWithDeprecationTime,
		"imageCriteriaComplete":            testAccEC2AllowedImagesSettings_imageCriteriaComplete,
		"stateUpdate":                      testAccEC2AllowedImagesSettings_stateUpdate,
		"imageCriteriaUpdate":              testAccEC2AllowedImagesSettings_imageCriteriaUpdate,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2AllowedImagesSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceAllowedImagesSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_auditMode(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_auditMode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "audit-mode"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteria(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.image_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.0.image_providers.*", "amazon"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.image_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.0.image_providers.*", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.1.image_providers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.1.image_providers.*", "aws-marketplace"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaWithNames(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaWithNames(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.image_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.0.image_names.*", "al2023-ami-*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.0.image_names.*", "ubuntu/images/*"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaWithMarketplace(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaWithMarketplace(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.marketplace_product_codes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_criterion.0.marketplace_product_codes.*", "abcdef123456"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaWithCreationDate(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaWithCreationDate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.creation_date_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.creation_date_condition.0.maximum_days_since_created", "365"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaWithDeprecationTime(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaWithDeprecationTime(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.deprecation_time_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.deprecation_time_condition.0.maximum_days_since_deprecated", "30"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaComplete(t *testing.T) {
	ctx := acctest.Context(t)
	var settings ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaComplete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.image_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.image_providers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.marketplace_product_codes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.creation_date_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.0.deprecation_time_condition.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_stateUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var settings1, settings2 ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings1),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, names.AttrEnabled),
				),
			},
			{
				Config: testAccAllowedImagesSettingsConfig_auditMode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings2),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "audit-mode"),
				),
			},
		},
	})
}

func testAccEC2AllowedImagesSettings_imageCriteriaUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var settings1, settings2 ec2.GetAllowedImagesSettingsOutput
	resourceName := "aws_ec2_allowed_images_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteria(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings1),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
				),
			},
			{
				Config: testAccAllowedImagesSettingsConfig_imageCriteriaMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAllowedImagesSettingsExists(ctx, resourceName, &settings2),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAllowedImagesSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_allowed_images_settings" {
				continue
			}

			_, err := tfec2.FindAllowedImagesSettings(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("EC2 Allowed Images Settings still exists")
		}

		return nil
	}
}

func testAccCheckAllowedImagesSettingsExists(ctx context.Context, n string, v *ec2.GetAllowedImagesSettingsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindAllowedImagesSettings(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAllowedImagesSettingsConfig_basic() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"
}
`
}

func testAccAllowedImagesSettingsConfig_auditMode() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "audit-mode"
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteria() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaMultiple() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]
  }

  image_criterion {
    image_providers = ["aws-marketplace"]
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaWithNames() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_names = [
      "al2023-ami-*",
      "ubuntu/images/*"
    ]
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaWithMarketplace() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers           = ["aws-marketplace"]
    marketplace_product_codes = ["abcdef123456"]
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaWithCreationDate() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]

    creation_date_condition {
      maximum_days_since_created = 365
    }
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaWithDeprecationTime() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]

    deprecation_time_condition {
      maximum_days_since_deprecated = 30
    }
  }
}
`
}

func testAccAllowedImagesSettingsConfig_imageCriteriaComplete() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_names               = ["al2023-ami-*"]
    image_providers           = ["amazon"]
    marketplace_product_codes = ["abc123def456"]

    creation_date_condition {
      maximum_days_since_created = 180
    }

    deprecation_time_condition {
      maximum_days_since_deprecated = 60
    }
  }
}
`
}
