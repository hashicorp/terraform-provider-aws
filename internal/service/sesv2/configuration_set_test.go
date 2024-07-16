// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ConfigurationSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", regexache.MustCompile(`configuration-set/.+`)),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceConfigurationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_tlsPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_tlsPolicy(rName, string(types.TlsPolicyRequire)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(types.TlsPolicyRequire)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetConfig_tlsPolicy(rName, string(types.TlsPolicyOptional)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", string(types.TlsPolicyOptional)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_reputationMetricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_reputationMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.#", acctest.Ct1),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "reputation_options.0.reputation_metrics_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_sendingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_sendingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_options.#", acctest.Ct1),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sending_options.0.sending_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_suppressedReasons(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_suppressedReasons(rName, string(types.SuppressionListReasonBounce)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.#", acctest.Ct1),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "suppression_options.0.suppressed_reasons.0", string(types.SuppressionListReasonComplaint)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_engagementMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_engagementMetrics(rName, string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.#", acctest.Ct1),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.dashboard_options.0.engagement_metrics", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_optimizedSharedDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_optimizedSharedDelivery(rName, string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.#", acctest.Ct1),
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
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vdm_options.0.guardian_options.0.optimized_shared_delivery", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
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
				Config: testAccConfigurationSetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConfigurationSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConfigurationSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_configuration_set" {
				continue
			}

			_, err := tfsesv2.FindConfigurationSetByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameConfigurationSet, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConfigurationSetExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSet, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSet, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindConfigurationSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSet, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccConfigurationSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}
`, rName)
}

func testAccConfigurationSetConfig_tlsPolicy(rName, tlsPolicy string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  delivery_options {
    tls_policy = %[2]q
  }
}
`, rName, tlsPolicy)
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

func testAccConfigurationSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConfigurationSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
