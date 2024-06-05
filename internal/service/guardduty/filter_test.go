// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	acmpca_types "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "test-filter"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "ARCHIVE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "rank", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "guardduty", regexache.MustCompile("detector/[0-9a-z]{32}/filter/test-filter$")),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "service.additionalInfo.threatListName",
						"not_equals.#":  acctest.Ct2,
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
					testAccCheckFilterExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "test-filter"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "NOOP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a NOOP"),
					resource.TestCheckResourceAttr(resourceName, "rank", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", acctest.Ct3),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", acctest.Ct3),
				),
			},
			{
				Config: testAccFilterConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: names.AttrRegion,
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						names.AttrField: "service.additionalInfo.threatListName",
						"not_equals.#":  acctest.Ct2,
						"not_equals.0":  "some-threat",
						"not_equals.1":  "yet-another-threat",
					}),
				),
			},
		},
	})
}

func testAccFilter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_multipleTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Value"),
				),
			},
			{
				Config: testAccFilterConfig_updateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Updated"),
				),
			},
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMPCACertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfguardduty.ResourceFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

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

			_, err = conn.GetFilterWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
					return nil
				}
				return err
			}

			return fmt.Errorf("Expected GuardDuty Filter to be destroyed, %s found", rs.Primary.Attributes["filter_name"])
		}

		return nil
	}
}

func testAccCheckFilterExists(ctx context.Context, name string, filter *guardduty.GetFilterOutput) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)
		input := guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(name),
		}
		filter, err = conn.GetFilterWithContext(ctx, &input)

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
      equals = [data.aws_region.current.name]
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
      equals = [data.aws_region.current.name]
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

func testAccFilterConfig_multipleTags() string {
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
      equals = [data.aws_region.current.name]
    }
  }

  tags = {
    Name = "test-filter"
    Key  = "Value"
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
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
      equals = [data.aws_region.current.name]
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

func testAccFilterConfig_updateTags() string {
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
      equals = [data.aws_region.current.name]
    }
  }

  tags = {
    Key = "Updated"
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccCheckACMPCACertificateAuthorityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

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

			if output != nil && output.CertificateAuthority != nil && aws.StringValue(output.CertificateAuthority.Arn) == rs.Primary.ID && output.CertificateAuthority.Status != acmpca_types.CertificateAuthorityStatusDeleted {
				return fmt.Errorf("ACM PCA Certificate Authority %q still exists in non-DELETED state: %s", rs.Primary.ID, string(output.CertificateAuthority.Status))
			}
		}

		return nil
	}
}
