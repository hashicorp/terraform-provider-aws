// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ConfigurationSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`configuration-set/.+`)),
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

func TestAccSESV2ConfigurationSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceConfigurationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_deliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, string(types.TlsPolicyRequire)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(types.TlsPolicyRequire)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions_maxDeliverySeconds(rName, 300, string(types.TlsPolicyRequire)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.max_delivery_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(types.TlsPolicyRequire)),
				),
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions_maxDeliverySeconds(rName, 800, string(types.TlsPolicyOptional)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.max_delivery_seconds", "800"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(types.TlsPolicyOptional)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_reputationMetricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_reputationMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.0.reputation_metrics_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_reputationMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.0.reputation_metrics_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_sendingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_sendingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sending_options.0.sending_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_sendingEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sending_options.0.sending_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_suppressedReasons(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_suppressedReasons(rName, string(types.SuppressionListReasonBounce)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.0", string(types.SuppressionListReasonBounce)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_suppressedReasons(rName, string(types.SuppressionListReasonComplaint)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.0", string(types.SuppressionListReasonComplaint)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_suppressedReasonsEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_suppressedReasonsEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.#", "0"),
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

func TestAccSESV2ConfigurationSet_engagementMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_engagementMetrics(rName, string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.0.engagement_metrics", string(types.FeatureStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_engagementMetrics(rName, string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.0.engagement_metrics", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_optimizedSharedDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_optimizedSharedDelivery(rName, string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.0.optimized_shared_delivery", string(types.FeatureStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_optimizedSharedDelivery(rName, string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.0.optimized_shared_delivery", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func testAccConfigurationSetConfig_suppressedReasonsEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
  suppression_options {
    suppressed_reasons = []
  }
}
`, rName)
}

func testAccCheckConfigurationSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_configuration_set" {
				continue
			}

			_, err := tfsesv2.FindConfigurationSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Configuration Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConfigurationSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindConfigurationSetByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccConfigurationSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}
`, rName)
}

func testAccConfigurationSetConfig_deliveryOptions(rName string, tlsPolicy string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  delivery_options {
    tls_policy = %[2]q
  }
}
`, rName, tlsPolicy)
}

func testAccConfigurationSetConfig_deliveryOptions_maxDeliverySeconds(rName string, maxDeliverySeconds int, tlsPolicy string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  delivery_options {
    max_delivery_seconds = %[2]d
    tls_policy           = %[3]q
  }
}
`, rName, maxDeliverySeconds, tlsPolicy)
}

func testAccConfigurationSetConfig_reputationMetricsEnabled(rName string, reputationMetricsEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  reputation_options {
    reputation_metrics_enabled = %[2]t
  }
}
`, rName, reputationMetricsEnabled)
}

func testAccConfigurationSetConfig_sendingEnabled(rName string, sendingEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  sending_options {
    sending_enabled = %[2]t
  }
}
`, rName, sendingEnabled)
}

func testAccConfigurationSetConfig_suppressedReasons(rName, suppressedReason string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  suppression_options {
    suppressed_reasons = [%[2]q]
  }
}
`, rName, suppressedReason)
}

func testAccConfigurationSetConfig_engagementMetrics(rName, engagementMetrics string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  vdm_options {
    dashboard_options {
      engagement_metrics = %[2]q
    }
  }
}
`, rName, engagementMetrics)
}

func testAccConfigurationSetConfig_optimizedSharedDelivery(rName, optimizedSharedDelivery string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  vdm_options {
    guardian_options {
      optimized_shared_delivery = %[2]q
    }
  }
}
`, rName, optimizedSharedDelivery)
}
