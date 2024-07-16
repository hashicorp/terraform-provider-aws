// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "guardduty", regexache.MustCompile("detector/.+$")),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_disable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtFalse),
				),
			},
			{
				Config: testAccDetectorConfig_enable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", acctest.CtTrue),
				),
			},
			{
				Config: testAccDetectorConfig_findingPublishingFrequency,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "FIFTEEN_MINUTES"),
				),
			},
		},
	})
}

func testAccDetector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDetectorConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDetector_datasources_s3logs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDetector_datasources_kubernetes_audit_logs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDetector_datasources_malware_protection(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesMalwareProtection(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesMalwareProtection(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDetector_datasources_all(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesAll(true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtTrue),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtFalse),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckDetectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_detector" {
				continue
			}

			_, err := tfguardduty.FindDetectorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GuardDuty Detector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDetectorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		_, err := tfguardduty.FindDetectorByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

const testAccDetectorConfig_basic = `
resource "aws_guardduty_detector" "test" {}
`

const testAccDetectorConfig_disable = `
resource "aws_guardduty_detector" "test" {
  enable = false
}
`

const testAccDetectorConfig_enable = `
resource "aws_guardduty_detector" "test" {
  enable = true
}
`

const testAccDetectorConfig_findingPublishingFrequency = `
resource "aws_guardduty_detector" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
}
`

func testAccDetectorConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccDetectorConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDetectorConfig_datasourcesS3Logs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    s3_logs {
      enable = %[1]t
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesKubernetesAuditLogs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesMalwareProtection(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = %[1]t
        }
      }
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesAll(enableK8s, enableS3, enableMalware bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
    s3_logs {
      enable = %[2]t
    }

    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = %[3]t
        }
      }
    }
  }
}
`, enableK8s, enableS3, enableMalware)
}
