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

func testAccAwsMacie2Member_basic(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520433213222"
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
					resource.TestCheckResourceAttr(resourceName, "relationship_status", macie2.RelationshipStatusCreated),
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

func testAccAwsMacie2Member_disappears(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520433213222"
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
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Member(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsMacie2Member_withTags(t *testing.T) {
	var macie2Output macie2.GetMemberOutput
	resourceName := "aws_macie2_member.test"
	accountID := "520433213222"
	email := "test@test.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2MemberDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieMemberConfigWithTags(accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2MemberExists(resourceName, &macie2Output),
					testAccCheckResourceAttrRfc3339(resourceName, "invited_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
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

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "member accounts are associated with your account") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "account is not associated with your account") {
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
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_member" "test" {
  account_id = %[1]q
  email      = %[2]q
  depends_on = [aws_macie2_account.test]
}
`, accountID, email)
}

func testAccAwsMacieMemberConfigWithTags(accountID, email string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_member" "test" {
  account_id = %[1]q
  email      = %[2]q
  tags = {
    Key = "value"
  }
  depends_on = [aws_macie2_account.test]
}
`, accountID, email)
}
