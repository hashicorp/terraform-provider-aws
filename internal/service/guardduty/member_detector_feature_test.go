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

func testAccMemberDetectorFeature_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member_detector_feature.test"
	accountID := testAccMemberAccountFromEnv(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_basic("RDS_LOGIN_EVENTS", "ENABLED", accountID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
				),
			},
		},
	})
}

func testAccMemberDetectorFeature_additionalConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_member_detector_feature.test"
	accountID := testAccMemberAccountFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_additionalConfiguration(accountID, "DISABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.1.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "additional_configuration.1.name", "ECS_FARGATE_AGENT_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "name", "RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccMemberDetectorFeatureConfig_multiple(accountID, "ENABLED", "DISABLED", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccMemberDetectorFeatureExists(ctx, resource1Name),
					testAccMemberDetectorFeatureExists(ctx, resource2Name),
					testAccMemberDetectorFeatureExists(ctx, resource3Name),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resource1Name, "additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resource1Name, "status", "ENABLED"),
					resource.TestCheckResourceAttr(resource1Name, "name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resource1Name, "account_id", accountID),
					resource.TestCheckResourceAttr(resource2Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource2Name, "status", "DISABLED"),
					resource.TestCheckResourceAttr(resource2Name, "name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resource2Name, "account_id", accountID),
					resource.TestCheckResourceAttr(resource3Name, "additional_configuration.#", "0"),
					resource.TestCheckResourceAttr(resource3Name, "status", "ENABLED"),
					resource.TestCheckResourceAttr(resource3Name, "name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resource3Name, "account_id", accountID),
				),
			},
		},
	})
}

func testAccMemberDetectorFeatureExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyClient(ctx)

		_, err := tfguardduty.FindMemberDetectorFeatureByThreePartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["account_id"], rs.Primary.Attributes["name"])

		return err
	}
}

func testAccMemberDetectorFeatureConfig_basic(name, status, accountID string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test" {
  detector_id = data.aws_guardduty_detector.test.id
  name        = "%[1]s"
  status      = "%[2]s"
  account_id  = "%[3]s"
}
`, name, status, accountID))
}

func testAccMemberDetectorFeatureConfig_additionalConfiguration(accountID, eksStatus, ecsStatus string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test" {
	detector_id = data.aws_guardduty_detector.test.id
	name        = "RUNTIME_MONITORING"
	status      = "ENABLED"
	account_id  = "%[1]s"

	additional_configuration {
		name        = "EKS_ADDON_MANAGEMENT"
		status      = "%[2]s"
	}

	additional_configuration {
		name        = "ECS_FARGATE_AGENT_MANAGEMENT"
		status      = "%[3]s"
	}
}
`, accountID, eksStatus, ecsStatus))
}

func testAccMemberDetectorFeatureConfig_multiple(accountID, status1, status2, status3 string) string {
	return acctest.ConfigCompose(testAccMemberDetectorFeatureConfig_base, fmt.Sprintf(`
resource "aws_guardduty_member_detector_feature" "test1" {
	detector_id = data.aws_guardduty_detector.test.id
	name        = "EKS_RUNTIME_MONITORING"
	status      = "%[2]s"
	account_id  = "%[1]s"

	additional_configuration {
		name        = "EKS_ADDON_MANAGEMENT"
		status      = "%[2]s"
	}
}

resource "aws_guardduty_member_detector_feature" "test2" {
	detector_id = data.aws_guardduty_detector.test.id
	name        = "S3_DATA_EVENTS"
	status      = "%[3]s"
	account_id  = "%[1]s"
}

resource "aws_guardduty_member_detector_feature" "test3" {
	detector_id = data.aws_guardduty_detector.test.id
	name        = "LAMBDA_NETWORK_LOGS"
	status      = "%[4]s"
	account_id  = "%[1]s"
}

`, accountID, status1, status2, status3))
}

const testAccMemberDetectorFeatureConfig_base = `

data "aws_guardduty_detector" "test" {}

`
