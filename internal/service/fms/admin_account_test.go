package fms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_fms_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckAdmin(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
				),
			},
		},
	})
}

func testAccCheckAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderAdmin.Meta().(*conns.AWSClient).FMSConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fms_admin_account" {
				continue
			}

			output, err := conn.GetAdminAccountWithContext(ctx, &fms.GetAdminAccountInput{})

			if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
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
}

func testAccAdminAccountConfig_basic() string {
	return acctest.ConfigCompose(
		testAccAdminRegionProviderConfig(),
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
