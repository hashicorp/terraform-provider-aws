// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDetectorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:                acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_basicResource(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccDetectorDataSourceConfig_basicResource2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					acctest.CheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.#", "9"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.0.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.1.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.2.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.3.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.4.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.5.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.6.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.enable", "false"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.8.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDetectorDataSource_ID(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_explicit(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					acctest.CheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.#", "9"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.0.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.1.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.2.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.3.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.4.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.5.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.6.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.enable", "false"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "features.8.enable", "true"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDetectorDataSourceConfig_basicResource() string {
	return `
resource "aws_guardduty_detector" "test" {}
`
}

func testAccDetectorDataSourceConfig_basicResource2() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {}
`
}

func testAccDetectorDataSourceConfig_explicit() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {
  id = aws_guardduty_detector.test.id
}
`
}
