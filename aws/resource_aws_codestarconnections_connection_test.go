package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCodeStarConnectionsConnection_Basic(t *testing.T) {
	resourceName := "aws_codestarconnections_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsConnectionConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("connection/.+")),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("connection/.+")),
					testAccMatchResourceAttrRegionalARN(resourceName, "connection_arn", "codestar-connections", regexp.MustCompile("connection/.+")),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeBitbucket),
					resource.TestCheckResourceAttr(resourceName, "connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "connection_status", codestarconnections.ConnectionStatusPending),
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

func testAccCheckAWSCodeStarConnectionsConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codestarconnectionsconn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarconnections_connection":
			_, err := conn.GetConnection(&codestarconnections.GetConnectionInput{
				ConnectionArn: aws.String(rs.Primary.ID),
			})

			if err != nil && !isAWSErr(err, codestarconnections.ErrCodeResourceNotFoundException, "") {
				return err
			}
		}
	}

	return nil
}

func testAccAWSCodeStarConnectionsConnectionConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
    connection_name    = %[1]q
    provider_type = "Bitbucket"
}
`, rName)
}
