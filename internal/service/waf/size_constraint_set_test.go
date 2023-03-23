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

func TestAccWAFSizeConstraintSet_basic(t *testing.T) {
	var v waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`sizeconstraintset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						"size":                "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
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

func TestAccWAFSizeConstraintSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	sizeConstraintSetNewName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
				),
			},
			{
				Config: testAccSizeConstraintSetChangeNameConfig(sizeConstraintSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSetNewName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
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

func TestAccWAFSizeConstraintSet_disappears(t *testing.T) {
	var v waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &v),
					testAccCheckSizeConstraintSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFSizeConstraintSet_changeConstraints(t *testing.T) {
	var before, after waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						"size":                "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
					}),
				),
			},
			{
				Config: testAccSizeConstraintSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "GE",
						"field_to_match.#":    "1",
						"size":                "1024",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
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

func TestAccWAFSizeConstraintSet_noConstraints(t *testing.T) {
	var contraints waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(resourceName, &contraints),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "0"),
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

func testAccCheckSizeConstraintSetDisappears(v *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSizeConstraintSetInput{
				ChangeToken:         token,
				SizeConstraintSetId: v.SizeConstraintSetId,
			}

			for _, sizeConstraint := range v.SizeConstraints {
				sizeConstraintUpdate := &waf.SizeConstraintSetUpdate{
					Action: aws.String("DELETE"),
					SizeConstraint: &waf.SizeConstraint{
						FieldToMatch:       sizeConstraint.FieldToMatch,
						ComparisonOperator: sizeConstraint.ComparisonOperator,
						Size:               sizeConstraint.Size,
						TextTransformation: sizeConstraint.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, sizeConstraintUpdate)
			}
			return conn.UpdateSizeConstraintSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating SizeConstraintSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSizeConstraintSetInput{
				ChangeToken:         token,
				SizeConstraintSetId: v.SizeConstraintSetId,
			}
			return conn.DeleteSizeConstraintSet(opts)
		})

		return err
	}
}

func testAccCheckSizeConstraintSetExists(n string, v *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SizeConstraintSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetSizeConstraintSet(&waf.GetSizeConstraintSetInput{
			SizeConstraintSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.SizeConstraintSet.SizeConstraintSetId == rs.Primary.ID {
			*v = *resp.SizeConstraintSet
			return nil
		}

		return fmt.Errorf("WAF SizeConstraintSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSizeConstraintSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_size_contraint_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetSizeConstraintSet(
			&waf.GetSizeConstraintSetInput{
				SizeConstraintSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.SizeConstraintSet.SizeConstraintSetId == rs.Primary.ID {
				return fmt.Errorf("WAF SizeConstraintSet %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			continue
		}

		return err
	}

	return nil
}

func testAccSizeConstraintSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetChangeNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "GE"
    size                = "1024"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"
}
`, name)
}
