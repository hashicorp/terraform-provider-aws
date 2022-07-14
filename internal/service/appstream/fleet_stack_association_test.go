package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appstream"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAppStreamFleetStackAssociation_basic(t *testing.T) {
	resourceName := "aws_appstream_fleet_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetStackAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "fleet_name", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
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

func TestAccAppStreamFleetStackAssociation_disappears(t *testing.T) {
	resourceName := "aws_appstream_fleet_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetStackAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceFleetStackAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetStackAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		fleetName, stackName, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding AppStream Fleet Stack Association ID (%s): %w", rs.Primary.ID, err)
		}

		err = tfappstream.FindFleetStackAssociation(context.TODO(), conn, fleetName, stackName)

		if tfresource.NotFound(err) {
			return fmt.Errorf("AppStream Fleet Stack Association %q does not exist", rs.Primary.ID)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckFleetStackAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_fleet_stack_association" {
			continue
		}

		fleetName, stackName, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding AppStream Fleet Stack Association ID (%s): %w", rs.Primary.ID, err)
		}

		err = tfappstream.FindFleetStackAssociation(context.TODO(), conn, fleetName, stackName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("AppStream Fleet Stack Association %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFleetStackAssociationConfig_basic(name string) string {
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

resource "aws_appstream_fleet_stack_association" "test" {
  fleet_name = aws_appstream_fleet.test.name
  stack_name = aws_appstream_stack.test.name
}
`, name)
}
