// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFindingsFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionArchive)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
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

func testAccFindingsFilter_nameGenerated(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func testAccFindingsFilter_namePrefix(t *testing.T) {
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
				Config: testAccFindingsFilterConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmacie2.ResourceFindingsFilter(), resourceName),
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
				Config: testAccFindingsFilterConfig_complete(description, string(awstypes.FindingsFilterActionArchive), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionArchive)),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
						"eq.0":          acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, string(awstypes.FindingsFilterActionNoop), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionNoop)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
						"eq.0":          acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_complete(descriptionUpdated, string(awstypes.FindingsFilterActionNoop), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionNoop)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
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
				Config: testAccFindingsFilterConfig_complete(description, string(awstypes.FindingsFilterActionArchive), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionArchive)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
						"eq.0":          acctest.Region(),
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterion(descriptionUpdated, string(awstypes.FindingsFilterActionNoop), startDate, endDate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionNoop)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
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
						"eq.#":          "1",
						"eq.0":          acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"finding_criteria"},
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
				Config: testAccFindingsFilterConfig_complete(description, string(awstypes.FindingsFilterActionArchive), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionArchive)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "finding_criteria.0.criterion.*.eq.*", dataSourceRegion, names.AttrName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
						"eq.0":          acctest.Region(),
					}),
				),
			},
			{
				Config: testAccFindingsFilterConfig_completeMultipleCriterionNumber(descriptionUpdated, string(awstypes.FindingsFilterActionNoop), firstNumber, secondNumber, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, string(awstypes.FindingsFilterActionNoop)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"eq.#":          "1",
						"eq.0":          acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "count",
						"gte":           firstNumber,
						"lt":            secondNumber,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "sample",
						"eq.#":          "1",
						"eq.0":          acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"finding_criteria"},
			},
		},
	})
}

func testAccFindingsFilter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingsFilterDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingsFilterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFindingsFilterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccFindingsFilterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingsFilterExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckFindingsFilterExists(ctx context.Context, n string, v *macie2.GetFindingsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		output, err := tfmacie2.FindFindingsFilterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFindingsFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_findings_filter" {
				continue
			}

			_, err := tfmacie2.FindFindingsFilterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Macie Findings Filter %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFindingsFilterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  name   = %[1]q
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, rName)
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

func testAccFindingsFilterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  name   = %[1]q
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_macie2_account.test]
}
`, rName, tag1Key, tag1Value)
}

func testAccFindingsFilterConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  name   = %[1]q
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_macie2_account.test]
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
