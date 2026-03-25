// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDetectorFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_basic("RDS_LOGIN_EVENTS", "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
				),
			},
		},
	})
}

func testAccDetectorFeature_additionalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("DISABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DISABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
				),
			},
		},
	})
}

func testAccDetectorFeature_additionalConfiguration_newOrder(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_additionalConfiguration_multiple([]string{"EKS_ADDON_MANAGEMENT", "EC2_AGENT_MANAGEMENT", "ECS_FARGATE_AGENT_MANAGEMENT"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "EKS_ADDON_MANAGEMENT", names.AttrStatus: "ENABLED"}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "EC2_AGENT_MANAGEMENT", names.AttrStatus: "ENABLED"}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "ECS_FARGATE_AGENT_MANAGEMENT", names.AttrStatus: "ENABLED"}),
				),
			},
			{
				// Change the order of the additional_configuration blocks and ensure that there is an empty plan
				Config: testAccDetectorFeatureConfig_additionalConfiguration_multiple([]string{"EC2_AGENT_MANAGEMENT", "ECS_FARGATE_AGENT_MANAGEMENT", "EKS_ADDON_MANAGEMENT"}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "EKS_ADDON_MANAGEMENT", names.AttrStatus: "ENABLED"}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "EC2_AGENT_MANAGEMENT", names.AttrStatus: "ENABLED"}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "additional_configuration.*",
						map[string]string{names.AttrName: "ECS_FARGATE_AGENT_MANAGEMENT", names.AttrStatus: "ENABLED"}),
				),
			},
		},
	})
}

func testAccDetectorFeature_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_guardduty_detector_feature.test1"
	resource2Name := "aws_guardduty_detector_feature.test2"
	resource3Name := "aws_guardduty_detector_feature.test3"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorFeatureConfig_multiple("ENABLED", "DISABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resource1Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource2Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrStatus, "ENABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_multiple("DISABLED", "ENABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resource1Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource2Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrStatus, "ENABLED"),
				),
			},
			{
				Config: testAccDetectorFeatureConfig_multiple("DISABLED", "DISABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resource1Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource2Name),
					testAccCheckDetectorFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrStatus, "DISABLED"),
				),
			},
		},
	})
}

func testAccDetectorFeature_additionalConfiguration_migrateListToSet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector_feature.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.GuardDutyServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.16.0", // version before List -> Set change
					},
				},
				Config: testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "1"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccDetectorFeatureConfig_additionalConfiguration("ENABLED", "ENABLED"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckDetectorFeatureExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindDetectorFeatureByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrName])

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

func testAccDetectorFeatureConfig_additionalConfiguration_multiple(configNames []string) string {
	var additionalConfigs strings.Builder
	for _, name := range configNames {
		fmt.Fprintf(&additionalConfigs, `
  additional_configuration {
    name   = %[1]q
    status = "ENABLED"
  }`, name)
	}

	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "RUNTIME_MONITORING"
  status      = "ENABLED"
  %[1]s
}
`, additionalConfigs.String())
}

func testAccDetectorFeatureConfig_multiple(status1, status2, status3 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  enable = true
}

resource "aws_guardduty_detector_feature" "test1" {
  detector_id = aws_guardduty_detector.test.id
  name        = "EBS_MALWARE_PROTECTION"
  status      = %[1]q
}

resource "aws_guardduty_detector_feature" "test2" {
  detector_id = aws_guardduty_detector.test.id
  name        = "LAMBDA_NETWORK_LOGS"
  status      = %[2]q
}

resource "aws_guardduty_detector_feature" "test3" {
  detector_id = aws_guardduty_detector.test.id
  name        = "S3_DATA_EVENTS"
  status      = %[3]q
}
`, status1, status2, status3)
}
