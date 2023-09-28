// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
)

func testAccDetectorFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_basic("RDS_LOGIN_EVENTS", "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
		},
	})
}

func testAccDetectorFeature_additionalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("DISABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "status", "DISABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
				),
			},
		},
	})
}

func testAccCheckDetectorFeatureExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		_, err := tfguardduty.FindDetectorFeatureByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["name"])

		return err
	}
}

func testAccDetectorFeatureConfig_basic(name, status string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = %[1]q
  status      = %[2]q
}
`, name, status)
}

func testAccDetectorFeatureConfig_additionalConfiguration(featureStatus, additionalConfigurationStatus string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "EKS_RUNTIME_MONITORING"
  status      = %[1]q

  additional_configuration {
    name   = "EKS_ADDON_MANAGEMENT"
    status = %[2]q
  }
}
`, featureStatus, additionalConfigurationStatus)
}
