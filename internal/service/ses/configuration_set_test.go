// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESConfigurationSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", fmt.Sprintf("configuration-set/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tracking_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
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

func TestAccSESConfigurationSet_sendingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_sending(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_sending(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
			{
				Config: testAccConfigurationSetConfig_sending(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_reputationMetricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_reputationMetrics(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_reputationMetrics(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccConfigurationSetConfig_reputationMetrics(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_deliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, string(awstypes.TlsPolicyRequire)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(awstypes.TlsPolicyRequire)),
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

func TestAccSESConfigurationSet_Update_deliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, string(awstypes.TlsPolicyRequire)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(awstypes.TlsPolicyRequire)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, string(awstypes.TlsPolicyOptional)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(awstypes.TlsPolicyOptional)),
				),
			},
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
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

func TestAccSESConfigurationSet_emptyDeliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_emptyDeliveryOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(awstypes.TlsPolicyOptional)),
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

func TestAccSESConfigurationSet_Update_emptyDeliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
				),
			},
			{
				Config: testAccConfigurationSetConfig_emptyDeliveryOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(awstypes.TlsPolicyOptional)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
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

/*
// TestAccSESConfigurationSet_trackingOptions requires a verified domain
// which poses a problem for testing.
func TestAccSESConfigurationSet_trackingOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_trackingOptions(rName, "wn011su7.test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tracking_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracking_options.0.custom_redirect_domain", rName),
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
*/

func TestAccSESConfigurationSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceConfigurationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		_, err := tfses.FindConfigurationSetByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckConfigurationSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_configuration_set" {
				continue
			}

			_, err := tfses.FindConfigurationSetByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Configuration Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigurationSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}
`, rName)
}

func testAccConfigurationSetConfig_sending(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name            = %[1]q
  sending_enabled = %[2]t
}
`, rName, enabled)
}

func testAccConfigurationSetConfig_reputationMetrics(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name                       = %[1]q
  reputation_metrics_enabled = %[2]t
}
`, rName, enabled)
}

func testAccConfigurationSetConfig_deliveryOptions(rName, tlsPolicy string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  delivery_options {
    tls_policy = %[2]q
  }
}
`, rName, tlsPolicy)
}

/*
// this cannot currently be tested without a verified domain
func testAccConfigurationSetConfig_trackingOptions(rName, customRedirect string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  tracking_options {
    custom_redirect_domain = %[2]q
  }
}
`, rName, customRedirect)
}
*/

func testAccConfigurationSetConfig_emptyDeliveryOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  delivery_options {}
}
`, rName)
}
