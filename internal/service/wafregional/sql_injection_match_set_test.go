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

func TestAccWAFRegionalSQLInjectionMatchSet_basic(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", sqlInjectionMatchSet),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
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

func TestAccWAFRegionalSQLInjectionMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))
	sqlInjectionMatchSetNewName := fmt.Sprintf("sqlInjectionMatchSetNewName-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", sqlInjectionMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", "1"),
				),
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", sqlInjectionMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", "1"),
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

func TestAccWAFRegionalSQLInjectionMatchSet_disappears(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &v),
					testAccCheckSQLInjectionMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalSQLInjectionMatchSet_changeTuples(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
					}),
				),
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
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

func TestAccWAFRegionalSQLInjectionMatchSet_noTuples(t *testing.T) {
	var ipset waf.SqlInjectionMatchSet
	resourceName := "aws_wafregional_sql_injection_match_set.sql_injection_match_set"
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(
						resourceName, "name", setName),
					resource.TestCheckResourceAttr(
						resourceName, "sql_injection_match_tuple.#", "0"),
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

func testAccCheckSQLInjectionMatchSetDisappears(v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		region := acctest.Provider.Meta().(*conns.AWSClient).Region

		wr := tfwafregional.NewRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}

			for _, sqlInjectionMatchTuple := range v.SqlInjectionMatchTuples {
				sqlInjectionMatchTupleUpdate := &waf.SqlInjectionMatchSetUpdate{
					Action: aws.String("DELETE"),
					SqlInjectionMatchTuple: &waf.SqlInjectionMatchTuple{
						FieldToMatch:       sqlInjectionMatchTuple.FieldToMatch,
						TextTransformation: sqlInjectionMatchTuple.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, sqlInjectionMatchTupleUpdate)
			}
			return conn.UpdateSqlInjectionMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Failed updating SQL Injection Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}
			return conn.DeleteSqlInjectionMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting SQL Injection Match Set: %s", err)
		}
		return nil
	}
}

func testAccCheckSQLInjectionMatchSetExists(n string, v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SqlInjectionMatchSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetSqlInjectionMatchSet(&waf.GetSqlInjectionMatchSetInput{
			SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
			*v = *resp.SqlInjectionMatchSet
			return nil
		}

		return fmt.Errorf("WAF SqlInjectionMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSQLInjectionMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_sql_injection_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetSqlInjectionMatchSet(
			&waf.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF SqlInjectionMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the SqlInjectionMatchSet is already destroyed
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccSQLInjectionMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"

  sql_injection_match_tuple {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"

  sql_injection_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "HEADER"
      data = "User-Agent"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"
}
`, name)
}
