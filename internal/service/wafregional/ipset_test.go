// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalIPSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_ipset.ipset"
	var v awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipsetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptor.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "waf-regional", regexache.MustCompile("ipset/.+$")),
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

func TestAccWAFRegionalIPSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_ipset.ipset"
	var v awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipsetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceIPSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalIPSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_ipset.ipset"
	var before, after awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	ipsetNewName := fmt.Sprintf("ip-set-new-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipsetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptor.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
				),
			},
			{
				Config: testAccIPSetConfig_changeName(ipsetNewName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetNewName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptor.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
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

func TestAccWAFRegionalIPSet_changeDescriptors(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_ipset.ipset"
	var before, after awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(ipsetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptor.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptor.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
				),
			},
			{
				Config: testAccIPSetConfig_changeDescriptors(ipsetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptor.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptor.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.8.0/24",
					}),
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

func TestAccWAFRegionalIPSet_IPSetDescriptors_1000UpdateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_ipset.ipset"

	incrementIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	// Generate 2048 IPs
	ip, ipnet, err := net.ParseCIDR("10.0.0.0/21")
	if err != nil {
		t.Fatal(err)
	}
	ipSetDescriptors := make([]string, 0, 2048)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ipSetDescriptors = append(ipSetDescriptors, fmt.Sprintf("ip_set_descriptor {\ntype=\"IPV4\"\nvalue=\"%s/32\"\n}", ip))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_ipSetDescriptors(ipsetName, strings.Join(ipSetDescriptors, "\n")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptor.#", "2048"),
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

func TestAccWAFRegionalIPSet_noDescriptors(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_wafregional_ipset.ipset"
	var ipset awstypes.IPSet
	ipsetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_noDescriptors(ipsetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ipsetName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptor.#", acctest.Ct0),
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

func testAccCheckIPSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_ipset" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindIPSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional IPSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIPSetExists(ctx context.Context, n string, v *awstypes.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindIPSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIPSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_changeDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.8.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_ipSetDescriptors(name, ipSetDescriptors string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q
  %[2]s
}
`, name, ipSetDescriptors)
}

func testAccIPSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q
}
`, name)
}
