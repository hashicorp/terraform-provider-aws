package waf_test

import (
	"fmt"
	"regexp"
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

func TestAccWAFXSSMatchSet_basic(t *testing.T) {
	var v waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`xssmatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
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

func TestAccWAFXSSMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	xssMatchSetNewName := fmt.Sprintf("xssMatchSetNewName-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
				),
			},
			{
				Config: testAccXSSMatchSetChangeNameConfig(xssMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", xssMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
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

func TestAccWAFXSSMatchSet_disappears(t *testing.T) {
	var v waf.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwaf.ResourceXSSMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFXSSMatchSet_changeTuples(t *testing.T) {
	var before, after waf.XssMatchSet
	setName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccXSSMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"text_transformation":   "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "BODY",
						"text_transformation":   "CMD_LINE",
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

func TestAccWAFXSSMatchSet_noTuples(t *testing.T) {
	var ipset waf.XssMatchSet
	setName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckXSSMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "0"),
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
			return fmt.Errorf("No WAF XSS Match Set ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
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

		return fmt.Errorf("WAF XssMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckXSSMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_xss_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetXssMatchSet(
			&waf.GetXssMatchSetInput{
				XssMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF XssMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the XssMatchSet is already destroyed
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccXSSMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccXSSMatchSetChangeNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccXSSMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuples {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccXSSMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q
}
`, name)
}
