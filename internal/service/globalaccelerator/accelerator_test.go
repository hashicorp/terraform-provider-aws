// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorAccelerator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ipRegex := regexache.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	dnsNameRegex := regexache.MustCompile(`^a[0-9a-f]{16}\.awsglobalaccelerator\.com$`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", ""),
					resource.TestMatchResourceAttr(resourceName, names.AttrDNSName, dnsNameRegex),
					resource.TestCheckResourceAttr(resourceName, "dual_stack_dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrHostedZoneID, "Z2BJ6XQ5FK7U4H"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "ip_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", acctest.Ct2),
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

func TestAccGlobalAcceleratorAccelerator_ipAddressType_dualStack(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dualStackDNSNameRegex := regexache.MustCompile(`^a[0-9a-f]{16}\.dualstack\.awsglobalaccelerator\.com$`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_ipAddressTypeDualStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "dual_stack_dns_name", dualStackDNSNameRegex),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "DUAL_STACK"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.1.ip_addresses.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.1.ip_family", "IPv6"),
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

func TestAccGlobalAcceleratorAccelerator_byoip(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	requestedAddr := os.Getenv("GLOBALACCELERATOR_BYOIP_IPV4_ADDRESS")
	matches := 0

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccCheckBYOIPExists(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_byoip(rName, requestedAddr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					// requested address may be index 0 or index 1 in ip_sets.0.ip_addresses. Test framework
					// does not have a mechanism to test against a list directly. We collect the number of
					// matches individually and then validate that there is only one match.
					resource.TestCheckResourceAttrWith(resourceName, "ip_sets.0.ip_addresses.0", func(value string) error {
						if requestedAddr == value {
							matches += 1
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith(resourceName, "ip_sets.0.ip_addresses.1", func(value string) error {
						if requestedAddr == value {
							matches += 1
						}
						return nil
					}),
					func(_ *terraform.State) error {
						if matches == 1 {
							return nil
						}
						return fmt.Errorf("Requested address %s should be present exactly once in %s", requestedAddr, resourceName)
					},
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorAccelerator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglobalaccelerator.ResourceAccelerator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorAccelerator_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	newName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAcceleratorConfig_enabled(newName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, newName),
				),
			},
			{
				Config: testAccAcceleratorConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorAccelerator_attributes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_globalaccelerator_accelerator.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_attributes(rName, false, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAcceleratorConfig_attributes(rName, true, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
			{
				Config: testAccAcceleratorConfig_attributes(rName, true, "flow-logs-updated/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs-updated/"),
				),
			},
			{
				Config: testAccAcceleratorConfig_attributes(rName, false, "flow-logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorAccelerator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_accelerator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccAcceleratorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAcceleratorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAcceleratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	input := &globalaccelerator.ListAcceleratorsInput{}

	_, err := conn.ListAccelerators(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckBYOIPExists(ctx context.Context, t *testing.T) {
	requestedAddr := os.Getenv("GLOBALACCELERATOR_BYOIP_IPV4_ADDRESS")

	if requestedAddr == "" {
		t.Skip("Environment variable GLOBALACCELERATOR_BYOIP_IPV4_ADDRESS not set")
	}

	parsedAddr := net.ParseIP(requestedAddr)

	conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

	input := &globalaccelerator.ListByoipCidrsInput{}
	cidrs := make([]awstypes.ByoipCidr, 0)

	pages := globalaccelerator.NewListByoipCidrsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		cidrs = append(cidrs, page.ByoipCidrs...)
	}

	if len(cidrs) == 0 {
		t.Skip("skipping acceptance testing: no BYOIP Global Accelerator CIDR found")
	}

	matches := false

	for _, cidr := range cidrs {
		_, network, _ := net.ParseCIDR(aws.ToString(cidr.Cidr))
		if network.Contains(parsedAddr) {
			matches = true
			break
		}
	}

	if !matches {
		t.Skipf("skipping acceptance testing: requested address %s not available via BYOIP", requestedAddr)
	}
}

func testAccCheckAcceleratorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		_, err := tfglobalaccelerator.FindAcceleratorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckAcceleratorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_accelerator" {
				continue
			}

			_, err := tfglobalaccelerator.FindAcceleratorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Accelerator %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccAcceleratorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAcceleratorConfig_byoip(rName string, ipAddress string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name         = %[1]q
  ip_addresses = [%[2]q]
}
`, rName, ipAddress)
}

func testAccAcceleratorConfig_ipAddressTypeDualStack(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "DUAL_STACK"
}
`, rName)
}

func testAccAcceleratorConfig_enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = %[2]t
}
`, rName, enabled)
}

func testAccAcceleratorConfig_attributes(rName string, flowLogsEnabled bool, flowLogsPrefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  attributes {
    flow_logs_enabled   = %[2]t
    flow_logs_s3_bucket = aws_s3_bucket.test.bucket
    flow_logs_s3_prefix = %[3]q
  }
}
`, rName, flowLogsEnabled, flowLogsPrefix)
}

func testAccAcceleratorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAcceleratorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
