package waf_test

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

func TestAccWAFIPSet_basic(t *testing.T) {
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile("ipset/.+$")),
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
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					testAccCheckIPSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFIPSet_changeNameForceNew(t *testing.T) {
	var before, after waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
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
					testAccCheckIPSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", uName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
				),
			},
		},
	})
}

func TestAccWAFIPSet_changeDescriptors(t *testing.T) {
	var before, after waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
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
					testAccCheckIPSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.8.0/24",
					}),
				),
			},
		},
	})
}

func TestAccWAFIPSet_noDescriptors(t *testing.T) {
	var ipset waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_noDescriptors(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	var ipset waf.IPSet
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_IPSetDescriptors(rName, strings.Join(ipSetDescriptors, "\n")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &ipset),
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
	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates [][]*waf.IPSetUpdate
	}{
		{
			// Change
			Old: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.7.0/24"},
			},
			New: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.8.0/24"},
			},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
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
				map[string]interface{}{"type": "IPV4", "value": "10.0.1.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "10.0.2.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "10.0.3.0/24"},
			},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.1.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.2.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.3.0/24"),
						},
					},
				},
			},
		},
		{
			// Deletion
			Old: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.7.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "192.0.8.0/24"},
			},
			New: []interface{}{},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String(waf.IPSetDescriptorTypeIpv4),
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String(waf.IPSetDescriptorTypeIpv4),
							Value: aws.String("192.0.8.0/24"),
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			updates := tfwaf.DiffIPSetDescriptors(tc.Old, tc.New)
			if !reflect.DeepEqual(updates, tc.ExpectedUpdates) {
				t.Fatalf("IPSet updates don't match.\nGiven: %s\nExpected: %s",
					updates, tc.ExpectedUpdates)
			}
		})
	}
}

func TestAccWAFIPSet_ipv6(t *testing.T) {
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig_ipV6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
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

func testAccCheckIPSetDisappears(v *waf.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}

			for _, IPSetDescriptor := range v.IPSetDescriptors {
				IPSetUpdate := &waf.IPSetUpdate{
					Action: aws.String("DELETE"),
					IPSetDescriptor: &waf.IPSetDescriptor{
						Type:  IPSetDescriptor.Type,
						Value: IPSetDescriptor.Value,
					},
				}
				req.Updates = append(req.Updates, IPSetUpdate)
			}

			return conn.UpdateIPSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF IPSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}
			return conn.DeleteIPSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF IPSet: %s", err)
		}
		return nil
	}
}

func testAccCheckIPSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_ipset" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetIPSet(
			&waf.GetIPSetInput{
				IPSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.IPSet.IPSetId == rs.Primary.ID {
				return fmt.Errorf("WAF IPSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the IPSet is already destroyed
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckIPSetExists(n string, v *waf.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF IPSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetIPSet(&waf.GetIPSetInput{
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

func testAccIPSetConfig_IPSetDescriptors(name, ipSetDescriptors string) string {
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
