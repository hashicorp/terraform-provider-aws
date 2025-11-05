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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AllowedImagesSettings_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		"basic":                            testAccEC2AllowedImagesSettings_basic,
		"disappears":                       testAccEC2AllowedImagesSettings_disappears,
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_auditMode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteria(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithNames(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithMarketplace(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithCreationDate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithDeprecationTime(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaComplete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings1),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "enabled"),
				),
			},
			{
				Config: testAccEC2AllowedImagesSettingsConfig_auditMode(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings2),
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
		CheckDestroy:             testAccCheckEC2AllowedImagesSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteria(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings1),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "1"),
				),
			},
			{
				Config: testAccEC2AllowedImagesSettingsConfig_imageCriteriaMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEC2AllowedImagesSettingsExists(ctx, resourceName, &settings2),
					resource.TestCheckResourceAttr(resourceName, "image_criterion.#", "2"),
				),
			},
		},
	})
}

func testAccCheckEC2AllowedImagesSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_allowed_images_settings" {
				continue
			}

			_, err := tfec2.FindAllowedImagesSettings(ctx, conn)

			if tfresource.NotFound(err) {
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

func testAccCheckEC2AllowedImagesSettingsExists(ctx context.Context, n string, v *ec2.GetAllowedImagesSettingsOutput) resource.TestCheckFunc {
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

func testAccEC2AllowedImagesSettingsConfig_basic() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"
}
`
}

func testAccEC2AllowedImagesSettingsConfig_auditMode() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "audit-mode"
}
`
}

func testAccEC2AllowedImagesSettingsConfig_imageCriteria() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]
  }
}
`
}

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaMultiple() string {
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

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithNames() string {
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

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithMarketplace() string {
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

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithCreationDate() string {
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

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaWithDeprecationTime() string {
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

func testAccEC2AllowedImagesSettingsConfig_imageCriteriaComplete() string {
	return `
resource "aws_ec2_allowed_images_settings" "test" {
  state = "enabled"

  image_criterion {
    image_names              = ["al2023-ami-*"]
    image_providers          = ["amazon"]
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
