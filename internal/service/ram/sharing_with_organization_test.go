package ram_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRAMSharingWithOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_sharing_with_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ram.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// RAM sharing with organization cannot be deleted separately.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSharingWithOrganizationConfig_basic(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_organization" {
				continue
			}

			params := &organizations.DescribeOrganizationInput{}

			resp, err := conn.DescribeOrganizationWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
				return nil
			}

			if err != nil {
				return err
			}

			if resp != nil && resp.Organization != nil {
				return fmt.Errorf("Bad: Organization still exists: %q", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccSharingWithOrganizationConfig_basic() string {
	return `
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["ram.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_ram_sharing_with_organization" "test" {
  depends_on = [
    aws_organizations_organization.test
  ]
}
`
}
