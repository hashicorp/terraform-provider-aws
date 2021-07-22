package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	RegisterServiceErrorCheckFunc(transfer.EndpointsID, testAccErrorCheckSkipTransfer)
}

func testAccAWSTransferAccess_basic(t *testing.T) {
	var conf transfer.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferAccessBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferAccessExists(resourceName, &conf),
					//testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "external_id", ""),
					//TODO: ...
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferAccessUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferAccessExists(resourceName, &conf),
					//testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestCheckResourceAttr(resourceName, "external_id", ""),
					//TODO: ...
				),
			},
		},
	})
}

func testAccCheckAWSTransferAccessExists(n string, v *transfer.DescribedAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Access ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		externalID := rs.Primary.Attributes["external_id"]
		serverID := rs.Primary.Attributes["server_id"]
		output, err := finder.AccessByID(conn, serverID, externalID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSTransferAccessDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_access" {
			continue
		}

		externalID := rs.Primary.Attributes["external_id"]
		serverID := rs.Primary.Attributes["server_id"]
		_, err := finder.AccessByID(conn, serverID, externalID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Transfer Access %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSTransferAccessBasicConfig() string {
	//TODO: Migrate from Server to Access
	return `
resource "aws_server_access" "test" {}
`
}

func testAccAWSTransferAccessUpdatedConfig(rName string) string {
	//TODO: Migrate from Server to Access

	return composeConfig(
		testAccAWSTransferServerConfigBaseLoggingRole(rName),
		`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.test.arn
}
`)
}
