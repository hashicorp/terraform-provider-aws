package aws

import (
	"testing"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"fmt"
)

func testACCAwsGuardDutyInvite_basic(t *testing.T) {
	resourceName := "aws_guardduty_invite.test"
	accountID := "111111111111"
	email := "required@example.com"
	message := "foobar"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t)},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep {
			{
				Config: testAccGuardDutyInviteConfig_basic(accountID, email, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyInviteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "message", message),
				),
			},
		},
	})
	return
}


func testAccCheckAwsGuardDutyInviteExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		//TODO Write test

		return nil
	}
}

func testAccGuardDutyInviteConfig_basic(accountID, email, message string) string {
	//TODO do something
	return "Do Somthing"
}