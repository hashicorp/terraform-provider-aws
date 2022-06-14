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

func TestAccWAFGeoMatchSet_basic(t *testing.T) {
	var v waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`geomatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CA",
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

func TestAccWAFGeoMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	geoMatchSetNewName := fmt.Sprintf("geoMatchSetNewName-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", geoMatchSet),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeName(geoMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", geoMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
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

func TestAccWAFGeoMatchSet_disappears(t *testing.T) {
	var v waf.GeoMatchSet
	geoMatchSet := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(geoMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &v),
					testAccCheckGeoMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFGeoMatchSet_changeConstraints(t *testing.T) {
	var before, after waf.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "US",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CA",
					}),
				),
			},
			{
				Config: testAccGeoMatchSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "RU",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "geo_match_constraint.*", map[string]string{
						"type":  "Country",
						"value": "CN",
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

func TestAccWAFGeoMatchSet_noConstraints(t *testing.T) {
	var ipset waf.GeoMatchSet
	setName := fmt.Sprintf("geoMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_geo_match_set.geo_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGeoMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGeoMatchSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "geo_match_constraint.#", "0"),
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

func testAccCheckGeoMatchSetDisappears(v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateGeoMatchSetInput{
				ChangeToken:   token,
				GeoMatchSetId: v.GeoMatchSetId,
			}

			for _, geoMatchConstraint := range v.GeoMatchConstraints {
				geoMatchConstraintUpdate := &waf.GeoMatchSetUpdate{
					Action: aws.String("DELETE"),
					GeoMatchConstraint: &waf.GeoMatchConstraint{
						Type:  geoMatchConstraint.Type,
						Value: geoMatchConstraint.Value,
					},
				}
				req.Updates = append(req.Updates, geoMatchConstraintUpdate)
			}
			return conn.UpdateGeoMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating GeoMatchSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteGeoMatchSetInput{
				ChangeToken:   token,
				GeoMatchSetId: v.GeoMatchSetId,
			}
			return conn.DeleteGeoMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting GeoMatchSet: %s", err)
		}
		return nil
	}
}

func testAccCheckGeoMatchSetExists(n string, v *waf.GeoMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF GeoMatchSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetGeoMatchSet(&waf.GetGeoMatchSetInput{
			GeoMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.GeoMatchSet.GeoMatchSetId == rs.Primary.ID {
			*v = *resp.GeoMatchSet
			return nil
		}

		return fmt.Errorf("WAF GeoMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckGeoMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_geo_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetGeoMatchSet(
			&waf.GetGeoMatchSetInput{
				GeoMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.GeoMatchSet.GeoMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF GeoMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the GeoMatchSet is already destroyed
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccGeoMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"

  geo_match_constraint {
    type  = "Country"
    value = "RU"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CN"
  }
}
`, name)
}

func testAccGeoMatchSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "%s"
}
`, name)
}
