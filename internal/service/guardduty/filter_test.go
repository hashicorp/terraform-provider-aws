// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	acmpca_types "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"
	detectorResourceName := "aws_guardduty_detector.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "test-filter"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "ARCHIVE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "guardduty", "detector/{detector_id}/filter/{name}"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"equals.#":      "1",
						"equals.0":      acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "service.additionalInfo.threatListName",
						"not_equals.#":  "2",
						"not_equals.0":  "some-threat",
						"not_equals.1":  "another-threat",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField:         "updatedAt",
						"greater_than_or_equal": startDate,
						"less_than":             endDate,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFilterConfig_noopfull(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "test-filter"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "NOOP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a NOOP"),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
		},
	})
}

func testAccFilter_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
			{
				Config: testAccFilterConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"equals.#":      "1",
						"equals.0":      acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "service.additionalInfo.threatListName",
						"not_equals.#":  "2",
						"not_equals.0":  "some-threat",
						"not_equals.1":  "yet-another-threat",
					}),
				),
			},
			{
				Config: testAccFilterConfig_matches(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.matches.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.matches.0", "us-*"),
				),
			},
			{
				Config: testAccFilterConfig_notMatches(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.not_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.not_matches.0", "us-*"),
				),
			},
		},
	})
}

func testAccFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMPCACertificateAuthorityDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfguardduty.ResourceFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_filter" {
				continue
			}

			detectorID, filterName, err := tfguardduty.FilterParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &guardduty.GetFilterInput{
				DetectorId: aws.String(detectorID),
				FilterName: aws.String(filterName),
			}

			_, err = conn.GetFilter(ctx, input)
			if err != nil {
				if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
					return nil
				}
				return err
			}

			return fmt.Errorf("Expected GuardDuty Filter to be destroyed, %s found", rs.Primary.Attributes["filter_name"])
		}

		return nil
	}
}

func testAccCheckFilterExists(ctx context.Context, t *testing.T, name string, filter *guardduty.GetFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GuardDuty filter is set")
		}

		detectorID, name, err := tfguardduty.FilterParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)
		input := guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(name),
		}
		filter, err = conn.GetFilter(ctx, &input)

		return err
	}
}

func testAccFilterConfig_full(startDate, endDate string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.region]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "another-threat"]
    }

    criterion {
      field                 = "updatedAt"
      greater_than_or_equal = %[1]q
      less_than             = %[2]q
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`, startDate, endDate)
}

func testAccFilterConfig_noopfull(startDate, endDate string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "NOOP"
  description = "This is a NOOP"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.region]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "another-threat"]
    }

    criterion {
      field                 = "updatedAt"
      greater_than_or_equal = %[1]q
      less_than             = %[2]q
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`, startDate, endDate)
}

func testAccFilterConfig_update() string {
	return `
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.region]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "yet-another-threat"]
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccFilterConfig_matches() string {
	return `
resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field   = "region"
      matches = ["us-*"]
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccFilterConfig_notMatches() string {
	return `
resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field       = "region"
      not_matches = ["us-*"]
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccCheckACMPCACertificateAuthorityDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ACMPCAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acmpca_certificate_authority" {
				continue
			}

			input := &acmpca.DescribeCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeCertificateAuthority(ctx, input)

			if err != nil {
				if errs.IsA[*acmpca_types.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			if output != nil && output.CertificateAuthority != nil && aws.ToString(output.CertificateAuthority.Arn) == rs.Primary.ID && output.CertificateAuthority.Status != acmpca_types.CertificateAuthorityStatusDeleted {
				return fmt.Errorf("ACM PCA Certificate Authority %q still exists in non-DELETED state: %s", rs.Primary.ID, string(output.CertificateAuthority.Status))
			}
		}

		return nil
	}
}
