// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFindingsFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionArchive),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
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

func testAccFindingsFilter_Name_Generated(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
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

func testAccFindingsFilter_NamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_namePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, namePrefix),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, namePrefix),
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

func testAccFindingsFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccFindingsFilter_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
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

func testAccFindingsFilter_WithDate(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionArchive),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterion(descriptionUpdated, macie2.FindingsFilterActionNoop, startDate, endDate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "updatedAt",
						"gte":           startDate,
						"lt":            endDate,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]*regexp.Regexp{
						"gte": regexache.MustCompile(acctest.RFC3339RegexPattern),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]*regexp.Regexp{
						"lt": regexache.MustCompile(acctest.RFC3339RegexPattern),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "sample",
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.CtTrue,
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

func testAccFindingsFilter_WithNumber(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	firstNumber := "12"
	secondNumber := "13"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionArchive),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterionNumber(descriptionUpdated, macie2.FindingsFilterActionNoop, firstNumber, secondNumber, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "count",
						"gte":           firstNumber,
						"lt":            secondNumber,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "sample",
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.CtTrue,
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

func testAccFindingsFilter_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	description := "this is a description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_tags(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", names.AttrValue),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "position", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          acctest.Ct1,
						"eq.0":          acctest.Region(),
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

func testAccCheckFindingsFilterExists(ctx context.Context, resourceName string, macie2Session *macie2.GetFindingsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)
		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}

		resp, err := conn.GetFindingsFilterWithContext(ctx, input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie FindingsFilter %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccCheckFindingsFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_findings_filter" {
				continue
			}

			input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}
			resp, err := conn.GetFindingsFilterWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
				tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil {
				return fmt.Errorf("macie FindingsFilter %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccFindingsFilterConfig_nameGenerated() string {
	return `
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }
  depends_on = [aws_macie2_account.test]
}
`
}

func testAccFindingsFilterConfig_namePrefix(name string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  name_prefix = %[1]q
  action      = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, name)
}

func testAccFindingsFilterConfig_complete(description, action string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position)
}

func testAccFindingsFilterConfig_completeMultipleCriterion(description, action, startDate, endDate string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
    criterion {
      field = "sample"
      eq    = ["true"]
    }
    criterion {
      field = "updatedAt"
      gte   = %[4]q
      lt    = %[5]q
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position, startDate, endDate)
}

func testAccFindingsFilterConfig_completeMultipleCriterionNumber(description, action, firstNum, secondNum string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
    criterion {
      field = "sample"
      eq    = ["true"]
    }
    criterion {
      field = "count"
      gte   = %[4]q
      lt    = %[5]q
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position, firstNum, secondNum)
}

func testAccFindingsFilterConfig_tags(description, action string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
  }
  tags = {
    Key = "value"
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position)
}
