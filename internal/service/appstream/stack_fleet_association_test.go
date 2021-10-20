package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamStackFleetAssociation_basic(t *testing.T) {
	resourceName := "aws_appstream_stack_fleet_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackFleetAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackFleetAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackFleetAssociationExists(resourceName),
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

func TestAccAppStreamStackFleetAssociation_disappears(t *testing.T) {
	resourceName := "aws_appstream_stack_fleet_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackFleetAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackFleetAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackFleetAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceStackFleetAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStackFleetAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		fleetName, _, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding id appstream stack fleet association (%s): %w", rs.Primary.ID, err)
		}

		resp, err := conn.ListAssociatedStacksWithContext(context.TODO(), &appstream.ListAssociatedStacksInput{FleetName: aws.String(fleetName)})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.Names) == 0 {
			return fmt.Errorf("appstream stack fleet association %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackFleetAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_stack_fleet_association" {
			continue
		}

		fleetName, _, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding id appstream stack fleet association (%s): %w", rs.Primary.ID, err)
		}

		resp, err := conn.ListAssociatedStacksWithContext(context.TODO(), &appstream.ListAssociatedStacksInput{FleetName: aws.String(fleetName)})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.Names) > 0 {
			return fmt.Errorf("appstream stack fleet association %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccStackFleetAssociationConfig(name string) string {
	// "Amazon-AppStream2-Sample-Image-02-04-2019" is not available in GovCloud
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = "stream.standard.small"

  compute_capacity {
    desired_instances = 1
  }
}

resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_stack_fleet_association" "test" {
  fleet_name = aws_appstream_fleet.test.name
  stack_name = aws_appstream_stack.test.name
}
`, name)
}
