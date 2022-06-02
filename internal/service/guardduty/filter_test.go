package guardduty_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
)

func testAccFilter_basic(t *testing.T) {
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"
	detectorResourceName := "aws_guardduty_detector.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "action", "ARCHIVE"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/[a-z0-9]{32}/filter/test-filter$")),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":    "region",
						"equals.#": "1",
						"equals.0": acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":        "service.additionalInfo.threatListName",
						"not_equals.#": "2",
						"not_equals.0": "some-threat",
						"not_equals.1": "another-threat",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":                 "updatedAt",
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
					testAccCheckFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "action", "NOOP"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a NOOP"),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
		},
	})
}

func testAccFilter_update(t *testing.T) {
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
			{
				Config: testAccFilterConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":    "region",
						"equals.#": "1",
						"equals.0": acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":        "service.additionalInfo.threatListName",
						"not_equals.#": "2",
						"not_equals.0": "some-threat",
						"not_equals.1": "yet-another-threat",
					}),
				),
			},
		},
	})
}

func testAccFilter_tags(t *testing.T) {
	var v1, v2, v3 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_multipleTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Value"),
				),
			},
			{
				Config: testAccFilterConfig_updateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Updated"),
				),
			},
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccFilter_disappears(t *testing.T) {
	var v guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckACMPCACertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFilterExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfguardduty.ResourceFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn

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

		_, err = conn.GetFilter(input)
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

func testAccCheckFilterExists(name string, filter *guardduty.GetFilterOutput) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn
		input := guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(name),
		}
		filter, err = conn.GetFilter(&input)
		if err != nil {
			return err
		}

		return nil
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

func testAccCheckACMPCACertificateAuthorityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_certificate_authority" {
			continue
		}

		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCertificateAuthority(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}

		if output != nil && output.CertificateAuthority != nil && aws.StringValue(output.CertificateAuthority.Arn) == rs.Primary.ID && aws.StringValue(output.CertificateAuthority.Status) != acmpca.CertificateAuthorityStatusDeleted {
			return fmt.Errorf("ACM PCA Certificate Authority %q still exists in non-DELETED state: %s", rs.Primary.ID, aws.StringValue(output.CertificateAuthority.Status))
		}
	}

	return nil
}
