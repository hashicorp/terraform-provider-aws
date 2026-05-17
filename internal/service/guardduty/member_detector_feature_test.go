// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMemberDetectorFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member_detector_feature.test"
	accountID := testAccMemberAccountFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_basic("RDS_LOGIN_EVENTS", "ENABLED", accountID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
				),
			},
			{
				Config: testAccMemberDetectorFeatureConfig_basic("RDS_LOGIN_EVENTS", "DISABLED", accountID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
				),
			},
		},
	})
}

func testAccMemberDetectorFeature_additionalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member_detector_feature.test"
	accountID := testAccMemberAccountFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_additionalConfiguration(accountID, "DISABLED", "ENABLED", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.1.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.1.name", "ECS_FARGATE_AGENT_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.2.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.2.name", "EC2_AGENT_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
				),
			},
		},
	})
}

func testAccMemberDetectorFeature_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_guardduty_member_detector_feature.test1"
	resource2Name := "aws_guardduty_member_detector_feature.test2"
	resource3Name := "aws_guardduty_member_detector_feature.test3"
	accountID := testAccMemberAccountFromEnv(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_multiple(accountID, "ENABLED", "DISABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, t, resource1Name),
					testAccMemberDetectorFeatureExists(ctx, t, resource2Name),
					testAccMemberDetectorFeatureExists(ctx, t, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrName, "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrAccountID, accountID),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrName, "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrAccountID, accountID),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrStatus, "ENABLED"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrName, "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrAccountID, accountID),
				),
			},
		},
	})
}

func testAccMemberDetectorFeatureExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindMemberDetectorFeatureByThreePartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccMemberDetectorFeatureConfig_basic(name, status, accountID string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = %[1]q
  status      = %[2]q
  account_id  = %[3]q
}
`, name, status, accountID))
}

func testAccMemberDetectorFeatureConfig_additionalConfiguration(accountID, eksStatus, ecsStatus, ec2Status string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = "RUNTIME_MONITORING"
  status      = "ENABLED"
  account_id  = %[1]q

  additional_configuration {
    name   = "EKS_ADDON_MANAGEMENT"
    status = %[2]q
  }

  additional_configuration {
    name   = "ECS_FARGATE_AGENT_MANAGEMENT"
    status = %[3]q
  }

  additional_configuration {
    name   = "EC2_AGENT_MANAGEMENT"
    status = %[4]q
  }
}
`, accountID, eksStatus, ecsStatus, ec2Status))
}

func testAccMemberDetectorFeatureConfig_multiple(accountID, status1, status2, status3 string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test1" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = "EKS_RUNTIME_MONITORING"
  status      = %[2]q
  account_id  = %[1]q

  additional_configuration {
    name   = "EKS_ADDON_MANAGEMENT"
    status = %[2]q
  }
}

resource "aws_guardduty_member_detector_feature" "test2" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = "S3_DATA_EVENTS"
  status      = %[3]q
  account_id  = %[1]q
}

resource "aws_guardduty_member_detector_feature" "test3" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = "LAMBDA_NETWORK_LOGS"
  status      = %[4]q
  account_id  = %[1]q
}
`, accountID, status1, status2, status3))
}

const testAccMemberDetectorFeatureConfig_base = `
data "aws_guardduty_detector" "test" {}
`
