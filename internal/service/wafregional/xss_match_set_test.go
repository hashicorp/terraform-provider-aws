package wafregional_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
)

func TestAccWAFRegionalXSSMatchSet_basic(t *testing.T) {
	var v waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
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

func TestAccWAFRegionalXSSMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.XssMatchSet
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccXSSMatchSetConfig_changeName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_disappears(t *testing.T) {
	var v waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafregional.ResourceXSSMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_changeTuples(t *testing.T) {
	var before, after waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccXSSMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"text_transformation":   "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "BODY",
						"text_transformation":   "CMD_LINE",
					}),
				),
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_noTuples(t *testing.T) {
	var ipset waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "0"),
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

func testAccCheckXSSMatchSetExists(n string, v *waf.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No regional WAF XSS Match Set ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetXssMatchSet(&waf.GetXssMatchSetInput{
			XssMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
			*v = *resp.XssMatchSet
			return nil
		}

		return fmt.Errorf("Regional WAF XSS Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckXSSMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_xss_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetXssMatchSet(
			&waf.GetXssMatchSetInput{
				XssMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
				return fmt.Errorf("Regional WAF XSS Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the regional WAF XSS Match Set is already destroyed
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccXSSMatchSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_changeName(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_changeTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuple {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_noTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q
}
`, rName)
}
