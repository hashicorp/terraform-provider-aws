// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rum_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/rum/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchrum "github.com/hashicorp/terraform-provider-aws/internal/service/rum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRUMAppMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "cw_log_group"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_customEvents(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "DISABLED"),
				),
			},
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "ENABLED"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_domainList(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_domainList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "domain_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "domain_list.0", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "domain_list.1", "terraform.*"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Updating by removing the domain_list and adding the domain
				Config: testAccAppMonitorConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "cw_log_group"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
			{
				// Updating by removing the domain and adding the domain list again
				Config: testAccAppMonitorConfig_domainList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "domain_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "domain_list.0", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "domain_list.1", "terraform.*"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppMonitorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_deobfuscationConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_deobfuscationConfiguration(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.0.s3_uri", fmt.Sprintf("s3://%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_deobfuscationConfiguration(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deobfuscation_configuration.0.javascript_source_maps.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", "1"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon awstypes.AppMonitor
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, t, resourceName, &appMon),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppMonitorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RUMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rum_app_monitor" {
				continue
			}

			_, err := tfcloudwatchrum.FindAppMonitorByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch RUM App Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppMonitorExists(ctx context.Context, t *testing.T, n string, v *awstypes.AppMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RUMClient(ctx)

		output, err := tfcloudwatchrum.FindAppMonitorByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"
}
`, rName)
}

func testAccAppMonitorConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name           = %[1]q
  domain         = "localhost"
  cw_log_enabled = true
}
`, rName)
}

func testAccAppMonitorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppMonitorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppMonitorConfig_customEvents(rName, enabled string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  custom_events {
    status = %[2]q
  }
}
`, rName, enabled)
}

func testAccAppMonitorConfig_domainList(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name        = %[1]q
  domain_list = ["localhost", "terraform.*"]
}
`, rName)
}

func testAccAppMonitorConfig_deobfuscationConfiguration(rName, status string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = "example-source-map.js.map"
  content = "dummy content for source map"
}

data "aws_iam_policy_document" "test" {
  statement {
    principals {
      identifiers = ["rum.amazonaws.com"]
      type        = "Service"
    }
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceArn"
      values   = ["arn:${data.aws_partition.current.partition}:rum:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:appmonitor/%[1]s"]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"
  deobfuscation_configuration {
    javascript_source_maps {
      status = %[2]q
      s3_uri = "s3://${aws_s3_object.test.bucket}"
    }
  }
}
`, rName, status)
}
