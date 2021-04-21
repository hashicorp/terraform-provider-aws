package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccAwsMacie2FindingsFilter_basic(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	name := acctest.RandomWithPrefix("testacc-findings-filter")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieFindingsFilterconfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
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

func TestAccAwsMacie2FindingsFilter_complete(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	name := acctest.RandomWithPrefix("testacc-findings-filter")
	clientToken := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieFindingsFilterconfigComplete(clientToken, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "action"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "position"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
			},
		},
	})
}

func TestAccAwsMacie2FindingsFilter_withTags(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	name := acctest.RandomWithPrefix("testacc-findings-filter")
	clientToken := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieFindingsFilterconfigWithTags(clientToken, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "action"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "position"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_token"},
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
			return fmt.Errorf("macie2 FindingsFilter %q does not exist", rs.Primary.ID)
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

		if isAWSErr(err, macie2.ErrCodeAccessDeniedException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie2 FindingsFilter %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testaccawsmacieFindingsFilterconfigBasic(name string) string {
	return fmt.Sprintf(`
	resource "aws_macie2_account" "test" {}

	resource "aws_macie2_findings_filter" "test" {
		name = "%s"
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

func testaccawsmacieFindingsFilterconfigComplete(clientToken, name string) string {
	return fmt.Sprintf(`
	data "aws_region" "current" {}

	resource "aws_macie2_account" "test" {
		client_token = "%s"
	}

	resource "aws_macie2_findings_filter" "test" {
		name = "%s"
		client_token = aws_macie2_account.test.client_token
		description = "this a description"
		position = 1
		action = "ARCHIVE"
		finding_criteria {
			criterion {
				field  = "region"
      			eq = [data.aws_region.current.name]
			}
		}
		depends_on = [aws_macie2_account.test]
	}
`, clientToken, name)
}

func testaccawsmacieFindingsFilterconfigWithTags(clientToken, name string) string {
	return fmt.Sprintf(`
	data "aws_region" "current" {}

	resource "aws_macie2_account" "test" {
		client_token = "%s"
	}

	resource "aws_macie2_findings_filter" "test" {
		name = "%s"
		client_token = aws_macie2_account.test.client_token
		description = "this a description"
		position = 1
		action = "ARCHIVE"
		finding_criteria {
			criterion {
				field  = "region"
      			eq = [data.aws_region.current.name]
			}
		}
		tags = {
    		Key = "value"
		}
		depends_on = [aws_macie2_account.test]
	}
`, clientToken, name)
}
