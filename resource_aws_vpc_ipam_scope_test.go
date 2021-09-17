package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcIpamScope_basic(t *testing.T) {
	resourceName := "aws_vpc_ipam_scope.test"
	ipamName := "aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsVpcIpamScopeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamScope("test"),
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
				Config: testAccAwsVpcIpamScope("test2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "test2"),
					// resource.TestCheckResourceAttr(rName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsVpcIpamScopeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_ipam_scope" {
			continue
		}

		id := aws.String(rs.Primary.ID)

		if _, err := waitIpamScopeDeleted(conn, *id, IpamScopeDeleteTimeout); err != nil {
			if isResourceNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("error waiting for IPAM Scope (%s) to be deleted: %w", *id, err)
		}
	}

	return nil
}

func testAccAwsVpcIpamScope(desc string) string {
	return testAccAwsVpcIpam + fmt.Sprintf(`
resource "aws_vpc_ipam_scope" "test" {
	ipam_id    =  aws_vpc_ipam.test.id
	description = %[1]q
}
`, desc)
}
