// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFIPSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile("ipset/.+$")),
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

func TestAccWAFIPSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
					testAccCheckIPSetDisappears(ctx, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFIPSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
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
			{
				Config: testAccIPSetConfig_changeName(uName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, uName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.7.0/24",
					}),
				),
			},
		},
	})
}

func TestAccWAFIPSet_changeDescriptors(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
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
			{
				Config: testAccIPSetConfig_changeDescriptors(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						names.AttrType:  "IPV4",
						names.AttrValue: "192.0.8.0/24",
					}),
				),
			},
		},
	})
}

func TestAccWAFIPSet_noDescriptors(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_noDescriptors(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "0"),
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

func TestAccWAFIPSet_IPSetDescriptors_1000UpdateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

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
		ipSetDescriptors = append(ipSetDescriptors, fmt.Sprintf("ip_set_descriptors {\ntype=\"IPV4\"\nvalue=\"%s/32\"\n}", ip))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_ipSetDescriptors(rName, strings.Join(ipSetDescriptors, "\n")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "2048"),
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

func TestDiffIPSetDescriptors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates [][]awstypes.IPSetUpdate
	}{
		{
			// Change
			Old: []interface{}{
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "192.0.7.0/24"},
			},
			New: []interface{}{
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "192.0.8.0/24"},
			},
			ExpectedUpdates: [][]awstypes.IPSetUpdate{
				{
					{
						Action: awstypes.ChangeActionDelete,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: awstypes.ChangeActionInsert,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("192.0.8.0/24"),
						},
					},
				},
			},
		},
		{
			// Fresh IPSet
			Old: []interface{}{},
			New: []interface{}{
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "10.0.1.0/24"},
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "10.0.2.0/24"},
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "10.0.3.0/24"},
			},
			ExpectedUpdates: [][]awstypes.IPSetUpdate{
				{
					{
						Action: awstypes.ChangeActionInsert,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("10.0.1.0/24"),
						},
					},
					{
						Action: awstypes.ChangeActionInsert,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("10.0.2.0/24"),
						},
					},
					{
						Action: awstypes.ChangeActionInsert,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("10.0.3.0/24"),
						},
					},
				},
			},
		},
		{
			// Deletion
			Old: []interface{}{
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "192.0.7.0/24"},
				map[string]interface{}{names.AttrType: "IPV4", names.AttrValue: "192.0.8.0/24"},
			},
			New: []interface{}{},
			ExpectedUpdates: [][]awstypes.IPSetUpdate{
				{
					{
						Action: awstypes.ChangeActionDelete,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: awstypes.ChangeActionDelete,
						IPSetDescriptor: &awstypes.IPSetDescriptor{
							Type:  awstypes.IPSetDescriptorTypeIpv4,
							Value: aws.String("192.0.8.0/24"),
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			updates := tfwaf.DiffIPSetDescriptors(tc.Old, tc.New)
			if !reflect.DeepEqual(updates, tc.ExpectedUpdates) {
				t.Fatalf("IPSet updates don't match.\nGiven: %v\nExpected: %v",
					updates, tc.ExpectedUpdates)
			}
		})
	}
}

func TestAccWAFIPSet_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_ipV6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(ctx, resourceName, &v),
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

func testAccCheckIPSetDisappears(ctx context.Context, v *awstypes.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}

			for _, IPSetDescriptor := range v.IPSetDescriptors {
				IPSetUpdate := awstypes.IPSetUpdate{
					Action: awstypes.ChangeAction("DELETE"),
					IPSetDescriptor: &awstypes.IPSetDescriptor{
						Type:  IPSetDescriptor.Type,
						Value: IPSetDescriptor.Value,
					},
				}
				req.Updates = append(req.Updates, IPSetUpdate)
			}

			return conn.UpdateIPSet(ctx, req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF IPSet: %s", err)
		}

		_, err = wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			opts := &waf.DeleteIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}
			return conn.DeleteIPSet(ctx, opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF IPSet: %s", err)
		}
		return nil
	}
}

func testAccCheckIPSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_ipset" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)
			resp, err := conn.GetIPSet(ctx, &waf.GetIPSetInput{
				IPSetId: aws.String(rs.Primary.ID),
			})

			if err == nil {
				if *resp.IPSet.IPSetId == rs.Primary.ID {
					return fmt.Errorf("WAF IPSet %s still exists", rs.Primary.ID)
				}
			}

			// Return nil if the IPSet is already destroyed
			if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
				return nil
			}

			return err
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF IPSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)
		resp, err := conn.GetIPSet(ctx, &waf.GetIPSetInput{
			IPSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.IPSet.IPSetId == rs.Primary.ID {
			*v = *resp.IPSet
			return nil
		}

		return fmt.Errorf("WAF IPSet (%s) not found", rs.Primary.ID)
	}
}

func testAccIPSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_changeDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.8.0/24"
  }
}
`, name)
}

func testAccIPSetConfig_ipSetDescriptors(name, ipSetDescriptors string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  %s
}
`, name, ipSetDescriptors)
}

func testAccIPSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q
}
`, name)
}

func testAccIPSetConfig_ipV6(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV6"
    value = "1234:5678:9abc:6811:0000:0000:0000:0000/64"
  }
}
`, name)
}
