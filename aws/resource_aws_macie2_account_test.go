package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsMacie2Account_basic(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieaccountconfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttrSet(resourceName, "finding_publishing_frequency"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func TestAccAwsMacie2Account_WithFinding(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"
	findingFreq := "FIFTEEN_MINUTES"
	findingFreqUpdated := "ONE_HOUR"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieaccountconfigWithfinding(findingFreq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", findingFreq),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testaccawsmacieaccountconfigWithfinding(findingFreqUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", findingFreqUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func TestAccAwsMacie2Account_WithStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"
	status := "ENABLED"
	statusUpdated := "PAUSED"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieaccountconfigWithstatus(status),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "status", status),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testaccawsmacieaccountconfigWithstatus(statusUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "status", statusUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func TestAccAwsMacie2Account_WithFindingAndStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"
	findingFreq := "FIFTEEN_MINUTES"
	status := "ENABLED"
	findingFreqUpdated := "ONE_HOUR"
	statusUpdated := "PAUSED"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieaccountconfigWithfindingandstatus(findingFreq, status),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", findingFreq),
					resource.TestCheckResourceAttr(resourceName, "status", status),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testaccawsmacieaccountconfigWithfindingandstatus(findingFreqUpdated, statusUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", findingFreqUpdated),
					resource.TestCheckResourceAttr(resourceName, "status", statusUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func TestAccAwsMacie2Account_disappears(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testaccawsmacieaccountconfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsMacie2AccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_account" {
			continue
		}

		input := &macie2.GetMacieSessionInput{}
		resp, err := conn.GetMacieSession(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie2 account %q still enabled", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsMacie2AccountExists(resourceName string, macie2Session *macie2.GetMacieSessionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetMacieSessionInput{}

		resp, err := conn.GetMacieSession(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie2 account %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testaccawsmacieaccountconfigBasic() string {
	return `
resource "aws_macie2_account" "test" {}
`
}

func testaccawsmacieaccountconfigWithfinding(finding string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
}
`, finding)
}

func testaccawsmacieaccountconfigWithstatus(status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  status = "%s"
}
`, status)
}

func testaccawsmacieaccountconfigWithfindingandstatus(finding, status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
  status                       = "%s"
}
`, finding, status)
}
