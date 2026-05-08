// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRaySamplingRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSamplingRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSamplingRuleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "xray", fmt.Sprintf("sampling-rule/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "5"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(resourceName, "reservoir_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, "*"),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccXRaySamplingRule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	updatedPriority := acctest.RandIntRange(t, 0, 9999)
	updatedReservoirSize := acctest.RandIntRange(t, 0, 2147483647)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSamplingRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_update(rName, acctest.RandIntRange(t, 0, 9999), acctest.RandIntRange(t, 0, 2147483647)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSamplingRuleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "xray", fmt.Sprintf("sampling-rule/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPriority),
					resource.TestCheckResourceAttrSet(resourceName, "reservoir_size"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, "*"),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
				),
			},
			{ // Update attributes
				Config: testAccSamplingRuleConfig_update(rName, updatedPriority, updatedReservoirSize),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSamplingRuleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "xray", fmt.Sprintf("sampling-rule/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, strconv.Itoa(updatedPriority)),
					resource.TestCheckResourceAttr(resourceName, "reservoir_size", strconv.Itoa(updatedReservoirSize)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(resourceName, "url_path", "*"),
					resource.TestCheckResourceAttr(resourceName, "host", "*"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "fixed_rate", "0.3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceARN, "*"),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "*"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "*"),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
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

func TestAccXRaySamplingRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.SamplingRule
	resourceName := "aws_xray_sampling_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSamplingRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSamplingRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSamplingRuleExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfxray.ResourceSamplingRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSamplingRuleExists(ctx context.Context, t *testing.T, n string, v *types.SamplingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Sampling Rule ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).XRayClient(ctx)

		output, err := tfxray.FindSamplingRuleByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSamplingRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_xray_sampling_rule" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).XRayClient(ctx)

			_, err := tfxray.FindSamplingRuleByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("XRay Sampling Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSamplingRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = %[1]q
  priority       = 5
  reservoir_size = 10
  url_path       = "*"
  host           = "*"
  http_method    = "GET"
  service_type   = "*"
  service_name   = "*"
  fixed_rate     = 0.3
  resource_arn   = "*"
  version        = 1

  attributes = {
    Hello = "World"
  }
}
`, rName)
}

func testAccSamplingRuleConfig_update(rName string, priority, reservoirSize int) string {
	return fmt.Sprintf(`
resource "aws_xray_sampling_rule" "test" {
  rule_name      = %[1]q
  priority       = %[2]d
  reservoir_size = %[3]d
  url_path       = "*"
  host           = "*"
  http_method    = "GET"
  service_type   = "*"
  service_name   = "*"
  fixed_rate     = 0.3
  resource_arn   = "*"
  version        = 1
}
`, rName, priority, reservoirSize)
}
