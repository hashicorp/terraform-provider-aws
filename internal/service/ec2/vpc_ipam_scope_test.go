package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVPCIpamScope_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCIpamScopeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIpamScope("test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "pool_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_arn", ipamName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "ipam_id", ipamName, "id"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCIpamScope("test2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test2"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckVPCIpamScopeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_scope" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := tfec2.WaitIpamScopeDeleted(conn, *id, tfec2.IpamScopeDeleteTimeout); err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Scope (%s) to be deleted: %w", *id, err)
		}
	}

	return nil
}

func testAccVPCIpamScope(desc string) string {
	return testAccVPCIpam + fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
	ipam_id    =  aws_vpc_ipam.test.id
	description = %[1]q
}
`, desc)
}
