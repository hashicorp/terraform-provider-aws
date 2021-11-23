package fms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccAdminAccount_basic(t *testing.T) {
	resourceName := "aws_fms_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckFmsAdmin(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, fms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFmsAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
				),
			},
		},
	})
}

func testAccCheckFmsAdminAccountDestroy(s *terraform.State) error {
	conn := testAccProviderFmsAdmin.Meta().(*conns.AWSClient).FMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_admin_account" {
			continue
		}

		output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

		if tfawserr.ErrMessageContains(err, fms.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusDeleted {
			continue
		}

		return fmt.Errorf("FMS Admin Account (%s) still exists with status: %s", aws.StringValue(output.AdminAccount), aws.StringValue(output.RoleStatus))
	}

	return nil
}

func testAccFmsAdminAccountConfig_basic() string {
	return acctest.ConfigCompose(
		testAccFmsAdminRegionProviderConfig(),
		`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["fms.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_fms_admin_account" "test" {
  account_id = aws_organizations_organization.test.master_account_id
}
`)
}
