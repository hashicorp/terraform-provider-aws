// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESConfigurationSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", fmt.Sprintf("configuration-set/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tracking_options.#", acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_sending(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
			{
				Config: testAccConfigurationSetConfig_sending(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_reputationMetricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_reputationMetrics(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccConfigurationSetConfig_reputationMetrics(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_deliveryOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, ses.TlsPolicyRequire),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyRequire),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
				),
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, ses.TlsPolicyRequire),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyRequire),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_deliveryOptions(rName, ses.TlsPolicyOptional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
				),
			},
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_emptyDeliveryOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct0),
				),
			},
			{
				Config: testAccConfigurationSetConfig_emptyDeliveryOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceConfigurationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationSetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES configuration set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES configuration set ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		response, err := conn.DescribeConfigurationSetWithContext(ctx, &ses.DescribeConfigurationSetInput{
			ConfigurationSetName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if aws.StringValue(response.ConfigurationSet.Name) != rs.Primary.ID {
			return fmt.Errorf("The configuration set was not created")
		}
		return nil
	}
}

func testAccCheckConfigurationSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_configuration_set" {
				continue
			}

			_, err := conn.DescribeConfigurationSetWithContext(ctx, &ses.DescribeConfigurationSetInput{
				ConfigurationSetName: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if tfawserr.ErrCodeEquals(err, ses.ErrCodeConfigurationSetDoesNotExistException) {
					return nil
				}
				return err
			}
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
