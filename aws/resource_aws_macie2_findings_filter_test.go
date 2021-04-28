package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsMacie2FindingsFilter_Name_Generated(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
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

func TestAccAwsMacie2FindingsFilter_NamePrefix(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNamePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
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

func TestAccAwsMacie2FindingsFilter_disappears(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsMacie2FindingsFilter_complete(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
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

func TestAccAwsMacie2FindingsFilter_withTags(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigWithTags(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigWithTags(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
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

func testAccCheckAwsMacie2FindingsFilterExists(resourceName string, macie2Session *macie2.GetFindingsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
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

func testAccCheckAwsMacie2FindingsFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_findings_filter" {
			continue
		}

		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetFindingsFilter(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
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

func testAccAwsMacieFindingsFilterconfigNameGenerated() string {
	return `resource "aws_macie2_account" "test" {}

	resource "aws_macie2_findings_filter" "test" {
		action = "ARCHIVE"
		finding_criteria {
			criterion {
				field  = "region"
			}
		}
		depends_on = [aws_macie2_account.test]
	}
`
}

func testAccAwsMacieFindingsFilterconfigNamePrefix(name string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

	resource "aws_macie2_findings_filter" "test" {
		name_prefix = %q
		action = "ARCHIVE"
		finding_criteria {
			criterion {
				field  = "region"
			}
		}
		depends_on = [aws_macie2_account.test]
	}
`, name)
}

func testAccAwsMacieFindingsFilterconfigComplete(description, action string, position int) string {
	return fmt.Sprintf(`data "aws_region" "current" {}

	resource "aws_macie2_account" "test" {}

	resource "aws_macie2_findings_filter" "test" {
		description = "%s"
		action = "%s"
		position = %d
		finding_criteria {
			criterion {
				field  = "region"
      			eq = [data.aws_region.current.name]
			}
		}
		depends_on = [aws_macie2_account.test]
	}
`, description, action, position)
}

func testAccAwsMacieFindingsFilterconfigWithTags(description, action string, position int) string {
	return fmt.Sprintf(`data "aws_region" "current" {}

	resource "aws_macie2_account" "test" {}

	resource "aws_macie2_findings_filter" "test" {
		description = "%s"
		action = "%s"
		position = %d
		finding_criteria {
			criterion {
				field  = "region"
      			eq = [data.aws_region.current.name]
			}
		}
		tags = {
    		key = "value"
		}
		depends_on = [aws_macie2_account.test]
	}
`, description, action, position)
}
