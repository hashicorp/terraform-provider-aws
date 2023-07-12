// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlobalAcceleratorCustomRoutingAccelerator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	dnsNameRegex := regexp.MustCompile(`^a[a-f0-9]{16}\.awsglobalaccelerator\.com$`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", ""),
					resource.TestMatchResourceAttr(resourceName, "dns_name", dnsNameRegex),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", "Z2BJ6XQ5FK7U4H"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.0", ipRegex),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.1", ipRegex),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglobalaccelerator.ResourceCustomRoutingAccelerator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingAccelerator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomRoutingAcceleratorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCustomRoutingAcceleratorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingAccelerator_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
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
					testAccCheckCustomRoutingAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccCheckCustomRoutingAcceleratorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Global Accelerator Custom Routing Accelerator ID is set")
		}

		_, err := tfglobalaccelerator.FindCustomRoutingAcceleratorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCustomRoutingAcceleratorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_custom_routing_accelerator" {
				continue
			}

			_, err := tfglobalaccelerator.FindCustomRoutingAcceleratorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
