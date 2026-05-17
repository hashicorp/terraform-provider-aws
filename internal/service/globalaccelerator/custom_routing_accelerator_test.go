// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorCustomRoutingAccelerator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ipRegex := regexache.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	dnsNameRegex := regexache.MustCompile(`^a[0-9a-f]{16}\.awsglobalaccelerator\.com$`)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "globalaccelerator", regexache.MustCompile(`accelerator/`+verify.UUIDRegexPattern+`$`)),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, dnsNameRegex),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrHostedZoneID, "Z2BJ6XQ5FK7U4H"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.0", ipRegex),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.1", ipRegex),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccGlobalAcceleratorCustomRoutingAccelerator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglobalaccelerator.ResourceCustomRoutingAccelerator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingAccelerator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccCustomRoutingAcceleratorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCustomRoutingAcceleratorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingAccelerator_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func testAccCheckCustomRoutingAcceleratorExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		_, err := tfglobalaccelerator.FindCustomRoutingAcceleratorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCustomRoutingAcceleratorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_custom_routing_accelerator" {
				continue
			}

			_, err := tfglobalaccelerator.FindCustomRoutingAcceleratorByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Custom Routing Accelerator %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCustomRoutingAcceleratorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCustomRoutingAcceleratorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCustomRoutingAcceleratorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
