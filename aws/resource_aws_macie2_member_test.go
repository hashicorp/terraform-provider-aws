package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsMacie2Member_basic(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520983883852" //os.Getenv("AWS_ACCOUNT_ID")
	email := "test@test.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2MemberDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieMemberConfigBasic(accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2MemberExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func TestAccAwsMacie2Member_disappears(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520983883852" //os.Getenv("AWS_ACCOUNT_ID")
	email := "test@test.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2MemberDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieMemberConfigBasic(accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2MemberExists(resourceName, &macie2Output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsMacie2Member_withTags(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520983883852" //os.Getenv("AWS_ACCOUNT_ID")
	email := "test@test.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2MemberDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieMemberConfigWithTags(accountID, email, macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2MemberExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAwsMacieMemberConfigWithTags(accountID, email, macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2MemberExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusPaused),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func testAccCheckAwsMacie2MemberExists(resourceName string, macie2Session *macie2.GetMemberOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetMemberInput{Id: aws.String(rs.Primary.ID)}

		resp, err := conn.GetMember(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie Member %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccCheckAwsMacie2MemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_member" {
			continue
		}

		input := &macie2.GetMemberInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetMember(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie Member %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsMacieMemberConfigBasic(accountID, email string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

	resource "aws_macie2_member" "test" {
		account_id = "%s"
		email = "%s"
		depends_on = [aws_macie2_account.test]
	}
`, accountID, email)
}

func testAccAwsMacieMemberConfigWithTags(accountID, email, status string) string {
	return fmt.Sprintf(`resource "aws_macie2_account" "test" {}

	resource "aws_macie2_member" "test" {
		account_id = "%s"
		email = "%s"
		status = "%s"
		tags = {
    		key = "value"
		}
		depends_on = [aws_macie2_account.test]
	}
`, accountID, email, status)
}
