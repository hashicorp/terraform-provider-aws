package macie2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
)

func testAccFindingsFilter_basic(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_namePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					acctest.CheckResourceDisappears(acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccFindingsFilter_complete(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterion(descriptionUpdated, macie2.FindingsFilterActionNoop, startDate, endDate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "updatedAt",
						"gte":   startDate,
						"lt":    endDate,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]*regexp.Regexp{
						"gte": regexp.MustCompile(acctest.RFC3339RegexPattern),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]*regexp.Regexp{
						"lt": regexp.MustCompile(acctest.RFC3339RegexPattern),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "sample",
						"eq.#":  "1",
						"eq.0":  "true",
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	firstNumber := "12"
	secondNumber := "13"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_complete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, "name"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterionNumber(descriptionUpdated, macie2.FindingsFilterActionNoop, firstNumber, secondNumber, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "count",
						"gte":   firstNumber,
						"lt":    secondNumber,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "sample",
						"eq.#":  "1",
						"eq.0":  "true",
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
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	description := "this is a description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingsFilterDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_tags(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(resourceName, &macie2Output),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  acctest.Region(),
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

func testAccCheckFindingsFilterExists(resourceName string, macie2Session *macie2.GetFindingsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn
		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}

		resp, err := conn.GetFindingsFilter(input)

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

func testAccCheckFindingsFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_findings_filter" {
			continue
		}

		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetFindingsFilter(input)

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
